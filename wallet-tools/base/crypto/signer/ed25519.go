package signer

import (
	"fmt"

	"upex-wallet/wallet-tools/base/crypto/key"

	"golang.org/x/crypto/ed25519"
)

const (
	Ed25519Class Class = "ed25519"
)

type Ed25519 struct{}

func NewEd25519() Signer {
	return &Ed25519{}
}

func (*Ed25519) Class() Class {
	return Ed25519Class
}

func (*Ed25519) Sign(k key.Key, data []byte) ([]byte, error) {
	if len(k.PrivateKey()) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key len: %d", len(k.PrivateKey()))
	}

	return ed25519.Sign(k.PrivateKey(), data), nil
}

func (*Ed25519) Verify(k key.Key, sigData []byte, data []byte) (bool, error) {
	if len(k.PublicKey()) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key len: %d", len(k.PublicKey()))
	}

	return ed25519.Verify(k.PublicKey(), data, sigData), nil
}
