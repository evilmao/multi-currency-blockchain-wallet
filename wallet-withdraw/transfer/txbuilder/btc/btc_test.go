package btc

import (
	"fmt"
	"testing"

	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc/rpc"

	"github.com/shopspring/decimal"
)

func assert(t *testing.T, ok bool, msg string) {
	if !ok {
		t.Fatal(msg)
	}
}

type BTCTest struct {
	*txbuilder.TestParams
	preTx       string
	preTxOutIdx uint32
	rpcCreator  func(string) gbtc.RPC
}

func NewBTCTest(params *txbuilder.TestParams, preTx string, preTxOutIdx uint32, rpcCreator func(string) gbtc.RPC) *BTCTest {
	return &BTCTest{
		params,
		preTx,
		preTxOutIdx,
		rpcCreator,
	}
}

func (t *BTCTest) PublicKey() []byte {
	return t.TestParams.FromPublicKey()
}

func (t *BTCTest) PreTx() string {
	return t.preTx
}

func (t *BTCTest) PreTxOutIdx() uint32 {
	return t.preTxOutIdx
}

func (t *BTCTest) RPCClient() gbtc.RPC {
	return t.rpcCreator(t.RPCURL())
}

var (
	btcTest = NewBTCTest(
		txbuilder.NewTestParams("BTC", "BTC",
			"http://111:111@127.0.0.1:6012",
			"cb483073d4ff08983099c24fceed646b9faab865b9c5785edf638d3257ca2612",
			"18RbyoRa5RVkzd3oPKhUh8eBuESMh4pRUT",
			0.1),
		"4887b61e6a2e1411d1c2df391f7bb71efc2775a7ebe30502a50231bdc2646114",
		0,
		rpc.NewBTCRPC)

	etpTest = NewBTCTest(
		txbuilder.NewTestParams("ETP", "ETPtest",
			"http://127.0.0.1:6021/rpc/v3/",
			"0f6055b44781882c04d6683d8e11a8282d068ef139a78f9d45e9ba290d1ce25f",
			"tHptS49T9AWopJtNLFXGx2cuSCcWQRuQU5",
			0.1),
		"4d13d26b8347d318d70b00d7e6eb956f9aad57fedd8926ba1be4b65243d760b6",
		0,
		rpc.NewETPRPC)

	qtumTest = NewBTCTest(
		txbuilder.NewTestParams("QTUM", "QTUM",
			"http://111:111@127.0.0.1:6031/",
			"c32697d43e132cce3fe51115278991976ccc17c638bfc7fc21fc15db9fce4023",
			"QfqUSBqpVEiB2yPfzCgJY4J2f35ffLUvJB",
			0.01),
		"43382190f2b885a801e90dfc21fc5fcc43f220c7c10b0fceafb0917bbe3739d2",
		1,
		rpc.NewBTCRPC)

	fabTest = NewBTCTest(
		txbuilder.NewTestParams("FAB", "FAB",
			"http://111:111@127.0.0.1:6035/",
			"5482dd97cb880aac553e2d67c74c98b1fb8c518aa9e57d3bd356e629a40db34a",
			"1HwwYNUaC7FXANxXK75cMqQ1Fj9Gtynw3U",
			0.01),
		"51ae3d1809207cc33ccf223b406af480418b831e8e823c1b9dd68511641f4267",
		1,
		rpc.NewBTCRPC)

	monaTest = NewBTCTest(
		txbuilder.NewTestParams("MONA", "MONA",
			"http://111:111@127.0.0.1:6029/",
			"1b2192d060ca607e356603369c4f26ba3882a6a9e3cbb11ddf5f86b3695321e3",
			"MVWRrFSdBPbDwRLo3pyoTbC9ZbEdufAeuN",
			0.01),
		"5e9aafe82324413f718d4b6d1ab0d3f7535d5f882fef737dc07495238c40991e",
		1,
		rpc.NewBTCRPC)

	test = monaTest
)

func TestSendTx(t *testing.T) {
	err := test.Init()
	if err != nil {
		t.Fatal(err)
	}

	metaData, _ := txbuilder.FindMetaData(test.Symbol())

	rpcClient := test.RPCClient()

	tx, err := rpcClient.GetRawTransaction(test.PreTx())
	if err != nil {
		t.Fatal(err)
	}

	tx.MakeHash()

	totalIn := decimal.New(int64(tx.Outputs[test.PreTxOutIdx()].Value), -int32(metaData.Precision))
	var outputs []*gbtc.Output
	txbuilder.MakeOutputs(
		totalIn, decimal.NewFromFloat(test.Amount()), decimal.Zero,
		test.ToAddress(), test.FromAddress(),
		metaData,
		func(address string, value uint64) {
			outputs = append(outputs, gbtc.NewOutput(address, value))
		},
	)
	newTx, err := rpcClient.CreateRawTransaction(metaData.TxVersion, []*gbtc.OutputPoint{
		&gbtc.OutputPoint{
			Hash:  tx.Hash,
			Index: test.PreTxOutIdx(),
		},
	}, outputs)
	if err != nil {
		t.Fatal(err)
	}

	txJSON, _ := newTx.JSON()
	fmt.Println(string(txJSON))
	fmt.Println()

	err = gbtc.Sign(rpcClient, newTx, gbtc.SigVersionBase, test)
	if err != nil {
		t.Fatal(err)
	}

	newTx.MakeHash()

	txJSON, _ = newTx.JSON()
	fmt.Println(string(txJSON))
	fmt.Println()

	txHash, err := rpcClient.SendRawTransaction(newTx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("success:", txHash)
}

func TestGetTx(t *testing.T) {
	rpcClient := test.RPCClient()
	rawTx, err := rpcClient.GetTransactionDetail(test.PreTx())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(rawTx))
}

func TestMakeOutputs(t *testing.T) {
	var (
		totalIn       = decimal.NewFromFloat(123.456)
		outAddress    = "outAddress"
		changeAddress = "changeAddress"
		outAmount     = decimal.NewFromFloat(120.45)
		maxOutAmount  = decimal.NewFromFloat(50)

		metaData = txbuilder.NewMetaData(
			8,
			decimal.NewFromFloat(0.001),
			1,
			30)

		outputs []*gbtc.Output
	)

	txbuilder.MakeOutputs(
		totalIn, outAmount, decimal.Zero,
		outAddress, changeAddress,
		metaData,
		func(address string, value uint64) {
			outputs = append(outputs, gbtc.NewOutput(address, value))
		},
	)
	assert(t, len(outputs) == 2, "check1 output len failed")
	assert(t, outputs[0].Value == 12045000000, "check1 output value failed")
	assert(t, outputs[1].Value == 300500000, "check1 change value failed")

	outputs = nil
	txbuilder.MakeOutputs(
		totalIn, outAmount, maxOutAmount,
		outAddress, changeAddress,
		metaData,
		func(address string, value uint64) {
			outputs = append(outputs, gbtc.NewOutput(address, value))
		},
	)
	assert(t, len(outputs) == 4, "check2 output len failed")
	assert(t, outputs[0].Value == 5000000000, "check2 output value at index 0 failed")
	assert(t, outputs[1].Value == 5000000000, "check2 output value at index 1 failed")
	assert(t, outputs[2].Value == 2045000000, "check2 output value at index 2 failed")
	assert(t, outputs[3].Value == 300500000, "check2 change value failed")
}
