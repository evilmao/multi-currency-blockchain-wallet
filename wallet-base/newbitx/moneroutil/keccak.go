package moneroutil

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"

	"github.com/ebfe/keccak"
)

const (
	ChecksumLength = 4
	HashLength     = 32
)

type Hash [HashLength]byte
type Checksum [ChecksumLength]byte

var (
	NullHash = [HashLength]byte{}
)

func Keccak256(data ...[]byte) (result Hash) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

func GetChecksum(data ...[]byte) (result Checksum) {
	keccak256 := Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

func Keccak512(data ...[]byte) (result Hash) {
	h := keccak.New512()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func HashesEqual(h1, h2 Hash) bool {
	return bytes.Equal(h1[:], h2[:])
}

func ParseHash(buf io.Reader) (Hash, error) {
	h := Hash{}
	_, err := io.ReadFull(buf, h[:])
	return h, err
}

func HexToHash(h string) (Hash, error) {
	result := Hash{}
	if len(h) != HashLength*2 {
		return result, errors.New("hash hex string must be 64 bytes long")
	}

	byteSlice, err := hex.DecodeString(h)
	if err != nil {
		return result, err
	}

	copy(result[:], byteSlice)
	return result, nil
}
