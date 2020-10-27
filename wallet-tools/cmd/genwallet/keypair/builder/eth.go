package builder

import (
	"upex-wallet/wallet-tools/base/crypto/addrprovider"
	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/crypto/signer"
)

func init() {
	registerETHLike("ETH", "ETC", "WAN", "DCAR", "SMT", "FT", "IONC")
}

func registerETHLike(classes ...string) {
	registerBatch(classes, key.NewSecp256k1, signer.NewSecp256k1Recoverable(false), addrprovider.NewETH())
}
