package currency

import "strings"

type CurrencyInfo struct {
	BlockchainName   string
	Symbol           string
	Address          string
	MinDepositAmount float64
	Confirm          int
	Decimal          int
}

// IsToken returns whether c is a token.
func (c *CurrencyInfo) IsToken() bool {
	if len(c.Address) > 0 {
		return true
	}

	if strings.EqualFold(c.Symbol, "usdt") {
		return true
	}

	return false
}

// BelongChainName returns the blockchain name that c belong to.
func (c *CurrencyInfo) BelongChainName() string {
	return c.BlockchainName
}

// ChainBelongTo returns whether c belong to the blockchain.
func (c *CurrencyInfo) ChainBelongTo(blockchain string) bool {
	return strings.EqualFold(c.BelongChainName(), blockchain)
}
