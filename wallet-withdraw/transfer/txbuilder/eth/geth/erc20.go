package geth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// _erc20ABI defines the ERC20 standard token abi.
const _erc20ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

var (
	_abi *abi.ABI
)

// PackABIParams is a wrapper of abi.Pack.
func PackABIParams(method string, args ...interface{}) ([]byte, error) {
	if _abi == nil {
		a, err := abi.JSON(strings.NewReader(_erc20ABI))
		if err != nil {
			return nil, fmt.Errorf("parse token api failed, %v", err)
		}

		_abi = &a
	}

	return _abi.Pack(method, args...)
}

func CallERC20(client *ethclient.Client, contractAddress string, method string, args ...interface{}) ([]byte, error) {
	payload, err := PackABIParams(method, args...)
	if err != nil {
		return nil, err
	}

	contractAddr := common.HexToAddress(contractAddress)
	return client.CallContract(context.Background(), ethereum.CallMsg{
		From:  common.Address{},
		To:    &contractAddr,
		Value: big.NewInt(0),
		Data:  payload,
	}, nil)
}

func GetERC20Symbol(client *ethclient.Client, contractAddress string) (string, error) {
	result, err := CallERC20(client, contractAddress, "symbol")
	if err != nil {
		return "", err
	}

	const paddedSize = 32
	if len(result) != paddedSize*3 {
		return "", fmt.Errorf("invalid erc20 symbol format")
	}

	n := new(big.Int).SetBytes(result[paddedSize : paddedSize*2]).Int64()
	symbol := result[paddedSize*2 : paddedSize*2+n]
	return string(symbol), nil
}

func GetERC20Precision(client *ethclient.Client, contractAddress string) (int64, error) {
	result, err := CallERC20(client, contractAddress, "decimals")
	if err != nil {
		return 0, err
	}

	return new(big.Int).SetBytes(result).Int64(), nil
}

func GetERC20Balance(client *ethclient.Client, contractAddress, address string, precision int32) (decimal.Decimal, error) {
	addr := common.HexToAddress(address)
	result, err := CallERC20(client, contractAddress, "balanceOf", addr)
	if err != nil {
		return decimal.Zero, err
	}

	balance := new(big.Int).SetBytes(result)
	return decimal.NewFromBigInt(balance, -int32(precision)), nil
}
