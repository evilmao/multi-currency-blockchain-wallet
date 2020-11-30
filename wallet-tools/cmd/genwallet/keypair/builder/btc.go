package builder

import (
	"upex-wallet/wallet-tools/base/crypto/addrprovider"
	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/crypto/signer"
)

func init() {
	registerBTCLike(map[string][]byte{
		"BTC":            []byte{0},
		"BCH":            []byte{0},
		"BSV":            []byte{0},
		"LTC":            []byte{0x30},
		"DASH":           []byte{0x4c},
		"ZEC":            []byte{0x1c, 0xb8},
		"ETP":            []byte{0x32},
		"ETPtest":        []byte{0x7f},
		"DOGE":           []byte{0x1e},
		"BCX":            []byte{0x4b},
		"MONA":           []byte{0x32},
		"VIPS":           []byte{0x46},
		"QTUM":           []byte{0x3a},
		"FAB":            []byte{0},
		"KMD":            []byte{0x3c},
		"ABBCdeprecated": []byte{0x17},
	})
}

func registerBTCLike(prefixs map[string][]byte) {
	for c, prefix := range prefixs {
		registerCommon(c, key.NewSecp256k1, signer.NewSecp256k1(), addrprovider.NewBTC(prefix))
	}
}
