package crypto

import (
	"hash"
)

// Sum calculate the sum hash of hasher over buf.
func Sum(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}
