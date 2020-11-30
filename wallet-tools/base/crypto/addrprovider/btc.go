package addrprovider

import (
	"upex-wallet/wallet-tools/base/crypto"
)

const (
	BTCClass Class = "btc"
)

type BTC struct {
	prefix        string
	addressPrefix []byte
}

func NewBTC(addressPrefix []byte) AddrProvider {
	return &BTC{
		addressPrefix: addressPrefix,
	}
}

func NewINT() AddrProvider {
	return &BTC{
		prefix:        "INT",
		addressPrefix: []byte{0},
	}
}

func NewPTN() AddrProvider {
	return &BTC{
		prefix:        "P",
		addressPrefix: []byte{0},
	}
}

func (*BTC) Class() Class {
	return BTCClass
}

func (*BTC) Address(k Key) []byte {
	return crypto.Hash160(k.PublicKey())
}

func (p *BTC) AddressString(k Key) string {
	return p.prefix + crypto.Base58Check(p.Address(k), p.addressPrefix, false)
}
