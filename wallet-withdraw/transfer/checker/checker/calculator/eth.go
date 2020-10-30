package calculator

import (
	"fmt"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/checker/checker"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/eth/geth"

	"github.com/shopspring/decimal"
)

func init() {
	checker.Add("eth", checker.NewFeeReadJuster(EthCalc))
	checker.Add("etc", checker.NewFeeReadJuster(EthCalc))
	checker.Add("smt", checker.NewFeeReadJuster(EthCalc))
	checker.Add("ionc", checker.NewFeeReadJuster(EthCalc))
}

func EthCalc(cfg *config.Config, txHash string) (*checker.ReadjustFeeInfo, error) {
	client := geth.NewClient(cfg.RPCUrl)
	tx, err := client.GetTransactionByHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("get tx %s failed, %v", txHash, err)
	}

	blockNumber, _ := geth.JSONHexToDecimal(tx, "blockNumber")
	if blockNumber.Equal(decimal.Zero) { // pending
		return nil, nil
	}

	receipt, err := client.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("get tx %s receipt failed, %v", txHash, err)
	}

	gasLimit, err := geth.JSONHexToDecimal(tx, "gas")
	if err != nil {
		return nil, fmt.Errorf("parse tx %s gas limit failed, %v", txHash, err)
	}

	gasPrice, err := geth.JSONHexToDecimal(tx, "gasPrice")
	if err != nil {
		return nil, fmt.Errorf("parse tx %s gas price failed, %v", txHash, err)
	}

	gasUsed, err := geth.JSONHexToDecimal(receipt, "gasUsed")
	if err != nil {
		return nil, fmt.Errorf("parse tx %s gas used failed, %v", txHash, err)
	}

	remainGas := gasLimit.Sub(gasUsed)
	info := &checker.ReadjustFeeInfo{
		RemainFee: remainGas.Mul(gasPrice).Mul(decimal.New(1, -geth.Precision)),
		FeeSymbol: cfg.Currency,
	}
	return info, nil
}
