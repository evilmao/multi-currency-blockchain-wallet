package txbuilder

import (
	"encoding/hex"
	"fmt"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
)

type TestParams struct {
	symbol     string
	class      string
	rpcURL     string
	privKeyHex string
	kp         keypair.KeyPair
	toAddress  string
	amount     float64
}

func NewTestParams(symbol, class, rpcURL, privKeyHex, toAddr string, amount float64) *TestParams {
	return &TestParams{
		symbol:     symbol,
		class:      class,
		rpcURL:     rpcURL,
		privKeyHex: privKeyHex,
		toAddress:  toAddr,
		amount:     amount,
	}
}

func (f *TestParams) Init() error {
	privKey, err := hex.DecodeString(f.privKeyHex)
	if err != nil {
		return fmt.Errorf("decode private key failed, %v", err)
	}

	kp, err := keypair.Build(f.class, privKey)
	if err != nil {
		return err
	}

	f.kp = kp
	return nil
}

func (f *TestParams) Symbol() string {
	return f.symbol
}

func (f *TestParams) Class() string {
	return f.class
}

func (f *TestParams) RPCURL() string {
	return f.rpcURL
}

func (f *TestParams) FromAddress() string {
	return f.kp.AddressString()
}

func (f *TestParams) FromPublicKey() []byte {
	return f.kp.PublicKey()
}

func (f *TestParams) ToAddress() string {
	return f.toAddress
}

func (f *TestParams) Amount() float64 {
	return f.amount
}

func (f *TestParams) Sign(hash []byte) ([]byte, error) {
	return f.kp.Sign(hash)
}

func (f *TestParams) KeyPair() keypair.KeyPair {
	return f.kp
}
