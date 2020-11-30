package txbuilder

import (
	"strings"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"

	"github.com/shopspring/decimal"
)

type TxIn struct {
	Account   *bmodels.Account
	Cost      decimal.Decimal
	CostUTXOs []*bmodels.UTXO
}

type TxInfo struct {
	Inputs         []*TxIn
	SigPubKeys     []string
	SigDigests     []string
	TxHex          string
	Nonce          *uint64
	Fee            decimal.Decimal
	DiscardAddress bool
}

type Model string

const (
	AccountModel Model = "AccountModel"
	UTXOModel    Model = "UTXOModel"
)

type Builder interface {
	Model() Model
	BuildWithdraw(*models.Tx) (*TxInfo, error)
	BuildGather(*models.Tx) (*TxInfo, error)
}

type SupplementaryFeeBuilder interface {
	FeeSymbol() string
	BuildSupplementaryFee(*models.Tx) (*TxInfo, error)
}

type BuilderCreator func(*config.Config) Builder

var (
	builderCreator = make(map[string]BuilderCreator)
)

func Register(currencyType string, creater BuilderCreator) {
	currencyType = strings.ToUpper(currencyType)
	if _, ok := Find(currencyType); ok {
		log.Errorf("tx-builder.Register, duplicate of %s\n", currencyType)
		return
	}

	builderCreator[currencyType] = creater
}

func Find(currencyType string) (BuilderCreator, bool) {
	currencyType = strings.ToUpper(currencyType)
	c, ok := builderCreator[currencyType]
	return c, ok
}
