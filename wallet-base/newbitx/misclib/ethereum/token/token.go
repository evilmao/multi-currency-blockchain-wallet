package token

import (
	"fmt"
	"math/big"
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/ethereum/types"
	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TokenABI defines the ERC20 standard token abi.
const TokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

var (
	TransferMethodID = []byte{0xa9, 0x05, 0x9c, 0xbb}
	TransferEventID  = []byte{0xdd, 0xf2, 0x52, 0xad, 0x1b, 0xe2, 0xc8, 0x9b, 0x69, 0xc2, 0xb0, 0x68, 0xfc, 0x37, 0x8d, 0xaa, 0x95, 0x2b, 0xa7, 0xf1, 0x63, 0xc4, 0xa1, 0x16, 0x28, 0xf5, 0x5a, 0x4d, 0xf5, 0x23, 0xb3, 0xef}
)

var (
	tokenABI abi.ABI
)

func init() {
	tokenABI, _ = abi.JSON(strings.NewReader(TokenABI))
}

// PackParams is a wrapper of abi.Pack.
func PackParams(method string, args ...interface{}) ([]byte, error) {
	return tokenABI.Pack(method, args...)
}

// UnpackTransferInput unpacks the erc20 transfer params.
func UnpackTransferInput(input []byte) (common.Address, *big.Int, error) {
	values, err := tokenABI.Methods["transfer"].Inputs.UnpackValues(input)
	if err != nil {
		return common.Address{}, big.NewInt(0), err
	}
	if len(values) == 2 {
		return values[0].(common.Address), values[1].(*big.Int), nil
	}
	return common.Address{}, big.NewInt(0), fmt.Errorf("Not correct transfer parameters")
}

// Transfer makes an erc20 token transfer transaction.
func Transfer(
	nonce uint64, contract, to common.Address,
	amount, gasLimit, gasPrice *big.Int) *types.Transaction {

	payload, err := PackParams("transfer", to, amount)
	if err != nil {
		log.Errorf("Token.PackParams error %+v", err)
		return nil
	}
	return types.NewTransaction(nonce, contract, big.NewInt(0), gasLimit, gasPrice, payload)
}

/*
// TransferMethodID returns the transfer 4-byte id.
func TransferMethodID() []byte {
	//return tokenABI.Methods["transfer"].Id()
	return []byte{0xa9, 0x05, 0x9c, 0xbb}
}

// TransferEventID returns the Transfer event siganture.
func TransferEventID() []byte {
	//return tokenABI.Event["Transfer"].Id()
	//topic[0]="0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	return []byte{0xdd, 0xf2, 0x52, 0xad, 0x1b, 0xe2, 0xc8, 0x9b, 0x69, 0xc2, 0xb0, 0x68, 0xfc, 0x37, 0x8d, 0xaa, 0x95, 0x2b, 0xa7, 0xf1, 0x63, 0xc4, 0xa1, 0x16, 0x28, 0xf5, 0x5a, 0x4d, 0xf5, 0x23, 0xb3, 0xef}
}
*/
