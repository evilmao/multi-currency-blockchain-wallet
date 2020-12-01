package main

import (
	"fmt"
	"os"

	"upex-wallet/wallet-deposit/cmd"
	_ "upex-wallet/wallet-deposit/cmd/imports"
)

// func main() {
// 	if err := cmd.Exec(createRPCClient); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// func createRPCClient(cfg *config.Config) rpc.RPC {
// 	if cfg == nil {
// 		return nil
// 	}
//
// 	creator, ok := rpc.Find(strings.ToUpper(cfg.Currency))
// 	if !ok {
// 		return nil
// 	}
// 	return creator(cfg)
// }
