package config

import (
	"strings"
	"time"

	bviper "upex-wallet/wallet-base/viper"
)

// Config defines configurations of the sync.
type Config struct {
	// Outer config.
	DSN                   string
	RPCURL                string
	RPCToken              string
	BrokerURL             string
	BrokerAccessKey       string
	BrokerPrivateKey      string
	Currency              string
	Code                  int
	ListenAddress         string
	StatusAddress         string
	MaxConfirm            int
	SecConfirm            int
	StartHeight           int64
	StartHash             string
	StartDate             time.Time
	IgnoreNotifyAudit     bool
	IgnoreBlockStuckCheck bool
	IsNeedTag             bool
	UseBlockTable         bool
	MinAmount             float64

	TrxTokenAirDropAddress string
	ChainID                string

	// Inner config.
	ForceTxs []string
}

// DefaultConfig returns a default Wallet Config.
func DefaultConfig() *Config {
	return &Config{
		DSN:                   "root:@tcp(127.0.0.1:3306)/btc?charset=utf8mb4&parseTime=True&loc=Local",
		RPCURL:                "http://111:222@127.0.0.1:8332",
		Currency:              "BTC",
		MaxConfirm:            2,
		SecConfirm:            1,
		StartHeight:           0,
		StartHash:             "",
		StartDate:             time.Date(2018, 6, 11, 6, 50, 0, 0, time.UTC),
		IgnoreNotifyAudit:     false,
		IgnoreBlockStuckCheck: false,
		IsNeedTag:             false,
		UseBlockTable:         true,
		ListenAddress:         ":8051",
		StatusAddress:         "127.0.0.1:8125",
	}
}

// New returns a new config instance.
func New() *Config {
	cfg := DefaultConfig()

	cfg.DSN = bviper.GetString("dsn", cfg.DSN)
	cfg.RPCURL = bviper.GetString("rpcUrl", cfg.RPCURL)
	cfg.RPCToken = bviper.GetString("rpcToken", cfg.RPCToken)
	cfg.BrokerURL = bviper.GetString("brokerUrl", cfg.BrokerURL)
	cfg.BrokerAccessKey = bviper.GetString("brokerAccessKey", cfg.BrokerAccessKey)
	cfg.BrokerPrivateKey = bviper.GetString("brokerPrivateKey", cfg.BrokerPrivateKey)
	cfg.Currency = strings.ToUpper(bviper.GetString("currency", cfg.Currency))
	cfg.MaxConfirm = int(bviper.GetInt64("maxConfirmations", int64(cfg.MaxConfirm)))
	cfg.SecConfirm = int(bviper.GetInt64("securityConfirmations", int64(cfg.SecConfirm)))
	cfg.StartHeight = bviper.GetInt64("startHeight", cfg.StartHeight)
	cfg.StartHash = bviper.GetString("startHash", cfg.StartHash)
	cfg.ListenAddress = bviper.GetString("listenAddress", cfg.ListenAddress)
	cfg.StatusAddress = bviper.GetString("statusAddress", cfg.StatusAddress)
	cfg.IgnoreNotifyAudit = bviper.GetBool("ignoreNotifyAudit", cfg.IgnoreNotifyAudit)
	cfg.IgnoreBlockStuckCheck = bviper.GetBool("ignoreBlockStuckCheck", cfg.IgnoreBlockStuckCheck)
	cfg.IsNeedTag = bviper.GetBool("isNeedTag", cfg.IsNeedTag)
	cfg.UseBlockTable = bviper.GetBool("useBlockTable", cfg.UseBlockTable)
	cfg.TrxTokenAirDropAddress = bviper.GetString("trxAirDropAddress", cfg.TrxTokenAirDropAddress)
	cfg.ChainID = bviper.GetString("chainID", cfg.ChainID)
	cfg.MinAmount = bviper.GetFloat64("minAmount", 0)

	startDate := bviper.GetInt64("startDate", 0)
	if startDate > 0 {
		cfg.StartDate = time.Unix(startDate, 0)
	}

	return cfg
}
