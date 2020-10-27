package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"upex-wallet/wallet-tools/base/crypto"
)

const (
	NISTP256Class Class = "NISTP256"
)

type NISTP256 struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  *ecdsa.PublicKey
}

func NewNISTP256() Key {
	return &NISTP256{}
}

func (k *NISTP256) Class() Class {
	return NISTP256Class
}

func (k *NISTP256) Random() error {
	var err error
	k.PrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	k.PubKey = &k.PrivKey.PublicKey
	return nil
}

func (k *NISTP256) SetPrivateKey(privKey []byte) error {
	if len(privKey) == 0 {
		return fmt.Errorf("invalid private key")
	}

	k.PrivKey = crypto.PrivKeyFromBytes(privKey, elliptic.P256())
	k.PubKey = &k.PrivKey.PublicKey
	return nil
}

func (k *NISTP256) PrivateKey() []byte {
	if k.PrivKey == nil {
		return nil
	}

	return k.PrivKey.D.Bytes()
}

func (k *NISTP256) PublicKey() []byte {
	if k.PubKey == nil {
		return nil
	}

	return crypto.PublicKey(*k.PubKey).SerializeCompressed()
}

func (k *NISTP256) PublicKeyUncompressed() []byte {
	if k.PubKey == nil {
		return nil
	}

	return crypto.PublicKey(*k.PubKey).SerializeUncompressed()
}
