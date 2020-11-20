package models

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-base/db"
)

// Currency represents a digital currency information.
type Currency struct {
	ID         uint   `gorm:"AUTO_INCREMENT" json:"id"`
	Decimals   uint   `gorm:"type:int;default:18" json:"decimals"`
	Blockchain string `gorm:"index"`
	Symbol     string `gorm:"index;" json:"symbol"`
	Address    string `gorm:"size:100;" json:"address"`
	MinBalance string `gorm:"size:32;"`
	MaxBalance string `gorm:"size:32;"`
}

// TableName defines the table name of currency.
func (c Currency) TableName() string { return "currency" }

func (c *Currency) whereQuery() string {
	return "blockchain = ? and symbol = ?"
}

func (c *Currency) whereArgs() []interface{} {
	return []interface{}{c.Blockchain, c.Symbol}
}

func (c *Currency) whereQueryAndArgs() []interface{} {
	return append([]interface{}{c.whereQuery()}, c.whereArgs()...)
}

// Insert inserts token information.
func (c *Currency) Insert() error {
	return db.Default().FirstOrCreate(c, c.whereQueryAndArgs()...).Error
}

func (c *Currency) Delete() error {
	return db.Default().Where(c.whereQuery(), c.whereArgs()...).Delete(c).Error
}

// Update updates the currency.
func (c *Currency) Update() error {
	data := map[string]interface{}{
		"blockchain":  c.Blockchain,
		"decimals":    c.Decimals,
		"symbol":      c.Symbol,
		"address":     c.Address,
		"min_balance": c.MinBalance,
		"max_balance": c.MaxBalance,
	}
	return db.Default().Model(c).
		Where(c.whereQuery(), c.whereArgs()...).
		Updates(data).Error
}

// GetCurrencies gets token currency list.
func GetCurrencies() []Currency {
	var (
		currencies []Currency
	)
	db.Default().Find(&currencies)
	return currencies
}

// GetCurrencies gets token currency list.
func GetCurrency(mainCurrency, symbol string) *Currency {
	var currency Currency
	err := db.Default().Where("blockchain = ? and symbol= ?", mainCurrency, symbol).First(&currency).Error
	if err != nil {
		return nil
	}
	return &currency
}

// GetCurrencies gets token currency list.
func GetCurrencyByContractAddress(address string) *Currency {
	var currency Currency
	err := db.Default().Where("address= ?", address).First(&currency).Error
	if err != nil {
		return nil
	}
	return &currency
}

func CurrencyExistedBySymbol(blockchain, symbol string) bool {
	var c Currency
	db.Default().Where(c.whereQuery(), blockchain, symbol).First(&c)
	if c.Symbol == "" {
		return false
	}

	return c.Symbol == symbol
}

func GetCurrencyBySymbol(symbol string) *Currency {
	var currency Currency
	err := db.Default().Where("symbol= ?", symbol).First(&currency).Error
	if err != nil {
		return nil
	}
	return &currency
}

func BulkInsert(symbols []*Currency) {
	valueStrings := make([]string, 0, len(symbols))
	valueArgs := make([]interface{}, 0, len(symbols)*4)
	for _, s := range symbols {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, s.Blockchain)
		valueArgs = append(valueArgs, s.Decimals)
		valueArgs = append(valueArgs, s.Symbol)
		valueArgs = append(valueArgs, s.Address)
	}
	stmt := fmt.Sprintf("INSERT INTO currency (blcokchain, decimals, symbol, address) VALUES %s", strings.Join(valueStrings, ","))
	_ = db.Default().Exec(stmt, valueArgs...)
}
