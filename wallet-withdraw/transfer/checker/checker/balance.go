package checker

import (
	"upex-wallet/wallet-base/api"
	bmodels "upex-wallet/wallet-base/models"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/shopspring/decimal"

	"upex-wallet/wallet-config/withdraw/transfer/config"
)

func init() {
	Add("eth", NewUSDTBalanceChecker("eth", "erc20", decimal.New(300000, 0)))
	Add("trx", NewUSDTBalanceChecker("trx", "trx_trc20", decimal.New(300000, 0)))
}

type USDTBalanceChecker struct {
	mainCurrency   string
	blockchainName string
	minBalance     decimal.Decimal

	cfg             *config.Config
	brokerAPI       *api.BrokerAPI
	withdrawEnabled bool
}

func NewUSDTBalanceChecker(mainCurrency, blockchainName string, minBalance decimal.Decimal) *USDTBalanceChecker {
	return &USDTBalanceChecker{
		mainCurrency:   mainCurrency,
		blockchainName: blockchainName,
		minBalance:     minBalance,
	}
}

func (c *USDTBalanceChecker) Name() string {
	return "USDTBalanceChecker"
}

func (c *USDTBalanceChecker) Init(cfg *config.Config) {
	c.cfg = cfg
	c.brokerAPI = api.NewBrokerAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey)
}

func (c *USDTBalanceChecker) Check() error {
	if c.cfg.Currency != c.mainCurrency {
		return nil
	}

	const usdtCode = 105

	balance := bmodels.GetBalance(usdtCode)
	if balance.LessThan(c.minBalance) && c.withdrawEnabled {
		c.brokerAPI.ChangeWithdrawStatus("usdt", c.blockchainName, false)
		c.withdrawEnabled = false
		log.Warnf("checker, %s usdt balance (%s) less than minBalance (%s), disable withdraw",
			c.blockchainName, balance, c.minBalance)
	}

	if balance.GreaterThan(c.minBalance) && !c.withdrawEnabled {
		c.brokerAPI.ChangeWithdrawStatus("usdt", c.blockchainName, true)
		c.withdrawEnabled = true
		log.Warnf("checker, %s usdt balance (%s) greater than minBalance (%s), enable withdraw",
			c.blockchainName, balance, c.minBalance)
	}

	return nil
}
