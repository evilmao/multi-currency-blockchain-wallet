package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ripemd160"
)

const (
	// HashSize represents the hash length.
	HashSize = 32
)

type (
	// Hash represents the 32 byte hash of arbitrary data.
	Hash [HashSize]byte
	// AddressFunc represents the hash method.
	AddressFunc func([]byte) string
)

// String returns the hex of the hash.
func (h Hash) String() string { return hex.EncodeToString(h[:]) }

// Bytes returns the bytes respresentation of the hash.
func (h Hash) Bytes() []byte { return h[:] }

// Reverse sets self byte-reversed hash.
func (h *Hash) Reverse() Hash {
	for i, b := range h[:HashSize/2] {
		h[i], h[HashSize-1-i] = h[HashSize-1-i], b
	}
	return *h
}

// Equal reports whether h1 equal h2.
func (h Hash) Equal(h1 Hash) bool {
	for i := 0; i < HashSize; i++ {
		if h[i] != h1[i] {
			return false
		}
	}
	return true
}

// SetBytes sets the hash to the value of b.
func (h *Hash) SetBytes(b []byte) {
	if len(*h) == len(b) {
		copy(h[:], b[:HashSize])
	}
}

// NewHash constructs a hash use the giving bytes.
func NewHash(data []byte) Hash {
	h := Hash{}
	h.SetBytes(data)
	return h
}

// HexToHash coverts string `s` to hash.
func HexToHash(s string) Hash {
	buf, _ := hex.DecodeString(s)
	return NewHash(buf)
}

// Sha256 calculates and returns sha256 hash of the input data.
func Sha256(data []byte) Hash {
	h := sha256.Sum256(data)

	return NewHash(h[:])
}

// DoubleSha256 calculates and returns double sha256 hash of the input data.
func DoubleSha256(data []byte) Hash {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])

	return NewHash(h[:])
}

// Hash160 calculates hash for bitcoin/litecoin/bcc address.
func Hash160(d []byte) []byte {
	h := sha256.Sum256(d)
	return Ripemd160(h[:])
}

// Ripemd160 calculates and returns ripemd160 hash of the input data.
func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}

// ethereum hash.

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Keccak512 calculates and returns the Keccak512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := sha3.NewKeccak512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Blake2b224 calculates and returns the blake2b hash of the input data.
// with 28 digest size.
func Blake2b224(data ...[]byte) []byte {
	h, _ := blake2b.New(28, nil)
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}

// // CreateAddress creates an ethereum address given the bytes and the nonce.
// func CreateAddress(b common.Address, nonce uint64) common.Address {
// 	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
// 	return common.BytesToAddress(Keccak256(data)[12:])
// }

// EthSignHash is a helper function that calculates a hash for the given message
// that can be safely used to calculate a signature from.
//
// The hash is calulcated as
//   keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func EthSignHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return Keccak256([]byte(msg))
}
