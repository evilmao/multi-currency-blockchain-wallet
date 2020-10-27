package crypto

import (
	"fmt"

	"upex-wallet/wallet-tools/base/libs/base58"
)

type WIFKey struct {
	prefix     []byte
	privKey    []byte
	compressed bool
}

func NewWIFKey(prefix []byte, privKey []byte, compressed bool) *WIFKey {
	return &WIFKey{
		prefix:     prefix,
		privKey:    privKey,
		compressed: compressed,
	}
}

func DecodeWIFKey(key string, prefixLen uint8) (*WIFKey, error) {
	var (
		normalLen     = int(prefixLen + 32 + CheckSumLen)
		compressedLen = normalLen + 1
	)

	n := len(base58.StdEncoding.Decode(key))
	switch n {
	case normalLen, compressedLen:
		compressed := (n == compressedLen)
		prefix, privKey := DeBase58Check(key, prefixLen, compressed)
		return NewWIFKey(prefix, privKey, compressed), nil
	default:
		return nil, fmt.Errorf("invalid WIF key format")
	}
}

func (k *WIFKey) Prefix() []byte {
	return k.prefix
}

func (k *WIFKey) PrivateKey() []byte {
	return k.privKey
}

func (k *WIFKey) Compressed() bool {
	return k.compressed
}
func (k *WIFKey) Data() []byte {
	return []byte(k.String())
}

func (k *WIFKey) String() string {
	return Base58Check(k.privKey, k.prefix, k.compressed)
}
