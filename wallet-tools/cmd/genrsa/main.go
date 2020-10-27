package main

import (
	"bufio"
	"fmt"
	"os"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
)

const (
	KeySize = 2048
)

var (
	fileNames []string
)

func main() {
	c := cmd.New(
		"genrsa",
		"generate rsa key pair",
		"./genrsa test",
		run)
	c.Flags().StringSliceVarP(&fileNames, "filename", "f", nil, "name of the rsa file")

	if err := c.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cmd.Command) error {
	if len(fileNames) == 0 {
		fileNames = append(fileNames, "test")
	}

	for _, fileName := range fileNames {
		priv, err := rsa.GenerateKey(KeySize)
		if err != nil {
			return err
		}

		err = util.WithWriteFile(fmt.Sprintf("./%s_rsa", fileName), func(w *bufio.Writer) error {
			_, err := w.Write(rsa.ExportKey(priv))
			return err
		})
		if err != nil {
			return err
		}

		pubBytes, err := rsa.ExportPubkey(&priv.PublicKey)
		if err != nil {
			return err
		}

		err = util.WithWriteFile(fmt.Sprintf("./%s_rsa.pub", fileName), func(w *bufio.Writer) error {
			_, err := w.Write(pubBytes)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
