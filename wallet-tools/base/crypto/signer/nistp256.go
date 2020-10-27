package signer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"

	"upex-wallet/wallet-tools/base/crypto/key"
)

const (
	NISTP256Class Class = "NISTP256"
)

type NISTP256 struct{}

func NewNISTP256() Signer {
	return &NISTP256{}
}

func (*NISTP256) Class() Class {
	return NISTP256Class
}

func (*NISTP256) Sign(k key.Key, data []byte) ([]byte, error) {
	nistpKey, ok := k.(*key.NISTP256)
	if !ok {
		return nil, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if nistpKey.PrivKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	r, s, err := ecdsa.Sign(rand.Reader, nistpKey.PrivKey, data)
	if err != nil {
		return nil, err
	}

	sig := &NISTP256Signature{
		R: r,
		S: s,
	}
	return sig.Serialize(), nil
}

func (*NISTP256) Verify(k key.Key, sigData []byte, data []byte) (bool, error) {
	nistpKey, ok := k.(*key.NISTP256)
	if !ok {
		return false, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if nistpKey.PubKey == nil {
		return false, fmt.Errorf("public key is nil")
	}

	sig, err := ParseNISTP256Signature(sigData)
	if err != nil {
		return false, err
	}

	return sig.Verify(data, nistpKey.PubKey), nil
}

// NISTP256Signature is a type representing an ecdsa signature.
type NISTP256Signature struct {
	R *big.Int
	S *big.Int
}

func (sig *NISTP256Signature) Serialize() []byte {
	curve := elliptic.P256()
	size := (curve.Params().BitSize + 7) >> 3
	res := make([]byte, size*2)

	r := sig.R.Bytes()
	s := sig.S.Bytes()
	copy(res[size-len(r):], r)
	copy(res[size*2-len(s):], s)
	return res
}

func (sig *NISTP256Signature) Verify(data []byte, pubKey *ecdsa.PublicKey) bool {
	return ecdsa.Verify(pubKey, data, sig.R, sig.S)
}

func ParseNISTP256Signature(sigData []byte) (*NISTP256Signature, error) {
	curve := elliptic.P256()
	size := (curve.Params().BitSize + 7) >> 3
	if len(sigData) != size*2 {
		return nil, fmt.Errorf("invalid signature data len: %d", len(sigData))
	}

	sig := &NISTP256Signature{
		R: new(big.Int).SetBytes(sigData[:size]),
		S: new(big.Int).SetBytes(sigData[size:]),
	}
	return sig, nil
}
