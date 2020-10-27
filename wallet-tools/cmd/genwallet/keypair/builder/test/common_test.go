package test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"
)

var (
	builders = keypair.AllBuilderClasses()
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

func testFramework(t *testing.T, do func(string) bool) {
	var failedN int
	for _, b := range builders {
		ok := do(b)
		if !ok {
			failedN++
			continue
		}

		fmt.Println(b, "pass.")
	}

	if failedN > 0 {
		t.Fatal("failed")
	}
}

func testInterfaceN(class string) bool {
	for i := 0; i < 300; i++ {
		kp1, err := keypair.Random(class)
		if err != nil {
			failed("%s, random failed, %v", class, err)
			return false
		}

		kp2, err := keypair.Build(class, kp1.PrivateKey())
		if err != nil {
			failed("%s, set private key failed, %v", class, err)
			return false
		}

		if !assert(bytes.Equal(kp1.PrivateKey(), kp2.PrivateKey()), fmt.Sprintf("%s, private key not equal", class)) {
			return false
		}

		if !assert(bytes.Equal(kp1.PublicKey(), kp2.PublicKey()), fmt.Sprintf("%s, public key not equal", class)) {
			return false
		}

		if !assert(bytes.Equal(kp1.Address(), kp2.Address()), fmt.Sprintf("%s, address not equal", class)) {
			return false
		}

		if !assert(kp1.AddressString() == kp2.AddressString(), fmt.Sprintf("%s, address string not equal", class)) {
			return false
		}
	}

	return true
}

// TestInterface tests keypair interface, except of Sign & Verify.
func TestInterface(t *testing.T) {
	testFramework(t, testInterfaceN)
}

func testCryptographyN(class string) bool {
	for i := 0; i < 300; i++ {
		standard, err := keypair.Random(class)
		if err != nil {
			failed("%s, random failed, %v", class, err)
			return false
		}

		if standard.Cryptography() == keypair.InvalidCryptoClass {
			failed("%s, invalid cryptography name", class)
			return false
		}

		for _, b := range builders {
			builder, ok := keypair.FindBuilder(b)
			if !ok {
				failed("%s, can't find builder", b)
				return false
			}

			kp := builder.Build()
			if kp.Class() == class {
				continue
			}

			if kp.Cryptography() != standard.Cryptography() {
				continue
			}

			err = kp.SetPrivateKey(standard.PrivateKey())
			if err != nil {
				failed("%s, set standard (%s) private key failed, %v", b, standard.Class(), err)
				return false
			}

			if !assert(bytes.Equal(kp.PrivateKey(), standard.PrivateKey()), fmt.Sprintf("%s, private key not equal to standard (%s)", b, standard.Class())) {
				return false
			}

			if !assert(bytes.Equal(kp.PublicKey(), standard.PublicKey()), fmt.Sprintf("%s, public key not equal to standard (%s)", b, standard.Class())) {
				return false
			}
		}
	}

	return true
}

// TestCryptography tests keypair cryptography class, except of Sign & Verify.
func TestCryptography(t *testing.T) {
	testFramework(t, testCryptographyN)
}

func testSignN(class string) bool {
	for i := 0; i < 100; i++ {
		standard, err := keypair.Random(class)
		if err != nil {
			failed("%s, random keypair failed, %v", class, err)
			return false
		}

		data := make([]byte, 32)
		io.ReadFull(rand.Reader, data)
		sigData, err := standard.Sign(data)
		if err != nil {
			failed("%s, sign failed, %v", class, err)
			return false
		}

		ok, err := standard.Verify(sigData, data)
		if err != nil {
			failed("%s, verify failed, %v", class, err)
			return false
		}

		if !ok {
			failed("%s, verify failed", class)
			return false
		}

		for _, b := range builders {
			builder, ok := keypair.FindBuilder(b)
			if !ok {
				failed("%s, can't find builder", b)
				return false
			}

			kp := builder.Build()
			if kp.Class() == class {
				continue
			}

			if kp.Cryptography() != standard.Cryptography() {
				continue
			}

			err = kp.SetPrivateKey(standard.PrivateKey())
			if err != nil {
				failed("%s, set standard (%s) private key failed, %v", b, standard.Class(), err)
				return false
			}

			ok, err = kp.Verify(sigData, data)
			if err != nil {
				failed("%s, verify failed, standard (%s), %v", b, standard.Class(), err)
				return false
			}

			if !ok {
				failed("%s, verify failed, standard (%s)", b, standard.Class())
				return false
			}
		}
	}

	return true
}

// TestSign tests keypair Sign & Verify.
func TestSign(t *testing.T) {
	testFramework(t, testSignN)
}

func printKeyPair(kp keypair.KeyPair) {
	fmt.Println("class:", kp.Class())
	fmt.Println("cryptography:", kp.Cryptography())
	fmt.Printf("private key:\n%s\n", hex.EncodeToString(kp.PrivateKey()))
	fmt.Printf("public key:\n%s\n", hex.EncodeToString(kp.PublicKey()))
	fmt.Printf("address:\n%s\n", kp.AddressString())
	if kp, ok := kp.(keypair.WithExtData); ok {
		for k, v := range kp.ExtData() {
			fmt.Printf("%s:\n%s\n", k, v.String())
		}
	}
	fmt.Println()
}

func TestPrintKeyPair(t *testing.T) {
	kp, err := keypair.Random("FAB")
	if err != nil {
		t.Fatal(err)
	}

	printKeyPair(kp)
}

func TestWIFKey(t *testing.T) {
	const (
		class     = "FAB"
		wifStr    = "L5acjcWyEiuqHAPNpJ5czA2KFoJvUNQPd8hQMRGSBWnaVy5JgwEv"
		prefixLen = 1
		address   = "1AqHKyr5zufhEpquTYuQqt34HSWsJFZ9Au"
	)

	wifKey, err := crypto.DecodeWIFKey(wifStr, prefixLen)
	if err != nil {
		t.Fatal(err)
	}

	kp, err := keypair.Build(class, wifKey.PrivateKey())
	if err != nil {
		t.Fatal(err)
	}

	if kp.AddressString() != address {
		printKeyPair(kp)
		t.Fatal("failed")
	}
}
