package generator

import (
	"fmt"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
)

type Convert struct {
	sourceGenerator keypair.Generator
	targetClass     string
	builder         keypair.Builder
}

func NewConvert(sourceGenerator keypair.Generator, targetClass string) *Convert {
	return &Convert{
		sourceGenerator: sourceGenerator,
		targetClass:     targetClass,
	}
}

func (g *Convert) Init() error {
	var ok bool
	g.builder, ok = keypair.FindBuilder(g.targetClass)
	if !ok {
		return fmt.Errorf("invalid keypair class: %s", g.targetClass)
	}
	return nil
}

func (g *Convert) Generate(idx int) (keypair.KeyPair, error) {
	kp, err := g.sourceGenerator.Generate(idx)
	if err != nil {
		return nil, err
	}

	newKP := g.builder.Build()
	err = newKP.SetPrivateKey(kp.PrivateKey())
	if err != nil {
		return nil, err
	}

	return newKP, nil
}
