package signer

import (
	"upex-wallet/wallet-tools/base/crypto/key"
)

type Class string

type Signer interface {
	Class() Class

	Sign(k key.Key, data []byte) ([]byte, error)
	Verify(k key.Key, sigData []byte, data []byte) (bool, error)
}
