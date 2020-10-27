package keypair

type Generator interface {
	Init() error
	Generate(idx int) (KeyPair, error)
}
