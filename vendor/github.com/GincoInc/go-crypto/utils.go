package crypto

import (
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
)

// Sha256 ...
func Sha256(data []byte) []byte {
	sha := sha256.New()
	sha.Write(data)

	return sha.Sum(nil)
}

// Compress ...
func Compress(curve elliptic.Curve, x, y *big.Int) []byte {
	return Marshal(curve, x, y, true)
}

// Marshal ...
func Marshal(curve elliptic.Curve, x, y *big.Int, compress bool) []byte {
	byteLen := (curve.Params().BitSize + 7) >> 3

	if compress {
		ret := make([]byte, 1+byteLen)
		if y.Bit(0) == 0 {
			ret[0] = 2
		} else {
			ret[0] = 3
		}
		xBytes := x.Bytes()
		copy(ret[1+byteLen-len(xBytes):], xBytes)
		return ret
	}

	ret := make([]byte, 1+2*byteLen)
	ret[0] = 4 // uncompressed point
	xBytes := x.Bytes()
	copy(ret[1+byteLen-len(xBytes):], xBytes)
	yBytes := y.Bytes()
	copy(ret[1+2*byteLen-len(yBytes):], yBytes)
	return ret
}
