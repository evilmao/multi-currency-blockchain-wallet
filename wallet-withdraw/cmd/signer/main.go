package main

import (
	"fmt"
	"os"
	"syscall"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/withdraw/signer/config"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"
	lcmd "upex-wallet/wallet-withdraw/cmd"
	"upex-wallet/wallet-withdraw/signer"

	"github.com/sevlyar/go-daemon"
	"github.com/spf13/viper"
)

const (
	EnvPwdKey = "SIGNER_PWD"
)

var (
	password string
	cfgFile  string

	cfg *config.Config
)

func main() {
	c := cmd.New("signer", "signer is a hot wallet signature server", "", run)
	c.Flags().StringVarP(&password, "password", "p", "", "server password")
	c.Flags().StringVarP(&cfgFile, "config", "c", "app.yml", "config file (default is app.yml)")

	if err := c.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("app")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config failed, %v", err)
	}

	cfg = config.NewConfig()
	return nil
}

func forkAndExit() error {
	dm := &daemon.Context{
		PidFileName: "pid",
		Env:         append(os.Environ(), fmt.Sprintf("%s=%s", EnvPwdKey, password)),
	}

	child, err := dm.Reborn()
	if err != nil {
		return err
	}

	if child != nil {
		fmt.Printf("start signer success, pid: %d\n", child.Pid)
		os.Exit(0)
	}

	return nil
}

func run(c *cmd.Command) error {
	const serviceName = "wallet-signer"
	defer util.DeferRecover(serviceName, nil)()

	if !daemon.WasReborn() {
		err := c.Password("password")
		if err != nil {
			panic(err)
		}
	}

	err := forkAndExit()
	if err != nil {
		panic(err)
	}

	err = util.InitDaysJSONRotationLogger("./log/", serviceName+".log", 60)
	if err != nil {
		panic(err)
	}

	err = initConfig()
	if err != nil {
		panic(err)
	}

	srv := signer.NewServer(cfg)
	err = srv.SetPassPhrase(os.Getenv(EnvPwdKey))
	if err != nil {
		panic(err)
	}
	os.Setenv(EnvPwdKey, "")

	util.RegisterSignalHandler(func(s os.Signal) {
		log.Infof("%s service stop", serviceName)
		os.Exit(0)
	}, syscall.SIGINT, syscall.SIGTERM)

	log.Infof("%s %s service start", serviceName, lcmd.Version())

	err = srv.Start()
	if err != nil {
		panic(err)
	}

	return nil
}
