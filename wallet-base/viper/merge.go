package viper

import (
	"bufio"

	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/spf13/viper"
)

const (
	extConfigsKey = "extConfigs"
)

// MergeExtIfNecessary merges external configs if necessary.
func MergeExtIfNecessary() error {
	exts := GetStringSlice(extConfigsKey, nil)
	for _, ext := range exts {
		if !util.FileExist(ext) {
			log.Errorf("merge config %s failed, file not exist", ext)
			continue
		}

		err := util.WithReadFile(ext, func(reader *bufio.Reader) error {
			return viper.MergeConfig(reader)
		})
		if err != nil {
			log.Errorf("merge config %s failed, %v", ext, err)
		}
	}
	return nil
}
