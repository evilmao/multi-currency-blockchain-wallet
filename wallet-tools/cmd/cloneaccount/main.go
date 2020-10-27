package main

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/util"
)

var (
	sourceFile      string
	depositSQLFile  string
	addrsFile       string
	exchangeSQLFile string
	withdrawSQLFile string

	prefix   []byte
	symbolID uint
	pubKey   string
	convert  string

	convertTo       keypair.KeyPair
	isSystemAddress bool
)

func main() {
	c := cmd.New(
		"cloneaccount",
		"clone & transform address from address.txt or sql",
		"",
		run,
	)
	c.Flags().StringVarP(&sourceFile, "sourcefile", "s", "", "the source address file or sql file")
	c.Flags().StringVarP(&depositSQLFile, "depositsqlfile", "t", "", "the output sql file for deposit")
	c.Flags().StringVarP(&addrsFile, "addrsfile", "a", "", "the output address file")
	c.Flags().StringVarP(&exchangeSQLFile, "exchangesqlfile", "q", "", "the output sql file for the exchange")
	c.Flags().StringVarP(&withdrawSQLFile, "withdrawsqlfile", "Q", "", "the output sql file for withdraw service")

	c.Flags().BytesHexVarP(&prefix, "prefix", "p", nil, "the new prefix of base58 format address")
	c.Flags().UintVarP(&symbolID, "symbolid", "i", 0, "the symbol id of the currency")
	c.Flags().StringVarP(&pubKey, "pubkey", "k", "", "the pubkey to encrypt address")
	c.Flags().StringVarP(&convert, "convert", "c", "", "convert address class from sql file, e.g.: btc->atom")
	c.Execute()
}

func initConvert() error {
	if len(convert) == 0 {
		return nil
	}

	convertPath := strings.Split(convert, "->")
	if len(convertPath) != 2 {
		return fmt.Errorf("invalid convert format: %s", convert)
	}

	build, ok := keypair.FindBuilder(convertPath[0])
	if !ok {
		return fmt.Errorf("can't find keypair of %s", convertPath[0])
	}

	fromKP := build.Build()

	build, ok = keypair.FindBuilder(convertPath[1])
	if !ok {
		return fmt.Errorf("can't find keypair of %s", convertPath[1])
	}

	toKP := build.Build()

	if fromKP.Cryptography() != toKP.Cryptography() {
		return fmt.Errorf("can't convert %s keypair to %s keypair",
			fromKP.Cryptography(), toKP.Cryptography())
	}

	convertTo = toKP
	return nil
}

func run(*cmd.Command) error {
	err := initConvert()
	if err != nil {
		return err
	}

	addrs, err := read()
	if err != nil {
		return err
	}

	util.TraceMemStats()

	if len(addrs) == 0 {
		return fmt.Errorf("read no data")
	}

	err = write(addrs)
	if err != nil {
		return err
	}

	util.TraceMemStats()

	return nil
}
