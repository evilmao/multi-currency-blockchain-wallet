package signer

import (
	"fmt"

	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/libs/secp256k1"
)

const (
	Secp256k1Class Class = "secp256k1"
)

type Secp256k1 struct{}

func NewSecp256k1() Signer {
	return &Secp256k1{}
}

func (*Secp256k1) Class() Class {
	return Secp256k1Class
}

func (*Secp256k1) Sign(k key.Key, data []byte) ([]byte, error) {
	secpKey, ok := k.(*key.Secp256k1)
	if !ok {
		return nil, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if secpKey.PrivKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	sig, err := secpKey.PrivKey.Sign(data)
	if err != nil {
		return nil, err
	}

	return sig.Serialize(), nil
}

func (*Secp256k1) Verify(k key.Key, sigData []byte, data []byte) (bool, error) {
	secpKey, ok := k.(*key.Secp256k1)
	if !ok {
		return false, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if secpKey.PubKey == nil {
		return false, fmt.Errorf("public key is nil")
	}

	sig, err := secp256k1.ParseSignature(sigData, secp256k1.S256())
	if err != nil {
		return false, err
	}

	return sig.Verify(data, secpKey.PubKey), nil
}
