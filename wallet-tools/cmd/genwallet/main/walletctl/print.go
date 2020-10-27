package main

import (
	"encoding/json"
)

type WalletPrintInfoItem struct {
	ID         int               `json:"id"`
	PrivateKey string            `json:"privateKey"`
	PublicKey  string            `json:"publicKey"`
	Address    string            `json:"address"`
	ExtData    map[string]string `json:"extData"`
}

func NewWalletPrintInfoItem() *WalletPrintInfoItem {
	return &WalletPrintInfoItem{
		ExtData: make(map[string]string),
	}
}

type WalletPrintInfo struct {
	Class        string                 `json:"class"`
	Cryptography string                 `json:"cryptography"`
	Total        int                    `json:"total"`
	Keypairs     []*WalletPrintInfoItem `json:"keypairs"`
}

func (w *WalletPrintInfo) Add(item *WalletPrintInfoItem) {
	w.Keypairs = append(w.Keypairs, item)
}

func (w *WalletPrintInfo) String() string {
	data, _ := json.Marshal(w)
	return string(data)
}
