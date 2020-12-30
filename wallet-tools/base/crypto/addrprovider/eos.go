package addrprovider

import (
	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/base/libs/base58"
)

const (
	EOSClass Class = "eos"
)

type EOS struct {
	prefix string
}

func NewEOS() AddrProvider {
	return &EOS{
		prefix: "EOS",
	}
}

func NewBTS() AddrProvider {
	return &EOS{
		prefix: "BTS",
	}
}

func NewABBC() AddrProvider {
	return &EOS{
		prefix: "ABBC",
	}
}

func (*EOS) Class() Class {
	return EOSClass
}

func (*EOS) Address(k Key) []byte {
	return k.PublicKey()
}

func (p *EOS) AddressString(k Key) string {
	b := p.Address(k)
	checkSum := crypto.SumRipemd160(b)
	a := append(b, checkSum[:4]...)
	return p.prefix + base58.StdEncoding.Encode(a)
}
