package generator

import (
	"io"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
)

type KeyPairIndex interface {
	KeyPairAtIndex(index int) (keypair.KeyPair, bool)
}

type FromWallet struct {
	sourceWallet KeyPairIndex
}

func NewFromWallet(sourceWallet KeyPairIndex) *FromWallet {
	return &FromWallet{
		sourceWallet: sourceWallet,
	}
}

func (g *FromWallet) Init() error { return nil }

func (g *FromWallet) Generate(idx int) (keypair.KeyPair, error) {
	kp, ok := g.sourceWallet.KeyPairAtIndex(idx)
	if !ok {
		return nil, io.EOF
	}

	return kp, nil
}

type KeyPairIndexV2 interface {
	KeyPairAtIndex(password string, index int) (keypair.KeyPair, error)
	Len() int
}

type FromWalletV2 struct {
	password     string
	sourceWallet KeyPairIndexV2
}

func NewFromWalletV2(password string, sourceWallet KeyPairIndexV2) *FromWalletV2 {
	return &FromWalletV2{
		password:     password,
		sourceWallet: sourceWallet,
	}
}

func (g *FromWalletV2) Init() error { return nil }

func (g *FromWalletV2) Generate(idx int) (keypair.KeyPair, error) {
	if idx >= g.sourceWallet.Len() {
		return nil, io.EOF
	}

	kp, err := g.sourceWallet.KeyPairAtIndex(g.password, idx)
	if err != nil {
		return nil, err
	}

	return kp, nil
}
