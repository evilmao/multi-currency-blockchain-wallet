package api

import (
	"fmt"
	"strings"
	"testing"
)

var (
	sandboxFcoinAPI = NewBrokerAPI("https://broker-www-sandbox.fcoin.com", "xx", "xx")
	sandboxBcoinAPI = NewBrokerAPI("https://broker-www-sandbox.bcoin.com", "xx", "xx")

	prdFcoinAPI = NewBrokerAPI("https://broker-api.fcoin.com", "xx", "xx")
	prdFJPAPI   = NewBrokerAPI("https://broker-api.fcoinjp.com", "xx", "xx")

	api = sandboxBcoinAPI
)

func TestBrokerAPI(t *testing.T) {
	resp, err := api.Currencies()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("total:", len(resp.Data))

	sIndex := make(map[string]bool)
	cIndex := make(map[string]bool)
	for _, v := range resp.Data {
		if _, ok := sIndex[v.Symbol]; ok {
			fmt.Println("symbol dunplicate:", v.Symbol)
		}
		sIndex[v.Symbol] = true

		if _, ok := cIndex[v.Code]; ok {
			fmt.Println("code dunplicate:", v.Code)
		}
		cIndex[v.Code] = true

		if v.Symbol == "zil" {
			fmt.Printf("%#v\n", v)
		}
	}
}

func TestBrokerTokenAPI(t *testing.T) {
	resp, err := api.CurrencyDetails()
	if err != nil {
		t.Fatal(err)
	}

	toString := func(v *CurrencyDetail) string {
		return fmt.Sprintf("blockchainName: %s, symbol: %s, kind: %s, belongChain: %s, address: %s, confirm: %d, minAmount: %s, precision: %d, withTag: %v, maxWithdraw: %v",
			v.BlockchainName, v.Symbol, v.Kind, v.BelongChainName(), v.Address, v.Confirm, v.MinDepositAmount, v.Decimal, v.WithTag, v.MaxWithdrawAmount)
	}

	like := func(d *CurrencyDetail, symbols ...string) bool {
		for _, s := range symbols {
			if strings.EqualFold(d.Symbol, s) || d.ChainBelongTo(s) || d.UseAddressOf(s) {
				return true
			}
		}
		return false
	}

	for _, v := range resp.Data {
		if !v.Valid() {
			fmt.Printf("error: invalid currency, %s\n", toString(v))
		}

		if like(v, "secs", "trx") {
			fmt.Println(toString(v))
		}
	}
}

func TestChangeWithdrawStatus(t *testing.T) {
	resp, err := api.ChangeWithdrawStatus("usdt", "default", true)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.Status)
}
