package currency

import (
    "fmt"
    "math"
    "strings"
    "sync"
    "time"

    "upex-wallet/wallet-base/models"
    "upex-wallet/wallet-base/newbitx/misclib/log"
    "upex-wallet/wallet-base/util"
    "upex-wallet/wallet-config/deposit/config"

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
}

var (
    rwMutex                 sync.RWMutex
    symbolDetailIndex       = map[string][]*CurrencyInfo{}
    addressIndex            = map[string]*CurrencyInfo{}
    defaultMinDepositAmount = decimal.New(math.MaxUint32, 0)
    scheduleUpdateOn        bool
)

func Init(cfg *config.Config) {

    log.Infof("start init currency...")

    Update(cfg)

    log.Infof("finish init currency successfully")
}

func ScheduleUpdate(cfg *config.Config, interval time.Duration) {
    rwMutex.Lock()
    defer rwMutex.Unlock()

    if scheduleUpdateOn {
        return
    }

    util.Go("currency-schedule-update", func() {
        for {
            time.Sleep(interval)

            Update(cfg)
        }
    }, nil)

    scheduleUpdateOn = true
}

// Update, update currency from config to db
func Update(cfg *config.Config) {

    err := updateCurrencyDetails(cfg)
    if err != nil {
        log.Errorf("update currency fail,%v,", err)
    }

    updateCurrencyTable(cfg)
}

func updateCurrencyDetails(cfg *config.Config) error {
    var (
        symbols      = cfg.Symbols
        mainCurrency = cfg.Currency
        tmpC         = &CurrencyInfo{}
    )

    rwMutex.Lock()
    defer rwMutex.Unlock()

    if mainCurrency == "" {
        return fmt.Errorf("currency can not be empty")
    }

    symbolDetailIndex = map[string][]*CurrencyInfo{}
    addressIndex = map[string]*CurrencyInfo{}

    // main currency detail
    symbolDetailIndex[mainCurrency] = append(symbolDetailIndex[mainCurrency],
        &CurrencyInfo{
            BlockchainName:   mainCurrency,
            Symbol:           mainCurrency,
            Confirm:          cfg.MaxConfirm,
            MinDepositAmount: cfg.MinAmount,
        })

    // token details
    for _, s := range symbols {
    	symbol := s.Symbol
        tmpC = &CurrencyInfo{
            BlockchainName:   mainCurrency,
            Symbol:           symbol,
            Address:          s.Address,
            Confirm:          cfg.MaxConfirm,
            Decimal:          int(s.Precision),
            MinDepositAmount: s.MinDepositAmount,
        }

        symbolDetailIndex[symbol] = append(symbolDetailIndex[symbol], tmpC)

        if s.Address != "" {
            address := s.Address
            addressIndex[address] = tmpC
        }
    }

    log.Infof("update token, total: %d", len(symbolDetailIndex))
    return nil
}

func updateCurrencyTable(cfg *config.Config) {

    if cfg.Currency == "" {
        return
    }

    deleteCurrencyTable(cfg)

    insertOrUpdateCurrencyTable(cfg)
}

func CurrencyDetail(symbol string) *CurrencyInfo {

    c := models.GetCurrencyBySymbol(symbol)
    if symbol == "" || c == nil {
        return nil
    }

    return &CurrencyInfo{
        BlockchainName: c.Blockchain,
        Address:        c.Address,
        Decimal:        int(c.Decimals),
        Symbol:         c.Symbol,
    }
}

func CurrencyDetailByAddress(address string) (*CurrencyInfo, bool) {

    address = strings.ToLower(address)
    c := models.GetCurrencyByContractAddress(address)
    if c == nil {
        return nil, false
    }

    return &CurrencyInfo{
        BlockchainName: c.Blockchain,
        Symbol:         c.Symbol,
        Address:        c.Address,
        Decimal:        int(c.Decimals),
    }, true

}

func firstDetailWithBlockchainName(symbol string) (*CurrencyInfo, bool) {

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

func MinAmount(symbol string) (decimal.Decimal, bool) {
    rwMutex.RLock()
    defer rwMutex.RUnlock()

    detail, ok := firstDetailWithBlockchainName(symbol)
    if !ok {
        return defaultMinDepositAmount, false
    }

    amount := decimal.NewFromFloat(detail.MinDepositAmount)
    return amount, true
}

func deleteCurrencyTable(cfg *config.Config) {
    var (
        dbCurrencies = models.GetCurrencies()
        symbols      = cfg.Symbols
    )

    Delete := func(c *models.Currency) {
        err := c.Delete()
        if err != nil {
            log.Errorf("delete currency item, blockchain: %s, symbol: %s fail,%v", c.Blockchain, c.Symbol, err)
        }
        log.Warnf("delete currency item, blockchain: %s, symbol: %s", cfg.Currency, strings.ToUpper(c.Symbol))
    }

    // Can't find currency in config then delete from currency table
    for i := 0; i < len(dbCurrencies); i++ {
        var (
            c         = dbCurrencies[i]
            canDelete = true
        )

        if len(symbols) == 0 {
            Delete(&c)
            continue
        }

        for _, s := range symbols {
            if c.Symbol == s.Symbol && c.Blockchain == s.Blockchain {
                canDelete = false
                break
            }
        }

        if canDelete {
            Delete(&c)
            continue
        }
    }
}

func insertOrUpdateCurrencyTable(cfg *config.Config) {
    symbols := cfg.Symbols

    for i := 0; i < len(symbols); i++ {
        symbolDetail := symbols[i]
        c := &models.Currency{
            Blockchain: symbolDetail.Blockchain,
            Symbol:     symbolDetail.Symbol,
        }
        // insert a new symbol for currency
        err := c.Insert()
        if err != nil {
            log.Errorf("Init symbol %s fail, %v", symbolDetail.Symbol, err)
        }

        // update existed symbol
        c.Address = symbolDetail.Address
        c.Decimals = symbolDetail.Precision
        c.MinBalance = fmt.Sprintf("%"+fmt.Sprintf(".%d", symbolDetail.Precision)+"f", symbolDetail.MinBalanceRemain)
        c.MaxBalance = fmt.Sprintf("%"+fmt.Sprintf(".%d", symbolDetail.Precision)+"f", symbolDetail.MaxBalanceRemain)

        err = c.Update()
        if err != nil {
            log.Errorf("update symbol %s fail, %v", symbols[i].Symbol, err)
        } else {
            log.Warnf("init symbol [%s] successfully", symbols[i].Symbol)
        }
    }
}

func Symbols(mainCurrency string) []string {

    var (
        currencies = models.GetCurrencies()
        symbols    = []string{mainCurrency}
    )

    for i := 0; i < len(currencies); i++ {
        symbols = append(symbols, strings.ToLower(currencies[i].Symbol))
    }

    return symbols
}
