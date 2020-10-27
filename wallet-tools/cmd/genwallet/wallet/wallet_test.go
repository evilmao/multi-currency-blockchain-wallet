package wallet

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"
	walletv1 "upex-wallet/wallet-tools/cmd/genwallet/wallet/v1"
	walletv2 "upex-wallet/wallet-tools/cmd/genwallet/wallet/v2"
)

type fakeGenerator struct{}

func (g *fakeGenerator) Init() error { return nil }

func (g *fakeGenerator) Generate(int) (keypair.KeyPair, error) {
	return keypair.Random("BTC")
}

const (
	password = "123s456"
	dataPath = "./"
	fileName = "fake-wallet.dat"
)

func TestWalletV1(t *testing.T) {
	w1 := walletv1.New(password, dataPath, fileName, &fakeGenerator{}, nil)
	err := w1.Generate(100)
	if err != nil {
		t.Fatal(err)
	}

	err = w1.Store()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(dataPath + fileName)

	w2 := walletv1.New(password, dataPath, fileName, nil, nil)
	err = w2.Load()
	if err != nil {
		t.Fatal(err)
	}

	if w2.Len() != w1.Len() {
		t.Fatal(fmt.Sprintf("wallet len not equal, %d vs %d", w2.Len(), w1.Len()))
	}

	err = w2.Foreach(func(idx int, kp2 keypair.KeyPair) (bool, error) {
		kp1, _ := w1.KeyPairAtIndex(idx)
		if !bytes.Equal(kp2.PrivateKey(), kp1.PrivateKey()) {
			return false, fmt.Errorf("keypair not equal at index %d", idx)
		}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	w3 := walletv1.New(password+"1", dataPath, fileName, nil, nil)
	err = w3.Load()
	if err == nil {
		t.Fatal(fmt.Sprintf("incorrect password must don't work"))
	}
}

func TestWalletV2LoadV1(t *testing.T) {
	w1 := walletv1.New(password, dataPath, fileName, &fakeGenerator{}, nil)
	err := w1.Generate(100)
	if err != nil {
		t.Fatal(err)
	}

	err = w1.Store()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(dataPath + fileName)

	w2 := walletv2.New(dataPath, fileName, fileName, nil, nil)
	err = w2.LoadV1(password)
	if err != nil {
		t.Fatal(err)
	}

	if w2.Len() != w1.Len() {
		t.Fatal(fmt.Sprintf("wallet len not equal, %d vs %d", w2.Len(), w1.Len()))
	}

	err = w1.Foreach(func(idx int, kp1 keypair.KeyPair) (bool, error) {
		kp2, _ := w2.KeyPairAtIndex(password, idx)
		if !bytes.Equal(kp2.PrivateKey(), kp1.PrivateKey()) {
			return false, fmt.Errorf("keypair not equal at index %d", idx)
		}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	w3 := walletv2.New(dataPath, fileName, fileName, nil, nil)
	err = w3.LoadV1(password + "1")
	if err == nil {
		t.Fatal(fmt.Sprintf("incorrect password must don't work"))
	}
}

func TestWalletV2(t *testing.T) {
	w1 := walletv2.New(dataPath, fileName, fileName, &fakeGenerator{}, nil)
	err := w1.Generate(password, 100)
	if err != nil {
		t.Fatal(err)
	}

	err = w1.Store(password)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(dataPath + fileName)

	w2 := walletv2.New(dataPath, fileName, fileName, nil, nil)
	err = w2.Load()
	if err != nil {
		t.Fatal(err)
	}

	if w2.Len() != w1.Len() {
		t.Fatal(fmt.Sprintf("wallet len not equal, %d vs %d", w2.Len(), w1.Len()))
	}

	for idx := 0; idx < w1.Len(); idx++ {
		kp1, _ := w1.KeyPairAtIndex(password, idx)
		kp2, _ := w2.KeyPairAtIndex(password, idx)
		if !bytes.Equal(kp2.PrivateKey(), kp1.PrivateKey()) {
			t.Fatal(fmt.Errorf("keypair not equal at index %d", idx))
		}
	}

	w3 := walletv2.New(dataPath, fileName, fileName, nil, nil)
	err = w3.LoadV1(password + "1")
	if err == nil {
		t.Fatal(fmt.Sprintf("incorrect password must don't work"))
	}
}
