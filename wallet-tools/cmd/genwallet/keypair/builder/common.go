package builder

import (
	"upex-wallet/wallet-tools/base/crypto/addrprovider"
	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/crypto/signer"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
)

func init() {
	registerSecp256k1Recoverable(map[string]addrprovider.AddrProvider{
		"PTNdeprecated": addrprovider.NewPTNdeprecated(),
		"TRX":           addrprovider.NewTRX(),
	})

}

func registerCommon(class string, kc func() key.Key, s signer.Signer, ap addrprovider.AddrProvider) {
	keypair.RegisterBuilder(NewCommon(class, kc, s, ap))
}

func registerBatch(classes []string, kc func() key.Key, s signer.Signer, ap addrprovider.AddrProvider) {
	for _, c := range classes {
		registerCommon(c, kc, s, ap)
	}
}

func registerEd25519(aps map[string]addrprovider.AddrProvider) {
	for c, ap := range aps {
		registerCommon(c, key.NewEd25519, signer.NewEd25519(), ap)
	}
}

func registerSecp256k1Normal(aps map[string]addrprovider.AddrProvider) {
	for c, ap := range aps {
		registerCommon(c, key.NewSecp256k1, signer.NewSecp256k1(), ap)
	}
}

func registerSecp256k1Recoverable(aps map[string]addrprovider.AddrProvider) {
	for c, ap := range aps {
		registerCommon(c, key.NewSecp256k1, signer.NewSecp256k1Recoverable(false), ap)
	}
}

func registerSecp256k1Canonical(aps map[string]addrprovider.AddrProvider) {
	for c, ap := range aps {
		registerCommon(c, key.NewSecp256k1, signer.NewSecp256k1Canonical(true), ap)
	}
}

func registerSecp256k1Ecschnorr(aps map[string]addrprovider.AddrProvider) {
	for c, ap := range aps {
		registerCommon(c, key.NewSecp256k1, signer.NewEcschnorr(), ap)
	}
}

type Common struct {
	class string

	kc func() key.Key
	key.Key
	s  signer.Signer
	ap addrprovider.AddrProvider
}

func NewCommon(class string, kc func() key.Key, s signer.Signer, ap addrprovider.AddrProvider) *Common {
	return &Common{
		class: class,
		kc:    kc,
		Key:   kc(),
		s:     s,
		ap:    ap,
	}
}

func (kp *Common) Build() keypair.KeyPair {
	return NewCommon(kp.class, kp.kc, kp.s, kp.ap)
}

func (kp *Common) Class() string {
	return kp.class
}

func (kp *Common) Cryptography() keypair.CryptoClass {
	return keypair.CryptoClass(kp.s.Class())
}

func (kp *Common) Address() []byte {
	return kp.ap.Address(kp.Key)
}

func (kp *Common) AddressString() string {
	return kp.ap.AddressString(kp.Key)
}

func (kp *Common) AddressProvider() addrprovider.AddrProvider {
	return kp.ap
}

func (kp *Common) Sign(data []byte) ([]byte, error) {
	return kp.s.Sign(kp.Key, data)
}

func (kp *Common) Verify(sigData []byte, data []byte) (bool, error) {
	return kp.s.Verify(kp.Key, sigData, data)
}
