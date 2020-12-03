package cmd

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/base/models"
)

var (
	cfgFile string
	cfg     = config.DefaultConfig()
)

var rootCmd = &cobra.Command{
	Use:   "wallet deposit sync",
	Short: "wallet deposit sync",
	Long:  `sync fetch block from blockchain node`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "app.yml", "config file (default is app.yml)")
	rootCmd.PersistentFlags().StringVarP(&cfg.RPCURL, "rpcurl", "u", cfg.RPCURL, "node rpc request url")
}

func initConfig() {
	if cfgFile != "" && utils.FileExist(cfgFile) {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("app")
		viper.AddConfigPath("./config")
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("read config failed, %v", err)
		log.Warnf("run with default config")
	} else {
		cfg = config.New()
	}
}

func initLogger(name string) error {
	filePath := "./log/"
	symbol := strings.ToLower(cfg.Currency)
	return util.InitDefaultRotationLogger(filePath, fmt.Sprintf("wallet-%s-%s.log", name, symbol))
}

// Execute executes run.
func Execute(serviceType string) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		serviceName := fmt.Sprintf("wallet-%s-%s", serviceType, strings.ToLower(cfg.Currency))

		defer util.DeferRecover(serviceName, nil)()

		err := initLogger(serviceType)
		if err != nil {
			panic(fmt.Errorf("init logger failed, %v", err))
		}

		log.Infof("%s %s service start", serviceName, Version())

		go heartbeat()

		// initial db
		dbInstance, err := db.New(cfg.DSN, serviceName)
		if err != nil {
			panic(err)
		}
		defer dbInstance.Close()
		err = models.InitDB()
		if err != nil {
			panic(err)
		}

		// init currency
		currency.Init(cfg)

		// chose runnable
		err = choseRunnable(cfg)
		if err != nil {
			panic(err)
		}
	}
	return rootCmd.Execute()
}

func choseRunnable(cfg *config.Config) error {

	runType, ok := Find(cfg.Currency)
	if !ok {
		return fmt.Errorf("chose runnable fail, currency %s is not exits", cfg.Currency)
	}

	runType0 := func(run Runnable) {
		restartTimes := 0
		for {
			util.WithRecover("deposit-run", func() {
				run(cfg, restartTimes)
			}, nil)

			time.Sleep(2 * time.Second)
			restartTimes++
			log.Errorf("%s deposit Service Restart %d Times", strings.ToUpper(cfg.Currency), restartTimes)
		}
	}

	switch runType.Type {
	case 0:
		runType0(runType.Runnable)
	case 1:
		runType.Runnable(cfg, 1)
	default:
		return fmt.Errorf("runnable type is wrong,should 1 or 0")
	}

	return nil
}

func heartbeat() {
	for {
		log.Info("heartbeat...")

		time.Sleep(time.Minute * 10)
	}
}
