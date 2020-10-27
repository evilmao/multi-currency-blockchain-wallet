package util

import (
	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/crypto/aes"
)

// Encryptor encrypt / decrypt data.
type Encryptor struct {
	salt, iv   []byte
	derivedKey []byte
	mac        []byte
}

// NewEncryptor returns a new Encryptor.
func NewEncryptor(password string) *Encryptor {
	salt, iv := aes.GenEncryptInfo()
	return MakeEncryptor(password, salt, iv)
}

// MakeEncryptor returns a new Encryptor.
func MakeEncryptor(password string, salt, iv []byte) *Encryptor {
	var en Encryptor
	en.salt, en.iv = salt, iv
	en.derivedKey = aes.GetDerivedKey(password, salt)
	en.mac = crypto.Keccak256(en.derivedKey[16:32])
	return &en
}

// Salt returns salt.
func (en *Encryptor) Salt() []byte {
	return en.salt
}

// IV returns iv.
func (en *Encryptor) IV() []byte {
	return en.iv
}

// Mac returns mac.
func (en *Encryptor) Mac() []byte {
	return en.mac
}

// Encrypt encrypts data.
func (en *Encryptor) Encrypt(data []byte) ([]byte, error) {
	return aes.Encrypt(en.derivedKey[:16], data, en.iv)
}

// Decrypt decrypts data.
func (en *Encryptor) Decrypt(data []byte) ([]byte, error) {
	return aes.Decrypt(en.derivedKey[:16], data, en.iv, en.mac)
}
