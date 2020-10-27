package generator

import "upex-wallet/wallet-tools/cmd/genwallet/keypair"

type Random struct {
	builderClass string
}

func NewRandom(builderClass string) *Random {
	return &Random{
		builderClass: builderClass,
	}
}

func (g *Random) Init() error { return nil }

// 生成随机密钥对
func (g *Random) Generate(idx int) (keypair.KeyPair, error) {
	return keypair.Random(g.builderClass)
}
