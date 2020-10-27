package v2

import (
	"bytes"
	"fmt"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

type EncPrivateKey struct {
	encPrivKey []byte
	pubKeyData []byte
}

func NewEncPrivateKey(encPrivKeyData, pubKeyData []byte) *EncPrivateKey {
	return &EncPrivateKey{
		encPrivKey: encPrivKeyData,
		pubKeyData: pubKeyData,
	}
}

func (k *EncPrivateKey) KeyPair(password string, salt, iv, mac []byte, class string) (keypair.KeyPair, error) {
	encryptor := util.MakeEncryptor(password, salt, iv)
	if !bytes.Equal(encryptor.Mac(), mac) {
		return nil, fmt.Errorf("incorrect password")
	}

	privateKey, err := encryptor.Decrypt(k.encPrivKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key failed, %v", err)
	}

	kp, err := keypair.Build(class, privateKey)
	if err != nil {
		return nil, fmt.Errorf("build keypair failed, %v", err)
	}

	return kp, nil
}

func (k *EncPrivateKey) PublicKey(password string, salt, iv, mac []byte, class string) (keypair.PublicKey, error) {
	builder, ok := keypair.FindBuilder(class)
	if !ok {
		return nil, fmt.Errorf("can't find keypair of %s", class)
	}

	kp := builder.Build()
	ap := kp.AddressProvider()
	if ap != nil {
		pubKey, err := keypair.CreatePublicKey(kp, k.pubKeyData)
		if err != nil {
			return nil, err
		}

		return pubKey, nil
	}

	kp, err := k.KeyPair(password, salt, iv, mac, class)
	if err != nil {
		return nil, err
	}

	return kp.(keypair.PublicKey), nil
}
