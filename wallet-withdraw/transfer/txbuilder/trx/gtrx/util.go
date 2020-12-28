package gtrx

import (
	"encoding/hex"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"

	"github.com/sasaxie/go-client-api/common/base58"
)

func AddressToHex(address string) (string, error) {
	if _, err := base58.Decode(address); err != nil {
		return "", err
	}

	return hex.EncodeToString(base58.DecodeCheck(address)), nil
}

func JSONHexToDecimal(data []byte, path ...string) (decimal.Decimal, error) {
	value, err := jsonparser.GetInt(data, path...)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromInt(value), nil
}
