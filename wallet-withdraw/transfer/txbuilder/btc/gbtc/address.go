package gbtc

import (
	"bytes"
	"errors"

	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/base/libs/bech32"

	"golang.org/x/crypto/ripemd160"
)

var (
	errBase58AddressFormat = errors.New("invalid base58 address format")
	errBase58AddressPrefix = errors.New("invalid base58 address prefix")

	errBech32AddressFormat  = errors.New("invalid bech32 address format")
	errBech32AddressHRP     = errors.New("invalid bech32 address HRP")
	errBech32AddressVersion = errors.New("invalid bech32 address version")
)

type Address interface {
	Data() []byte
	Script() []byte
	String() string
}

type AddressParam struct {
	P2PKHPrefix   []byte
	P2SHPrefix    []byte
	Bech32HRP     string
	Bech32Version byte
}

var (
	AddressParamBTC = &AddressParam{
		P2PKHPrefix:   []byte{0},
		P2SHPrefix:    []byte{5},
		Bech32HRP:     "bc",
		Bech32Version: 0,
	}
)

func ParseAddress(address string, param *AddressParam) (Address, error) {
	if hrp, data, err := bech32.Decode(address); err == nil {
		if len(data) == 0 {
			return nil, errBech32AddressFormat
		}

		if hrp != param.Bech32HRP {
			return nil, errBech32AddressHRP
		}

		version := data[0]
		if version != param.Bech32Version {
			return nil, errBech32AddressVersion
		}

		data, err = crypto.Bech32ConvertBits5To8(data[1:])
		if err != nil {
			return nil, errBech32AddressFormat
		}

		return &bech32Address{
			hrp:     hrp,
			version: version,
			data:    data,
		}, nil
	}

	prefix, data := crypto.DeBase58Check(address, 1, false)
	if len(data) != ripemd160.Size {
		return nil, errBase58AddressFormat
	}

	var scriptFunc func([]byte) []byte
	switch {
	case bytes.Equal(prefix, param.P2PKHPrefix):
		scriptFunc = P2PKHScript
	case bytes.Equal(prefix, param.P2SHPrefix):
		scriptFunc = P2SHScript
	default:
		return nil, errBase58AddressPrefix
	}

	return &base58Address{
		prefix:     prefix,
		data:       data,
		scriptFunc: scriptFunc,
	}, nil
}

type base58Address struct {
	prefix     []byte
	data       []byte
	scriptFunc func([]byte) []byte
}

func (a *base58Address) Prefix() []byte {
	return a.prefix
}

func (a *base58Address) Data() []byte {
	return a.data
}

func (a *base58Address) Script() []byte {
	return a.scriptFunc(a.data)
}

func (a *base58Address) String() string {
	return crypto.Base58Check(a.data, a.prefix, false)
}

type bech32Address struct {
	hrp     string
	version byte
	data    []byte
}

func (a *bech32Address) HRP() string {
	return a.hrp
}

func (a *bech32Address) Version() byte {
	return a.version
}

func (a *bech32Address) Data() []byte {
	return a.data
}

func (a *bech32Address) Script() []byte {
	return P2Bech32(a.version, a.data)
}

func (a *bech32Address) String() string {
	data, err := crypto.Bech32ConvertBits8To5(a.data)
	if err != nil {
		return ""
	}

	data = append([]byte{a.version}, data...)
	s, _ := bech32.Encode(a.hrp, data)
	return s
}
