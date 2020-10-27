package crypto_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"upex-wallet/wallet-tools/base/crypto/key"
	"upex-wallet/wallet-tools/base/crypto/signer"
)

func failed(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
}

func assert(ok bool, msg string) bool {
	if !ok {
		failed(msg)
	}
	return ok
}

type Key struct {
	key.Key
	signer.Signer
}

func (k *Key) Class() string {
	return fmt.Sprintf("%s|%s", k.Key.Class(), k.Signer.Class())
}

func testKeyN(k *Key) bool {
	data := make([]byte, 32)
	for i := 0; i < 3000; i++ {
		err := k.Random()
		if err != nil {
			failed("%s, random failed, %v", k.Class(), err)
			return false
		}

		io.ReadFull(rand.Reader, data)
		sigData, err := k.Sign(k.Key, data)
		if err != nil {
			failed("%s, sign failed, %v", k.Class(), err)
			return false
		}

		ok, err := k.Verify(k.Key, sigData, data)
		if err != nil {
			failed("%s, verify failed, %v", k.Class(), err)
			return false
		}

		if !ok {
			failed("%s, verify failed", k.Class())
			return false
		}
	}
	return true
}

func TestKey(t *testing.T) {
	keys := []*Key{
		&Key{key.NewEd25519(), signer.NewEd25519()},
		&Key{key.NewEd25519blk(), signer.NewEd25519blk()},
		&Key{key.NewNISTP256(), signer.NewNISTP256()},
		&Key{key.NewSecp256k1(), signer.NewSecp256k1()},
		&Key{key.NewSecp256k1(), signer.NewSecp256k1Recoverable(false)},
		&Key{key.NewSecp256k1(), signer.NewSecp256k1Recoverable(true)},
		&Key{key.NewSecp256k1(), signer.NewSecp256k1Canonical(false)},
		&Key{key.NewSecp256k1(), signer.NewSecp256k1Canonical(true)},
		&Key{key.NewSecp256k1(), signer.NewEcschnorr()},
	}

	var failedN int
	for _, k := range keys {
		ok := testKeyN(k)
		if !ok {
			failedN++
			continue
		}

		fmt.Println(k.Class(), "pass.")
	}

	if failedN > 0 {
		t.Fatal("failed")
	}
}
