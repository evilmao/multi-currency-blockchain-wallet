package config

import (
	"fmt"

	bviper "upex-wallet/wallet-base/viper"

	"github.com/spf13/viper"
)

func Init(cfgFile string) error {
	viper.SetConfigFile(cfgFile)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config failed, %v", err)
	}

	err = bviper.MergeExtIfNecessary()
	if err != nil {
		return fmt.Errorf("merge config failed, %v", err)
	}
	return nil
}
