package currency_test

import (
	"fmt"
	"testing"

	"upex-wallet/wallet-base/blockchain"
	"upex-wallet/wallet-base/currency"
)

func TestSandboxCurrency(t *testing.T) {
	err := currency.Init("https://broker-www-sandbox.fcoin.com", "ak8XfeRpD", "pk9MnCpDFg", "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(currency.MinAmount("usdt"))
	fmt.Println(blockchain.CorrectName("usdt", "usdt"))
	fmt.Println(blockchain.CorrectName("usdt", "eth"))
	fmt.Println(blockchain.CorrectName("usdt", "btc"))
	fmt.Println(blockchain.CorrectName("gas", "neo"))
	fmt.Println(blockchain.CorrectName("mxm", "eth"))
	fmt.Println(blockchain.CorrectName("ft", "ft"))
}

func TestPrdCurrency(t *testing.T) {
	err := currency.Init("https://broker-api.fcoin.com", "p5JbYWB6dU6JeUyN", "titN@hS*AfGIZynP", "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(currency.MinAmount("usdt"))
	fmt.Println(blockchain.CorrectName("usdt", "usdt"))
	fmt.Println(blockchain.CorrectName("usdt", "eth"))
	fmt.Println(blockchain.CorrectName("usdt", "btc"))
	fmt.Println(blockchain.CorrectName("gas", "neo"))
	fmt.Println(blockchain.CorrectName("mxm", "eth"))
	fmt.Println(blockchain.CorrectName("ft", "ft"))
}
