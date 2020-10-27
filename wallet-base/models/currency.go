package models

import (
	"upex-wallet/wallet-base/db"
)

// Currency represents a digital currency infomations.
type Currency struct {
	ID         uint   `gorm:"AUTO_INCREMENT" json:"id"`
	Blockchain string `gorm:"index"`
	Decimals   uint   `gorm:"type:int;default:18" json:"decimals"`
	Symbol     string `gorm:"index" json:"symbol"`
	Address    string `gorm:"size:100" json:"address"`
	Code       int    `gorm:"default:0" json:"code"`
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

// Insert inserts token infomations.
func (c *Currency) Insert() error {
	return db.Default().FirstOrCreate(c, c.whereQueryAndArgs()...).Error
}

func (c *Currency) Delete() error {
	return db.Default().Where(c.whereQuery(), c.whereArgs()...).Delete(c).Error
}

// GetCurrencies gets token currency list.
func GetCurrencies() []Currency {
	var (
		currencies []Currency
	)
	db.Default().Find(&currencies)
	return currencies
}

func CurrencyExistedBySymbol(blockchain, symbol string) bool {
	var c Currency
	db.Default().Where(c.whereQuery(), blockchain, symbol).First(&c)
	if c.Symbol == "" {
		return false
	}

	return c.Symbol == symbol
}

// Update updates the currency.
func (c *Currency) Update() error {
	data := map[string]interface{}{
		"blockchain": c.Blockchain,
		"decimals":   c.Decimals,
		"symbol":     c.Symbol,
		"address":    c.Address,
		"code":       c.Code,
	}
	return db.Default().Model(c).
		Where(c.whereQuery(), c.whereArgs()...).
		Updates(data).Error
}
