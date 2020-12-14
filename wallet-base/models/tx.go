package models

import (
	"fmt"
	"strings"
	"time"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/monitor"

	log "upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// Transaction type define
const (
	TxDeposit = iota
)

// Retry config
const (
	MaxRetryTimes   = 10000
	RetryExpiration = time.Hour * 24 * 2

	// MaxUpdateAccountRetryTimes is the max retry times to update account table
	MaxUpdateAccountRetryTimes = 3
)

// Tx represents a simple transaction info.
type Tx struct {
	SequenceID       string          `gorm:"primary_key;size:32" json:"sequence_id"`
	Hash             string          `gorm:"column:txid;size:90;index" json:"hash"`
	Address          string          `gorm:"size:256;index" json:"address"`
	Extra            string          `gorm:"size:100" json:"extra"`
	Confirm          uint16          `gorm:"type:int" json:"confirm_times"`
	Symbol           string          `gorm:"size:10" json:"symbol"`
	Type             uint16          `gorm:"column:tx_type;type:tinyint" json:"tx_type"`
	NotifyRetryCount uint16          `gorm:"type:int" json:"notify_retry_count"`
	NotifyStatus     uint16          `gorm:"type:tinyint" json:"notify_status"`
	Version          uint16          `gorm:"type:bigint;default:0" json:"version"`
	InnerIndex       uint16          `gorm:"type:int" json:"inner_index"`
	Amount           decimal.Decimal `gorm:"type:decimal(32,20);default:0" json:"amount"`
	CreatedAt        time.Time
	BlockchainTime   *time.Time
	// EncAddress       string          `gorm:"size:500" json:"enc_address"`
	// PrivData         string          `gorm:";size:512" json:"priv_data"`

}

// TableName defines the table name of deposit_tx.
func (tx Tx) TableName() string { return "deposit_tx" }

// Insert inserts tx to transaction table.
func (tx *Tx) Insert() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("insert transaction %s failed, %v",
				tx.Hash, err)
			log.Error(err)
		}
	}()

	if TxExistedBySeqID(tx.SequenceID) {
		return nil
	}

	err = db.Default().FirstOrCreate(tx, "sequence_id = ?", tx.SequenceID).Error
	if err != nil {
		return err
	}

	if len(tx.Address) == 0 {
		return nil
	}

	var acc Account
	err = db.Default().Where(Account{Symbol: tx.Symbol, Address: tx.Address}).First(&acc).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		addressInfo := GetAddressInfo(nil, tx.Address)
		if addressInfo.Address != tx.Address {
			return fmt.Errorf("can't find address for tx %s", tx.Hash)
		}

		zero := decimal.New(0, 0)
		acc = Account{
			Address: tx.Address,
			Balance: &zero,
			Nonce:   0,
			Symbol:  tx.Symbol,
			Version: addressInfo.Version,
			Type:    addressInfo.Type,
		}
		err = acc.Insert()
		if err != nil {
			return fmt.Errorf("db insert account failed, %v", err)
		}
	}

	err = acc.ForUpdate(map[string]interface{}{
		"balance": tx.Amount,
		"op":      "add",
	})
	if err != nil {
		return fmt.Errorf("update account balance failed, %v", err)
	}

	amount, _ := tx.Amount.Float64()
	if amount > 0 {
		tags := monitor.MetricsTags{"currency": tx.Symbol}
		monitor.MetricsHistogram("deposit_amount", amount, tags)
		monitor.MetricsCount("deposit_txnum", 1, tags)
	}

	return nil
}

// Update updates the tx status.
func (tx *Tx) Update(data map[string]interface{}) error {
	data["version"] = gorm.Expr("`version` + 1")
	return db.Default().Model(tx).Where("version = ?", tx.Version).Updates(data).Error
}

// IsFinished returns tx notify task status.
func (tx *Tx) IsFinished() bool {
	var tmpTx Tx
	err := db.Default().First(&tmpTx, "sequence_id = ?", tx.SequenceID).Error
	if err != nil {
		return false
	}

	return (tmpTx.NotifyStatus == 1) || tmpTx.NotifyRetryCount > MaxRetryTimes ||
		tmpTx.CreatedAt.Before(time.Now().Add(-RetryExpiration))
}

func (tx *Tx) DepositNotifyFormat() map[string]interface{} {
	data := make(map[string]interface{})
	data["address_to"] = tx.Address
	data["amount"] = tx.Amount
	data["app_id"] = ""
	data["confirm"] = tx.Confirm
	data["symbol"] = strings.ToLower(tx.Symbol)
	data["txid"] = tx.Hash
	data["timestamp"] = time.Now().Unix()
	data["is_expired"] = 0
	data["is_mining"] = 0
	return data
}

// GetUnfinishedTxs returns failed deposit tx.
func GetUnfinishedTxs(symbol string) []Tx {
	if symbol == "" {
		return nil
	}

	var (
		txs    []Tx
		tokens []Tx

		conditions = strings.Join([]string{
			"(notify_status = 0)",
			fmt.Sprintf("(notify_retry_count < %d)", MaxRetryTimes),
			fmt.Sprintf("(created_at > '%s')", time.Now().Add(-RetryExpiration).Format("2006-01-02 15:04:05")),
		}, " and ")
	)

	db.Default().
		Find(&txs, fmt.Sprintf("symbol = ? and %s", conditions), symbol).
		Limit(500)

	db.Default().
		Find(&tokens, fmt.Sprintf("(symbol in (select symbol from currency where blockchain = '%s')) and %s", symbol, conditions)).
		Limit(500)

	txs = append(txs, tokens...)
	return txs
}

// WithdrawIDExists checks whether withdraw tasks is processed.
func WithdrawIDExists(id float64) bool {
	var (
		tx Tx
	)
	return db.Default().Where("trans_id = ?", id).First(&tx).RecordNotFound()
}

func GetTxByHash(txid string) bool {
	var (
		tx Tx
	)
	db.Default().Where("txid = ?", txid).First(&tx)
	if tx.Hash == "" {
		return false
	}
	return tx.Hash == txid
}

func LoadTxByHash(txHash string) (*Tx, error) {
	var tx Tx
	err := db.Default().Where("txid = ?", txHash).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func TxExistedBySeqID(sequenceID string) bool {
	var tx Tx
	db.Default().Where("sequence_id = ?", sequenceID).First(&tx)
	if tx.SequenceID == "" {
		return false
	}
	return tx.SequenceID == sequenceID
}

func GetTxByAddress(addr string) (*Tx, error) {
	var tx Tx
	err := db.Default().Where("address = ?", addr).First(&tx).Error
	return &tx, err
}

func DeleteTxs() {
	db.Default().Where("address= '' and tx_type = 100").Delete(&Tx{})
}

func GetTxsByAddressWithDB(db *gorm.DB, addr string, offset, limit int) []*Tx {
	var txs []*Tx
	db.Limit(limit).Offset(offset).Order("created_at desc").Find(&txs, "address = ? and amount > 0", addr)
	return txs
}
