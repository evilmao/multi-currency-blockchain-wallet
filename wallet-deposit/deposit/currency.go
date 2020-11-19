package deposit

import (
	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-config/deposit/config"
)

// CurrencyInit, initial currency from config
func CurrencyInit(cfg *config.Config) {

	currency.Init(cfg)
}
