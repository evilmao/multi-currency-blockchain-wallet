package key

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strconv"

	"upex-wallet/wallet-tools/base/libs/edwards25519"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

const (
	Ed25519blkClass Class = "ed25519blk"
)

type Ed25519blk struct {
	PrivKey ed25519.PrivateKey
	PubKey  ed25519.PublicKey
}

func NewEd25519blk() Key {
	return &Ed25519blk{}
}

func (k *Ed25519blk) Class() Class {
	return Ed25519blkClass
}

func (k *Ed25519blk) Random() error {
	var err error
	k.PubKey, k.PrivKey, err = generateKey(rand.Reader)
	return err
}

func generateKey(reader io.Reader) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if reader == nil {
		reader = rand.Reader
	}

	seed := make([]byte, ed25519.SeedSize)
	if _, err := io.ReadFull(reader, seed); err != nil {
		return nil, nil, err
	}

	privateKey := newKeyFromSeed(seed)
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])

	return publicKey, privateKey, nil
}

func newKeyFromSeed(seed []byte) ed25519.PrivateKey {
	if l := len(seed); l != ed25519.SeedSize {
		panic("ed25519: bad seed length: " + strconv.Itoa(l))
	}

	var digest []byte
	h, _ := blake2b.New(64, nil)
	h.Write(seed)
	digest = h.Sum(digest)
	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	var A edwards25519.ExtendedGroupElement
	var hBytes [32]byte
	copy(hBytes[:], digest[:])
	edwards25519.GeScalarMultBase(&A, &hBytes)
	var publicKeyBytes [32]byte
	A.ToBytes(&publicKeyBytes)

	privateKey := make([]byte, ed25519.PrivateKeySize)
	copy(privateKey, seed)
	copy(privateKey[32:], publicKeyBytes[:])

	return privateKey
}

func (k *Ed25519blk) SetPrivateKey(privKey []byte) error {
	if len(privKey) == ed25519.SeedSize || len(privKey) == ed25519.PrivateKeySize {
		var err error
		k.PubKey, k.PrivKey, err = generateKey(bytes.NewReader(privKey))
		return err
	}
	return fmt.Errorf("invalid private key len: %d", len(privKey))
}

func (k *Ed25519blk) PrivateKey() []byte {
	return k.PrivKey
}

func (k *Ed25519blk) PublicKey() []byte {
	return k.PubKey
}

func (k *Ed25519blk) PublicKeyUncompressed() []byte {
	return k.PublicKey()
}
