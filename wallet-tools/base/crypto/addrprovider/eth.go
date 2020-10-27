package addrprovider

import (
	"encoding/hex"

	"upex-wallet/wallet-tools/base/crypto"
)

const (
	ETHClass Class = "eth"
)

type ETH struct {
	prefix string
}

func NewETH() AddrProvider {
	return &ETH{
		prefix: "0x",
	}
}

func NewPTNdeprecated() AddrProvider {
	return &ETH{
		prefix: "px",
	}
}

func (*ETH) Class() Class {
	return ETHClass
}

func (*ETH) Address(k Key) []byte {
	uncompressedPubKey := k.PublicKeyUncompressed()
	hash := crypto.SumLegacyKeccak256(uncompressedPubKey[1:])
	return hash[12:]
}

func (p *ETH) AddressString(k Key) string {
	return p.prefix + hex.EncodeToString(p.Address(k))
}
