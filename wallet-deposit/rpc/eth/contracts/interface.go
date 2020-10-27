package contracts

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// contractCallers is a whitelist of eth internal transfer contract addresses.
	// Note: address must be lower case.
	contractCallers = map[string]ContractCaller{
		"0xfd26d9ec46172759ff9d2039a4e841cd82e835e9": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x1522900b6dafac587d499a862861c0869be6e428": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0xd4f5bf184bebfd53ac276ec6e091d051d0ed459e": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0xb890abef585b25536e8c32a738ce1fd9b4ccbda6": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x045befa474588abfca096b3086c44421d0d09716": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x30b71d015f60e2f959743038ce0aaec9b4c1ea44": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x75db8b92937f8f86213e523788ab9f066efde3fe": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x72298bd3ed75e0aa289ad61c6390596d60ffdcef": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0xd6a062cae6123c158768a5c444ca0896cc60d6b1": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x371016430c6d0e230d51a7b88e8f9f1d49c43a8a": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x3fbe1f8fc5ddb27d428aa60f661eaaab0d2000ce": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x121effb8160f7206444f5a57d13c7a4424a237a4": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0xe5cafdf7cc3c74a33fa0e829680b1760f3003fb4": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0x560e389a2b032319e742a59ae8bafa62671089fe": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
		"0xd1560b3984b7481cd9a8f40435a53c860187174d": WalletSimpleCaller{
			MethodName: "sendMultiSig",
		},
	}
)

// ContractCaller represents a contract have method that transfer eth.
type ContractCaller interface {
	MethodID() []byte
	UnpackParams([]byte) (common.Address, *big.Int, error)
}

// Caller returns contract caller.
func Caller(address string) (ContractCaller, bool) {
	address = strings.ToLower(address)
	caller, ok := contractCallers[address]
	return caller, ok
}
