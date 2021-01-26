package main

import (
	_ "upex-wallet/wallet-base/api"
	_ "upex-wallet/wallet-base/cmd"
	_ "upex-wallet/wallet-base/currency"
	_ "upex-wallet/wallet-base/db"
	_ "upex-wallet/wallet-base/jsonrpc"
	_ "upex-wallet/wallet-base/models"
	_ "upex-wallet/wallet-base/monitor"
	_ "upex-wallet/wallet-base/service"
	_ "upex-wallet/wallet-base/util"
	_ "upex-wallet/wallet-base/viper"
)

// Just for checking any compile error.
func main() {}
