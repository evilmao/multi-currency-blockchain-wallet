package addrprovider

import (
	"upex-wallet/wallet-tools/base/crypto"
)

const (
	TRXClass Class = "trx"
)

type TRX struct{}

func NewTRX() AddrProvider {
	return &TRX{}
}

func (*TRX) Class() Class {
	return TRXClass
}

func (*TRX) Address(k Key) []byte {
	uncompressedPubKey := k.PublicKeyUncompressed()
	hash := crypto.SumLegacyKeccak256(uncompressedPubKey[1:])
	return hash[12:]
}

func (p *TRX) AddressString(k Key) string {
	return crypto.Base58Check(p.Address(k), []byte{0x41}, false)
}
