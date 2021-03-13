package trx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/trx/gtrx"

	"github.com/shopspring/decimal"
)

func init() {
	txbuilder.Register("trx", New)
}

type TRXBuilder struct {
	cfg       *config.Config
	rpcClient *gtrx.Client
}

func New(cfg *config.Config) txbuilder.Builder {
	err := InitSupportAssets(cfg)
	if err != nil {
		panic(fmt.Sprintf("init support assets failed, %v", err))
	}

	return txbuilder.NewAccountModelBuilder(cfg, &TRXBuilder{
		cfg:       cfg,
		rpcClient: gtrx.NewClient(cfg.RPCUrl),
	})
}

func (b *TRXBuilder) Support(currency string) bool {
	return b.cfg.Currency == currency
}

func (b *TRXBuilder) DefaultFeeMeta() txbuilder.FeeMeta {
	return txbuilder.FeeMeta{
		Fee: decimal.NewFromFloat(gtrx.NormalTransferFee).Div(decimal.New(gtrx.TRX, 0)),
	}
}

func (b *TRXBuilder) EstimateFeeMeta(symbol string, txType int8) *txbuilder.FeeMeta {
	return &txbuilder.FeeMeta{
		Fee: b.DefaultFeeMeta().Fee,
	}
}

func (b *TRXBuilder) DoBuild(info *txbuilder.AccountModelBuildInfo) (*txbuilder.TxInfo, error) {
	if info.Task.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("can't build tx with 0 amount")
	}

	assetInfo, ok := SupportAssetInfo(info.Task.Symbol)
	if !ok {
		return nil, txbuilder.NewErrUnsupportedCurrency(info.Task.Symbol)
	}

	var tx *gtrx.Transaction
	var err error
	if assetInfo.Type == NORMAL {
		amount := uint64(info.Task.Amount.Mul(decimal.New(1, int32(assetInfo.Precision))).IntPart())
		tx, err = b.rpcClient.CreateTransaction(info.FromAccount.Address, info.Task.Address, amount, assetInfo.ID)
		if err != nil {
			return nil, fmt.Errorf("create transaction failed, %v", err)
		}
	} else if assetInfo.Type == TRC20 {
		transferReq, err := gtrx.CreateTrc20TransferReq(info.FromAccount.Address, info.Task.Address, assetInfo.ContractAddress, info.Task.Amount, int32(assetInfo.Precision))
		if err != nil {
			log.Errorf("444----info:%#v,assetInfo:%#v",info,assetInfo)
			return nil, fmt.Errorf("create trc20 transfer req falied,%v", err)
		}

		txInfoJson, err := json.Marshal(transferReq)
		if err != nil {
			return nil, fmt.Errorf("marshal transferInfo failed,%v", err)
		}

		transferResJson, err := b.rpcClient.TriggerSmartContract(string(txInfoJson))
		if err != nil {
			return nil, fmt.Errorf("transfer smart contract failed,%v", err)
		}

		transferResStr := string(transferResJson)
		transferRes := &gtrx.TransferResult{}
		err = json.Unmarshal(transferResJson, transferRes)
		if err != nil {
			return nil, fmt.Errorf("unmarshal transfer result failed, %v, result:%s", err, transferResStr)
		}

		if !transferRes.Result.Result {
			return nil, fmt.Errorf("transfer smart contract failed,result is %s", transferResStr)
		}

		tx = &transferRes.Transaction

	} else {
		return nil, fmt.Errorf("assert type %s is not support", assetInfo.Type)
	}

	_, err = tx.UpdateTimestamp()
	if err != nil {
		return nil, fmt.Errorf("update tx hash failed, %v", err)
	}

	rawData, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("json marshal tx failed, %v", err)
	}

	cost := info.Task.Amount
	if info.Task.Symbol == b.cfg.Currency {
		cost = cost.Add(info.FeeMeta.Fee)
	}

	return &txbuilder.TxInfo{
		Inputs: []*txbuilder.TxIn{
			&txbuilder.TxIn{
				Account: info.FromAccount,
				Cost:    cost,
			},
		},
		SigPubKeys: []string{hex.EncodeToString(info.FromPubKey)},
		SigDigests: []string{tx.TxID},
		TxHex:      string(rawData),
		Fee:        info.FeeMeta.Fee,
	}, nil

}
