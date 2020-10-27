package main

import (
	"encoding/hex"
	"fmt"

	"upex-wallet/wallet-tools/base/libs/secp256k1"

	"upex-wallet/wallet-base/cmd"
)

var (
	pubKeys []string
)

func main() {
	c := cmd.New(
		"ecdsautil",
		"util for ecdsa.",
		",/ecdsautil -k <pubkey>",
		run,
	)
	c.Flags().StringSliceVarP(&pubKeys, "pubkey", "k", nil, "the pubkeys")
	c.Execute()
}

func run(*cmd.Command) error {
	for i, k := range pubKeys {
		data, err := hex.DecodeString(k)
		if err != nil {
			return fmt.Errorf("hex decode pubkey at index %d failed, %v", i, err)
		}

		pubKey, err := secp256k1.ParsePubKey(data)
		if err != nil {
			return fmt.Errorf("parse pubkey at index %d failed, %v", i, err)
		}

		fmt.Printf("compressed:\n%s\n", hex.EncodeToString(pubKey.SerializeCompressed()))
		fmt.Printf("uncompressed:\n%s\n", hex.EncodeToString(pubKey.SerializeUncompressed()))
		fmt.Println()
	}
	return nil
}
