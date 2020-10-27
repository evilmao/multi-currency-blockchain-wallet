package signer

import (
	"fmt"
	"math/big"

	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/libs/secp256k1"
)

const (
	Secp256k1RecoverableClass              Class = "secp256k1recoverable"
	Secp256k1RecoverableCompressedKeyClass Class = "secp256k1recoverablecompressedkey"

	Secp256k1CanonicalClass              Class = "secp256k1canonical"
	Secp256k1CanonicalCompressedKeyClass Class = "secp256k1canonicalcompressedkey"
)

type Secp256k1Recoverable struct {
	isCompressedKey bool
	isCanonical     bool
}

func NewSecp256k1Recoverable(isCompressedKey bool) Signer {
	return &Secp256k1Recoverable{
		isCompressedKey: isCompressedKey,
		isCanonical:     false,
	}
}

func NewSecp256k1Canonical(isCompressedKey bool) Signer {
	return &Secp256k1Recoverable{
		isCompressedKey: isCompressedKey,
		isCanonical:     true,
	}
}

func (signer *Secp256k1Recoverable) Class() Class {
	if signer.isCanonical {
		if signer.isCompressedKey {
			return Secp256k1CanonicalCompressedKeyClass
		}
		return Secp256k1CanonicalClass
	} else {
		if signer.isCompressedKey {
			return Secp256k1RecoverableCompressedKeyClass
		}
		return Secp256k1RecoverableClass
	}
}

func (signer *Secp256k1Recoverable) Sign(k key.Key, data []byte) ([]byte, error) {
	secpKey, ok := k.(*key.Secp256k1)
	if !ok {
		return nil, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if secpKey.PrivKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	if len(data) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(data))
	}

	var (
		sig []byte
		err error
	)
	if signer.isCanonical {
		sig, err = secp256k1.SignCanonical(secpKey.PrivKey, data, signer.isCompressedKey)
	} else {
		sig, err = secp256k1.SignCompact(secpKey.PrivKey, data, signer.isCompressedKey)

		// Convert to Ethereum signature format with 'recovery id' v at the end.
		v := sig[0] - 27
		copy(sig, sig[1:])
		sig[64] = v
	}
	if err != nil {
		return nil, err
	}

	return sig, nil
}

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

func (signer *Secp256k1Recoverable) Verify(k key.Key, sigData []byte, data []byte) (bool, error) {
	secpKey, ok := k.(*key.Secp256k1)
	if !ok {
		return false, fmt.Errorf("unsupport key class: %s", k.Class())
	}

	if secpKey.PubKey == nil {
		return false, fmt.Errorf("public key is nil")
	}

	if len(sigData) != 64 && len(sigData) != 65 {
		return false, fmt.Errorf("invalid signature len: %d", len(sigData))
	}

	if len(sigData) == 65 {
		if signer.isCanonical {
			sigData = sigData[1:]
		} else {
			sigData = sigData[:64]
		}
	}

	sig := &secp256k1.Signature{
		R: new(big.Int).SetBytes(sigData[:32]),
		S: new(big.Int).SetBytes(sigData[32:]),
	}
	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	if sig.S.Cmp(secp256k1halfN) > 0 {
		return false, nil
	}

	return sig.Verify(data, secpKey.PubKey), nil
}
