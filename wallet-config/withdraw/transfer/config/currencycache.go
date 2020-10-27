package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

var (
	CC = NewCurrencyCache()
)

type CurrencyCache struct {
	sync.RWMutex

	mainCurrency   string
	mainCode       int
	currencyToCode map[string]int
	codeToCurrency map[int]string
	codes          []int

	scheduleUpdateOn bool
}

func NewCurrencyCache() *CurrencyCache {
	cache := &CurrencyCache{}
	cache.reset()
	return cache
}

func (c *CurrencyCache) deferAutoRLock() func() {
	c.RLock()
	return c.RUnlock
}

func (c *CurrencyCache) deferAutoLock() func() {
	c.Lock()
	return c.Unlock
}

func (c *CurrencyCache) SetMainCurrency(currency string, code int) {
	defer c.deferAutoLock()()

	c.mainCurrency = strings.ToLower(currency)
	c.mainCode = code
}

func (c *CurrencyCache) ScheduleUpdate(interval time.Duration) error {
	c.Lock()
	if c.scheduleUpdateOn {
		c.Unlock()
		return nil
	}

	c.scheduleUpdateOn = true
	c.Unlock()

	err := c.Update()
	if err != nil {
		return err
	}

	util.Go("currency-cache-schedule-update", func() {
		time.Sleep(interval)

		err := c.Update()
		if err != nil {
			log.Errorf("currency cache schedule update failed, %v", err)
		}
	}, nil)
	return nil
}

func (c *CurrencyCache) Update() error {
	defer c.deferAutoLock()()

	c.reset()
	c.currencyToCode[c.mainCurrency] = c.mainCode
	c.codeToCurrency[c.mainCode] = c.mainCurrency
	c.codes = append(c.codes, c.mainCode)
	return currency.ForeachCurrencyDetail(func(detail *api.CurrencyDetail) (bool, error) {
		if !detail.IsToken() {
			return true, nil
		}

		if !detail.ChainBelongTo(c.mainCurrency) {
			return true, nil
		}

		if ok, _ := currency.WithdrawEnabled(detail.Symbol); !ok {
			return true, nil
		}

		code, ok := currency.Code(detail.Symbol)
		if !ok {
			return false, fmt.Errorf("can't find code of currency %s", detail.Symbol)
		}

		symbol := strings.ToLower(detail.Symbol)
		c.currencyToCode[symbol] = code
		c.codeToCurrency[code] = symbol
		c.codes = append(c.codes, code)
		return true, nil
	})
}

func (c *CurrencyCache) Code(currency string) (int, bool) {
	defer c.deferAutoRLock()()

	currency = strings.ToLower(currency)
	code, ok := c.currencyToCode[currency]
	return code, ok
}

func (c *CurrencyCache) Currency(code int) (string, bool) {
	defer c.deferAutoRLock()()

	currency, ok := c.codeToCurrency[code]
	return currency, ok
}

func (c *CurrencyCache) Foreach(f func(string, int) error) error {
	if f == nil {
		return nil
	}

	defer c.deferAutoRLock()()
	for currency, code := range c.currencyToCode {
		err := f(currency, code)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CurrencyCache) Codes() []int {
	defer c.deferAutoRLock()()

	codes := make([]int, len(c.codes))
	copy(codes, c.codes)
	return codes
}

func (c *CurrencyCache) String() string {
	defer c.deferAutoRLock()()

	return fmt.Sprintf("%v", c.currencyToCode)
}

func (c *CurrencyCache) reset() {
	c.currencyToCode = make(map[string]int, len(c.currencyToCode))
	c.codeToCurrency = make(map[int]string, len(c.codeToCurrency))
	c.codes = make([]int, 0, len(c.codes))
}
