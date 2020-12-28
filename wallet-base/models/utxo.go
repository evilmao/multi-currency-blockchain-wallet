package models

import (
	"fmt"

	"upex-wallet/wallet-base/db"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// UTXO status.
const (
	UTXOStatusNotRecord = 0
	UTXOStatusRecord    = 1

	UTXOStatusSpent = 10
)

// UTXO represents unspent tx out.
type UTXO struct {
	gorm.Model
	Symbol     string          `gorm:"column:symbol;"`
	TxHash     string          `gorm:"column:tx_hash;size:90;index"`
	BlockHash  string          `gorm:"column:block_hash;size:255"`
	Amount     decimal.Decimal `gorm:"type:decimal(32,20);default:0"`
	OutIndex   uint            `gorm:"type:int"`
	Address    string          `gorm:"size:255;index"`
	ScriptData string          `gorm:"size:500"`
	Status     uint            `gorm:"type:tinyint;index"`
	SpentID    string          `gorm:"column:spent_id;index"`
	Extra      string          `gorm:"column:extra;size:256"`
}

func (*UTXO) TableName() string { return "utxo" }

// FirstOrCreate find first matched record or create a new one.
func (u *UTXO) FirstOrCreate() error {
	if u.Symbol == "" {
		return fmt.Errorf("can't insert utxo with symbol empty")
	}

	if u.Status == UTXOStatusNotRecord {
		u.Status = UTXOStatusRecord
	}
	return db.Default().FirstOrCreate(u, "symbol = ? and tx_hash = ? and out_index = ?",
		u.Symbol, u.TxHash, u.OutIndex).Error
}

// Spend spends the utxo.
func (u *UTXO) Spend(spentID string) error {
	return db.Default().Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusSpent,
		"spent_id": spentID,
	}).Error
}

// UnSpend unSpends the UTXO.
func (u *UTXO) UnSpend() error {
	return db.Default().Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusRecord,
		"spent_id": "",
	}).Error
}

// GetUTXO gets utxo by tx hash and index.
func GetUTXO(symbol uint, txHash string, index int) (*UTXO, error) {
	var utxo UTXO
	err := db.Default().Where("symbol = ? and tx_hash = ? and out_index = ?",
		symbol, txHash, index).First(&utxo).Error
	return &utxo, err
}

// GetSpentUTXOs gets spent utxos by the tx hash.
func GetSpentUTXOs(spentID string) []*UTXO {
	var utxos []*UTXO
	db.Default().Where("spent_id = ? and status = ?",
		spentID, UTXOStatusSpent).Find(&utxos)
	return utxos
}

// GetUTXOsByAddress gets utxos by the address.
func GetUTXOsByAddress(address, symbol string) []*UTXO {
	var utxos []*UTXO
	db.Default().Where("address = ? and status = ? and symbol= ?",
		address, UTXOStatusRecord, symbol).Find(&utxos)
	return utxos
}

// GetSmallUTXOsByAddress gets utxos that amount < maxAmount by the address.
func GetSmallUTXOsByAddress(symbol, address string, maxAmount decimal.Decimal) []*UTXO {
	var utxos []*UTXO
	db.Default().Where("symbol= ? and address = ? and amount < ? and status = ?",
		symbol, address, maxAmount, UTXOStatusRecord).Find(&utxos)
	return utxos
}

func (u *UTXO) SpendWithDB(db *gorm.DB, spentID string) error {
	return db.Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusSpent,
		"spent_id": spentID,
	}).Error
}

func GetUTXOsByAddressWithDB(db *gorm.DB, symbol string, address string) []*UTXO {
	var utxos []*UTXO
	db.Where("symbol = ? and address = ? and status = ?",
		symbol, address, UTXOStatusRecord).Find(&utxos)
	return utxos
}

func GetUTXOById(db *gorm.DB, id uint) *UTXO {
	var utxo UTXO
	db.Where("id = ?", id).Find(&utxo)
	return &utxo
}

func GetUTXOUnspentById(db *gorm.DB, id uint) *UTXO {
	var utxo UTXO
	db.Where("id = ? and status = ?", id, UTXOStatusRecord).Find(&utxo)
	return &utxo
}
