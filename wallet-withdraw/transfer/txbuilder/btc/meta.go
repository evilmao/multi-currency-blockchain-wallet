package btc

import (
	"upex-wallet/wallet-withdraw/transfer/txbuilder"

	"github.com/shopspring/decimal"
)

func init() {
	txbuilder.AddMeta("BTC", 8, decimal.NewFromFloat(0.0005), 2, 0)
	txbuilder.AddMeta("ETP", 8, decimal.NewFromFloat(0.0002), 4, 0)
	txbuilder.AddMeta("ABBC", 8, decimal.NewFromFloat(0.0005), 1, 0)
	txbuilder.AddMeta("QTUM", 8, decimal.NewFromFloat(0.005), 2, 0)
	txbuilder.AddMeta("FAB", 8, decimal.NewFromFloat(0.004), 2, 0)
	txbuilder.AddMeta("MONA", 8, decimal.NewFromFloat(0.001), 2, 0)
	txbuilder.AddMeta("LTC", 8, decimal.NewFromFloat(0.00005), 2, 0)
}
