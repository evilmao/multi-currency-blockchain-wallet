package btc

import (
	"encoding/hex"
	"fmt"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc/rpc"

	"github.com/shopspring/decimal"
)

func init() {
	txbuilder.Register("btc", NewBTC)
	txbuilder.Register("etp", NewETP)
	txbuilder.Register("qtum", NewBTC)
	txbuilder.Register("fab", NewBTC)
	txbuilder.Register("mona", NewBTC)
	txbuilder.Register("ltc", NewBTC)
}

func NewBTC(cfg *config.Config) txbuilder.Builder {
	return newBTCLike(cfg, rpc.NewBTCRPC(cfg.RPCUrl))
}

func NewETP(cfg *config.Config) txbuilder.Builder {
	return newBTCLike(cfg, rpc.NewETPRPC(cfg.RPCUrl))
}

func newBTCLike(cfg *config.Config, client gbtc.RPC) txbuilder.Builder {
	return txbuilder.NewUTXOModelBuilder(cfg, &BTCBuilder{
		cfg:    cfg,
		client: client,
	})
}

type BTCBuilder struct {
	cfg    *config.Config
	client gbtc.RPC
}

func (b *BTCBuilder) Support(currency string) bool {
	return currency == b.cfg.Currency
}

func (b *BTCBuilder) DoBuild(metaData *txbuilder.MetaData, task *models.Tx, extInfo *txbuilder.BuildExtInfo) (*txbuilder.TxInfo, error) {
	if !b.Support(task.Symbol) {
		return nil, txbuilder.NewErrUnsupportedCurrency(task.Symbol)
	}

	if task.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("can't build tx with 0 amount")
	}

	if _, err := gbtc.ParseAddress(task.Address, gbtc.AddressParamBTC); err != nil {
		return nil, fmt.Errorf("invalid target address, %v", task.Address)
	}

	var (
		preOutPoints  []*gbtc.OutputPoint
		sigPubKeys    []string
		changeAddress string
	)
	for _, in := range extInfo.Inputs {
		pubKey, ok := bmodels.GetPubKey(nil, in.Account.Address)
		if !ok {
			return nil, fmt.Errorf("db get pubkey of %s failed", in.Account.Address)
		}

		for _, u := range in.CostUTXOs {
			if len(u.ScriptData) == 0 {
				return nil, fmt.Errorf("scriptPubKey of utxo (hash: %s, index: %d) is empty",
					u.TxHash, u.OutIndex)
			}

			scriptPubKey, err := hex.DecodeString(u.ScriptData)
			if err != nil {
				return nil, fmt.Errorf("hex decode scriptPubKey of utxo (hash: %s, index: %d) failed, %v",
					u.TxHash, u.OutIndex, err)
			}

			preOutPoints = append(preOutPoints, &gbtc.OutputPoint{
				Hash:       u.TxHash,
				Index:      uint32(u.OutIndex),
				Address:    u.Address,
				Amount:     uint64(u.Amount.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()),
				ScriptData: scriptPubKey,
			})

			sigPubKeys = append(sigPubKeys, pubKey)
		}

		if len(changeAddress) == 0 {
			changeAddress = in.Account.Address
		}
	}

	var outputs []*gbtc.Output
	txbuilder.MakeOutputs(
		extInfo.TotalInput, task.Amount, extInfo.MaxOutAmount,
		task.Address, changeAddress,
		metaData,
		func(address string, value uint64) {
			outputs = append(outputs, gbtc.NewOutput(address, value))
		},
	)

	tx, err := b.client.CreateRawTransaction(metaData.TxVersion, preOutPoints, outputs)
	if err != nil {
		return nil, err
	}

	sigDigests := make([]string, 0, len(tx.Inputs))
	for i, in := range tx.Inputs {
		hash := gbtc.SignatureHash(tx, i, in.PreOutput.ScriptData, in.PreOutput.Amount, gbtc.SigVersionBase)
		sigDigests = append(sigDigests, hex.EncodeToString(hash))
	}

	return &txbuilder.TxInfo{
		Inputs:     extInfo.Inputs,
		SigPubKeys: sigPubKeys,
		SigDigests: sigDigests,
		TxHex:      hex.EncodeToString(tx.Bytes()),
	}, nil
}

func (b *BTCBuilder) SupportFeeRate() (feeRate float64, ok bool) {
	feeRate, _ = b.client.EstimateSmartFee(6)
	return feeRate, feeRate != 0
}

func (b *BTCBuilder) CalculateFee(nIn, nOut int, feeRate float64, precision int32) (fee decimal.Decimal) {
	transSize := gbtc.CalculateTxSize(nIn, nOut)
	if nIn == 0 || nOut == 0 {
		return
	}

	return gbtc.CalculateTxFee(transSize, feeRate).Round(precision)
}
