package models

import (
	"fmt"

	"upex-wallet/wallet-base/db"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"
)

const (
	AddressTypeSystem = 0
	AddressTypeNormal = 1
)

const (
	AddressStatusRecord = 0

	AddressStatusDiscard = 100
)

// Address .
type Address struct {
	ID      uint   `gorm:"AUTO_INCREMENT" json:"id"`
	Address string `gorm:"index;unique_index:addr_ver" json:"address"`
	Type    uint   `gorm:"type:tinyint;index" json:"type"`
	Version string `gorm:"size:8;unique_index:addr_ver" json:"version"`
	PubKey  string `gorm:"size:512" json:"pubkey"`
	Status  uint   `gorm:"type:tinyint;default:0;index"`
}

// TableName defines the table name of Address.
func (a Address) TableName() string { return "address" }

func (a *Address) Discard() error {
	return db.Default().Model(a).Where("address = ?", a.Address).
		Updates(map[string]interface{}{
			"status": AddressStatusDiscard,
		}).Error
}

// GetAllAddresses gets all addresses.
func GetAllAddresses() []*Address {
	var (
		addrs []*Address
	)
	db.Default().Find(&addrs)
	return addrs
}

// GetAddressInfo get account by address.
func GetAddressInfo(dbInst *gorm.DB, addr string) *Address {
	if dbInst == nil {
		dbInst = db.Default()
	}

	var addressInfo Address
	dbInst.Where("address = ?", addr).First(&addressInfo)
	return &addressInfo
}

// HasAddress returns true is wallet address.
func HasAddress(addr string) (bool, error) {
	var addressInfo Address
	db := db.Default().Where("address = ?", addr).First(&addressInfo)
	if db.Error != nil {
		if db.RecordNotFound() {
			return false, nil
		}

		return false, db.Error
	}

	if addressInfo.Status == AddressStatusDiscard {
		return false, nil
	}

	return true, nil
}

// GetPubKey get pubkey by address.
func GetPubKey(dbInst *gorm.DB, addr string) (string, bool) {
	if len(addr) == 0 {
		return "", false
	}

	info := GetAddressInfo(dbInst, addr)
	if info.Address == addr {
		return info.PubKey, true
	}
	return "", false
}

// InitAddressTable just for update database start
func InitAddressTable() error {
	addressCount := -1
	err := db.Default().Table("address").Count(&addressCount).Error
	if err != nil {
		return fmt.Errorf("db get address count failed, %v", err)
	}

	if addressCount > 0 {
		return nil
	}

	// Need one minute if there are three million addresses
	log.Warnf("address table is empty, start copy addresses from account table...")

	tx := db.Default().Begin()
	tx.Exec("insert into address(address,type,version) select address,account_type,version from account")
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("copy address from account table failed, %v", err)
	}

	db.Default().Table("address").Count(&addressCount)
	log.Warnf("finish copy %d addresses", addressCount)
	return nil
}

// GetSystemAddress returns system address.
func GetSystemAddress() []*Address {
	var addrs []*Address
	db.Default().Where("`type` = ? and status = ?", AddressTypeSystem, AddressStatusRecord).Find(&addrs)
	return addrs
}

// IsSystemAddress returns whether the address is a system address.
func IsSystemAddress(dbInst *gorm.DB, addr string) bool {
	addrInfo := GetAddressInfo(dbInst, addr)
	if addrInfo.Address != addr {
		return false
	}

	return addrInfo.Type == AddressTypeSystem
}

// IsNormalAddress returns whether the address is a normal user address.
func IsNormalAddress(dbInst *gorm.DB, addr string) bool {
	addrInfo := GetAddressInfo(dbInst, addr)
	if addrInfo.Address != addr {
		return false
	}

	return addrInfo.Type == AddressTypeNormal
}

func (addr *Address) InsertWithDB(db *gorm.DB) error {
	return db.FirstOrCreate(addr, "address = ?", addr.Address).Error
}
