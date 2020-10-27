package main

import (
	"bufio"
	"fmt"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair/storer"

	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
)

func write(keypairs []keypair.PublicKey) error {
	defer util.DeferLogTimeCost("[write]")()

	if len(keypairs) == 0 {
		return nil
	}

	works := []func([]keypair.PublicKey) error{
		writeDepositSQLFile,
		writeAddressFile,
		writeExSQLFile,
		writeWithdrawSQLFile,
	}

	return util.BatchDo(len(works), func(i int) (interface{}, error) {
		err := works[i](keypairs)
		return nil, err
	}, nil)
}

func writeDepositSQLFile(keypairs []keypair.PublicKey) error {
	if len(depositSQLFile) == 0 {
		return nil
	}

	w, err := storer.OpenDepositAddressSQLWriter(depositSQLFile)
	if err != nil {
		return err
	}

	defer w.Close()

	return writeByWriter(w, keypairs)
}

func writeAddressFile(keypairs []keypair.PublicKey) error {
	if len(addrsFile) == 0 {
		return nil
	}

	w, err := storer.OpenAddressFileWriter(addrsFile)
	if err != nil {
		return err
	}

	defer w.Close()

	return writeByWriter(w, keypairs)
}

func encryptAddress(address, pubKey string) (string, error) {
	const version = "1"

	result, err := rsa.B64Encrypt(address, pubKey)
	if err != nil {
		return "", fmt.Errorf("encrypt address failed, %v", err)
	}
	return version + "," + result, nil
}

func writeExSQLFile(keypairs []keypair.PublicKey) error {
	if len(exchangeSQLFile) == 0 {
		return nil
	}

	err := util.WithWriteFile(exchangeSQLFile, func(writer *bufio.Writer) error {
		counter := 1
		sql := "INSERT INTO t_deposit_address (f_currency, f_address) VALUES "
		subsql := ""
		return util.BatchDo(len(keypairs), func(i int) (interface{}, error) {
			kp := keypairs[i]
			addrStr := kp.AddressString()
			addrStr, err := encryptAddress(addrStr, pubKey)
			if err != nil {
				return nil, err
			}

			return addrStr, nil
		}, func(i int, data interface{}) error {
			addrStr := data.(string)
			subsql += fmt.Sprintf("('%d', '%s')", symbolID, addrStr)
			if counter%100 == 0 || i == len(keypairs)-1 {
				subsql += ";\n"
				_, err := writer.WriteString(sql + subsql)
				if err != nil {
					return err
				}

				subsql = ""
			} else {
				subsql += ",\n"
			}

			counter++
			return nil
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func writeWithdrawSQLFile(keypairs []keypair.PublicKey) error {
	if len(withdrawSQLFile) == 0 {
		return nil
	}

	w, err := storer.OpenWithdrawAddressSQLWriter(withdrawSQLFile, isSystemAddress)
	if err != nil {
		return err
	}

	defer w.Close()

	return writeByWriter(w, keypairs)
}

func writeByWriter(w storer.Writer, keypairs []keypair.PublicKey) error {
	for _, pubKey := range keypairs {
		err := w.Write(pubKey)
		if err != nil {
			return err
		}
	}
	return nil
}
