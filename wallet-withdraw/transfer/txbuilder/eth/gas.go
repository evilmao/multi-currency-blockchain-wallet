package eth

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	_minGasLimit      = 21000
	_minTokenGasLimit = 100000

	_minGasPrice          = 1000000000
	_minHighPriorityPrice = _minGasPrice * 10
)

func estimateGasLimit(client *ethclient.Client,
	symbol, mainCurrency string,
	from common.Address, to *common.Address,
	amount *big.Int, payload []byte) (uint64, error) {

	symbol = strings.ToLower(symbol)
	if symbol == mainCurrency {
		gas, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:  from,
			To:    to,
			Value: amount,
			Data:  payload,
		})
		if err != nil {
			return 0, err
		}

		return gas, nil
	}

	return tokenGasLimit(symbol), nil
}

var (
	// Note: Use lower case key.
	_tokenGasLimits = map[string]int64{
		"ionc": 101000,
		"snt":  120000,
		"tnb":  120000,
		"cs":   200000,
		"hgt":  300000,
		"ss":   300000,
		"dnt":  300000,
		"tusd": 300000,
		"cdt":  110000,
	}
)

func tokenGasLimit(symbol string) uint64 {
	if limit, ok := _tokenGasLimits[symbol]; ok {
		return uint64(limit)
	}

	return _minTokenGasLimit
}

func suggestGasPrice(client *ethclient.Client, highPriority bool) (*big.Int, error) {
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	if gasPrice.Cmp(big.NewInt(_minGasPrice)) < 0 {
		gasPrice = big.NewInt(_minGasPrice)
	}

	if highPriority {
		// Use a bigger gasPrice in case of congestion.
		gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))
		if gasPrice.Cmp(big.NewInt(_minHighPriorityPrice)) < 0 {
			gasPrice = big.NewInt(_minHighPriorityPrice)
		}
	}

	return gasPrice, nil
}
