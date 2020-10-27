package eth

import (
	"fmt"
	"math/big"
	"testing"

	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/rpc/eth/geth"
)

var (
	cfg        = &config.Config{RPCURL: "http://127.0.0.1:6085", Currency: "IONC"}
	sandboxRPC = &RPC{cfg, geth.NewClient(cfg.RPCURL)}
)

func TestRPC_GetBlockByHeight(t *testing.T) {
	block, err := sandboxRPC.GetBlockByHeight(21)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(block.([]byte)))
}

func TestRPC_GetLastBlockHeight(t *testing.T) {
	lastBlockHeight, err := sandboxRPC.GetLastBlockHeight()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(lastBlockHeight)
}

func TestRPC_GetTx(t *testing.T) {
	hash := "0x55b004a3363f9ec3808f64c90aa720a4f7168bd10f51a773cbdaf7c28967610c"
	tx, err := sandboxRPC.GetTx(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(tx.([]byte)))
}

func TestRPC_GetTxConfirmations(t *testing.T) {
	hash := "0x55b004a3363f9ec3808f64c90aa720a4f7168bd10f51a773cbdaf7c28967610c"
	c, err := sandboxRPC.GetTxConfirmations(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(c)
}

func TestRPC_GetBlockHashByHeight(t *testing.T) {
	blockHash, err := sandboxRPC.GetBlockHashByHeight(11)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(blockHash)
}

func TestLen(t *testing.T) {
	s := "0000000000000000000000006a8b1e24967edeab380ca44f3413e0e865c0a2090000000000000000000000000000000000000000000000000000041568dae400"
	fmt.Println(len(s))

	a := big.NewInt(5015187)
	fmt.Println(a.Text(16))

	h := "4c88e2"
	a.SetString(h, 16)
	fmt.Println(a)
}
