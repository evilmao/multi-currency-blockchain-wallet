package keypair

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-tools/base/crypto/addrprovider"
	"upex-wallet/wallet-tools/base/crypto/signer"
	"upex-wallet/wallet-tools/base/libs/secp256k1"
)

type CryptoClass string

const (
	InvalidCryptoClass CryptoClass = ""
)

type KeyPair interface {
	Class() string
	Cryptography() CryptoClass

	Random() error
	SetPrivateKey(privKey []byte) error

	PrivateKey() []byte
	PublicKey() []byte

	Sign(data []byte) ([]byte, error)
	Verify(sigData []byte, data []byte) (bool, error)

	Address() []byte
	AddressString() string
	AddressProvider() addrprovider.AddrProvider
}

type DerivableKeyPair interface {
	KeyPair
	Derive(idx int) (DerivableKeyPair, error)
}

type ExtData interface {
	Data() []byte
	String() string
}

type WithExtData interface {
	ExtData() map[string]ExtData
}

type PublicKey interface {
	Class() string
	PublicKey() []byte
	AddressString() string
}

type publicKey struct {
	class              string
	pubKey             []byte
	uncompressedPubKey []byte
	address            string
}

func CreatePublicKey(sampleAP KeyPair, pubKey []byte) (PublicKey, error) {
	var k publicKey
	k.class = sampleAP.Class()
	k.pubKey = pubKey
	k.uncompressedPubKey = k.pubKey
	if strings.Index(string(sampleAP.Cryptography()), string(signer.Secp256k1Class)) >= 0 {
		pubKey, err := secp256k1.ParsePubKey(k.pubKey)
		if err != nil {
			return nil, fmt.Errorf("parse public key failed, %v", err)
		}

		k.uncompressedPubKey = pubKey.SerializeUncompressed()
	}

	ap := sampleAP.AddressProvider()
	if ap == nil {
		return nil, fmt.Errorf("%s has no address provider", k.class)
	}

	k.address = ap.AddressString(&k)
	k.uncompressedPubKey = nil
	return &k, nil
}

func (k *publicKey) Class() string                 { return k.class }
func (k *publicKey) PublicKey() []byte             { return k.pubKey }
func (k *publicKey) PublicKeyUncompressed() []byte { return k.uncompressedPubKey }
func (k *publicKey) AddressString() string         { return k.address }
