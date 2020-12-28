package gtrx

import (
	"fmt"
	"testing"
)

var (
	sandboxSolidityRPC = NewClient("http://127.0.0.1:6027") // solidity node
	mainnetRPC         = NewClient("http://127.0.0.1:6025") // solidity node
)

func TestClient_GetSolidityBlockByNum(t *testing.T) {
	var num uint64
	num = 6700849
	block, err := sandboxSolidityRPC.GetSolidityBlockByNum(num)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block))
}

func TestClient_GetSolidityCurrentBlock(t *testing.T) {
	block, err := sandboxSolidityRPC.GetSolidityCurrentBlock()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block))

}

func TestClient_GetSolidityTransactionById(t *testing.T) {
	tx, err := mainnetRPC.GetSolidityTransactionById("d390fbd745cd6ffafd0719d76e769cc2a967246457fcf333d2e982935acaf879")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx))
}

func TestClient_GetTransaction(t *testing.T) {
	txid := "cb718332957bcac1570e369c887e2e71f36a937c4baae28b7820a336a0abd0d7"
	tx, err := sandboxSolidityRPC.GetTransactionInfoById(txid) //可以获取tx所在高度
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx))
	tx1, err := sandboxSolidityRPC.GetSolidityTransactionById(txid) //包含交易明细，但是没有所在高度
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx1))
}
