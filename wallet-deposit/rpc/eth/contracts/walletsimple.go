package contracts

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// WalletSimple.sol
// Function: sendMultiSig(
//   address toAddress,
// 	 uint256 value,
// 	 bytes data,
// 	 uint256 expireTime,
// 	 uint256 sequenceId,
// 	 bytes signature)
// https://github.com/BitGo/eth-multisig-v2/blob/master/contracts/WalletSimple.sol

const walletSimpleABI = `[{"constant":false,"inputs":[{"name":"toAddress","type":"address"},{"name":"value","type":"uint256"},{"name":"tokenContractAddress","type":"address"},{"name":"expireTime","type":"uint256"},{"name":"sequenceId","type":"uint256"},{"name":"signature","type":"bytes"}],"name":"sendMultiSigToken","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"}],"name":"signers","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"forwarderAddress","type":"address"},{"name":"tokenContractAddress","type":"address"}],"name":"flushForwarderTokens","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"toAddress","type":"address"},{"name":"value","type":"uint256"},{"name":"data","type":"bytes"},{"name":"expireTime","type":"uint256"},{"name":"sequenceId","type":"uint256"},{"name":"signature","type":"bytes"}],"name":"sendMultiSig","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"signer","type":"address"}],"name":"isSigner","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"getNextSequenceId","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"createForwarder","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"safeMode","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"activateSafeMode","outputs":[],"payable":false,"type":"function"},{"inputs":[{"name":"allowedSigners","type":"address[]"}],"payable":false,"type":"constructor"},{"payable":true,"type":"fallback"},{"anonymous":false,"inputs":[{"indexed":false,"name":"from","type":"address"},{"indexed":false,"name":"value","type":"uint256"},{"indexed":false,"name":"data","type":"bytes"}],"name":"Deposited","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"msgSender","type":"address"}],"name":"SafeModeActivated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"msgSender","type":"address"},{"indexed":false,"name":"otherSigner","type":"address"},{"indexed":false,"name":"operation","type":"bytes32"},{"indexed":false,"name":"toAddress","type":"address"},{"indexed":false,"name":"value","type":"uint256"},{"indexed":false,"name":"data","type":"bytes"}],"name":"Transacted","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"msgSender","type":"address"},{"indexed":false,"name":"otherSigner","type":"address"},{"indexed":false,"name":"operation","type":"bytes32"},{"indexed":false,"name":"toAddress","type":"address"},{"indexed":false,"name":"value","type":"uint256"},{"indexed":false,"name":"tokenContractAddress","type":"address"}],"name":"TokenTransacted","type":"event"}]`

var (
	walletSimple abi.ABI
)

func init() {
	walletSimple, _ = abi.JSON(strings.NewReader(walletSimpleABI))
}

// WalletSimpleCaller represents a unpacker for WalletSimple.sol.
type WalletSimpleCaller struct {
	MethodName string
}

// MethodID returns the sendMultiSig method id.
func (c WalletSimpleCaller) MethodID() []byte {
	// return WalletSimple.Methods["sendMultiSig"].Id()
	return []byte{0x39, 0x12, 0x52, 0x15}
}

// UnpackParams unpacks the contract method inputs and returns the dest address and amount.
func (c WalletSimpleCaller) UnpackParams(input []byte) (common.Address, *big.Int, error) {
	values, err := walletSimple.Methods[c.MethodName].Inputs.UnpackValues(input)
	return values[0].(common.Address), values[1].(*big.Int), err
}
