package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/utils"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptR     = 6
	scryptN     = 1 << 2
	scryptP     = 1
	scryptDKLen = 32
)

var (
	// ErrDecrypt represents the decrypt error message.
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

// GenEncryptInfo generates salt and iv for aes encryption.
func GenEncryptInfo() (salt, iv []byte) {
	salt = utils.GetRandomBytes(32)
	iv = GenRandomIV()
	return
}

// GenRandomIV generates random iv.
func GenRandomIV() []byte {
	return utils.GetRandomBytes(aes.BlockSize)
}

// GetDerivedKey gets aes params according the auth.
func GetDerivedKey(auth string, salt []byte) []byte {
	authArray := []byte(auth)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil
	}
	return derivedKey
}

// Encrypt encrypts a key(plain text) using the specified key into bytes
// that can be decrypted later on.
func Encrypt(derivedKey []byte, key []byte, iv []byte) ([]byte, error) {
	cipherText, err := aesCTRXOR(derivedKey[:16], key, iv)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

// Decrypt decrypts a key from bytes, returning the private key bytes.
func Decrypt(derivedKey []byte, cipherText []byte, iv, mac []byte) ([]byte, error) {
	// check mac
	if !bytes.Equal(crypto.Keccak256(derivedKey[16:32]), mac) {
		return nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	return plainText, err
}

// PKCS7Padding or PKCS5UnPadding padding plain text.
func PKCS7Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, paddingText...)
}

// PKCS7UnPadding or PKCS5UnPadding unpadding origin plain text.
func PKCS7UnPadding(plainText []byte) []byte {
	length := len(plainText)
	unpadding := int(plainText[length-1])
	return plainText[:(length - unpadding)]
}

// CBCEncrypt encrypts data using the specified key, iv.
func CBCEncrypt(plainText, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	paddingText := PKCS7Padding(plainText, block.BlockSize())
	cbcMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(paddingText))
	cbcMode.CryptBlocks(cipherText, paddingText)

	return cipherText, nil
}

// CBCDecrypt decrypts data using the specified key, iv.
func CBCDecrypt(cipherText, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cbcMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cbcMode.CryptBlocks(plainText, cipherText)
	plainText = PKCS7UnPadding(plainText)

	return plainText, nil
}
