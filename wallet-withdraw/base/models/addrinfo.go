package models

import (
	"upex-wallet/wallet-base/db"

	"github.com/jinzhu/gorm"
)

// AddrInfo represents address ext info.
type AddrInfo struct {
	gorm.Model
	Address string `gorm:"index"`
	Code    int    `gorm:"type:int;default:0;index"`

	// Account's nonce in blockchain.
	BlockchainNonce uint64 `gorm:"column:blockchain_nonce;type:bigint;default:0"`
}

func (*AddrInfo) TableName() string { return "addr_info" }

func (info *AddrInfo) FirstOrCreate() error {
	return db.Default().FirstOrCreate(info, "address = ? and code = ?", info.Address, info.Code).Error
}

func (info *AddrInfo) Update(values M) error {
	return db.Default().Model(info).Updates(values).Error
}

func GetAddrInfo(address string, code int) (*AddrInfo, error) {
	var info AddrInfo
	err := db.Default().Where("address = ? and code = ?", address, code).First(&info).Error
	return &info, err
}

func NextBlockchainNonce(address string, code int) (uint64, error) {
	info, err := GetAddrInfo(address, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 1, nil
		}

		return 0, err
	}

	return info.BlockchainNonce + 1, nil
}

func SetBlockchainNonceIfGreater(address string, code int, nonce uint64) error {
	info, err := GetAddrInfo(address, code)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		info = &AddrInfo{
			Address:         address,
			Code:            code,
			BlockchainNonce: nonce,
		}
		return info.FirstOrCreate()
	}

	if info.BlockchainNonce >= nonce {
		return nil
	}

	return info.Update(M{
		"blockchain_nonce": nonce,
	})
}
