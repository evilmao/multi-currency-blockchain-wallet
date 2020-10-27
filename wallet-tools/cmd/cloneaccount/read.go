package main

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

type publicKey struct {
	pubKey  []byte
	address string
}

func (*publicKey) Class() string           { return "" }
func (k *publicKey) PublicKey() []byte     { return k.pubKey }
func (k *publicKey) AddressString() string { return k.address }

func NewPublicKey(pubKeyData []byte, addr string) (keypair.PublicKey, error) {
	if len(convert) > 0 {
		if len(pubKeyData) == 0 {
			return nil, fmt.Errorf("covert address %s failed, need address public key", addr)
		}

		return keypair.CreatePublicKey(convertTo, pubKeyData)
	}

	if len(prefix) > 0 {
		addr = crypto.ReplaceAddressPrefix(addr, 1, prefix)
	}
	return &publicKey{pubKeyData, addr}, nil
}

func read() ([]keypair.PublicKey, error) {
	defer util.DeferLogTimeCost("[read]")()

	if len(sourceFile) == 0 {
		return nil, fmt.Errorf("source file can't be empty")
	}

	if strings.HasSuffix(sourceFile, ".txt") {
		return readAddressFile()
	}

	if strings.HasSuffix(sourceFile, ".sql") {
		return readSQLFile()
	}

	return nil, fmt.Errorf("unsupport source file type")
}

func readAddressFile() ([]keypair.PublicKey, error) {
	var keypairs []keypair.PublicKey
	err := util.WithReadFileLineByLine(sourceFile, func(addr string) error {
		addr = strings.TrimSpace(addr)
		if len(addr) > 0 {
			pubKey, err := NewPublicKey(nil, addr)
			if err != nil {
				return err
			}

			keypairs = append(keypairs, pubKey)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return keypairs, nil
}

var (
	addrSQLReg = regexp.MustCompile(`\('(\w+)',\s+'(\w+)',\s+(\d)`)
)

func readSQLFile() ([]keypair.PublicKey, error) {
	var pubKeys []*publicKey
	err := util.WithReadFileLineByLine(sourceFile, func(line string) error {
		line = strings.TrimSpace(line)
		addrs := addrSQLReg.FindAllStringSubmatch(line, -1)
		for _, addr := range addrs {
			if len(addr) == 4 {
				pubKeyData, err := hex.DecodeString(addr[2])
				if err != nil {
					return err
				}

				pubKeys = append(pubKeys, &publicKey{
					pubKey:  pubKeyData,
					address: addr[1],
				})

				if !isSystemAddress {
					isSystemAddress = (addr[3] == "0")
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	keypairs := make([]keypair.PublicKey, len(pubKeys))
	err = util.BatchDo(len(pubKeys), func(i int) (interface{}, error) {
		pubKey := pubKeys[i]
		pubKeys[i] = nil
		return NewPublicKey(pubKey.pubKey, pubKey.address)
	}, func(i int, data interface{}) error {
		keypairs[i] = data.(keypair.PublicKey)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return keypairs, nil
}
