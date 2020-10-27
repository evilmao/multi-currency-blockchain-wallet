package geth

import (
	"fmt"
	"testing"
)

var (
	ioncSandboxClient = NewClient("http://127.0.0.1:6085")

	ethSandboxClient = NewClient("http://127.0.0.1:6016")
	ethPrdClient     = NewClient("http://127.0.0.1:5500")

	etcSandboxClient = NewClient("http://127.0.0.1:6017")
	etcPrdClient     = NewClient("http://127.0.0.1:5501")
)

func TestClient_GetBlockByNumber(t *testing.T) {
	block, err := ioncSandboxClient.GetBlockByNumber(6)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block))
}

func TestClient_GetBlockByHash(t *testing.T) {
	hash := "0xc7a30f38bab00900d5147c877b9ca579e62baf180fbf95cb5cbb4909c6a881a0"
	block, err := ioncSandboxClient.GetBlockByHash(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block))
}

func TestClient_GetTransactionByHash(t *testing.T) {
	//hash := "0x2c662a1af4d0e231c0145c32594fae59eec76c719a76421da2152165df102bda"
	hash := "0x847aae7a6c6aee26a9f2e24912862be92106b667f5a27f692c75cc3aa14c225c"
	tx, err := ioncSandboxClient.GetTransactionByHash(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx))
}

func TestClient_GetTransactionReceipt(t *testing.T) {
	hash := "0x309680da9e980e87f46a99d5d6106da4e05f9ffeb19d4a44eb5d205cbdc2ea08"
	res, err := ioncSandboxClient.GetTransactionReceipt(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(res))

}
