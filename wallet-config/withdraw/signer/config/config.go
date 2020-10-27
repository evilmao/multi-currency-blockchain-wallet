package config

import (
	bviper "upex-wallet/wallet-base/viper"
)

type Config struct {
	// RSAKey for decrypt the part of wallet.dat password from client
	RSAKey string

	// RSAPubKey for encrypt the signature
	RSAPubKey string

	ListenAddr string
	Currency   string
	DataPath   string
	FileNames  []string
}

func NewConfig() *Config {
	return &Config{
		RSAKey:     bviper.GetString("rsaKey", ""),
		RSAPubKey:  bviper.GetString("rsaPubKey", ""),
		ListenAddr: bviper.GetString("listen", ":8899"),
		Currency:   bviper.GetString("currency", "btc"),
		DataPath:   bviper.GetString("dataPath", "./"),
		FileNames:  bviper.GetStringSlice("fileName", nil),
	}
}
