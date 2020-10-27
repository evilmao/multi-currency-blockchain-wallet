package util

import (
	"encoding/hex"

	"upex-wallet/wallet-tools/base/crypto"
)

// HashString32 generates a hash160 hex string with length of 32.
func HashString32(datas ...[]byte) string {
	var buf []byte
	for _, data := range datas {
		buf = append(buf, data...)
	}
	return hex.EncodeToString(crypto.Hash160(buf))[:32]
}
