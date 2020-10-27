package signer

import (
	"fmt"
	"math/big"
	"testing"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
	"upex-wallet/wallet-base/newbitx/misclib/ethereum/types"

	"github.com/ethereum/go-ethereum/common"
)

func TestSignTx(t *testing.T) {
	// make tx
	tx := types.NewTransaction(
		0,
		common.HexToAddress("ee3304057099030f227c5668f1b078cfc81489e4"),
		big.NewInt(100),
		big.NewInt(210000),
		big.NewInt(20),
		nil,
	)
	// signature request
	pass, err := rsa.B64Encrypt("23", `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwNUOTn18XDvxYsnHssQ1
A422GrgaQF8C8/Nb3El9J+3K8qp0mrt6W+jlLt397I26bBt4cnue02VewC3uuzN9
miXxuG9dmYgFzmUIrihZDQZ7h2IeRQ112kyCJ5fDQL+3e1JBti+DYLCr7RU/njpx
cDmR+403xqYo0uDzQ9clyaOW6YVqAIj9ao6ANKzm3QrALdt9tufaMGsmkqOZA6WJ
/SoNy8ggtbuEOCe4cx5YZBinxWO43YH7awGRl/7D2AZHcbzhkEAwDmnWDEw+s9A6
OaVWRY64VAzFwlrFsuwQO1stDimRfMsUY1BaBvG6iZEUkTs+YAJ/XxM/3jefodYi
/wIDAQAB
-----END PUBLIC KEY-----`)
	fmt.Println(pass, err)
	signerClient := NewClient("http://localhost:9090", pass)
	fmt.Println(signerClient.Request(
		&Request{
			PubKeys:   []string{"ee3304057099030f227c5668f1b078cfc81489e4"},
			Digests:   []string{tx.Hash().Hex()[2:]},
			AuthToken: pass,
		},
	))
	// relay
}
