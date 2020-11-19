package checker

import (
	"fmt"
	"strings"
	"time"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/alarm"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	headContent = "【Balance Not Enough】 "
	content     = "%s balance(%s) less than minimum remain balance(%s),\n"
	tailContent = "pls deposit ASAP."
)

type BalanceChecker struct {
	cfg                    *config.Config
	lastBalanceCheckerTime time.Time
}

// NewBalanceChecker, check symbol balance
func NewBalanceChecker(cfg *config.Config, t time.Time) *BalanceChecker {
	return &BalanceChecker{
		lastBalanceCheckerTime: t,
		cfg:                    cfg,
	}
}

func (c *BalanceChecker) Name() string {
	return fmt.Sprintf("%sBalanceChecker", strings.ToUpper(c.cfg.Currency))
}

func (c *BalanceChecker) Init(cfg *config.Config) {
	return
}

func (c *BalanceChecker) Check() error {
	now := time.Now()
	if now.Sub(c.lastBalanceCheckerTime) < time.Minute*c.cfg.CoolDownTaskInterval {
		return nil
	}

	var (
		currency   = strings.ToLower(c.cfg.Currency)
		minBalance = decimal.NewFromFloat(c.cfg.MinAccountRemain)
	)

	if currency == "" || minBalance.LessThan(decimal.Zero) {
		err := fmt.Errorf("main currency or MinAccountRemain set wrong, check `currency` and `minAccountRemain` fields ")
		log.Errorf("Balance checker fail,%v", err)
		return err
	}

	c1 := mainCurrencyBalanceChecker(currency, minBalance)
	c2 := tokenCurrencyBalanceChecker()

	if c1 != "" || c2 != "" {
		emailContent := headContent + c1 + c2 + tailContent
		go alarm.SendEmailByText(c.cfg, emailContent)
	}

	c.lastBalanceCheckerTime = now
	return nil
}

func mainCurrencyBalanceChecker(currency string, minBalance decimal.Decimal) string {
	var (
		tmpContent = ""
	)

	mainCurrencyBalance := bmodels.GetBalance(currency)
	if mainCurrencyBalance.LessThan(minBalance) {
		tmpContent = fmt.Sprintf(content, currency, minBalance.String(), mainCurrencyBalance.String())
	}

	return tmpContent
}

func tokenCurrencyBalanceChecker() string {
	var (
		currencies = bmodels.GetCurrencies()
		tmpContent = ""
	)

	if len(currencies) == 0 {
		return ""
	}

	for _, c := range currencies {
		var (
			symbol              = strings.ToLower(c.Symbol)
			balance             = bmodels.GetBalance(symbol)
			minRemainBalance, _ = decimal.NewFromString(c.MinBalance)
		)
		if balance.LessThan(minRemainBalance) {
			tmpContent += fmt.Sprintf(content, symbol, balance, minRemainBalance)
		}
	}

	return tmpContent
}
