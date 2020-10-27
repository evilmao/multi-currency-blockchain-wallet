package crypto

import (
	"golang.org/x/crypto/sha3"
)

// SumLegacyKeccak256 returns the Keccak-256 digest of the data.
func SumLegacyKeccak256(data []byte) []byte {
	return Sum(data, sha3.NewLegacyKeccak256())
}
