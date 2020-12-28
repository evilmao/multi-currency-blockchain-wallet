package tron

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"upex-wallet/wallet-deposit/rpc/tron/gtrx"
	"upex-wallet/wallet-tools/base/crypto"

	"github.com/buger/jsonparser"
)

var (
	sandboxSolidityRPC = gtrx.NewClient("http://127.0.0.1:6027") // solidity node
	mainnetSolidityRPC = gtrx.NewClient("http://127.0.0.1:6025") // solidity node
	mainnetRPC         = gtrx.NewClient("http://127.0.0.1:5516")
	testRPC            = &RPC{client: sandboxSolidityRPC}
)

func TestRPC_GetBlockByHeight(t *testing.T) {
	num := uint64(4245518)
	block, err := testRPC.GetBlockByHeight(num)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block.([]byte)))
}

func TestRPC_GetLastBlockHeight(t *testing.T) {
	height, err := testRPC.GetLastBlockHeight()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(height)
}

func TestRPC_GetTx(t *testing.T) {
	tx, err := testRPC.GetTx("919a8ed53496f61834c8b517bdab2b08a3fc802cf594532eace6014e70d38a02")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx.([]byte)))
}

func TestClientGetBlockByNum(t *testing.T) {
	num := uint64(4245518)
	block, _ := mainnetSolidityRPC.GetSolidityBlockByNum(num)
	number, _ := jsonparser.GetInt(block, "block_header", "raw_data", "number")
	t.Log("number:", number)
	fmt.Println(string(block))
}

func TestGetTxConfirmations(t *testing.T) {
	rpc := &RPC{
		client: sandboxSolidityRPC,
	}
	confirm, err := rpc.GetTxConfirmations("919a8ed53496f61834c8b517bdab2b08a3fc802cf594532eace6014e70d38a02")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("confirm:", confirm)
}

func TestGetAssetIssue(t *testing.T) {
	name := "BitTorrent"
	name = hex.EncodeToString([]byte(name))
	fmt.Println(name) //426974546f7272656e74
	data, err := mainnetRPC.GetAssetIssueListByName(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))

	id := "1002000"
	id2 := hex.EncodeToString([]byte(id))
	fmt.Println(name, id, id2)

	id3, _ := hex.DecodeString(id2)
	fmt.Println(string(id3))

	addr, _ := hex.DecodeString("413aff82b4a8fc7d78df08301ff65af5b75bae72fb")
	addrstr := crypto.Base58Check(addr, nil, false)
	fmt.Println(addrstr)
}

func TestParseAddressAmount(t *testing.T) {
	data := `a9059cbb00000000000000000000004150b0c2b3bcad53eb45b57c4e5df8a9890d002cc8000000000000000000000000000000000000000000000000000000000010c8e0`
	contractAddress := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	addr, amt, err := parseAddressAmount(data, contractAddress)
	if err != nil {
		t.Log(err)
	}
	fmt.Println(addr)
	fmt.Println(amt)
}

func TestBase58Check(t *testing.T) {
	addr := "415a523b449890854c8fc460ab602df9f31fe4293f"
	base58addr := "TJCnKsPa7y5okkXvQAidZBzqx3QyQ6sxMW"
	buf, _ := hex.DecodeString(addr)
	base58addrcheck := crypto.Base58Check(buf, nil, false)
	fmt.Println(base58addr == base58addrcheck)
	_, addrBuf := crypto.DeBase58Check(base58addrcheck, 0, false)
	fmt.Println(hex.EncodeToString(addrBuf))
	amt := big.NewInt(100)
	fmt.Println(hex.EncodeToString(amt.Bytes()))

}
