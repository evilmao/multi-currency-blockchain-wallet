package crypto

import (
	"upex-wallet/wallet-tools/base/libs/bech32"

	"github.com/pkg/errors"
)

// Bech32ConvertAndEncode converts from 8-bit byte slice to 5-bit byte slice and then to bech32.
func Bech32ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := Bech32ConvertBits8To5(data)
	if err != nil {
		return "", errors.Wrap(err, "encoding bech32 failed")
	}
	return bech32.Encode(hrp, converted)

}

// Bech32DecodeAndConvert decodes a bech32 encoded string and converts from 5-bit byte slice to 8-bit byte slice.
func Bech32DecodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.Decode(bech)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	converted, err := Bech32ConvertBits5To8(data)
	if err != nil {
		return "", nil, errors.Wrap(err, "decoding bech32 failed")
	}
	return hrp, converted, nil
}

// Bech32ConvertBits8To5 converts from 8-bit byte slice to 5-bit byte slice.
func Bech32ConvertBits8To5(data []byte) ([]byte, error) {
	return bech32.ConvertBits(data, 8, 5, true)
}

// Bech32ConvertBits5To8 converts from 5-bit byte slice to 8-bit byte slice.
func Bech32ConvertBits5To8(data []byte) ([]byte, error) {
	return bech32.ConvertBits(data, 5, 8, false)
}
