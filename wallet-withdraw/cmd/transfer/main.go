package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"
	bviper "upex-wallet/wallet-base/viper"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	lcmd "upex-wallet/wallet-withdraw/cmd"
	"upex-wallet/wallet-withdraw/transfer/checker"
	"upex-wallet/wallet-withdraw/transfer/cooldown"
	"upex-wallet/wallet-withdraw/transfer/gather"
	"upex-wallet/wallet-withdraw/transfer/rollback"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/utxofee"
	"upex-wallet/wallet-withdraw/transfer/withdraw"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	cfg         = config.DefaultConfig()
	serviceName string
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cobra.OnInitialize(initConfig)

	c := cmd.New("withdraw", "withdraw is a hot wallet for crypto currency exchange", "", run)
	c.Flags().StringVarP(&cfgFile, "config", "c", "app.yml", "config file (default is app.yml)")

	if err := c.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig 初始化配置文件
func initConfig() {
	if cfgFile != "" && utils.FileExist(cfgFile) {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("app")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("read config failed, %v", err)
		log.Warnf("run with default config")
	} else {
		err = bviper.MergeExtIfNecessary()
		if err != nil {
			log.Errorf("merge config failed, %v", err)
		}

		cfg = config.New()
	}

	serviceName = fmt.Sprintf("wallet-withdraw-%s", strings.ToLower(cfg.Currency))
}

func createTxBuilder(cfg *config.Config) txbuilder.Builder {
	if cfg == nil {
		return nil
	}

	creator, ok := txbuilder.Find(strings.ToUpper(cfg.Currency))
	if !ok {
		log.Errorf("can't find transfer of %s", cfg.Currency)
		return nil
	}
	return creator(cfg)
}

// initFeeService , update UTXO-LIKE transaction fee to db
func initFeeService(txBuilder txbuilder.Builder, cfg *config.Config) (uf *utxofee.UpdateFee, err error) {
	log.Infof("suggest fee service for %s start...", cfg.Currency)

	if cfg == nil {
		return nil, fmt.Errorf("config did not loading ")
	}

	if txBuilder.Model() == txbuilder.UTXOModel {
		var (
			sf = models.SuggestFee{
				Symbol: strings.ToLower(cfg.Currency),
			}
		)

		err = sf.InitCurrencyFee()
		if err != nil {
			log.Errorf("Init suggest fee table for %s fail", cfg.Currency)
			return nil, err
		}
		return utxofee.NewUpdateFee(cfg), nil
	}

	return nil, nil
}

func run(*cmd.Command) error {
	defer util.DeferRecover(serviceName, nil)()

	err := util.InitDaysJSONRotationLogger("./log/", serviceName+".log", 60)
	if err != nil {
		panic(err)
	}

	log.Infof("%s %s service start", serviceName, lcmd.Version())

	// initial db
	dbInst, err := db.New(cfg.DSN, serviceName)
	if err != nil {
		panic(err)
	}
	defer dbInst.Close()
	err = models.Init(dbInst)
	if err != nil {
		panic(err)
	}

	txBuilder := createTxBuilder(cfg)
	if txBuilder == nil {
		panic("failed to create tx builder")
	}

	// update transaction fee
	suggestFee, err := initFeeService(txBuilder, cfg)
	if suggestFee != nil {
		go suggestFee.FeeService(strings.ToLower(cfg.Currency))
	} else if err != nil {
		panic(err)
	}

	// register service
	var services []*service.Service
	util.RegisterSignalHandler(func(s os.Signal) {
		close(cfg.ExitSignal)

		for _, srv := range services {
			srv.Stop()
		}
		os.Exit(0)
	}, syscall.SIGINT, syscall.SIGTERM)

	services = append(services, service.New(checker.New(cfg)))

	// withdraw service
	if cfg.Withdraw {
		services = append(services, service.NewWithInterval(withdraw.New(cfg, txBuilder), cfg.WithdrawInterval))
	}

	// coolDown service
	if cfg.Cooldown {
		services = append(services, service.New(cooldown.New(cfg, txBuilder)))
	}

	// gather service
	if cfg.Gather {
		services = append(services, service.NewWithInterval(gather.New(cfg, txBuilder), cfg.GatherInterval))
	}

	if cfg.AutoRollback {
		services = append(services, service.New(rollback.New(cfg)))
	}

	for _, s := range services {
		go s.Start()
	}

	select {}
}
