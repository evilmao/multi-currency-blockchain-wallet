package checker

import (
	"fmt"

	"upex-wallet/wallet-base/api"
	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"

	"github.com/shopspring/decimal"
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

// NewUSDTBalanceChecker, USD
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

	const symbol = "usdt"

	balance := bmodels.GetBalance(symbol)
	if balance.LessThan(c.minBalance) && c.withdrawEnabled {
		_, err := c.brokerAPI.ChangeWithdrawStatus("usdt", c.blockchainName, false)
		if err != nil {
			return fmt.Errorf("change Withdraw Status fail, %v", err.Error())
		}
		c.withdrawEnabled = false
		log.Warnf("checker, %s usdt balance (%s) less than minBalance (%s), disable withdraw",
			c.blockchainName, balance, c.minBalance)
	}

	if balance.GreaterThan(c.minBalance) && !c.withdrawEnabled {
		_, err := c.brokerAPI.ChangeWithdrawStatus("usdt", c.blockchainName, false)
		if err != nil {
			return fmt.Errorf("change Withdraw Status fail, %v", err.Error())
		}
		c.withdrawEnabled = true
		log.Warnf("checker, %s usdt balance (%s) greater than minBalance (%s), enable withdraw",
			c.blockchainName, balance, c.minBalance)
	}

	return nil
}
