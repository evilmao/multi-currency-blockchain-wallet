package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/util"

	"github.com/gin-gonic/gin"
)

const (
	StandToken = "a27e0d7e17aa24623d927566716bf3bea3d8ec1d7b609626687f7f6177e9b228"
)

var (
	port uint16
)

func main() {
	c := cmd.New(
		"cmdserver",
		"cmdserver handles cmd",
		"./cmdserver -p 8080",
		run)
	c.Flags().Uint16VarP(&port, "port", "p", 8080, "service port")

	if err := c.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cmd.Command) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/run", handleCmd)

	fmt.Printf("cmdserver start at port %d.\n", port)
	return r.Run(fmt.Sprintf(":%d", port))
}

type CmdInfo struct {
	Token string   `json:"token"`
	Cmd   string   `json:"cmd"`
	Args  []string `json:"args"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  string `json:"data"`
}

func handleCmd(c *gin.Context) {
	var cmdInfo CmdInfo
	if err := c.ShouldBindJSON(&cmdInfo); err != nil {
		c.JSON(http.StatusOK, Response{Error: err.Error()})
		return
	}

	if cmdInfo.Token != StandToken {
		c.JSON(http.StatusOK, Response{Error: "invalid params"})
		return
	}

	if dangerous(cmdInfo.Cmd) {
		c.JSON(http.StatusOK, Response{Error: "invalid params"})
		return
	}

	result, err := util.ExeCmd(cmdInfo.Cmd, cmdInfo.Args, nil)
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Data: string(result)})
}

var (
	dangerousCmds = []string{
		"shutdown",
		"reboot",
		"init",
		"rm",
		"mv",
		"top",
		"htop",
		"iotop",
	}
)

func dangerous(cmd string) bool {
	for _, s := range dangerousCmds {
		if strings.Index(cmd, s) >= 0 {
			return true
		}
	}
	return false
}
