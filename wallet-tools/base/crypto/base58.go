package crypto

import (
	"upex-wallet/wallet-tools/base/libs/base58"
)

func Base58Check(input []byte, prefix []byte, compressed bool) string {
	b := make([]byte, 0, len(prefix)+len(input)+CheckSumLen)
	b = append(b, prefix...)
	b = append(b, input...)
	if compressed {
		b = append(b, 0x01)
	}
	cksum := CheckSum(b)
	b = append(b, cksum[:]...)
	return base58.StdEncoding.Encode(b)
}

func DeBase58Check(s string, prefixLen uint8, compressed bool) ([]byte, []byte) {
	if len(s) == 0 {
		return nil, nil
	}

	b := base58.StdEncoding.Decode(s)

	tailLen := CheckSumLen
	if compressed {
		tailLen++
	}

	if len(b) < int(prefixLen)+tailLen {
		return nil, nil
	}
	return b[:int(prefixLen)], b[int(prefixLen) : len(b)-tailLen]
}

// ReplaceAddressPrefix replaces base58 format address's prefix.
func ReplaceAddressPrefix(addr string, prefixLen uint8, newPrefix []byte) string {
	if len(newPrefix) == 0 {
		return addr
	}
	_, data := DeBase58Check(addr, prefixLen, false)
	return Base58Check(data, newPrefix, false)
}
