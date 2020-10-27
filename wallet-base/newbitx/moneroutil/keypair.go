package moneroutil

// KeyPair consists of a ed25519 256-bit SpendKey and a ViewKey.
// In order to make the mnemonic seeds work, ViewKey is actually derived from
// SpendKey with Keccak256, also leading to SpendKey serving as the seed.
type KeyPair struct {
	SpendKey, ViewKey *Key
}

func (k *KeyPair) PublicSpendKey() *Key {
	return k.SpendKey.PubKey()
}

func (k *KeyPair) PublicViewKey() *Key {
	return k.ViewKey.PubKey()
}

// HalfToFull transfroms a Key construsted by HalfKeyFromSeed to a full key.
// Basically, it triggers the derivation of its view key from spend key.
func (k *KeyPair) HalfToFull() {
	hash := Keccak256(k.SpendKey[:])
	k.ViewKey = (*Key)(&hash)
	ScReduce32(k.ViewKey)
}

type Network []byte

var (
	MainNetwork = Network{0x12}
	TestNetwork = Network{0x35}
)

// Address encodes network, public spend key and view key in base58 format.
func (k *KeyPair) Address(network Network) string {
	spendPub := k.PublicSpendKey()
	viewPub := k.PublicViewKey()

	hash := Keccak256(network, spendPub[:], viewPub[:])
	address := EncodeMoneroBase58(network, spendPub[:], viewPub[:], hash[:4])

	return address
}

// KeyPairFromSeed construsts a Key with given seed.
func KeyPairFromSeed(seed [32]byte) *KeyPair {
	k := &KeyPair{new(Key), new(Key)}

	copy(k.SpendKey[:], seed[:])
	ScReduce32(k.SpendKey)

	k.HalfToFull()
	return k
}
