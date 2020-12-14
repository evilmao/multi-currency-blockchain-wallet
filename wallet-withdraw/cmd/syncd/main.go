package main

import (
	"fmt"
	"os"

	"upex-wallet/wallet-deposit/cmd"
	_ "upex-wallet/wallet-deposit/cmd/imports"
)

func main() {
	if err := cmd.Execute("syncd"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
