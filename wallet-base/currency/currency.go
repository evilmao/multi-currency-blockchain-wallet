package currency

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"

	"github.com/shopspring/decimal"
)

var (
	// MaxDecimalDivisionPrecision is the number of digits to the right of
	// the decimal point (the scale) for decimal division.
	MaxDecimalDivisionPrecision = 30
)

func init() {
	if decimal.DivisionPrecision < MaxDecimalDivisionPrecision {
		decimal.DivisionPrecision = MaxDecimalDivisionPrecision
	}
	models.GetCode = Code
}

var (
	brokerAPI         *api.BrokerAPI
	rwMutex           sync.RWMutex
	symbolIndex       = map[string]*api.CurrencyData{}
	symbolDetailIndex = map[string][]*api.CurrencyDetail{}
	addressIndex      = map[string]*api.CurrencyDetail{}

	scheduleUpdateOn bool
)

func Init(brokerURL, brokerAccessKey, brokerPrivateKey string, blockchain string) error {
	if len(brokerURL) == 0 {
		return fmt.Errorf("broker api url is empty")
	}

	log.Infof("start init currency...")

	brokerAPI = api.NewBrokerAPI(brokerURL, brokerAccessKey, brokerPrivateKey)
	// err := Update(blockchain)
	// if err != nil {
	// 	return err
	// }

	log.Infof("finish init currency success")

	return nil
}

func ScheduleUpdate(blockchain string, interval time.Duration) {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	if scheduleUpdateOn {
		return
	}

	util.Go("currency-schedule-update", func() {
		for {
			time.Sleep(interval)

			err := Update(blockchain)
			if err != nil {
				log.Errorf("currency schedule update failed, %v", err)
			}
		}
	}, nil)

	scheduleUpdateOn = true
}

func Update(blockchain string) error {
	var (
		err error
	)

	err = updateCurrency()
	if err != nil {
		return err
	}

	err = updateCurrencyDetails()
	if err != nil {
		return err
	}

	err = updateCurrencyTable(blockchain)
	if err != nil {
		return err
	}

	return nil
}

func updateCurrency() error {
	var (
		resp *api.CurrenciesResponse
		err  error
	)

	err = util.TryWithInterval(20, time.Second*3, func(int) error {
		resp, err = brokerAPI.Currencies()
		return err
	})
	if err != nil {
		return err
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()

	symbolIndex = map[string]*api.CurrencyData{}
	for _, v := range resp.Data {
		symbolIndex[strings.ToUpper(v.Symbol)] = v
	}

	log.Infof("update currency, total: %d", len(symbolIndex))

	return nil
}

func updateCurrencyDetails() error {
	var (
		resp *api.CurrencyDetailResponse
		err  error
	)

	err = util.TryWithInterval(20, time.Second*3, func(int) error {
		resp, err = brokerAPI.CurrencyDetails()
		return err
	})
	if err != nil {
		return err
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()

	symbolDetailIndex = map[string][]*api.CurrencyDetail{}
	addressIndex = map[string]*api.CurrencyDetail{}
	for _, v := range resp.Data {
		symbol := strings.ToUpper(v.Symbol)
		symbolDetailIndex[symbol] = append(symbolDetailIndex[symbol], v)

		if v.Address != "" {
			address := strings.ToLower(v.Address)
			addressIndex[address] = v
		}
	}

	log.Infof("update token, total: %d", len(symbolDetailIndex))

	return nil
}

func updateCurrencyTable(blockchain string) error {
	var (
		code int
		ok   bool
		err  error
	)

	if blockchain == "" {
		return nil
	}

	// Can't find currency in dw api returns data then delete from currency table
	currencies := models.GetCurrencies()
	for i := 0; i < len(currencies); i++ {
		if _, isIn := CurrencyDetailByAddress(currencies[i].Address); !isIn {
			err = deleteCurrency(blockchain, currencies[i].Symbol)
			if err != nil {
				log.Errorf("Cant find currency, delete currency item Error: %s(%d)",
					strings.ToUpper(currencies[i].Symbol), currencies[i].Code)
				continue
			}
		}
	}

	for k, vs := range symbolDetailIndex {
		for _, v := range vs {
			// Check token.
			if !v.IsToken() {
				continue
			}

			// Check token belong blockchain.
			if !v.ChainBelongTo(blockchain) {
				continue
			}

			// Can't find code then delete from currency table and continue
			if code, ok = Code(k); !ok {
				err = deleteCurrency(v.BelongChainName(), v.Symbol)
				if err != nil {
					log.Errorf("cant find code, delete currency item failed: %s(%d)",
						strings.ToUpper(v.Symbol), code)
				}
				continue
			}

			// Deposit Disabled then delete from currency table and continue
			if isEnable, _ := DepositEnabled(v.Symbol); !isEnable {
				err = deleteCurrency(v.BelongChainName(), v.Symbol)
				if err != nil {
					log.Errorf("deposit disabled, delete currency item failed: %s(%d)",
						strings.ToUpper(v.Symbol), code)
				}
				continue
			}

			cur := models.Currency{
				Blockchain: v.BelongChainName(),
				Decimals:   uint(v.Decimal),
				Symbol:     strings.ToUpper(v.Symbol),
				Address:    v.Address,
				Code:       code,
			}
			if !models.CurrencyExistedBySymbol(v.BelongChainName(), strings.ToUpper(v.Symbol)) {
				err = cur.Insert()
				if err != nil {
					return err
				}
			} else {
				err = cur.Update()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func Code(symbol string) (int, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	symbol = strings.ToUpper(symbol)
	c, ok := symbolIndex[symbol]
	if !ok {
		return 0, false
	}

	code, err := strconv.Atoi(c.Code)
	if err != nil {
		return 0, false
	}

	return code, true
}

func CurrencyDetail(symbol string) ([]*api.CurrencyDetail, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	symbol = strings.ToUpper(symbol)
	cs, ok := symbolDetailIndex[symbol]
	if !ok {
		return nil, false
	}
	return cs, ok
}

func CurrencyDetailByAddress(address string) (*api.CurrencyDetail, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	address = strings.ToLower(address)
	c, ok := addressIndex[address]
	if !ok {
		return nil, false
	}
	return c, ok
}

func ForeachCurrencyDetail(h func(*api.CurrencyDetail) (bool, error)) error {
	for _, details := range symbolDetailIndex {
		for _, detail := range details {
			ok, err := h(detail)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}
		}
	}
	return nil
}

func firstDetailWithBlockchainName(symbol string) (*api.CurrencyDetail, bool) {
	symbol = strings.ToUpper(symbol)
	cs, ok := symbolDetailIndex[symbol]
	if !ok || len(cs) == 0 {
		return nil, false
	}

	var idx int
	for i, c := range cs {
		if len(c.BlockchainName) > 0 {
			idx = i
			break
		}
	}

	return cs[idx], true
}

var defaultMinDepositAmount = decimal.New(math.MaxUint32, 0)

func MinAmount(symbol string) (decimal.Decimal, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	detail, ok := firstDetailWithBlockchainName(symbol)
	if !ok {
		return defaultMinDepositAmount, false
	}

	amount, err := decimal.NewFromString(detail.MinDepositAmount)
	return amount, err == nil
}

func DepositEnabled(symbol string) (bool, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	symbol = strings.ToUpper(symbol)
	c, ok := symbolIndex[symbol]
	if !ok {
		return false, false
	}
	return c.DepositEnabled, ok
}

func WithdrawEnabled(symbol string) (bool, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	symbol = strings.ToUpper(symbol)
	c, ok := symbolIndex[symbol]
	if !ok {
		return false, false
	}
	return c.WithdrawEnabled, ok
}

func MaxWithdrawAmount(symbol string) (decimal.Decimal, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	detail, ok := firstDetailWithBlockchainName(symbol)
	if !ok {
		return decimal.Zero, false
	}

	amount, err := decimal.NewFromString(detail.MaxWithdrawAmount)
	return amount, err == nil
}

func deleteCurrency(blockchain, symbol string) error {
	if models.CurrencyExistedBySymbol(blockchain, symbol) {
		cur := models.Currency{
			Blockchain: blockchain,
			Symbol:     strings.ToUpper(symbol),
		}
		err := cur.Delete()
		log.Warnf("delete currency item, blockchain: %s, symbol: %s, error: %v",
			blockchain, strings.ToUpper(symbol), err)
		return err
	}
	return nil
}
