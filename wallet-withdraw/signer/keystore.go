package signer

import (
	"fmt"

	wallet "upex-wallet/wallet-tools/cmd/genwallet/wallet/v2"
)

type KeyStore struct {
	wallets []*wallet.Wallet
}

func NewKeyStore(dataPath string, fileNames []string) *KeyStore {
	wallets := make([]*wallet.Wallet, 0, len(fileNames))
	for _, fileName := range fileNames {
		wallets = append(wallets, wallet.New(dataPath, fileName, "", nil, nil))
	}

	return &KeyStore{
		wallets: wallets,
	}
}

func (k *KeyStore) Load() error {
	for _, w := range k.wallets {
		err := w.Load()
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *KeyStore) Sign(password, pubkey string, hash []byte) ([]byte, error) {
	for _, w := range k.wallets {
		if !w.Contains(pubkey) {
			continue
		}

		kp, err := w.KeypairByPubkey(password, pubkey)
		if err != nil {
			return nil, fmt.Errorf("get keypair by public key failed, %v", err)
		}

		return kp.Sign(hash)
	}

	return nil, fmt.Errorf("can't find keypair for pubkey %s", pubkey)
}
