package gtrx

import (
	"encoding/hex"

	"github.com/sasaxie/go-client-api/common/base58"
)

func AddressToHex(address string) (string, error) {
	if _, err := base58.Decode(address); err != nil {
		return "", err
	}

	return hex.EncodeToString(base58.DecodeCheck(address)), nil
}
