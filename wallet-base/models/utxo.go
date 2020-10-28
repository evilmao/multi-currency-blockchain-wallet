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
	SymbolID   uint            `gorm:"column:symbol_id;index"`
	TxHash     string          `gorm:"column:tx_hash;size:90;index"`
	BlockHash  string          `gorm:"column:block_hash;size:256"`
	Amount     decimal.Decimal `gorm:"type:decimal(32,20);default:0"`
	OutIndex   uint            `gorm:"type:int"`
	Address    string          `gorm:"size:256;index"`
	ScriptData string          `gorm:"size:500"`
	Status     uint            `gorm:"type:tinyint;index"`
	SpentID    string          `gorm:"column:spent_id;index"`
	Extra      string          `gorm:"column:extra;size:256"`
}

func (*UTXO) TableName() string { return "utxo" }

// FirstOrCreate find first matched record or create a new one.
func (u *UTXO) FirstOrCreate() error {
	if u.SymbolID == 0 {
		return fmt.Errorf("can't insert utxo with symbol id 0")
	}

	if u.Status == UTXOStatusNotRecord {
		u.Status = UTXOStatusRecord
	}
	return db.Default().FirstOrCreate(u, "symbol_id = ? and tx_hash = ? and out_index = ?",
		u.SymbolID, u.TxHash, u.OutIndex).Error
}

// Spend spends the utxo.
func (u *UTXO) Spend(spentID string) error {
	return db.Default().Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusSpent,
		"spent_id": spentID,
	}).Error
}

// Unspend unspends the utxo.
func (u *UTXO) Unspend() error {
	return db.Default().Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusRecord,
		"spent_id": "",
	}).Error
}

// GetUTXO gets utxo by tx hash and index.
func GetUTXO(symbolID uint, txHash string, index int) (*UTXO, error) {
	var utxo UTXO
	err := db.Default().Where("symbol_id = ? and tx_hash = ? and out_index = ?",
		symbolID, txHash, index).First(&utxo).Error
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
func GetUTXOsByAddress(address string) []*UTXO {
	var utxos []*UTXO
	db.Default().Where("address = ? and status = ?",
		address, UTXOStatusRecord).Find(&utxos)
	return utxos
}

// GetSmallUTXOsByAddress gets utxos that amount < maxAmount by the address.
func GetSmallUTXOsByAddress(address string, maxAmount decimal.Decimal) []*UTXO {
	var utxos []*UTXO
	db.Default().Where("address = ? and amount < ? and status = ?",
		address, maxAmount, UTXOStatusRecord).Find(&utxos)
	return utxos
}

func (u *UTXO) SpendWithDB(db *gorm.DB, spentID string) error {
	return db.Model(u).Updates(map[string]interface{}{
		"status":   UTXOStatusSpent,
		"spent_id": spentID,
	}).Error
}

func GetUTXOsByAddressWithDB(db *gorm.DB, symbolID uint, address string) []*UTXO {
	var utxos []*UTXO
	db.Where("symbol_id = ? and address = ? and status = ?",
		symbolID, address, UTXOStatusRecord).Find(&utxos)
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
