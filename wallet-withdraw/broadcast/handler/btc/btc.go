package btc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	bviper "upex-wallet/wallet-base/viper"
	"upex-wallet/wallet-withdraw/broadcast/handler"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/btc/gbtc/rpc"

	"github.com/buger/jsonparser"
)

func init() {
	handler.Register("btc", &btcHandler{})
	handler.Register("etp", &btcHandler{
		initer: func(h *btcHandler) error {
			h.rsaKey = bviper.GetString("etp.rsaKey", "")
			h.rpcClient = rpc.NewETPRPC(bviper.GetString("etp.rpcUrl", ""))
			h.extRPCURL = bviper.GetString("etp.extRPCUrl", "")
			h.SetConfigPrefix("etp")
			return h.InitDB(bviper.GetString("etp.dsn", ""))
		},
		deserConf: gbtc.DeserializeConfig{
			WithAttachment: true,
		},
		txHashKey: "hash",
	})
	handler.Register("qtum", &btcHandler{
		initer: func(h *btcHandler) error {
			h.rsaKey = bviper.GetString("qtum.rsaKey", "")
			h.rpcClient = rpc.NewBTCRPC(bviper.GetString("qtum.rpcUrl", ""))
			h.extRPCURL = bviper.GetString("qtum.extRPCUrl", "")
			h.SetConfigPrefix("qtum")
			return h.InitDB(bviper.GetString("qtum.dsn", ""))
		},
	})
	handler.Register("fab", &btcHandler{
		initer: func(h *btcHandler) error {
			h.rsaKey = bviper.GetString("fab.rsaKey", "")
			h.rpcClient = rpc.NewBTCRPC(bviper.GetString("fab.rpcUrl", ""))
			h.extRPCURL = bviper.GetString("fab.extRPCUrl", "")
			h.SetConfigPrefix("fab")
			return h.InitDB(bviper.GetString("fab.dsn", ""))
		},
	})
	handler.Register("mona", &btcHandler{
		initer: func(h *btcHandler) error {
			h.rsaKey = bviper.GetString("mona.rsaKey", "")
			h.rpcClient = rpc.NewBTCRPC(bviper.GetString("mona.rpcUrl", ""))
			h.extRPCURL = bviper.GetString("mona.extRPCUrl", "")
			h.SetConfigPrefix("mona")
			return h.InitDB(bviper.GetString("mona.dsn", ""))
		},
	})
}

type btcHandler struct {
	handler.BaseHandler

	initer func(*btcHandler) error

	rsaKey    string
	rpcClient gbtc.RPC
	extRPCURL string
	deserConf gbtc.DeserializeConfig
	txHashKey string // default: "txid"
}

func (h *btcHandler) Init() error {
	if len(h.txHashKey) == 0 {
		h.txHashKey = "txid"
	}

	if h.initer != nil {
		return h.initer(h)
	}

	h.rsaKey = bviper.GetString("btc.rsaKey", "")
	h.rpcClient = rpc.NewBTCRPC(bviper.GetString("btc.rpcUrl", ""))
	h.extRPCURL = strings.TrimRight(bviper.GetString("btc.extRPCUrl", ""), "/")
	h.SetConfigPrefix("btc")
	return h.InitDB(bviper.GetString("btc.dsn", ""))
}

func (h *btcHandler) BuildTx(txHex string, signatures []string, pubKeys []string) (handler.Tx, string, error) {
	sigs := handler.DecryptSignatures(h.rsaKey, signatures)
	if len(sigs) == 0 {
		return nil, "", handler.ErrDecryptSignatureFail
	}

	txData, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, "", fmt.Errorf("%s, %v", handler.ErrTxHexForamt, err)
	}

	tx := gbtc.Transaction{
		DeserConf: h.deserConf,
	}
	err = tx.SetBytes(txData)
	if err != nil {
		return nil, "", fmt.Errorf("%s, %v", handler.ErrTxHexForamt, err)
	}

	if len(sigs) != len(tx.Inputs) {
		return nil, "", fmt.Errorf("%s, got: %d, need: %d", handler.ErrSignatureCountMismatch, len(sigs), len(tx.Inputs))
	}

	for i, in := range tx.Inputs {
		pubKey, err := hex.DecodeString(pubKeys[i])
		if err != nil {
			return nil, "", handler.ErrDecodePubKeyFail
		}

		buffer := new(bytes.Buffer)
		util.WriteVarBytes(buffer, append(sigs[i], gbtc.SigHashAll))
		util.WriteVarBytes(buffer, pubKey)
		in.ScriptData = buffer.Bytes()
	}

	tx.MakeHash()

	return &tx, tx.Hash, nil
}

func (h *btcHandler) BroadcastTransaction(tx handler.Tx, txHash string) (string, error) {
	btcTx := tx.(*gbtc.Transaction)
	txHash, err := h.rpcClient.SendRawTransaction(btcTx)
	if err != nil {
		txHash = btcTx.Hash
		if !h.VerifyTxBroadCasted(txHash) {
			return "", fmt.Errorf("%s, txid: %s, %v", handler.ErrBroadcastFail, btcTx.Hash, err)
		}
	}

	if len(h.extRPCURL) > 0 {
		_, err := util.RestRawPost(h.extRPCURL, map[string]string{
			"tx": hex.EncodeToString(btcTx.Bytes()),
		})
		if err != nil {
			log.Errorf("post tx %s to extRPC %s failed, %v", txHash, h.extRPCURL, err)
		}
	}

	return txHash, nil
}

func (h *btcHandler) VerifyTxBroadCasted(txHash string) bool {
	tx, err := h.rpcClient.GetTransactionDetail(txHash)
	if err != nil {
		return false
	}

	hash, _ := jsonparser.GetString(tx, h.txHashKey)
	return hash == txHash
}
