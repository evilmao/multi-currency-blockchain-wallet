package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	bviper "upex-wallet/wallet-base/viper"
	"upex-wallet/wallet-withdraw/broadcast/handler"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

func init() {
	handler.Register("eth", &ethHandler{configPrefix: "eth"})
	handler.Register("etc", &ethHandler{configPrefix: "etc"})
	handler.Register("smt", &ethHandler{configPrefix: "smt"})
	handler.Register("ionc", &ethHandler{configPrefix: "ionc"})
}

type ethHandler struct {
	handler.BaseHandler
	configPrefix string
	rsaKey       string
	rpcClient    *ethclient.Client
	chainID      *big.Int
}

func (h *ethHandler) Init() error {
	h.rsaKey = bviper.GetString(h.configPrefix+".rsaKey", "")

	rpcURL := bviper.GetString(h.configPrefix+".rpcUrl", "")
	if len(rpcURL) > 0 {
		var err error
		h.rpcClient, err = ethclient.Dial(rpcURL)
		if err != nil {
			log.Warnf("eth client dial %s failed, %v", rpcURL, err)
		}
	}

	chainID := bviper.GetString(h.configPrefix+".chainID", "")
	if len(chainID) > 0 {
		id, ok := new(big.Int).SetString(chainID, 10)
		if ok {
			h.chainID = id
		}
	}

	h.SetConfigPrefix(h.configPrefix)
	return h.InitDB(bviper.GetString(h.configPrefix+".dsn", ""))
}

func (h *ethHandler) BuildTx(txHex string, signatures []string, pubKeys []string) (handler.Tx, string, error) {
	if h.rpcClient == nil {
		return nil, "", fmt.Errorf("rpc client is nil")
	}

	// if h.chainID == nil {
	// 	chainID, err := h.rpcClient.ChainID(context.Background())
	// 	if err != nil {
	// 		return nil, "", fmt.Errorf("get chain id failed, %v", err)
	// 	}
	//
	// 	h.chainID = chainID
	// }

	sigs := handler.DecryptSignatures(h.rsaKey, signatures)
	if len(sigs) == 0 {
		return nil, "", handler.ErrDecryptSignatureFail
	}

	txData, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, "", fmt.Errorf("decode tx hex failed, %s, %v", txHex, err)
	}

	tx := new(types.Transaction)
	err = rlp.DecodeBytes(txData, tx)
	if err != nil {
		return nil, "", fmt.Errorf("decode bytes to tx failed, %v", err)
	}

	signer := types.NewEIP155Signer(h.chainID)
	tx, err = tx.WithSignature(signer, sigs[0])
	if err != nil {
		return nil, "", fmt.Errorf("set tx signature failed, %v", err)
	}

	return tx, tx.Hash().Hex(), nil
}

func (h *ethHandler) BroadcastTransaction(tx handler.Tx, txHash string) (string, error) {
	if h.rpcClient == nil {
		return "", fmt.Errorf("rpc client is nil")
	}

	ethTx := tx.(*types.Transaction)
	log.Warnf("2222-----------ethTX:%v",ethTx)
	err := h.rpcClient.SendTransaction(context.Background(), ethTx)
	if err != nil {
		if h.VerifyTxBroadCasted(txHash) {
			return txHash, nil
		}

		return "", fmt.Errorf("%s, txid: %s, %v", handler.ErrBroadcastFail, txHash, err)
	}

	return txHash, nil
}

func (h *ethHandler) VerifyTxBroadCasted(txHash string) bool {
	if h.rpcClient == nil {
		return false
	}

	tx, isPending, err := h.rpcClient.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil || isPending {
		return false
	}

	return strings.EqualFold(tx.Hash().Hex(), txHash)
}
