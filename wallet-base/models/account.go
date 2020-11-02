package models

import (
	"upex-wallet/wallet-base/db"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// Account represents a wrapper of address info.
type Account struct {
	ID      uint             `gorm:"AUTO_INCREMENT" json:"id"`
	Address string           `gorm:"index:idx_address_symbolid" json:"address"`
	Nonce   uint64           `gorm:"type:bigint;default:0" json:"nonce"`
	Balance *decimal.Decimal `gorm:"type:decimal(32, 20);default:0" json:"balance"`
	Symbol  string           `gorm:"column:symbol;size:32" json:"symbol"`
	Type    uint             `gorm:"column:account_type;type:tinyint" json:"account_type"`
	Version string           `gorm:"size:8" json:"version"`
}

// TableName defines the table name of account.
func (act Account) TableName() string { return "account" }

// ForUpdate updates account info with lock.
func (act *Account) ForUpdate(data M) error {
	if op, ok := data["op"].(string); ok {
		balance := decimal.Zero
		if b, ok := data["balance"].(decimal.Decimal); ok {
			balance = b
		}

		switch op {
		case "add":
			data["balance"] = gorm.Expr("`balance` + " + balance.String())
		case "sub":
			data["balance"] = gorm.Expr("`balance` - " + balance.String())
		}
	}

	delete(data, "op")
	err := db.Default().Model(act).Where("address = ?", act.Address).Updates(data).Error
	if err != nil {
		return err
	}
	return nil
}

// Insert inserts new address to table.
func (act *Account) Insert() error {
	return db.Default().FirstOrCreate(act, "symbol = ? and address = ?", act.Symbol, act.Address).Error
}

// DeprecatedGetAccountByAddress get account by address only.
func DeprecatedGetAccountByAddress(addr string) *Account {
	var (
		account Account
	)

	db.Default().Where("address = ?", addr).First(&account)
	return &account
}

// GetAccountByAddress get account by address and symbolID.
func GetAccountByAddress(addr string, symbol string) *Account {
	var (
		account Account
	)

	db.Default().Where("symbol = ? and address = ?", symbol, addr).First(&account)
	return &account
}

// IsContractAddress returns true if matchs contract configuration.
func IsContractAddress(addr string) (*Currency, error) {
	var (
		token Currency
		err   error
	)

	db := db.Default().Where("address = ?", addr).First(&token)
	if db.Error != gorm.ErrRecordNotFound {
		err = db.Error
	}
	return &token, err
}

// GetBalance returns the balance of the wallet.
func GetBalance(symbol string) *decimal.Decimal {
	var data struct {
		Balance decimal.Decimal
	}
	db.Default().Model(&Account{}).Select("sum(balance) as balance").Where("symbol = ?", symbol).Scan(&data)
	return &data.Balance
}

// GetSystemBalance returns the balance of system accounts.
func GetSystemBalance() *decimal.Decimal {
	var data struct {
		Balance decimal.Decimal
	}
	db.Default().Model(&Account{}).Select("sum(balance) as balance").Where("`account_type` = ?", AddressTypeSystem).Scan(&data)
	return &data.Balance
}

// GetMatchedAccount gets matched account for withdraw.
func GetMatchedAccount(amount string, addressType uint) *Account {
	account := Account{Balance: &decimal.Zero}
	db.Default().Where("balance > ? and `account_type` = ?", amount, addressType).First(&account)
	return &account
}

// GetAllMatchedAccounts gets matched accounts for withdraw.
func GetAllMatchedAccounts(amount string, addressType uint) []*Account {
	var accounts []*Account
	db.Default().Where("balance > ?  and `account_type` = ?", amount, addressType).Find(&accounts)
	return accounts
}

// GetAccounts return accounts info.
func GetAccounts(symbolID uint, index, pageSize int64) (*[]Account, int64) {
	var (
		count    int64
		accounts []Account
	)

	db.Default().Table("tx").Select("address, sum(amount) as balance").Where("symbol_id = ?", symbolID).
		Group("address").Count(&count).Offset(pageSize * (index - 1)).Limit(pageSize).Find(&accounts)

	return &accounts, count
}

// GetBalanceByAddress return account balance.
func GetBalanceByAddress(address string, symbolID int) decimal.Decimal {
	acc := GetAccountByAddress(address, symbolID)
	if acc.Balance == nil {
		return decimal.Zero
	}
	return *acc.Balance
}

func GetAccountByAddressWithDB(db *gorm.DB, addr string, symboolID int) *Account {
	var account Account
	db.Where("symbol_id = ? and address = ?", symboolID, addr).First(&account)
	return &account
}

func (act *Account) ForUpdateWithDB(db *gorm.DB, data map[string]interface{}) error {
	tx := db.Begin()
	if act.Balance == nil {
		d := decimal.New(0, 0)
		act.Balance = &d
	}

	switch data["op"].(string) {
	case "add":
		data["balance"] = act.Balance.Add(data["balance"].(decimal.Decimal)).String()
	case "sub":
		data["balance"] = act.Balance.Sub(data["balance"].(decimal.Decimal)).String()
	}
	err := tx.Model(act).Where("symbol=? and address = ?", act.Symbol, act.Address).Updates(data).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit().Error
	return err
}

// GetMaxBalanceAccount gets max balance account
func GetMaxBalanceAccount(addressType uint) *Account {
	account := Account{Balance: &decimal.Zero}
	db.Default().Where("`account_type` = ?", addressType).Order("balance DESC").Limit(1).Find(&account)
	return &account
}
