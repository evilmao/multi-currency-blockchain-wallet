package blockchain

import (
	"strings"

	"upex-wallet/wallet-base/currency"
)

// Blockchain name def.
const (
	OMNI  = "omni"
	ERC20 = "erc20"
)

const (
	MaxLen = 50
)

// CorrectName returns the correct blockchain name of currency.
func CorrectName(c, mainCurrency string) (name string, canEmpty bool) {
	if strings.EqualFold(c, "usdt") && strings.EqualFold(mainCurrency, "usdt") {
		return OMNI, true
	}

	details, _ := currency.CurrencyDetail(c)
	for _, detail := range details {
		if len(detail.BlockchainName) > 0 && detail.ChainBelongTo(mainCurrency) {
			return detail.BlockchainName, false
		}
	}

	return "", true
}
