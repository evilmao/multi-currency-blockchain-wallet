package eth

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/deposit"
	"upex-wallet/wallet-deposit/rpc"
	"upex-wallet/wallet-deposit/rpc/eth/contracts"
	"upex-wallet/wallet-deposit/rpc/eth/geth"

	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/ethereum/token"

	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
)

func init() {
	rpc.Register("ETH", New)
	rpc.Register("ETC", New)
	rpc.Register("SMT", New)
	rpc.Register("IONC", New)
}

type RPC struct {
	cfg    *config.Config
	client *geth.Client
}

func New(cfg *config.Config) rpc.RPC {
	client := geth.NewClient(cfg.RPCURL)
	return rpc.NewBlockScanRPC(cfg, &RPC{
		cfg:    cfg,
		client: client,
	}, true)
}

func (r *RPC) GetLastBlockHeight() (uint64, error) {
	return r.client.GetLatestBlockNumber()
}

func (r *RPC) GetBlockHashByHeight(height uint64) (string, error) {
	block, err := r.client.GetBlockByNumber(height)
	if err != nil {
		return "", err
	}
	if len(block) == 0 {
		return "", fmt.Errorf("block is nil")
	}
	return jsonparser.GetString(block, "hash")

}

func (r *RPC) GetBlockByHeight(height uint64) (interface{}, error) {
	return r.client.GetBlockByNumber(height)
}

func (r *RPC) ParsePreviousBlockHash(block interface{}) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block is nil")
	}
	return jsonparser.GetString(block.([]byte), "parentHash")
}

func (r *RPC) ParseBlockHash(block interface{}) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block is nil")
	}
	return jsonparser.GetString(block.([]byte), "hash")
}

func (r *RPC) ParseBlockTxs(height uint64, hash string, block interface{}, handleTx func(tx interface{}) error) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	return util.JSONParserArrayEach(block.([]byte), func(tx []byte, _ jsonparser.ValueType) error {
		return handleTx(tx)
	}, "transactions")
}

func (r *RPC) GetTx(hash string) (interface{}, error) {
	return r.client.GetTransactionByHash(hash)

}

func (r *RPC) GetTxConfirmations(hash string) (uint64, error) {
	tx, err := r.client.GetTransactionByHash(hash)
	if err != nil {
		return 0, err
	}
	if tx == nil {
		return 0, fmt.Errorf("tx is nil")
	}

	txHexHeight, err := jsonparser.GetString(tx, "blockNumber")
	txHeight, err := hexutil.DecodeUint64(txHexHeight)
	if err != nil {
		return 0, err
	}

	lastBlockHeight, err := r.GetLastBlockHeight()
	if err != nil {
		return 0, err
	}

	confirm := deposit.CalculateConfirm(int64(txHeight), int64(lastBlockHeight))
	return uint64(confirm), nil
}

func (r *RPC) ParseTx(tx interface{}) ([]*models.Tx, error) {
	var (
		txs     []*models.Tx
		address string
	)
	txData := tx.([]byte)
	hash, err := jsonparser.GetString(txData, "hash")
	if err != nil {
		return nil, fmt.Errorf("parse tx hash failed, %v", err)
	}

	address, _ = jsonparser.GetString(txData, "to")
	// The 'to' field is null in deploy contract transactions.
	if len(address) == 0 {
		return nil, nil
	}

	input, err := jsonparser.GetString(txData, "input")
	if err != nil {
		return nil, fmt.Errorf("parse tx input failed, %v", err)
	}

	if len(input) != 0 && input != geth.HEX_PREFIX &&
		bytes.Equal(crypto.Hash160([]byte(address + "FCoin Wallet"))[:8], hexutil.MustDecode(input)) {

		return nil, nil
	}

	ok, err := models.HasAddress(address)
	if err != nil {
		return nil, err
	}

	if ok {
		amount, err := jsonparser.GetString(txData, "value")
		if err != nil {
			return nil, fmt.Errorf("parse tx value failed, %v", err)
		}

		a := hexutil.MustDecodeBig(amount)
		amt := decimal.NewFromBigInt(a, -geth.PRECISION)
		txs = append(txs, &models.Tx{
			Hash:    hash,
			Address: address,
			Amount:  amt,
			Confirm: 1,
			Symbol:  r.cfg.Currency,
		})
		return txs, nil
	}

	if len(input) == 0 || input == geth.HEX_PREFIX {
		return nil, nil
	}

	receipt, err := r.client.GetTransactionReceipt(hash)
	if err != nil {
		return nil, fmt.Errorf("get receipt by hash :%s faield, %v", hash, err)
	}

	status, _ := jsonparser.GetString(receipt, "status")
	// The 'status' field may be null in some etc transactions.
	if status != geth.RECEIPT_STATUS_SUCCESS {
		return nil, nil
	}

	mts, err := r.parseTokenTx(address, hash, receipt)
	if err != nil {
		return nil, fmt.Errorf("parse token tx failed, %v", err)
	}

	if len(mts) != 0 {
		txs = append(txs, mts...)
	}

	mt, err := r.parseInternalTx(address, hash, input)
	if err != nil {
		return nil, fmt.Errorf("parse internal tx failed, %v", err)
	}

	if mt != nil {
		txs = append(txs, mt)
	}
	return txs, nil
}

func (r *RPC) parseTokenTx(contractAddress, hash string, receipt []byte) ([]*models.Tx, error) {
	var (
		symbol    string
		precision int32
		mtxs      []*models.Tx
	)

	// check contract address
	detail, ok := currency.CurrencyDetailByAddress(contractAddress)
	if !ok {
		return nil, nil
	}
	symbol = strings.ToLower(detail.Symbol)
	precision = int32(detail.Decimal)

	err := util.JSONParserArrayEach(receipt, func(log []byte, _ jsonparser.ValueType) error {
		eventId, err := jsonparser.GetString(log, "topics", "[0]")
		if err != nil {
			return fmt.Errorf("parse log topics[0] failed, %v", err)
		}
		if eventId != hexutil.Encode(token.TransferEventID) {
			return nil
		}

		address, err := jsonparser.GetString(log, "topics", "[2]")
		if err != nil {
			return fmt.Errorf("parse log topics[2] failed, %v", err)
		}
		address = geth.HEX_PREFIX + strings.TrimLeft(string(address[2:]), "0")
		ok, err := models.HasAddress(address)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		amount, err := jsonparser.GetString(log, "data")
		if err != nil {
			return fmt.Errorf("parse log data failed, %v", err)
		}
		amount = strings.TrimLeft(string(amount[2:]), "0")
		if amount == "" {
			amount = "0"
		}
		amount = geth.HEX_PREFIX + amount
		a := hexutil.MustDecodeBig(amount)
		amt := decimal.NewFromBigInt(a, -precision)

		mtxs = append(mtxs, &models.Tx{
			Hash:    hash,
			Address: address,
			Amount:  amt,
			Confirm: 1,
			Symbol:  symbol,
		})
		return nil
	}, "logs")

	if err != nil {
		return nil, err
	}
	return mtxs, nil

}

func (r *RPC) parseInternalTx(contractAddress, hash, txInput string) (*models.Tx, error) {
	to, amount, err := ParseInternalTx(contractAddress, txInput)
	if err != nil {
		return nil, err
	}

	if len(to) == 0 {
		return nil, nil
	}

	ok, err := models.HasAddress(to)
	if err != nil || !ok {
		return nil, err
	}

	return &models.Tx{
		Address: to,
		Amount:  decimal.NewFromBigInt(amount, -geth.PRECISION),
		Hash:    hash,
		Confirm: 1,
		Symbol:  r.cfg.Currency,
	}, nil
}

func ParseInternalTx(contractAddress, inputHex string) (string, *big.Int, error) {
	caller, ok := contracts.Caller(contractAddress)
	if !ok {
		return "", nil, nil
	}

	input, err := hexutil.Decode(inputHex)
	if err != nil {
		return "", nil, fmt.Errorf("decode txInput %s failed, %v", inputHex, err)
	}

	if !bytes.Equal(caller.MethodID(), input[:4]) {
		return "", nil, nil
	}

	addr, amount, err := caller.UnpackParams(input[4:])
	if err != nil {
		return "", nil, err
	}

	return addr.String(), amount, nil
}
