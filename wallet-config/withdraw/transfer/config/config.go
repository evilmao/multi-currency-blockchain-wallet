package config

import (
	"fmt"
	"strings"
	"time"

	bviper "upex-wallet/wallet-base/viper"
)

const (
	minWithdrawInterval = 5
	minGatherInterval   = 5
	minSignTimeout      = 5
	sendEmaiInterval    = 15
)

type ExchangeConfig struct {
	ExURL      string
	AuditURL   string
	PrivateKey string
}

type RequestFeeAPI struct {
	ApiFeeURL string
}

type LimitFeeRange struct {
	MinTxFee float64
	MaxTxFee float64
}

type Email struct {
	Host string
	Port string
	From string
	Pwd  string
	To   string
}

type TxError struct {
	TxType     string
	UpdateTime time.Time
	Error      error
}

// Config defines configurations of the exchange wallet.
type Config struct {
	Currency       string
	Code           int
	UpdateCurrency bool

	DSN      string
	RPCUrl   string
	RPCToken string
	ChainID  string

	// Withdraw interval in seconds.
	Withdraw         bool
	WithdrawInterval time.Duration

	Cooldown bool

	Gather bool
	// Gather interval in seconds.
	GatherInterval time.Duration

	AutoRollback bool

	ScheduleChecker []string

	// signer
	SignURL string
	// Part of the wallet.dat password, encrypted by signer-pubkey
	SignPass string
	// Sign timeout int second.
	SignTimeout time.Duration

	// wallet configuration
	BroadcastURL     string
	MaxFee           float64
	AlarmTimeout     int64
	MaxGasPrice      float64
	MaxGasLimit      float64
	MaxAccountRemain float64
	MinAccountRemain float64
	ColdAddress      string

	// broker api
	BrokerURL        string
	BrokerAccessKey  string
	BrokerPrivateKey string

	// Warning: use it carefully
	// TxTransIDForIgnoreReceiveNotify       string
	TxSequenceIDForUpdateTxHashToExchange string
	TxSequenceIDForRejectToExchange       string

	// Signal
	ExitSignal chan struct{}

	// get best fee api
	SuggestTransactionFees map[string]map[string]float64 // a map is used to store different currency
	GetFeeAPI              map[string]*RequestFeeAPI     // store all of currency questFeeAPI
	UpdateFeeInterval      time.Duration
	FeeFloatUp             float64                   // float up percent  for transaction fee
	FeeLimitMap            map[string]*LimitFeeRange // limit min fee and max fee for currency

	// email config
	EmailCfg *Email

	// tx error catch
	// {"btc":[{"withdraw","balance not enough",'2020-09-08 12:00:00'}]}
	ErrorCatch map[string][]*TxError

	// send email interval as same err
	ErrorAlarmInterval time.Duration
}

// DefaultConfig returns a default Wallet Config.
func DefaultConfig() *Config {
	return &Config{
		Currency:         "btc",
		Code:             0,
		UpdateCurrency:   false,
		RPCUrl:           "http://localhost:8545",
		RPCToken:         "",
		SignURL:          "http://localhost:8899",
		SignTimeout:      5,
		WithdrawInterval: time.Second * minWithdrawInterval,
		Cooldown:         false,
		Gather:           false,
		AutoRollback:     false,
		Withdraw:         true,
		GatherInterval:   time.Second * minGatherInterval,

		AlarmTimeout:     108000,
		MaxFee:           0,
		MaxGasPrice:      0,
		MaxGasLimit:      0,
		MaxAccountRemain: 0,
		MinAccountRemain: 0,
		ColdAddress:      "",
		ChainID:          "",
		DSN:              "",
		BroadcastURL:     "",
		ExitSignal:       make(chan struct{}),

		// btc fee api
		SuggestTransactionFees: make(map[string]map[string]float64),
		GetFeeAPI:              map[string]*RequestFeeAPI{"btc": &RequestFeeAPI{ApiFeeURL: "https://api.blockchain.info/mempool/fees"}},
		UpdateFeeInterval:      time.Second * 30,
		FeeFloatUp:             0.10,
		FeeLimitMap:            make(map[string]*LimitFeeRange),

		// email Config
		EmailCfg: &Email{},

		// error catch for send email
		ErrorCatch: make(map[string][]*TxError),

		ErrorAlarmInterval: time.Minute * 15,
	}
}

// New returns a new config instance.
func New() *Config {
	cfg := DefaultConfig()

	cfg.Currency = strings.ToLower(bviper.GetString("currency", cfg.Currency))
	cfg.UpdateCurrency = bviper.GetBool("updateCurrency", cfg.UpdateCurrency)

	cfg.RPCUrl = bviper.GetString("rpcUrl", cfg.RPCUrl)
	cfg.RPCToken = bviper.GetString("rpcToken", cfg.RPCToken)
	cfg.DSN = bviper.GetString("dsn", cfg.DSN)
	cfg.ChainID = bviper.GetString("chainID", cfg.ChainID)

	cfg.Withdraw = bviper.GetBool("withdraw", true)
	if cfg.Withdraw {
		interval := bviper.GetInt64("withdrawInterval", minWithdrawInterval)
		if interval < minWithdrawInterval {
			interval = minWithdrawInterval
		}
		cfg.WithdrawInterval = time.Second * time.Duration(interval)
	}

	cfg.Cooldown = bviper.GetBool("cooldown", false)

	cfg.Gather = bviper.GetBool("gather", false)
	if cfg.Gather {
		interval := bviper.GetInt64("gatherInterval", minGatherInterval)
		if interval < minGatherInterval {
			interval = minGatherInterval
		}
		cfg.GatherInterval = time.Second * time.Duration(interval)
	}

	cfg.AutoRollback = bviper.GetBool("autoRollback", false)
	cfg.ScheduleChecker = bviper.GetStringSlice("scheduleChecker", nil)

	cfg.SignURL = bviper.GetString("sign.url", cfg.SignURL)
	cfg.SignPass = bviper.GetString("sign.pass", cfg.SignPass)
	timeout := bviper.GetInt64("sign.timeout", minSignTimeout)
	if timeout < minSignTimeout {
		timeout = minSignTimeout
	}
	cfg.SignTimeout = time.Second * time.Duration(timeout)

	cfg.AlarmTimeout = bviper.GetInt64("wallet.alarmTimeout", cfg.AlarmTimeout)
	cfg.MaxFee = bviper.GetFloat64("wallet.maxFee", cfg.MaxFee)
	cfg.BroadcastURL = bviper.GetString("wallet.broadcastUrl", cfg.BroadcastURL)
	cfg.MaxGasPrice = bviper.GetFloat64("wallet.maxGasPrice", cfg.MaxGasPrice)
	cfg.MaxGasLimit = bviper.GetFloat64("wallet.maxGasLimit", cfg.MaxGasLimit)
	cfg.MaxAccountRemain = bviper.GetFloat64("wallet.maxAccountRemain", cfg.MaxAccountRemain)
	cfg.MinAccountRemain = bviper.GetFloat64("wallet.minAccountRemain", cfg.MinAccountRemain)
	cfg.ColdAddress = bviper.GetString("wallet.coldAddress", cfg.ColdAddress)

	cfg.BrokerURL = bviper.GetString("broker.url", cfg.BrokerURL)
	cfg.BrokerAccessKey = bviper.GetString("broker.accessKey", cfg.BrokerAccessKey)
	cfg.BrokerPrivateKey = bviper.GetString("broker.privateKey", cfg.BrokerPrivateKey)

	// cfg.TxTransIDForIgnoreReceiveNotify = bviper.GetString("txTransIDForIgnoreReceiveNotify", "")
	cfg.TxSequenceIDForUpdateTxHashToExchange = bviper.GetString("txSequenceIDForUpdateTxHashToExchange", "")
	cfg.TxSequenceIDForRejectToExchange = bviper.GetString("txSequenceIDForRejectToExchange", "")

	// fee apis
	feeAPIs := bviper.GetStringMap("transfer.feeapis", nil)

	if feeAPIs != nil {
		for currency := range feeAPIs {

			APIFeeURL := bviper.GetString(fmt.Sprintf("transfer.feeapis.%s.apiFeeUrl", currency), "")
			cfg.GetFeeAPI[currency] = &RequestFeeAPI{
				APIFeeURL,
			}
			// write currency fee range
			minTxFee := bviper.GetFloat64(fmt.Sprintf("transfer.feeapis.%s.minTxFee", currency), 5)
			maxTxFee := bviper.GetFloat64(fmt.Sprintf("transfer.feeapis.%s.maxTxFee", currency), 100)
			cfg.FeeLimitMap[currency] = &LimitFeeRange{
				MinTxFee: minTxFee,
				MaxTxFee: maxTxFee,
			}
		}
	}

	cfg.UpdateFeeInterval = bviper.GetDuration("transfer.btc.updateFeeInterval", cfg.UpdateFeeInterval)
	cfg.FeeFloatUp = bviper.GetFloat64("transfer.feeFloatUp", cfg.FeeFloatUp)

	// email config
	cfg.EmailCfg = &Email{
		Host: bviper.GetString("email.host", ""),
		Port: bviper.GetString("email.port", ""),
		From: bviper.GetString("email.from", ""),
		Pwd:  bviper.GetString("email.password", ""),
		To:   bviper.GetString("email.to", ""),
	}

	interval := bviper.GetFloat64("email.errorAlarmInterval", sendEmaiInterval)
	cfg.ErrorAlarmInterval = time.Minute * time.Duration(interval)
	return cfg
}
