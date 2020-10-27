package main

import (
	"fmt"
	"os"
	"strings"

	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/cmd"
	"upex-wallet/wallet-deposit/rpc"
)

func main() {
	if err := cmd.Exec(createRPCClient); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createRPCClient(cfg *config.Config) rpc.RPC {
	if cfg == nil {
		return nil
	}

	creator, ok := rpc.Find(strings.ToUpper(cfg.Currency))
	if !ok {
		return nil
	}
	return creator(cfg)
}
