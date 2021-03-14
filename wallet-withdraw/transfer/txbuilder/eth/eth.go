package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer/alarm"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/eth/geth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
)

// todo: 关闭 eth/erc20功能
func init() {
	// txbuilder.Register("ETH", New)
	// txbuilder.Register("ETC", New)
	// txbuilder.Register("SMT", New)
	// txbuilder.Register("IONC", New)
}

type ETHBuilder struct {
	cfg     *config.Config
	client  *ethclient.Client
	chainID *big.Int
}

func New(cfg *config.Config) txbuilder.Builder {
	client, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		log.Errorf("eth client dial %s failed, %v", cfg.RPCUrl, err)
		return nil
	}

	var chainID *big.Int
	if len(cfg.ChainID) > 0 {
		id, ok := new(big.Int).SetString(cfg.ChainID, 10)
		if ok {
			chainID = id
		}
	}

	return txbuilder.NewAccountModelBuilder(cfg, &ETHBuilder{
		cfg:     cfg,
		client:  client,
		chainID: chainID,
	})
}

func (b *ETHBuilder) DefaultFeeMeta() txbuilder.FeeMeta {
	return txbuilder.FeeMeta{
		Fee: decimal.New(_minGasLimit*_minGasPrice, -geth.Precision),
	}
}

func (b *ETHBuilder) EstimateFeeMeta(symbol string, txType int8) *txbuilder.FeeMeta {
	var (
		address = common.Address{}
		amount  = big.NewInt(0)
		payload []byte
		err     error
	)
	if symbol != b.cfg.Currency {
		payload, err = geth.PackABIParams("transfer", address, amount)
		if err != nil {
			return nil
		}
	}

	gasLimit, err := estimateGasLimit(b.client, symbol, b.cfg.Currency, address, &address, amount, payload)
	if err != nil {
		return nil
	}

	highPriority := txType == models.TxTypeWithdraw
	gasPrice, err := suggestGasPrice(b.client, highPriority)
	if err != nil {
		return nil
	}
	return &txbuilder.FeeMeta{
		Fee: decimal.New(int64(gasLimit), 0).Mul(decimal.NewFromBigInt(gasPrice, -geth.Precision)),
	}
}

func (b *ETHBuilder) DoBuild(info *txbuilder.AccountModelBuildInfo) (*txbuilder.TxInfo, error) {
	if b.chainID == nil {
		chainID, err := b.client.ChainID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("get chain id failed, %v", err)
		}

		b.chainID = chainID
	}

	localNextNonce, err := models.NextBlockchainNonce(info.FromAccount.Address, int(b.cfg.Code))
	if err != nil {
		return nil, fmt.Errorf("get local next nonce failed, %v", err)
	}

	fromAddress := common.HexToAddress(info.FromAccount.Address)
	remoteNextNonce, err := b.client.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("get nonce failed, %v", err)
	}

	var (
		nonce     = txbuilder.CalculateNextNonce(info.Task.TxType, info.Task.Nonce, localNextNonce, remoteNextNonce)
		toAddress = common.HexToAddress(info.Task.Address)
		cost      = info.Task.Amount
		taskType  = info.Task.TxType
		bigAmount *big.Int
		payload   []byte
		ok        bool
	)

	if info.Task.Symbol == b.cfg.Currency {

		if info.Task.TxType == models.TxTypeCold {
			info.Task.Amount = cost.Sub(info.FeeMeta.Fee)
			cost = info.Task.Amount
			// task amount less than fee
			if cost.LessThan(decimal.Zero) {
				return nil, nil
			}
		}
		cost = cost.Add(info.FeeMeta.Fee)
		bigAmount, ok = decimalToBigInt(info.Task.Amount.Mul(decimal.New(1, geth.Precision)))
		if !ok {
			return nil, fmt.Errorf("convert task amount %s to bigint failed", info.Task.Amount)
		}

	} else {
		// Token transfer.
		contractAddr, precision, err := contractAddress(info.Task.Symbol, b.cfg.Currency)
		if err != nil {
			return nil, err
		}

		bigAmount, ok = decimalToBigInt(info.Task.Amount.Mul(decimal.New(1, int32(precision))))
		if !ok {
			return nil, fmt.Errorf("convert task amount %s to bigint failed", info.Task.Amount)
		}

		payload, err = geth.PackABIParams("transfer", toAddress, bigAmount)
		if err != nil {
			return nil, fmt.Errorf("pack api params failed, %v", err)
		}

		toAddress = contractAddr
		bigAmount = big.NewInt(0)

	}

	gasLimit, err := estimateGasLimit(b.client, info.Task.Symbol, b.cfg.Currency, fromAddress, &toAddress, bigAmount, payload)
	if err != nil {
		return nil, fmt.Errorf("estimate gas failed, %v", err)
	}

	highPriority := taskType == models.TxTypeWithdraw
	gasPrice, err := suggestGasPrice(b.client, highPriority)
	if err != nil {
		return nil, fmt.Errorf("get suggest gas price failed, %v", err)
	}

	needFee := decimal.New(int64(gasLimit), 0).Mul(decimal.NewFromBigInt(gasPrice, -geth.Precision))
	if needFee.GreaterThan(info.FeeMeta.Fee) {
		return nil, alarm.NewErrFeeNotEnough(info.FeeMeta.Fee, needFee)
	}

	// fix: needFee less than fee, account of balance exchange is incorrect
	if needFee.LessThan(info.FeeMeta.Fee) && info.Task.Symbol == b.cfg.Currency {
		costExchange := info.FeeMeta.Fee.Sub(needFee)
		bigAmountExchange, _ := decimalToBigInt(costExchange.Mul(decimal.New(1, geth.Precision)))

		switch taskType {
		case models.TxTypeWithdraw:
			cost = cost.Sub(costExchange) // update cost
		case models.TxTypeCold, models.TxTypeGather:
			bigAmount = bigAmount.Add(bigAmount, bigAmountExchange) // update amount
		}
	}

	tx := types.NewTransaction(nonce, toAddress, bigAmount, gasLimit, gasPrice, payload)
	rawTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("encode tx failed, %v", err)
	}

	sigDigest := types.NewEIP155Signer(b.chainID).Hash(tx).Bytes()

	log.Warnf("---111--fee:%s,needFee:%s", info.FeeMeta.Fee.String(), needFee.String())
	return &txbuilder.TxInfo{
		Inputs: []*txbuilder.TxIn{&txbuilder.TxIn{
			Account: info.FromAccount,
			Cost:    cost,
		},
		},
		SigPubKeys: []string{hex.EncodeToString(info.FromPubKey)},
		SigDigests: []string{hex.EncodeToString(sigDigest)},
		TxHex:      hex.EncodeToString(rawTx),
		Nonce:      &nonce,
		Fee:        needFee,
	}, nil
}

func contractAddress(symbol, mainCurrency string) (addr common.Address, precision int, err error) {
	currentDetail := currency.CurrencyDetail(symbol)

	if currentDetail == nil {
		err = fmt.Errorf("can't find currency detail of %s", symbol)
		return
	}

	if currentDetail.IsToken() && currentDetail.ChainBelongTo(mainCurrency) {
		addr = common.HexToAddress(currentDetail.Address)
		precision = currentDetail.Decimal
		return

	}

	err = fmt.Errorf("can't find contract address of %s", symbol)
	return
}

func decimalToBigInt(v decimal.Decimal) (*big.Int, bool) {
	return new(big.Int).SetString(v.String(), 10)
}
