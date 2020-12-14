package bitcoin

import (
	"fmt"
	"testing"

	"upex-wallet/wallet-deposit/syncer/bitcoin/gbtc"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

var (
	// sandbox
	sandboxBTC  = gbtc.NewClient("http://111:111@127.0.0.1:6012/")
	sandboxDOGE = gbtc.NewClient("http://111:111@127.0.0.1:6020/")
	sandboxMONA = gbtc.NewClient("http://111:111@127.0.0.1:6029/")
	sandboxVIPS = gbtc.NewClient("http://111:111@127.0.0.1:6030/")
	sandboxQtum = gbtc.NewClient("http://111:111@127.0.0.1:6031/")
	sandboxFAB  = gbtc.NewClient("http://111:111@127.0.0.1:6035/")

	sandboxRPC = sandboxFAB

	// mainnet
	mainnetBTC  = gbtc.NewClient("http://111123232:343483943sdkfjdskfjdksfj@127.0.0.1:5508")
	mainnetBSV  = gbtc.NewClient("http://7ZZ4Dzyby9GcuSwr5p8P8momJ:7ZZ4Dzyby9GcuSwr5p8P8momJ@127.0.0.1:5515")
	mainnetUSDT = gbtc.NewClient("http://7ZZ4Dzyby9GcuSwr5p8P8momJ:G4OCmzzMPLtHWFINTnKkCVJz2@127.0.0.1:5509")
	mainnetMONA = gbtc.NewClient("http://8QbMPYCvZFO9:8QbMPYCvZFO9@127.0.0.1:6013")

	mainnetRPC = mainnetBSV
)

func TestGetCurrentBlock(t *testing.T) {
	hash, err := mainnetRPC.GetBestBlockHash()
	if err != nil {
		t.Fatal(err)
	}

	data, err := mainnetRPC.GetBlockByHash(hash)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestGetBlock(t *testing.T) {
	data, err := mainnetBSV.GetBlockByHeight(593964)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestGetFullBlock(t *testing.T) {
	data, err := mainnetRPC.GetFullBlockByHeight(1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestGetTransaction(t *testing.T) {
	data, err := sandboxFAB.GetRawTransaction("babc017e5b723efe45c73201007b8c2d04ed17adb40251917e6a248447e4ef95")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestGetTransactionConfirmations(t *testing.T) {
	data, err := mainnetRPC.GetRawTransaction("10067abeabcd96a1261bc542b16d686d083308304923d74cb8f3bab4209cc3b9")
	if err != nil {
		t.Fatal(err)
	}

	confirm, err := jsonparser.GetInt(data, "confirmations")
	fmt.Println(confirm, err)
}

func TestFloat64LossPrecision(t *testing.T) {
	const (
		a float64 = 10000
		b float64 = 30000.48187095
	)

	fmt.Println("correct:", decimal.NewFromFloat(a).Add(decimal.NewFromFloat(b)))
	fmt.Println("float64:", a+b)
}

func TestOmniListTxs(t *testing.T) {
	data, err := mainnetUSDT.OmniListBlockTransactions(500000)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestOmniGetTx(t *testing.T) {
	data, err := mainnetUSDT.OmniGetTransaction("8662e51ae3e9a8a050dc958a95ea27015107388ba212e483755121427e7ba2b7")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}
