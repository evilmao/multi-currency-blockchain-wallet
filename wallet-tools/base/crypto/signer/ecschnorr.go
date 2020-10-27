package signer

import (
	"bytes"
	"fmt"

	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/libs/secp256k1"

	zilCrypto "github.com/GincoInc/go-crypto"
	"github.com/GincoInc/zillean"
)

const (
	EcschnorrClass Class = "ecschnorr"
)

type Ecschnorr struct {
	Ecs *zillean.ECSchnorr
}

func NewEcschnorr() Signer {
	return &Ecschnorr{
		Ecs: &zillean.ECSchnorr{
			Curve: secp256k1.S256(),
		},
	}
}

func (*Ecschnorr) Class() Class {
	return EcschnorrClass
}

func (ecs *Ecschnorr) Sign(key key.Key, data []byte) ([]byte, error) {
	r, s := ecs.Ecs.Sign(key.PrivateKey(), key.PublicKey(), data)
	sigBytes := make([]byte, 0, 64)
	sigBytes = crypto.PaddedAppend(32, sigBytes, r)
	sigBytes = crypto.PaddedAppend(32, sigBytes, s)
	return sigBytes, nil
}

func (ecs *Ecschnorr) Verify(k key.Key, sigData []byte, data []byte) (bool, error) {
	if len(sigData) != 64 {
		return false, fmt.Errorf("invalid signature len: %d", len(sigData))
	}
	r := sigData[:32]
	s := sigData[32:]
	secpKey, ok := k.(*key.Secp256k1)
	if !ok {
		return false, fmt.Errorf("unsupport key class: %s", k.Class())
	}
	if secpKey.PubKey == nil {
		return false, fmt.Errorf("public key is nil")
	}
	rPubKeyX, rPubKeyY := ecs.Ecs.Curve.ScalarMult(secpKey.PubKey.GetX(), secpKey.PubKey.GetY(), r)
	sBasePointX, sBasePointY := ecs.Ecs.Curve.ScalarBaseMult(s)
	commitmentX, commitmentY := ecs.Ecs.Curve.Add(sBasePointX, sBasePointY, rPubKeyX, rPubKeyY)
	commitment := zilCrypto.Compress(ecs.Ecs.Curve, commitmentX, commitmentY)
	commitment = append(commitment, k.PublicKey()...)
	commitment = append(commitment, data...)
	return bytes.Equal(r, zilCrypto.Sha256(commitment)), nil
}
