package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/checker/checker/calculator"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/eth/geth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

var (
	ethTest = txbuilder.NewTestParams("ETH", "ETH",
		"http://127.0.0.1:6016",
		"35fe93539227d64a434c0aeec35007a0e32b3e74c8f20a2ddfa10bd15fa98431",
		"0xeca041473e92daeba47a000b943568da6cb7739e",
		0.1)

	ioncTest = txbuilder.NewTestParams("IONC", "ETH",
		"http://127.0.0.1:6085",
		"7cc47bc2bdf29d73bd7f033132ef7150a0bc7f29acb028ae7b5745df567d812e",
		"0xe34346b21c8e6a9a9c519f51de7ae1398e3f7af3",
		12.2)
	chainIds = map[string]int64{"ionc": 138111}

	test = ioncTest
)

func init() {
	test.Init()
}

func getETHBalance(client *ethclient.Client, address string) (decimal.Decimal, error) {
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(balance, -geth.Precision), nil
}
func TestGetBalance(t *testing.T) {
	rpcClient, err := ethclient.Dial(test.RPCURL())
	if err != nil {
		t.Fatal(err)
	}

	addr := test.FromAddress()
	fmt.Println(addr)

	balance, err := getETHBalance(rpcClient, addr)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("balance:", balance)

	nonce, err := rpcClient.NonceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("nonce:", nonce)
}

func getAllBalance(client *ethclient.Client, address string, contractAddrs []string) (decimal.Decimal, map[string]decimal.Decimal, error) {
	ethBalance, err := getETHBalance(client, address)
	if err != nil {
		return decimal.Zero, nil, err
	}

	erc20Balances := make(map[string]decimal.Decimal, len(contractAddrs))
	for _, contractAddr := range contractAddrs {
		precision, err := geth.GetERC20Precision(client, contractAddr)
		if err != nil {
			return decimal.Zero, nil, err
		}

		if contractAddr == "0x0fcf6ff5e0dcf87f50c52978d6fd89c01e0e106f" {
			precision = 6
		}

		balance, err := geth.GetERC20Balance(client, contractAddr, address, int32(precision))
		if err != nil {
			return decimal.Zero, nil, err
		}

		erc20Balances[contractAddr] = balance
	}

	return ethBalance, erc20Balances, nil
}

func TestGetERC20Balance(t *testing.T) {
	rpcClient, err := ethclient.Dial(test.RPCURL())
	if err != nil {
		t.Fatal(err)
	}

	const (
		contractAddr = "0x9e27e1c5ea6d1f4b0ac3ccc046824a325eb769c9"
		addr         = "0xa4573902c992b6e3b7df05fb61644c3c56885146"
	)

	symbol, err := geth.GetERC20Symbol(rpcClient, contractAddr)
	if err != nil {
		t.Fatal(err)
	}

	precision, err := geth.GetERC20Precision(rpcClient, contractAddr)
	if err != nil {
		t.Fatal(err)
	}

	balance, err := geth.GetERC20Balance(rpcClient, contractAddr, addr, int32(precision))
	if err != nil {
		t.Fatal((err))
	}

	fmt.Println(addr)
	fmt.Println("balance:", balance, symbol)
}

func TestTransfer(t *testing.T) {
	rpcClient, err := ethclient.Dial(test.RPCURL())
	if err != nil {
		t.Fatal(err)
	}

	fromAddress := common.HexToAddress(test.FromAddress())
	nonce, err := rpcClient.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	var (
		toAddress = common.HexToAddress(test.ToAddress())
		amt       = decimal.NewFromFloat(test.Amount()).Mul(decimal.New(1, geth.Precision))
		payload   []byte
	)
	bigAmount, ok := decimalToBigInt(amt)
	if !ok {
		t.Fatal("decimalToBigInt failed")
		return
	}

	gasLimit, err := estimateGasLimit(rpcClient, test.Symbol(), test.Symbol(), fromAddress, &toAddress, bigAmount, payload)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := suggestGasPrice(rpcClient, false)
	if err != nil {
		t.Fatal(err)
	}

	tx := types.NewTransaction(nonce, toAddress, bigAmount, gasLimit, gasPrice, payload)
	var chainID *big.Int
	chainid, ok := chainIds[strings.ToLower(test.Symbol())]
	if ok {
		chainID = big.NewInt(chainid)
	} else {
		chainID, err = rpcClient.ChainID(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	signer := types.NewEIP155Signer(chainID)
	sigDigest := signer.Hash(tx).Bytes()
	sig, err := test.Sign(sigDigest)
	if err != nil {
		t.Fatal(err)
	}

	tx, err = tx.WithSignature(signer, sig)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(tx.Hash().Hex())

	err = rpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("success:", tx.Hash().Hex())
}

func TestEstimateFee(t *testing.T) {
	var (
		rpcClient, _ = ethclient.Dial(test.RPCURL())
		address      = common.Address{}
		amount       = big.NewInt(0)
		payload      []byte
		err          error
	)
	payload, err = geth.PackABIParams("transfer", address, amount)
	if err != nil {
		t.Fatal(err)
	}

	gasLimit, err := estimateGasLimit(rpcClient, "usdt", "eth", address, &address, amount, payload)
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := suggestGasPrice(rpcClient, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("gasLimit: %d(0x%x), gasPrice: %d(0x%x)\n", gasLimit, gasLimit, gasPrice, gasPrice)

	fee := decimal.New(int64(gasLimit), 0).Mul(decimal.NewFromBigInt(gasPrice, -geth.Precision))
	fmt.Println("fee:", fee)
}

func TestReadjustFee(t *testing.T) {
	txHash := "0xf973cb7fc73e453d977d96edd1dc50621288a26b143373ecec61e459c35f8819"
	info, err := calculator.EthCalc(&config.Config{RPCUrl: test.RPCURL()}, txHash)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("remain:", info.RemainFee)
}
