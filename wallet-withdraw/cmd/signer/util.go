package main

import (
	"fmt"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
)

func main() {
	signStr, _ := rsa.B64Encrypt("password后3位", "signer密钥对公钥")
	fmt.Println(signStr)
}
