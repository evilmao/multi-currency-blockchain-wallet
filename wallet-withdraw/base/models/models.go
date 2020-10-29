package models

import (
	"upex-wallet/wallet-base/models"

	"github.com/jinzhu/gorm"
)

type M map[string]interface{}

func Init(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tx{},
		&TxIn{},
		&AddrInfo{},
		&BroadcastTask{},
		&SuggestFee{},
		&models.Address{},
		&models.Account{},
		&models.UTXO{},
		// &ColdInfo{},
	).Error
}
