package trx

import (
	"encoding/hex"
	"fmt"
	"time"

	"upex-wallet/wallet-base/util"

	bviper "upex-wallet/wallet-base/viper"

	"upex-wallet/wallet-withdraw/broadcast/handler"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/trx/gtrx"

	"github.com/buger/jsonparser"
)

func init() {
	handler.Register("trx", &trxHandler{})
}

type trxHandler struct {
	handler.BaseHandler
	rsaKey    string
	rpcClient *gtrx.Client
}

func (h *trxHandler) Init() error {
	h.rsaKey = bviper.GetString("trx.rsaKey", "")
	h.rpcClient = gtrx.NewClient(bviper.GetString("trx.rpcUrl", ""))
	h.SetConfigPrefix("trx")
	return h.InitDB(bviper.GetString("trx.dsn", ""))
}

func (h *trxHandler) BuildTx(txHex string, signatures []string, pubKeys []string) (handler.Tx, string, error) {
	sigs := handler.DecryptSignatures(h.rsaKey, signatures)
	if len(sigs) == 0 {
		return nil, "", handler.ErrDecryptSignatureFail
	}

	tx, err := gtrx.JSONUnmarshalTx([]byte(txHex))
	if err != nil {
		return nil, "", err
	}

	tx.Signature = append(tx.Signature, hex.EncodeToString(sigs[0]))

	return tx, tx.TxID, nil
}

func (h *trxHandler) BroadcastTransaction(tx handler.Tx, txHash string) (string, error) {
	trxTx := tx.(*gtrx.Transaction)
	_, err := h.rpcClient.BroadcastTransaction(trxTx)
	if err != nil {
		if h.VerifyTxBroadcasted(trxTx.TxID) {
			return trxTx.TxID, nil
		}

		return "", fmt.Errorf("%s, txid: %s, %v", handler.ErrBroadcastFail, trxTx.TxID, err)
	}

	return trxTx.TxID, nil
}

func (h *trxHandler) VerifyTxBroadcasted(txHash string) bool {
	var id string
	util.TryWithInterval(3, time.Second, func(int) error {
		txData, err := h.rpcClient.GetTransactionInfoByID(txHash)
		if err != nil {
			return err
		}

		result, _ := jsonparser.GetString(txData, "result")
		if result == "FAILED" {
			return fmt.Errorf("failed")
		}

		id, err = jsonparser.GetString(txData, "id")
		if err != nil {
			return err
		}

		return nil
	})

	return id == txHash
}
