package tron

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/deposit"
	"upex-wallet/wallet-deposit/rpc"
	"upex-wallet/wallet-deposit/rpc/tron/gtrx"
	"upex-wallet/wallet-tools/base/crypto"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

const (
	TypeTransferContract      = "TransferContract"      // trx交易类型
	TypeTransferAssetContract = "TransferAssetContract" // token 交易类型 如:btt
	TypeTriggerSmartContract  = "TriggerSmartContract"  // 调用智能合约类型 如：tron-usdt
	ContractRetSuccess        = "SUCCESS"
	TRC20Transfer             = "a9059cbb"
	TRC20TransferParamsLen    = 128
	PRECISION                 = 6
	RawAddressLen             = 20
	AddressPrefix             = 0x41
)

func init() {
	rpc.Register("TRX", New)
}

type RPC struct {
	cfg          *config.Config
	client       *gtrx.Client
	currentBlock *models.BlockInfo
}

func New(cfg *config.Config) rpc.RPC {
	client := gtrx.NewClient(cfg.RPCURL)
	return rpc.NewBlockScanRPC(cfg, &RPC{
		cfg:    cfg,
		client: client,
	}, true)
}

func (r *RPC) GetLastBlockHeight() (uint64, error) {
	block, err := r.client.GetSolidityCurrentBlock()
	if err != nil {
		return 0, err
	}

	if block == nil {
		return 0, fmt.Errorf("get last block is nil")
	}

	height, err := jsonparser.GetInt(block, "block_header", "raw_data", "number")
	if err != nil {
		return 0, fmt.Errorf("parse block_header->raw_data->number failed,%v", err)
	}
	return uint64(height), nil
}

func (r *RPC) GetBlockHashByHeight(height uint64) (string, error) {
	block, err := r.client.GetSolidityBlockByNum(height)
	if err != nil {
		return "", err
	}
	if block == nil {
		return "", fmt.Errorf("get last block is nil")
	}

	blockID, err := jsonparser.GetString(block, "blockID")
	if err != nil {
		return "", fmt.Errorf("parse blockID faield,%v", err)
	}

	return blockID, nil
}

func (r *RPC) GetBlockByHeight(height uint64) (interface{}, error) {
	return r.client.GetSolidityBlockByNum(height)
}

func (r *RPC) ParsePreviousBlockHash(block interface{}) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block is nil")
	}
	return jsonparser.GetString(block.([]byte), "block_header", "raw_data", "parentHash")
}

func (r *RPC) ParseBlockHash(block interface{}) (string, error) {
	if block == nil {
		return "", fmt.Errorf("block is nil")
	}
	return jsonparser.GetString(block.([]byte), "blockID")
}

func (r *RPC) ParseBlockTxs(height uint64, hash string, block interface{}, handleTx func(tx interface{}) error) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	if !strings.Contains(string(block.([]byte)), "transactions") {
		return nil
	}

	err := util.JSONParserArrayEach(block.([]byte), func(transaction []byte, _ jsonparser.ValueType) error {
		return handleTx(transaction)
	}, "transactions")

	return err
}

func (r *RPC) GetTx(hash string) (interface{}, error) {
	return r.client.GetSolidityTransactionById(hash)
}

func (r *RPC) ParseTx(tx interface{}) ([]*models.Tx, []*models.UTXO, error) {
	if tx == nil {
		return nil, nil, fmt.Errorf("tx is nil")
	}

	var mtx []*models.Tx
	modelTxs, err := ParseSingleTx(tx.([]byte), r.cfg.Currency, r.cfg.TrxTokenAirDropAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("parse transaction failed, %v", err)
	}

	for _, modelTx := range modelTxs {
		ok, err := models.HasAddress(modelTx.Address)
		if err != nil {
			return nil, nil, fmt.Errorf("HasAddress failed, %v", err)
		}

		if ok {
			mtx = append(mtx, modelTx)
		}
	}

	return mtx, nil, err
}

func ParseSingleTx(transaction []byte, symbol, trxTokenAirDropAddress string) ([]*models.Tx, error) {
	if len(transaction) == 0 {
		return nil, fmt.Errorf("transaction is nil")
	}

	if symbol == "" {
		return nil, fmt.Errorf("symbol is nil")
	}

	if trxTokenAirDropAddress == "" {
		return nil, fmt.Errorf("air drop address is nil")
	}

	var (
		toAddr58     string
		decAmount    decimal.Decimal
		txs          []*models.Tx
		index        = -1
		contractRets []bool
	)

	util.JSONParserArrayEach(transaction, func(ret []byte, _ jsonparser.ValueType) error {
		contractRet, err := jsonparser.GetString(ret, "contractRet")
		if contractRet != ContractRetSuccess || err != nil {
			contractRets = append(contractRets, false)
		} else {
			contractRets = append(contractRets, true)
		}
		return nil
	}, "ret")

	txId, err := jsonparser.GetString(transaction, "txID")
	if err != nil {
		return nil, fmt.Errorf("parse txID failed, %v", err)
	}

	err = util.JSONParserArrayEach(transaction, func(contract []byte, _ jsonparser.ValueType) error {
		index++

		txType, err := jsonparser.GetString(contract, "type")
		if err != nil {
			return fmt.Errorf("parse type failed, %v", err)
		}
		// 判断交易类型
		if !AcceptTxType(txType) {
			return nil
		}

		parameterValue, _, _, err := jsonparser.Get(contract, "parameter", "value")
		if err != nil {
			return fmt.Errorf("parse parameter value failed, %v", err)
		}

		isAirDrop, err := IsAirDrop(parameterValue, trxTokenAirDropAddress)
		if err != nil {
			return err
		}

		if isAirDrop {
			log.Infof("ignore air drop deposit tx, hash: %s", txId)
			return nil
		}

		if txType == TypeTriggerSmartContract {
			// Always check contractRets.
			if index >= len(contractRets) || !contractRets[index] {
				return nil
			}

			contractAddress, err := jsonparser.GetString(parameterValue, "contract_address")
			if err != nil {
				return fmt.Errorf("parse contract_address failed,%v", err)
			}

			addressBuf, err := hex.DecodeString(contractAddress)
			if err != nil {
				return fmt.Errorf("decodeString contract_address failed,%v", err)
			}

			// check contract address
			ca := crypto.Base58Check(addressBuf, nil, false)
			fmt.Println("1111------ca=", ca)
			if c, ok := currency.CurrencyDetailByAddress(ca); ok {
				symbol = c.Symbol
			} else {
				return nil
			}

			if symbol == "" {
				return nil
			}

			data, err := jsonparser.GetString(parameterValue, "data")
			if err != nil {
				return fmt.Errorf("parse smart contract data failed,%v", err)
			}

			// 解析data 拿到toAddress和amount
			toAddr58, decAmount, err = parseAddressAmount(data, ca)
			if err != nil {
				return fmt.Errorf("parse toAddress,amount from data failed,%v", err)
			}
			if toAddr58 == "" || decAmount.Equal(decimal.Zero) {
				return nil
			}
		} else {
			amount, err := jsonparser.GetInt(parameterValue, "amount")
			if err != nil {
				return fmt.Errorf("parse amount failed,%v", err)
			}

			to, err := jsonparser.GetString(parameterValue, "to_address") // 类似：4190cf84a92971d828e16e70890b5e175aa5710772
			if err != nil {
				return fmt.Errorf("jsonparser to address failed, %v", err)
			}

			toByte, err := hex.DecodeString(to)
			if err != nil {
				return fmt.Errorf("decodeString toaddr failed, %v", err)
			}
			toAddr58 = crypto.Base58Check(toByte, nil, false) // 类似：TPAtuJNY9sHpA9d2gqvgTbHR94GuPeNhPP

			if txType == TypeTransferAssetContract {
				assetId, err := jsonparser.GetString(parameterValue, "asset_name")
				if err != nil {
					return fmt.Errorf("decodeString asset_name failed, %v", err)
				}

				var precision int
				symbol, precision, err = GetTokenSymbol(assetId)
				if err != nil {
					return fmt.Errorf("get token symbol failed, %v", err)
				}

				if symbol == "" {
					return nil
				}

				decAmount = decimal.New(amount, -int32(precision))
			} else {
				decAmount = decimal.New(amount, -PRECISION)
			}
		}

		modelTx := &models.Tx{
			Hash:       txId,
			Amount:     decAmount,
			Address:    toAddr58,
			Confirm:    1,
			Symbol:     symbol,
			InnerIndex: uint16(index),
		}

		txs = append(txs, modelTx)
		return nil
	}, "raw_data", "contract")

	if err != nil {
		return nil, err
	}

	return txs, err
}

/**
example:
symbol:BTT
CurrencyDetail.address = tokenId
id:1002000->hexString 31303032303030
链上的数据类型是 hexString
*/
func GetTokenSymbol(tokenIdHex string) (string, int, error) {
	buf, err := hex.DecodeString(tokenIdHex)
	if err != nil {
		return "", 0, err
	}

	tokenId := string(buf)
	if c, ok := currency.CurrencyDetailByAddress(tokenId); ok {
		return c.Symbol, c.Decimal, nil
	}

	return "", 0, nil
}

/**
解析data形如
a9059cbb0000000000000000000000414948c2e8a756d9437037dcd8c7e0c73d560ca38d0000000000000000000000000000000000000000000000000000000000989680
a9059cbb hash3(transfer(address,uint256))[:4] 固定值
后面是两个参数 toAddress，amount
*/
func parseAddressAmount(data string, contractAddress string) (string, decimal.Decimal, error) {
	if !strings.HasPrefix(data, TRC20Transfer) {
		return "", decimal.Zero, nil
	}

	data = strings.TrimPrefix(data, TRC20Transfer)
	if len(data) != TRC20TransferParamsLen {
		return "", decimal.Zero, nil
	}

	m := TRC20TransferParamsLen / 2

	// Parse address.
	to := data[:m]
	toByte, err := hex.DecodeString(to)
	if err != nil {
		return "", decimal.Zero, fmt.Errorf("parse to address failed, %v", err)
	}

	toByte = toByte[len(toByte)-RawAddressLen:]
	to58 := crypto.Base58Check(toByte, []byte{AddressPrefix}, false)

	// fetch precision
	detail, ok := currency.CurrencyDetailByAddress(contractAddress)
	if !ok {
		return "", decimal.Zero, fmt.Errorf("can't find contractAddress: %s", contractAddress)
	}

	// Parse amount.
	amtHex := data[m:]
	amtBuf, err := hex.DecodeString(amtHex)
	if err != nil {
		return "", decimal.Zero, fmt.Errorf("parse amount from hex failed,%v", err)
	}

	b := big.NewInt(0)
	b.SetBytes(amtBuf)
	amount := decimal.NewFromBigInt(b, -int32(detail.Decimal))

	return to58, amount, nil

}

func AcceptTxType(txType string) bool {
	return txType == TypeTransferContract ||
		txType == TypeTransferAssetContract ||
		txType == TypeTriggerSmartContract
}

func IsAirDrop(rawValueData []byte, airDropAddress string) (bool, error) {
	ownerAddress, err := jsonparser.GetString(rawValueData, "owner_address")
	if err != nil {
		return false, fmt.Errorf("parse owner_address failed,%v", err)
	}

	ownerByte, err := hex.DecodeString(ownerAddress)
	if err != nil {
		return false, fmt.Errorf("hex decode ownerAddress failed,%v", err)
	}

	ownerAddr58 := crypto.Base58Check(ownerByte, nil, false)

	// 空投地址的转账不上账
	// 确认是否来自空投地址（有多个，不同的token有一个或者多个空投地址）
	return strings.Contains(airDropAddress, ownerAddr58), nil
}

func (r *RPC) GetTxConfirmations(hash string) (uint64, error) {
	if hash == "" {
		return 0, errors.New("hash is nil")
	}

	tx, err := r.client.GetTransactionInfoById(hash)
	if err != nil {
		return 0, fmt.Errorf("get transaction info %s failed,%v", hash, err)
	}

	thisHeight, err := jsonparser.GetInt(tx, "blockNumber")
	if err != nil {
		return 0, fmt.Errorf("parse blockNumber failed,%v", err)
	}

	nowBlock, err := r.client.GetSolidityCurrentBlock()
	if err != nil {
		return 0, fmt.Errorf("get solidity current block failed,%v", err)
	}

	currentHeight, err := jsonparser.GetInt(nowBlock, "block_header", "raw_data", "number")
	if err != nil {
		return 0, fmt.Errorf("parse block_header->raw_data->number failed,%v", err)
	}
	confirm := uint64(deposit.CalculateConfirm(thisHeight, currentHeight))

	return confirm, nil
}
