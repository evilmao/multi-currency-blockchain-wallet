package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/util"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"upex-wallet/wallet-base/models"
)

// Tx type.
const (
	// Base tx types.
	TxTypeDeposit          = 0
	TxTypeWithdraw         = 2
	TxTypeGather           = 4
	TxTypeCold             = 8
	TxTypeClaim            = 16 // Ext tx types.
	TxTypeSupplementaryFee = 32
)

func TxTypeName(txType int8) string {
	switch txType {
	case TxTypeDeposit:
		return "deposit"
	case TxTypeWithdraw:
		return "withdraw"
	case TxTypeGather:
		return "gather"
	case TxTypeCold:
		return "cold"
	case TxTypeClaim:
		return "claim"
	case TxTypeSupplementaryFee:
		return "supplementary-fee"
	default:
		return "invalid"
	}
}

// Tx status.
const (
	TxStatusNotRecord = 0
	TxStatusRecord    = 1

	TxStatusBroadcast        = 10
	TxStatusBroadcastSuccess = 11
	TxStatusBroadcastFailed  = 12

	TxStatusSuccess = 50

	TxStatusDiscard = 100
	TxStatusReject  = 127

	TxStatusMax = TxStatusReject
)

// Tx represents the txs that wallet create
// also it represents a wallet withdraw task.
type Tx struct {
	SequenceID        string          `gorm:"primary_key;size:32" json:"sequence_id"`
	Hash              string          `gorm:"column:txid;size:90" json:"hash"`
	Address           string          `gorm:"size:256;index" json:"address_to"`
	Confirm           uint16          `gorm:"type:int" json:"confirm_times"`
	TxType            int8            `gorm:"column:tx_type;type:tinyint;index" json:"tx_type"`
	TransID           string          `gorm:"size:50;default:0;index" json:"trans_id"`
	Symbol            string          `gorm:"size:10" json:"symbol"`
	BlockchainName    string          `gorm:"size:50"`
	Hex               string          `gorm:"type:mediumtext" json:"hex"`
	Nonce             uint64          `gorm:"type:bigint;default:0;index"`
	TxStatus          int8            `gorm:"type:tinyint;index" json:"tx_status"`
	BroadcastTryCount uint16          `gorm:"type:int" json:"broadcast_try_count"`
	ReadjustedFee     bool            `gorm:"column:readjusted_fee;default:0;index" json:"readjusted_fee"`
	ClaimHash         string          `gorm:"size:100;default:''" json:"claim_hash"`
	ClaimStatus       uint16          `gorm:"type:tinyint" json:"claim_status"`
	Fees              decimal.Decimal `gorm:"type:decimal(32,20);default:0" json:"fees"`
	Amount            decimal.Decimal `gorm:"type:decimal(32,20);default:0" json:"amount"`
	CreatedAt         time.Time
	// EncAddress     string          `gorm:"size:500" json:"en/c_address"`
	// Extra          string          `gorm:"size:100" json:"extra"`
	// Code           int             `gorm:"type:int;default:0" json:"code"`
}

func (wtx Tx) TableName() string { return "wallet_tx" }

// FirstOrCreate find first matched record or create a new one.
func (wtx *Tx) FirstOrCreate() error {
	if wtx.TxStatus == TxStatusNotRecord {
		wtx.TxStatus = TxStatusRecord
	}
	// fix same gather task repeat record
	return db.Default().FirstOrCreate(wtx, "txid = ? ", wtx.Hash).Error
}

// Update updates the tx status.
func (wtx *Tx) Update(values M, dbInst *gorm.DB) error {
	if len(wtx.SequenceID) == 0 {
		return fmt.Errorf("can't update tx with no sequenceID")
	}

	if dbInst == nil {
		dbInst = db.Default()
	}

	return dbInst.Model(wtx).Updates(values).Error
}

// WithdrawNotifyFormat returns a data structure for withdraw notify.
func (wtx *Tx) WithdrawNotifyFormat() map[string]interface{} {
	data := make(map[string]interface{})
	id, _ := strconv.ParseInt(wtx.TransID, 10, 64)
	data["id"] = id
	data["address"] = wtx.Address
	data["numbers"] = wtx.Amount.String()
	data["coinName"] = strings.ToLower(wtx.Symbol)
	data["txId"] = wtx.Hash
	data["chainType"]= wtx.BlockchainName
	data["code"] = 1
	data["remark"] = "Success"
	data["timestamp"] = time.Now().Unix()
	data["app_id"] = ""
	data["real_fee"] = wtx.Fees
	data["confirm"] = wtx.Confirm
	return data
}

// ClaimFormat return a data request for claim task
func (wtx *Tx) ClaimFormat(txHash string) map[string]string {
	data := make(map[string]string)
	data["symbol"] = wtx.Symbol
	data["address"] = wtx.Address
	data["amount"] = wtx.Amount.String()
	data["source"] = wtx.ClaimHash
	data["tx_hash"] = txHash
	data["sequence_id"] = wtx.SequenceID
	return data
}

func (wtx *Tx) CloneCore() *Tx {
	txCopy := Tx{}
	txCopy.SequenceID = wtx.SequenceID
	txCopy.TransID = wtx.TransID
	txCopy.Address = wtx.Address
	txCopy.BlockchainName = wtx.BlockchainName
	txCopy.Symbol = wtx.Symbol
	txCopy.TxType = wtx.TxType
	txCopy.Amount = wtx.Amount
	txCopy.Fees = wtx.Fees
	txCopy.Nonce = wtx.Nonce
	return &txCopy
}

// UpdateLocalTransIDSequenceID updates transID and sequenceID for local built task.
func (wtx *Tx) UpdateLocalTransIDSequenceID() {
	const MaxShortAddrLen = 12
	shortAddr := wtx.Address
	if len(shortAddr) > MaxShortAddrLen {
		shortAddr = shortAddr[:MaxShortAddrLen]
	}

	wtx.TransID = fmt.Sprintf("%s%d%s%d", wtx.Symbol, wtx.TxType, shortAddr, time.Now().UnixNano()/1e6)
	wtx.SequenceID = util.HashString32([]byte(wtx.TransID))
}

func (wtx *Tx) String() string {
	return fmt.Sprintf("blockchain: %s, currency: %s, txType: %s, sequenceID: %s, transID: %s, txid: %s, to: %s, amount: %s, nonce: %d",
		wtx.BlockchainName, wtx.Symbol, TxTypeName(wtx.TxType), wtx.SequenceID, wtx.TransID, wtx.Hash, wtx.Address, wtx.Amount, wtx.Nonce)
}

// GetTxsByStatus returns txs in the status.
func GetTxsByStatus(status uint) []*Tx {
	var txs []*Tx
	db.Default().Where("tx_status = ? ", status).Limit(10).Find(&txs)
	return txs
}

// GetTxHash, get TxHash by sequenceID
func GetTxHashBySequenceID(sequenceID string) string {

	var tx Tx
	_ = db.Default().Where("sequence_id = ?", sequenceID).First(&tx).Error

	return tx.Hash
}

// GetTxBySequenceID gets tx by sequence id.
func GetTxBySequenceID(dbInst *gorm.DB, sequenceID string) (*Tx, error) {
	if dbInst == nil {
		dbInst = db.Default()
	}

	var tx Tx
	err := dbInst.Where("sequence_id = ?", sequenceID).First(&tx).Error
	return &tx, err
}

// GetWithdrawTxByTransID gets withdraw tx by transID.
func GetWithdrawTxByTransID(transID string) (*Tx, error) {
	var tx Tx
	err := db.Default().Where("trans_id = ? and tx_type = ? and tx_status <> ?",
		transID, TxTypeWithdraw, TxStatusDiscard).First(&tx).Error
	return &tx, err
}

// GetUnfinishedWithdraws returns unfinished withdraws.
func GetUnfinishedWithdraws(symbols []string) []*Tx {
	var txs []*Tx
	db.Default().Where("tx_type = ? and tx_status = ? and symbol in (?)", TxTypeWithdraw, TxStatusRecord, symbols).Find(&txs)
	return txs
}

func GetLastSupplementaryFeeTxByAddress(symbol, toAddress string) (*Tx, error) {
	var tx Tx
	err := db.Default().
		Where("tx_type = ? and symbol = ? and address = ?", TxTypeSupplementaryFee, symbol, toAddress).
		Order("created_at desc").
		First(&tx).Error
	return &tx, err
}

func GetUnReadjustedFeeTxs() []*Tx {
	var txs []*Tx
	db.Default().
		Where("readjusted_fee = false and tx_status = ? ", TxStatusSuccess).
		Limit(10).
		Find(&txs)
	return txs
}

// SelectUTXOWithTransFee sort utxo according utxo.Amount
func SelectUTXOWithTransFee(address, symbol string, limitLen int, bigOrder bool) ([]*models.UTXO, bool) {
	utxos := models.GetUTXOsByAddress(address, symbol)

	if len(utxos) == 0 {
		return nil, false
	}

	if bigOrder {
		// u1 >= u2 >= u3
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Amount.GreaterThan(utxos[j].Amount)
		})
	} else {
		// u1 <= u2 <= u3
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Amount.LessThan(utxos[j].Amount)
		})
	}

	if len(utxos) >= limitLen {
		utxos = utxos[:limitLen]
	}
	return utxos, true
}

// TODO: 平台支持后,需修改
// TRC20 USDT should cover
func TaskSymbolCover(chainName string, taskSymbol string) string {
	if chainName == "trx" && taskSymbol == "usdt" {
		taskSymbol = strings.Join([]string{"trc", taskSymbol}, "_")
	}
	return taskSymbol
}
