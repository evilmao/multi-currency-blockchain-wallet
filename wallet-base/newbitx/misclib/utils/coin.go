package utils

import (
	"math/big"
)

func round(f float64) float64 {
	return float64(f*100000000) / 100000000
}

// CoinToFloat converts coin to float64.
func CoinToFloat(val, c *big.Int) float64 {
	bigval := new(big.Float)
	bigval.SetInt(val)

	coin := new(big.Float)
	coin.SetInt(c)
	bigval.Quo(bigval, coin)

	fCoin, _ := bigval.Float64()
	return round(fCoin)
}

// FloatToCoin coverts float64 to coin.
func FloatToCoin(val float64, c *big.Int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)

	coin := new(big.Float)
	coin.SetInt(c)
	coin.SetPrec(64)

	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result)

	return result
}
