package key

import (
	"fmt"

	"upex-wallet/wallet-tools/base/libs/secp256k1"
)

const (
	Secp256k1Class Class = "secp256k1"
)

type Secp256k1 struct {
	PrivKey *secp256k1.PrivateKey
	PubKey  *secp256k1.PublicKey
}

func NewSecp256k1() Key {
	return &Secp256k1{}
}

func (k *Secp256k1) Class() Class {
	return Secp256k1Class
}

func (k *Secp256k1) Random() error {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return err
	}

	k.PrivKey = priv
	k.PubKey = (*secp256k1.PublicKey)(&priv.PublicKey)
	return nil
}

func (k *Secp256k1) SetPrivateKey(privKey []byte) error {
	if len(privKey) == 0 {
		return fmt.Errorf("invalid private key")
	}

	k.PrivKey, k.PubKey = secp256k1.PrivKeyFromBytes(privKey)
	return nil
}

func (k *Secp256k1) PrivateKey() []byte {
	if k.PrivKey == nil {
		return nil
	}

	return k.PrivKey.Serialize()
}

func (k *Secp256k1) PublicKey() []byte {
	if k.PubKey == nil {
		return nil
	}

	return k.PubKey.SerializeCompressed()
}

func (k *Secp256k1) PublicKeyUncompressed() []byte {
	if k.PubKey == nil {
		return nil
	}

	return k.PubKey.SerializeUncompressed()
}
