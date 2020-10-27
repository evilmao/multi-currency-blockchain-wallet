package models

import (
	"strings"

	"upex-wallet/wallet-base/db"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// ColdInfo represents cooldown config info.
type ColdInfo struct {
	ID            uint            `gorm:"primary_key"`
	Currency      string          `gorm:"default:"";`
	Address       string          `gorm:"index"`
	MaxBalance    decimal.Decimal `gorm:"type:decimal(32,20);default:0"`
	RemainBalance decimal.Decimal `gorm:"type:decimal(32,20);default:0"`
}

func (*ColdInfo) TableName() string { return "cold_info" }

func GetColdInfo(dbInst *gorm.DB, currency string) (*ColdInfo, error) {
	if dbInst == nil {
		dbInst = db.Default()
	}

	var info ColdInfo
	err := dbInst.Where("currency = ?", strings.ToLower(currency)).First(&info).Error
	return &info, err
}
