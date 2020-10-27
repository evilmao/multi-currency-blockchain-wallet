# Zilliqa Golang SDK
[![GoDoc](https://godoc.org/github.com/GincoInc/zillean?status.svg)](https://godoc.org/github.com/GincoInc/zillean)
[![Go Report Card](https://goreportcard.com/badge/github.com/GincoInc/zillean)](https://goreportcard.com/report/github.com/GincoInc/zillean)
[![Build Status](https://travis-ci.org/GincoInc/zillean.svg?branch=develop)](https://travis-ci.org/GincoInc/zillean)

A Golang implementation of [Zilliqa](https://github.com/Zilliqa/Zilliqa) API.

The project is still under development. Please note that specifications may change greatly.

## Installation
```sh
go get -u github.com/GincoInc/zillean
go get -u github.com/GincoInc/go-crypto
```

## Getting started

```
cd $GOPATH/src/github.com/GincoInc/zillean/example
go run main.go
```

## Example
```go
package main

import (
	"fmt"

	"github.com/GincoInc/zillean"
)

func main() {
	// initialize the Zillean
	zil := zillean.NewZillean("https://api.zilliqa.com")

	// generate a private key
	privKey := zil.GeneratePrivateKey()
	fmt.Printf("private key: %s\n", privKey) // private key: b6c00064b10d33c4a9fadb5b473d834b1995f132acdbe4b831ab5343702c174e

	// get a public key
	pubKey, _ := zil.GetPublicKeyFromPrivateKey(privKey)
	fmt.Printf("public key: %s\n", pubKey) // public key: 03dcb21aaaa918f91a708858dc271343b4bee059e53202ce0358b68effa7e64378

	// get a address
	address, _ := zil.GetAddressFromPrivateKey(privKey)
	fmt.Printf("address: %s\n", address) // address: 5f0e26adf701bb6a4535f0485fe3400e6e90c9ae

	// sign the transaction
	rawTx := zillean.RawTransaction{
		Version:  0,
		Nonce:    2,
		To:       "to address",
		Amount:   "1000000000000",
		PubKey:   pubKey,
		GasPrice: big.NewInt(1000000000),
		GasLimit: 1,
	}
	signature, _ := zil.SignTransaction(rawTx, privKey)
	txID, _ := zil.RPC.CreateTransaction(rawTx, signature)
	fmt.Printf("txID: %s\n", txID)
}
```

## Supports
### Wallet API
- [x] GeneratePrivateKey
- [x] VerifyPrivateKey
- [x] GetPublicKeyFromPrivateKey
- [x] IsPublicjKey
- [x] GetAddressFromPrivateKey
- [x] GetAddressFromPublicKey
- [x] IsAddress
- [x] SignTransaction
- [x] VerifySignature

### JSON-RPC API
#### Blockchain-related methods
- [x] GetNetworkId
- [x] GetBlockchainInfo
- [x] GetShardingStructure
- [x] GetDsBlock
- [x] GetLatestDsBlock
- [x] GetNumDSBlocks
- [x] GetDSBlockRate
- [x] DSBlockListing
- [x] GetTxBlock
- [x] GetLatestTxBlock
- [x] GetNumTxBlocks
- [x] GetTxBlockRate
- [x] TxBlockListing
- [x] GetNumTransactions
- [x] GetTransactionRate
- [x] GetCurrentMiniEpoch
- [x] GetCurrentDSEpoch
- [x] GetPrevDifficulty
- [x] GetPrevDSDifficulty
#### Transaction-related methods
- [x] CreateTransaction
- [x] GetTransaction
- [x] GetRecentTransactions
- [x] GetTransactionsForTxBlock
- [x] GetNumTxnsTxEpoch
- [x] GetNumTxnsDSEpoch
- [x] GetMinimumGasPrice
#### Contract-related methods
- [x] GetSmartContractCode
- [x] GetSmartContractInit
- [x] GetSmartContractState
- [x] GetSmartContracts
- [x] GetContractAddressFromTransactionID
#### Account-related methods
- [x] GetBalance