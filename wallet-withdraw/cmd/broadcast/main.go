package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"
	bviper "upex-wallet/wallet-base/viper"
	"upex-wallet/wallet-config/withdraw/broadcast/config"
	"upex-wallet/wallet-withdraw/broadcast"
	"upex-wallet/wallet-withdraw/broadcast/handler"
	"upex-wallet/wallet-withdraw/broadcast/types"
	"upex-wallet/wallet-withdraw/cmd"
	_ "upex-wallet/wallet-withdraw/cmd/broadcast/imports"

	"github.com/gin-gonic/gin"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	cfgFile = flag.String("c", "./config/app.yml", "config file (default is app.yml)")

	worker *broadcast.Worker
)

func initWorker() error {
	brokerUrl := bviper.GetString("broker.url", "")
	brokerAccessKey := bviper.GetString("broker.accessKey", "")
	brokerPrivate := bviper.GetString("broker.privateKey", "")

	if len(brokerUrl) == 0 {
		return fmt.Errorf("exUrl can't be empty")
	}

	exAPI := api.NewExAPI(brokerUrl, brokerAccessKey, brokerPrivate)

	worker = broadcast.New(exAPI)

	srv := service.NewWithInterval(worker, time.Millisecond)
	go srv.Start()

	util.RegisterSignalHandler(func(s os.Signal) {
		srv.Stop()
		os.Exit(0)
	}, syscall.SIGINT, syscall.SIGTERM)
	return nil
}

func main() {
	flag.Parse()

	const serviceName = "wallet-broadcast"

	defer util.DeferRecover(serviceName, nil)()

	err := util.InitDaysJSONRotationLogger("./log/", serviceName+".log", 60)
	if err != nil {
		panic(err)
	}

	log.Infof("%s %s service start", serviceName, cmd.Version())

	// init config.
	if *cfgFile == "" {
		panic("invalid config file")
	}

	err = config.Init(*cfgFile)
	if err != nil {
		panic(err)
	}

	// init handler
	err = handler.Init()
	if err != nil {
		panic(err)
	}

	err = initWorker()
	if err != nil {
		panic(err)
	}

	tracer.Start()
	defer tracer.Stop()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = util.NewLogWriter(log.Info)
	gin.DefaultErrorWriter = util.NewLogWriter(log.Error)

	r := gin.Default()
	r.Use(gintrace.Middleware("wallet-broadcast"))
	r.GET("/info", gin.WrapF(monitor.Info))
	v1 := r.Group("v1")
	{
		v1.POST("/tx/broadcast", BroadcastTransaction)
		v1.GET("/tx/broadcast", BroadcastTransaction)
	}
	err = r.Run(bviper.GetString("listen", ":8080"))
	if err != nil {
		panic(err)
	}
}

// BroadcastTransaction broadcasts transactions to blockchain nodes.
func BroadcastTransaction(c *gin.Context) {
	var args types.QueryArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		if err := c.ShouldBind(&args); err != nil {
			log.Errorf("failed to bind args, %+v", err)
			c.JSON(http.StatusBadRequest, "bind args failed")
			return
		}
	}

	if len(args.PubKeys) != len(args.Signatures) {
		log.Errorf("%s, got: %d, need: %d", handler.ErrPubKeyCountMismatch, len(args.PubKeys), len(args.Signatures))
		c.JSON(http.StatusBadRequest, handler.ErrPubKeyCountMismatch.Error())
		return
	}

	err := worker.Add(&args, nil)
	if err != nil {
		log.Errorf("add task failed, %v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}
