package eos

import (
	"fmt"
	"testing"

	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-deposit/syncer/eos/geos"

	"github.com/buger/jsonparser"
)

/*
https://eos.greymass.com:443
http://eos.greymass.com
http://eu.eosdac.io
http://node.eosflare.io
http://api.eossweden.se
http://api.eostribe.io
http://peer1.eoshuobipool.com:8181
*/

var (
	sandboxRPC = geos.NewClient("http://127.0.0.1:6061")
	mainnetRPC = geos.NewClient("http://127.0.0.1:8888")
)

func TestCurrentBlock(t *testing.T) {
	data, err := sandboxRPC.GetInfo()
	if err != nil {
		t.Fatal(err)
	}

	height, _ := jsonparser.GetInt(data, "head_block_num")
	fmt.Println("height:", height)

	data, err = sandboxRPC.GetBlock(height)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestGetBlock(t *testing.T) {
	var (
		start    = 66670929
		minTxNum int
		maxTxNum int
	)
	util.WithLogTimeCost("get-block-100", func() {
		for i := 0; i < 100; i++ {
			data, _ := mainnetRPC.GetBlock(int64(start - i))
			height, _ := jsonparser.GetInt(data, "block_num")
			var txNum int
			util.JSONParserArrayEach(data, func([]byte, jsonparser.ValueType) error {
				txNum++
				return nil
			}, "transactions")

			if i == 0 {
				minTxNum, maxTxNum = txNum, txNum
			} else {
				if txNum > maxTxNum {
					maxTxNum = txNum
				}

				if txNum < minTxNum {
					minTxNum = txNum
				}
			}
			fmt.Println(i, height, txNum)
		}
	})

	fmt.Println(minTxNum, maxTxNum)
}

func TestGetTransaction(t *testing.T) {
	txHash := "2f7b71384cdafc7654ebe10c219999ae0bc0774ef8da80e0a58f320ca697b592"
	data, err := mainnetRPC.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}
