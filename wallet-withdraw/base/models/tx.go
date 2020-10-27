package models

import (
	"fmt"
	"sort"
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
	TxTypeDeposit  = 0
	TxTypeWithdraw = 2
	TxTypeGather   = 4
	TxTypeCold     = 8

	// Ext tx types.
	TxTypeClaim = 16

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
	Address           string          `gorm:"size:256;index" json:"address"`
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
	// Channel        string          `gorm:"size:50;index"`

}

func (wtx Tx) TableName() string { return "wallet_tx" }

// FirstOrCreate find first matched record or create a new one.
func (wtx *Tx) FirstOrCreate() error {
	if wtx.TxStatus == TxStatusNotRecord {
		wtx.TxStatus = TxStatusRecord
	}

	return db.Default().FirstOrCreate(wtx, "sequence_id = ? and tx_type = ?", wtx.SequenceID, wtx.TxType).Error
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
	data["trans_id"] = wtx.TransID
	data["address_to"] = wtx.Address
	data["amount"] = wtx.Amount.Sub(wtx.Fees).String()
	data["symbol"] = strings.ToLower(wtx.Symbol)
	data["txid"] = wtx.Hash
	data["confirm"] = wtx.Confirm
	data["real_fee"] = ""
	data["app_id"] = ""
	data["timestamp"] = time.Now().Unix()
	return data
}

// ReceiveFormat returns a data structure for receive task request.
// func (wtx *Tx) ReceiveFormat() map[string]string {
// 	return wtx.notifyFormat("receive", "", "", "")
// }

// PassFormat returns a pass request args for withdraw task.
// func (wtx *Tx) PassFormat(txHash string) map[string]string {
// 	return wtx.notifyFormat("pass", txHash, "", "")
// }

// RejectFormat returns a reject request args for withdraw task.
// func (wtx *Tx) RejectFormat() map[string]string {
// 	return wtx.notifyFormat("reject", "", "1", "wallet")
// }

// func (wtx *Tx) notifyFormat(state, txHash, errCode, errMsg string) map[string]string {
// 	data := make(map[string]string)
// 	data["id"] = wtx.TransID
// 	data["address"] = wtx.EncAddress
// 	data["state"] = state
// 	data["tx_hash"] = txHash
// 	data["error_code"] = errCode
// 	data["error_msg"] = errMsg
// 	return data
// }

// ClaimFormat return a data request for claim task
func (wtx *Tx) ClaimFormat(txHash string) map[string]string {
	data := make(map[string]string)
	data["symbol"] = wtx.Symbol
	data["address"] = wtx.Address
	data["amount"] = wtx.Amount.String()
	// data["extra"] = wtx.Extra
	data["source"] = wtx.ClaimHash
	data["tx_hash"] = txHash
	// data["currency"] = strconv.Itoa(wtx.Code)
	data["sequence_id"] = wtx.SequenceID
	return data
}

func (wtx *Tx) CloneCore() *Tx {
	txCopy := Tx{}
	txCopy.SequenceID = wtx.SequenceID
	// txCopy.Channel = wtx.Channel
	txCopy.TransID = wtx.TransID
	txCopy.Address = wtx.Address
	// txCopy.EncAddress = wtx.EncAddress
	txCopy.BlockchainName = wtx.BlockchainName
	txCopy.Symbol = wtx.Symbol
	// txCopy.Code = wtx.Code
	txCopy.TxType = wtx.TxType
	txCopy.Amount = wtx.Amount
	txCopy.Fees = wtx.Fees
	txCopy.Nonce = wtx.Nonce
	// txCopy.Extra = wtx.Extra
	return &txCopy
}

// UpdateLocalTransIDSequenceID updates transID and sequenceID for local built task.
func (wtx *Tx) UpdateLocalTransIDSequenceID() {
	const MaxShortAddrLen = 12
	shortAddr := wtx.Address
	if len(shortAddr) > MaxShortAddrLen {
		shortAddr = shortAddr[:MaxShortAddrLen]
	}

	wtx.TransID = fmt.Sprintf("%s%d%s%d", wtx.Symbol, wtx.TxType, shortAddr, time.Now().Unix())
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
func GetUnfinishedWithdraws() []*Tx {
	var txs []*Tx
	db.Default().Where("tx_type = ? and tx_status = ?", TxTypeWithdraw, TxStatusRecord).Find(&txs)
	return txs
}

func GetLastSupplementaryFeeTxByAddress(code int, toAddress string) (*Tx, error) {
	var tx Tx
	err := db.Default().
		Where("tx_type = ? and code = ? and address = ?", TxTypeSupplementaryFee, code, toAddress).
		Order("created_at desc").
		First(&tx).Error
	return &tx, err
}

func GetUnReadjustedFeeTxs(codes []int) []*Tx {
	var txs []*Tx
	db.Default().
		Where("readjusted_fee = false and tx_status = ? and code in (?)", TxStatusSuccess, codes).
		Limit(10).
		Find(&txs)
	return txs
}

// SelectUTXOWithTransFee sort utxo accoring utxo.Amount
func SelectUTXOWithTransFee(symbolID uint, address string, limitLen int, bigOrder bool) ([]*models.UTXO, bool) {
	// 获取 address下所有的UTXO详情--utxo 切片 (数据库集)
	utxos := models.GetUTXOsByAddress(symbolID, address)

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
