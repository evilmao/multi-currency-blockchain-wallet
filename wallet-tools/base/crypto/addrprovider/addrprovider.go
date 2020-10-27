package addrprovider

type Class string

type Key interface {
	PublicKey() []byte
	PublicKeyUncompressed() []byte
}

type AddrProvider interface {
	Class() Class

	Address(k Key) []byte
	AddressString(k Key) string
}
