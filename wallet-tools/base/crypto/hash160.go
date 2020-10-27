package crypto

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return Sum(Sum(buf, sha256.New()), ripemd160.New())
}

// DoubleSha256 calculates the hash sha256(sha256(b)).
func DoubleSha256(buf []byte) []byte {
	return Sum(Sum(buf, sha256.New()), sha256.New())
}

// SumRipemd160 returns the RIPEMD-160 digest of the data.
func SumRipemd160(data []byte) []byte {
	return Sum(data, ripemd160.New())
}
