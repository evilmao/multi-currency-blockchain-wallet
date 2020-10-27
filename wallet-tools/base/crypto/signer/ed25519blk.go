package signer

import (
	"bytes"
	"fmt"
	"strconv"

	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/libs/edwards25519"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

const (
	Ed25519blkClass Class = "ed25519blk"
)

type Ed25519blk struct{}

func NewEd25519blk() Signer {
	return &Ed25519blk{}
}

func (*Ed25519blk) Class() Class {
	return Ed25519blkClass
}

func (*Ed25519blk) Sign(k key.Key, data []byte) ([]byte, error) {
	if len(k.PrivateKey()) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key len: %d", len(k.PrivateKey()))
	}

	h, _ := blake2b.New(64, nil)

	h.Write(k.PrivateKey()[:32])

	var digest1, messageDigest, hramDigest [64]byte
	var expandedSecretKey [32]byte
	h.Sum(digest1[:0])
	copy(expandedSecretKey[:], digest1[:])
	expandedSecretKey[0] &= 248
	expandedSecretKey[31] &= 63
	expandedSecretKey[31] |= 64

	h.Reset()
	h.Write(digest1[32:])
	h.Write(data)
	h.Sum(messageDigest[:0])

	var messageDigestReduced [32]byte
	edwards25519.ScReduce(&messageDigestReduced, &messageDigest)
	var R edwards25519.ExtendedGroupElement
	edwards25519.GeScalarMultBase(&R, &messageDigestReduced)

	var encodedR [32]byte
	R.ToBytes(&encodedR)

	h.Reset()
	h.Write(encodedR[:])
	h.Write(k.PrivateKey()[32:])
	h.Write(data)
	h.Sum(hramDigest[:0])
	var hramDigestReduced [32]byte
	edwards25519.ScReduce(&hramDigestReduced, &hramDigest)

	var s [32]byte
	edwards25519.ScMulAdd(&s, &hramDigestReduced, &expandedSecretKey, &messageDigestReduced)

	signature := make([]byte, ed25519.SignatureSize)
	copy(signature[:], encodedR[:])
	copy(signature[32:], s[:])

	return signature, nil
}

func (*Ed25519blk) Verify(k key.Key, sig []byte, data []byte) (bool, error) {
	if l := len(k.PublicKey()); l != ed25519.PublicKeySize {
		return false, fmt.Errorf("ed25519blk: bad public key length: " + strconv.Itoa(l))
	}

	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return false, fmt.Errorf("ed25519blk: bad SignatureSize length: %d", len(sig))
	}

	var A edwards25519.ExtendedGroupElement
	var publicKeyBytes [32]byte
	copy(publicKeyBytes[:], k.PublicKey())
	if !A.FromBytes(&publicKeyBytes) {
		return false, fmt.Errorf("ed25519blk: pubkey mismatch ")
	}
	edwards25519.FeNeg(&A.X, &A.X)
	edwards25519.FeNeg(&A.T, &A.T)

	//h := sha512.New()
	h, _ := blake2b.New(64, nil)
	h.Write(sig[:32])
	h.Write(k.PublicKey())
	h.Write(data)
	var digest [64]byte
	h.Sum(digest[:0])

	var hReduced [32]byte
	edwards25519.ScReduce(&hReduced, &digest)

	var R edwards25519.ProjectiveGroupElement
	var s [32]byte
	copy(s[:], sig[32:])

	// https://tools.ietf.org/html/rfc8032#section-5.1.7 requires that s be in
	// the range [0, order) in order to prevent signature malleability.
	if !edwards25519.ScMinimal(&s) {
		return false, fmt.Errorf("ed25519blk: pubkey mismatch ")
	}

	edwards25519.GeDoubleScalarMultVartime(&R, &hReduced, &A, &s)

	var checkR [32]byte
	R.ToBytes(&checkR)
	return bytes.Equal(sig[:32], checkR[:]), nil
}
