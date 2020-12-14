package txbuilder

import (
	"strings"

	"github.com/shopspring/decimal"
)

const (
	defaultMaxTxInLen = 30
	// TODO: 方便测试
	// defaultMaxTxInLen = 6
)

type MetaData struct {
	Precision  int
	Fee        decimal.Decimal
	TxVersion  uint32
	MaxTxInLen int // only for utxo model.
}

func NewMetaData(precision int, fee decimal.Decimal, txVersion uint32, maxTxInLen int) *MetaData {
	return &MetaData{
		Precision:  precision,
		Fee:        fee,
		TxVersion:  txVersion,
		MaxTxInLen: maxTxInLen,
	}
}

func (m *MetaData) Clone() *MetaData {
	return NewMetaData(m.Precision, m.Fee, m.TxVersion, m.MaxTxInLen)
}

var (
	metaDatas = map[string]*MetaData{}
)

// AddMeta adds meta data.
// Params maxTxInLen are only for utxo model.
func AddMeta(currency string, precision int, fee decimal.Decimal, txVersion uint32, maxTxInLen int) {
	currency = strings.ToUpper(currency)
	if maxTxInLen <= 0 {
		maxTxInLen = defaultMaxTxInLen
	}
	metaDatas[currency] = NewMetaData(precision, fee, txVersion, maxTxInLen)
}

func FindMetaData(currency string) (*MetaData, bool) {
	currency = strings.ToUpper(currency)
	d, ok := metaDatas[currency]
	return d, ok
}

func (m *MetaData) UpdateFee(fee decimal.Decimal) {
	if fee.GreaterThan(m.Fee) {
		m.Fee = fee
	}
}
