package cmd

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/base/models"
	"upex-wallet/wallet-deposit/rpc"
	"upex-wallet/wallet-withdraw/syncd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	cfgFile string
	cfg     = config.DefaultConfig()
)

var rootCmd = &cobra.Command{
	Use:   "wallet deposit syncer",
	Short: "wallet deposit syncer",
	Long:  `syncer fetch block from blockchain node`,
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

func initLogger() error {
	filePath := "./log/"
	symbol := strings.ToLower(cfg.Currency)
	return util.InitDefaultRotationLogger(filePath, fmt.Sprintf("wallet-deposit-%s.log", symbol))
}

// Runnable def.
type Runnable func(*config.Config, int)

// Execute executes run.
func Execute(run Runnable) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		serviceName := fmt.Sprintf("wallet-withdraw-syncd-%s", strings.ToLower(cfg.Currency))

		defer util.DeferRecover(serviceName, nil)()

		err := initLogger()
		if err != nil {
			panic(fmt.Errorf("init logger failed, %v", err))
		}

		log.Infof("%s service start", serviceName)

		go heartbeat()

		// data-dog monitor and tracer.
		go monitor.ListenAndServe(cfg.ListenAddress)
		statusReporter := monitor.NewStatsdReporter(cfg.StatusAddress, "wallet-deposit", nil)
		go statusReporter.Start()

		tracer.Start(tracer.WithServiceName(serviceName))
		defer tracer.Stop()

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

		restartTimes := 0
		for {
			util.WithRecover("syncd-run", func() {
				run(cfg, restartTimes)
			}, nil)

			time.Sleep(2 * time.Second)
			restartTimes++
			log.Errorf("%s Syncer Service Restart %d Times", strings.ToUpper(cfg.Currency), restartTimes)
		}
	}
	return rootCmd.Execute()
}

// Exec def.
func Exec(createRPCClient rpc.RPCCreator) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		serviceName := fmt.Sprintf("wallet-deposit-%s", strings.ToLower(cfg.Currency))
		defer util.DeferRecover(serviceName, nil)()

		err := initLogger()
		if err != nil {
			panic(fmt.Errorf("init logger failed, %v", err))
		}

		log.Infof("%s service start", serviceName)

		go heartbeat()

		// data-dog monitor and tracer.
		go monitor.ListenAndServe(cfg.ListenAddress)
		statusReporter := monitor.NewStatsdReporter(cfg.StatusAddress, "wallet-deposit", nil)
		go statusReporter.Start()

		tracer.Start(tracer.WithServiceName(serviceName))
		defer tracer.Stop()

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

		rpcClient := createRPCClient(cfg)
		if rpcClient == nil {
			panic("failed to create rpc client")
		}

		syncdSrv := service.NewWithInterval(syncd.New(cfg, rpcClient), time.Millisecond)
		defer syncdSrv.Stop()
		if err = syncdSrv.Start(); err != nil {
			panic(err)
		}
	}
	return rootCmd.Execute()
}

func heartbeat() {
	for {
		log.Info("heartbeat...")

		time.Sleep(time.Minute * 10)
	}
}
