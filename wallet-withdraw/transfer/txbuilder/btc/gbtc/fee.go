package gbtc

import (
	"github.com/shopspring/decimal"
)

type FeeRate interface {
	OneOutPutFee(txType string, precision int32) decimal.Decimal
	OneInOutPutFee(txType string, precision int32) decimal.Decimal
	OneInputFee(txType string, precision int32) decimal.Decimal
}

func CalculateTxSize(nIn, nOut int) int {
	return nIn*148 + nOut*43
}

func CalculateTxFee(nByte int, feeRate float64) decimal.Decimal {
	if nByte == 0 {
		return decimal.Zero
	}

	fee := float64(nByte) * feeRate / 1000
	return decimal.NewFromFloat(fee)
}
