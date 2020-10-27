package key

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/ed25519"
)

const (
	Ed25519Class Class = "ed25519"
)

type Ed25519 struct {
	PrivKey ed25519.PrivateKey
	PubKey  ed25519.PublicKey
}

func NewEd25519() Key {
	return &Ed25519{}
}

func (k *Ed25519) Class() Class {
	return Ed25519Class
}

func (k *Ed25519) Random() error {
	var err error
	k.PubKey, k.PrivKey, err = ed25519.GenerateKey(rand.Reader)
	return err
}

func (k *Ed25519) SetPrivateKey(privKey []byte) error {
	if len(privKey) == ed25519.SeedSize || len(privKey) == ed25519.PrivateKeySize {
		var err error
		k.PubKey, k.PrivKey, err = ed25519.GenerateKey(bytes.NewReader(privKey))
		return err
	}
	return fmt.Errorf("invalid private key len: %d", len(privKey))
}

func (k *Ed25519) PrivateKey() []byte {
	return k.PrivKey
}

func (k *Ed25519) PublicKey() []byte {
	return k.PubKey
}

func (k *Ed25519) PublicKeyUncompressed() []byte {
	return k.PublicKey()
}
