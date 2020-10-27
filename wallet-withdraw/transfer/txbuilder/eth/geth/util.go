package geth

import (
	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
)

func JSONHexToDecimal(data []byte, path ...string) (decimal.Decimal, error) {
	value, err := jsonparser.GetString(data, path...)
	if err != nil {
		return decimal.Zero, err
	}

	bigInt, err := hexutil.DecodeBig(value)
	if err != nil || bigInt == nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(bigInt, 0), nil
}
