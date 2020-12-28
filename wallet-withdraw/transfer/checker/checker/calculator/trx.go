package calculator

import (
	"fmt"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/checker/checker"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/trx/gtrx"

	"github.com/shopspring/decimal"
)

func init() {
	checker.Add("trx", checker.NewFeeReadJuster(TRC20Calc))
}

func TRC20Calc(cfg *config.Config, txHash string) (*checker.ReadjustFeeInfo, error) {

	client := gtrx.NewClient(cfg.RPCUrl)
	tx, err := client.GetTransactionInfoByID(txHash)
	if err != nil {
		return nil, fmt.Errorf("get tx %s failed, %v", txHash, err)
	}

	blockNumber, _ := gtrx.JSONHexToDecimal(tx, "blockNumber")
	if blockNumber.Equal(decimal.Zero) { // pending
		return nil, nil
	}

	// 获取真实交易手续费
	fee, err := gtrx.JSONHexToDecimal(tx, "fee")
	if err != nil {
		return nil, fmt.Errorf("parse tx %s fee used failed, %v", txHash, err)
	}

	usedFee, err := feeMeta(cfg)
	if err != nil {
		return nil, fmt.Errorf("get currency %s fee used failed, %v", usedFee, err)
	}

	info := &checker.ReadjustFeeInfo{
		RemainFee: fee.Mul(decimal.New(1, -gtrx.Precision)).Add(usedFee),
		FeeSymbol: cfg.Currency,
	}
	return info, nil
}

func feeMeta(cfg *config.Config) (usedFee decimal.Decimal, err error) {
	builderFunc, ok := txbuilder.Find(cfg.Currency)
	if !ok {
		return decimal.Zero, fmt.Errorf("can not find builder for currency %s", cfg.Currency)
	}

	if builder, ok := builderFunc(cfg).(*txbuilder.AccountModelBuilder); ok {
		feeMeta := builder.FeeMeta()
		return feeMeta.Fee, nil
	}

	return decimal.Zero, fmt.Errorf("get %s builder fail ", cfg.Currency)
}
