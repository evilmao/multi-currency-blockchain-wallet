// Package zillean is golang library for Zilliqa blockchain.
package zillean

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	crypto "github.com/GincoInc/go-crypto"
)

var (
	curve = crypto.Secp256k1()
)

// Zillean represents the zillean object.
type Zillean struct {
	ECS *ECSchnorr
	RPC *RPC
}

// NewZillean returns a new Zillean.
func NewZillean(endpoint string) *Zillean {
	return &Zillean{
		ECS: NewECSchnorr(),
		RPC: NewRPC(endpoint),
	}
}

// GeneratePrivateKey returns string which represents a generated private key.
func (z *Zillean) GeneratePrivateKey() string {
	return fmt.Sprintf("%x", z.ECS.GeneratePrivateKey())
}

// VerifyPrivateKey verifies a EC-Schnorr private key.
func (z *Zillean) VerifyPrivateKey(privateKey string) (bool, error) {
	privKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return false, err
	}

	D := new(big.Int).SetBytes(privKey)

	// The D must < N
	if D.Cmp(z.ECS.Curve.Params().N) >= 0 {
		return false, errors.New("invalid private key, >= N")
	}

	// The D must not be zero or negative.
	if D.Sign() <= 0 {
		return false, errors.New("invalid private key, zero or negative")
	}

	pubx, _ := z.ECS.Curve.ScalarBaseMult(D.Bytes())
	if pubx == nil {
		return false, errors.New("invalid private key")
	}

	return true, nil
}

// GetPublicKeyFromPrivateKey returns the public key derived from a private key.
func (z *Zillean) GetPublicKeyFromPrivateKey(privateKey string) (string, error) {
	privKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", z.ECS.GetPublicKey(privKey, true)), nil
}

// IsPublicKey checks whether a given string is a public key or not.
func (z *Zillean) IsPublicKey(publicKey string) bool {
	return regexp.MustCompile(`^[0-9a-fA-F]{66}$`).MatchString(publicKey)
}

// GetAddressFromPrivateKey returns the address derived from a private key.
func (z *Zillean) GetAddressFromPrivateKey(privateKey string) (string, error) {
	privKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	return publicKeyToAddress(z.ECS.GetPublicKey(privKey, true)), nil
}

// GetAddressFromPublicKey returns the address derived from a public key.
func (z *Zillean) GetAddressFromPublicKey(publicKey string) (string, error) {
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	return publicKeyToAddress(publicKeyBytes), nil
}

// IsAddress checks whether a given string is an address or not.
func (z *Zillean) IsAddress(address string) bool {
	return regexp.MustCompile(`^[0-9a-fA-F]{40}$`).MatchString(address)
}

// SignTransaction returns the EC-Schnorr signature on a raw transaction.
func (z *Zillean) SignTransaction(rawTx RawTransaction, privateKey string) (string, error) {
	privKey, _ := hex.DecodeString(privateKey)
	pubKey, _ := hex.DecodeString(rawTx.PubKey)
	r, s := z.ECS.Sign(privKey, pubKey, encodeTransaction(rawTx))

	return fmt.Sprintf("%x%x", r, s), nil
}

// VerifySignature verifies a signature on a message.
func (z *Zillean) VerifySignature(r, s, publicKey, msg []byte) bool {
	return z.ECS.Verify(r, s, publicKey, msg)
}
