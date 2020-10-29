package models

import (
	"fmt"
	"math/big"

	"upex-wallet/wallet-base/db"

	"github.com/jinzhu/gorm"
)

const (
	maxBlockCount = 10
)

func blockInnerIdx(height uint64) uint {
	// We don't use the default value 0.
	return uint(height%maxBlockCount) + 1
}

// BlockInfo represents a block info.
type BlockInfo struct {
	ID         uint   `gorm:"AUTO_INCREMENT" json:"id"`
	Height     uint64 `gorm:"type:bigint;default:0;index" json:"height"`
	Hash       string `gorm:"size:90;index" json:"hash"`
	Symbol     string `gorm:"size:10;index" json:"symbol"`
	InnerIndex uint   `gorm:"type:int;default:0;index"`
}

// TableName defines the table name of status.
func (b BlockInfo) TableName() string { return "blockinfo" }

// Number returns the block height.
func (b BlockInfo) Number() *big.Int {
	return new(big.Int).SetUint64(b.Height)
}

// Insert inserts new block info to block table.
func (b *BlockInfo) Insert() error {
	if len(b.Hash) == 0 {
		b.Hash = fmt.Sprintf("%d", b.Height)
	}

	innerIdx := blockInnerIdx(b.Height)

	var b1 BlockInfo
	err := db.Default().First(&b1, "symbol=? and inner_index=?", b.Symbol, innerIdx).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		b.InnerIndex = innerIdx
		return db.Default().FirstOrCreate(b, "symbol=? and hash=?", b.Symbol, b.Hash).Error
	}

	return db.Default().Model(b1).Updates(M{
		"height": b.Height,
		"hash":   b.Hash,
	}).Error
}

// InsertByHeight inserts new blockinfo to block table.
func (b *BlockInfo) InsertByHeight() error {
	return db.Default().FirstOrCreate(b, "symbol=? and height=?", b.Symbol, b.Height).Error
}

// Update updates block hash
func (b *BlockInfo) Update() error {
	return db.Default().Model(b).Where("symbol=? and height=?", b.Symbol, b.Height).Update("hash", b.Hash).Error
}

// Delete deletes a block info.
func (b *BlockInfo) Delete() error {
	return db.Default().Where("symbol=? and hash=?", b.Symbol, b.Hash).Delete(b).Error
}

type Block struct {
	Height uint64 `gorm:"primary_key" json:"height"`
	Hash   string `gorm:"index" json:"hash"`
}

// TableName defines the table name of block.
func (b Block) TableName() string { return "block" }

// GetLastBlockInfo gets last block.
func GetLastBlockInfo(symbol string, useBlockTable bool) *BlockInfo {
	var block BlockInfo
	db.Default().Where("symbol=?", symbol).Order("height DESC").Limit(1).First(&block)
	if len(block.Symbol) > 0 {
		return &block
	}

	if useBlockTable {
		var tmp Block
		err := db.Default().Last(&tmp).Error
		if err == nil {
			block.Hash = tmp.Hash
			block.Height = tmp.Height
			block.Symbol = symbol
		}
	}
	return &block
}

// GetBlockInfoByHeight get block by height
func GetBlockInfoByHeight(height uint64, symbol string, useBlockTable bool) (*BlockInfo, bool) {
	var block BlockInfo
	db.Default().First(&block, "symbol=? and height = ?", symbol, height)
	if len(block.Symbol) > 0 {
		return &block, true
	}

	if useBlockTable {
		var tmp Block
		err := db.Default().First(&tmp, "height = ?", height).Error
		if err == nil {
			block.Hash = tmp.Hash
			block.Height = tmp.Height
			block.Symbol = symbol
			return &block, true
		}
	}

	return &block, false
}
