package models

import (
	"fmt"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/models"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

// InitDB initializes database.
func InitDB() error {
	err := db.Default().AutoMigrate(
		&models.Account{},
		&models.Tx{},
		&models.Currency{},
		&models.BlockInfo{},
		&models.Address{},
	).Error
	if err != nil {
		log.Errorf("db auto migrate failed, %v", err)
	}

	err = models.InitAddressTable()
	if err != nil {
		return fmt.Errorf("init address table failed, %v", err)
	}

	return nil
}
