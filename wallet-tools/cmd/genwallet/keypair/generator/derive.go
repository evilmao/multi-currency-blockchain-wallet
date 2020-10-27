package generator

import (
	"fmt"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
)

type Derive struct {
	builderClass string
	originKP     keypair.KeyPair

	builder   keypair.Builder
	originDKP keypair.DerivableKeyPair
}

func NewDerive(builderClass string) *Derive {
	return &Derive{
		builderClass: builderClass,
	}
}

func (g *Derive) SetOrigin(kp keypair.KeyPair) error {
	if g.originKP != nil {
		return fmt.Errorf("origin has exist")
	}

	g.originKP = kp
	return nil
}

func (g *Derive) Init() error {
	builder, ok := keypair.FindBuilder(g.builderClass)
	if !ok {
		return fmt.Errorf("invalid keypair class: %s", g.builderClass)
	}

	g.builder = builder

	var originAsFirst bool
	if g.originKP == nil {
		kp := g.builder.Build()
		err := kp.Random()
		if err != nil {
			return err
		}

		g.originKP = kp
		originAsFirst = true
	}

	dkp, ok := g.originKP.(keypair.DerivableKeyPair)
	if !ok {
		return fmt.Errorf("keypair class %s is not a DerivableKeyPair", g.originKP.Class())
	}

	if originAsFirst {
		g.originDKP = dkp
	} else {
		next, err := dkp.Derive(1)
		if err != nil {
			return err
		}
		g.originDKP = next
	}

	return nil
}

func (g *Derive) Generate(idx int) (keypair.KeyPair, error) {
	dkp, err := g.originDKP.Derive(idx)
	if err != nil {
		return nil, err
	}

	return dkp, nil
}
