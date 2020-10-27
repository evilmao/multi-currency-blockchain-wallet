package key

type Class string

type Key interface {
	Class() Class

	Random() error
	SetPrivateKey(privKey []byte) error

	PrivateKey() []byte
	PublicKey() []byte
	PublicKeyUncompressed() []byte
}
