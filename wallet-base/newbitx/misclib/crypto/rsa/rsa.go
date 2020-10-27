package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// GenerateKey generates rsa privatekey.
func GenerateKey(size int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// ExportKey exports key with pem encoding.
func ExportKey(priv *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// ExportPubkey exports pubkey in the format "ssh-rsa ...".
func ExportPubkey(pub *rsa.PublicKey) ([]byte, error) {
	derPkix, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}

	pubPEM := pem.EncodeToMemory(&pubBlock)
	return pubPEM, nil
}

// Encrypt encrypts data with publickey.
func Encrypt(plainText []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, plainText)
}

// Decrypt decrypts data with privatekey.
func Decrypt(cipherText []byte, privkey []byte) ([]byte, error) {
	block, _ := pem.Decode(privkey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherText)
}

// B64Encrypt encrypts data with rsa and return base64 encoding result.
func B64Encrypt(plainText string, publicKey string) (string, error) {
	data, err := Encrypt([]byte(plainText), []byte(publicKey))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), err
}

// B64Decrypt decrypts data with rsa and base64 encoding.
func B64Decrypt(plainText string, privateKey string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(plainText)
	if err != nil {
		return nil, err
	}

	return Decrypt(data, []byte(privateKey))
}
