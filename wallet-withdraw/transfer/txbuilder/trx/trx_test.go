package trx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/trx/gtrx"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

type testParams struct {
	*txbuilder.TestParams
	assetID int
}

func newTest(p *txbuilder.TestParams, assetID int) *testParams {
	return &testParams{
		TestParams: p,
		assetID:    assetID,
	}
}

var (
	baseTest = txbuilder.NewTestParams("TRX", "TRX",
		"http://127.0.0.1:6022",
		"D95611A9AF2A2A45359106222ED1AFED48853D9A44DEFF8DC7913F5CBA727366",
		"TRCFV4ZqbBciamdmdTxsYB6Ug6pgRynYbw",
		1.5)
	trxTest = newTest(baseTest, 0)
	bttTest = newTest(baseTest, 1000003)
	test    = bttTest

	testTRC20 = txbuilder.NewTestParams("USDT_TRON", "TRX",
		"http://127.0.0.1:6022",
		"D95611A9AF2A2A45359106222ED1AFED48853D9A44DEFF8DC7913F5CBA727366",
		"TRCFV4ZqbBciamdmdTxsYB6Ug6pgRynYbw",
		1.7)
	contractAddress = "TQwzAeUxxk5ofkf1wvb9FC7RtxBn2Z4vF6"
)

func TestGetAccount(t *testing.T) {
	test.Init()

	rpcClient := gtrx.NewClient(test.RPCURL())

	address := test.FromAddress()
	data, err := rpcClient.GetAccount(address)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))

	balance, _ := jsonparser.GetInt(data, "balance")
	bAmount := decimal.New(balance, -gtrx.Precision)
	fmt.Printf("balance: %s TRX\n", bAmount)

	data, err = rpcClient.GetAccountNet(address)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println()
	fmt.Println(string(data))
}

func TestSendTx(t *testing.T) {
	err := test.Init()
	if err != nil {
		t.Fatal(err)
	}

	rpcClient := gtrx.NewClient(test.RPCURL())

	rawAmount := uint64(decimal.NewFromFloat(test.Amount()).Mul(decimal.New(gtrx.TRX, 0)).IntPart())

	tx, err := rpcClient.CreateTransaction(test.FromAddress(), test.ToAddress(), rawAmount, test.assetID)
	if err != nil {
		t.Fatal(err)
	}

	sigDigest, err := tx.UpdateTimestamp()
	if err != nil {
		t.Fatal(err)
	}

	sig, err := test.Sign(sigDigest)
	if err != nil {
		t.Fatal(err)
	}

	tx.Signature = append(tx.Signature, hex.EncodeToString(sig))

	result, err := rpcClient.BroadcastTransaction(tx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(tx.TxID, string(result))
}

func TestSendTRC20(t *testing.T) {
	err := testTRC20.Init()
	if err != nil {
		t.Fatal(err)
	}
	rpcClient := gtrx.NewClient(testTRC20.RPCURL())
	rawAmount := decimal.NewFromFloat(testTRC20.Amount())
	transferInfo, err := gtrx.CreateTrc20TransferReq(testTRC20.FromAddress(), testTRC20.ToAddress(), contractAddress, rawAmount, 6)

	if err != nil {
		t.Fatal(err)
	}
	txInfoJson, err := json.Marshal(transferInfo)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(txInfoJson))

	transferResJson, err := rpcClient.TriggerSmartContract(string(txInfoJson))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(transferResJson))
	transferRes := &gtrx.TransferResult{}
	err = json.Unmarshal(transferResJson, transferRes)

	if err != nil {
		t.Fatal(err)
	}
	if !transferRes.Result.Result {
		t.Fatal(err)
	}
	tx := transferRes.Transaction

	digest, err := tx.UpdateTimestamp()
	signature, err := testTRC20.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}

	tx.Signature = append(tx.Signature, hex.EncodeToString(signature))
	//broadcast
	txJson, err := json.Marshal(tx)
	txStr := string(txJson)
	fmt.Println(txStr)

	broadCastRes, err := rpcClient.BroadcastTransaction(&tx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("broadCastRes:", string(broadCastRes))
}

func TestGetTx(t *testing.T) {
	rpcClient := gtrx.NewClient(test.RPCURL())

	txHash := "054bc796a474c5556c80b040ac0c37362fd207793c0cfbf9c12799d56dcf27a5"
	data, err := rpcClient.GetTransactionByID(txHash)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))

	data, err = rpcClient.GetTransactionInfoByID(txHash)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}
