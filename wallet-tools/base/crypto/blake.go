package crypto

import (
	"github.com/dchest/blake256"
	"golang.org/x/crypto/blake2b"
)

// Blake2b224 calculates and returns the blake2b hash of the input data.
// with 28 digest size.
func Blake2b224(data ...[]byte) []byte {
	h, _ := blake2b.New(28, nil)
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}

// Blake256 calculates and returns the blake256 hash of the input data.
func Blake256(data []byte) []byte {
	return Sum(data, blake256.New())
}
