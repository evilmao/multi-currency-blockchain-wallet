package bitcoin

import (
	"strings"
)

func init() {
	addSupportFullDataRPC("btc")
	addSupportFullDataRPC("bsv")
	addSupportFullDataRPC("bch")
	addFastRollbackBlock("abbc")
	addFastRollbackBlock("fab")
	addFastRollbackBlock("ltc")
}

// Slot is just a placeholder for map-set.
type Slot struct{}

var (
	slot = Slot{}

	useFullDataRPCCurrencies        = make(map[string]Slot)
	needFastRollbackBlockCurrencies = make(map[string]Slot)
)

func addSupportFullDataRPC(c string) {
	c = strings.ToUpper(c)
	useFullDataRPCCurrencies[c] = slot
}

func supportFullDataRPC(c string) bool {
	c = strings.ToUpper(c)
	_, ok := useFullDataRPCCurrencies[c]
	return ok
}

func addFastRollbackBlock(c string) {
	c = strings.ToUpper(c)
	needFastRollbackBlockCurrencies[c] = slot
}

func needFastRollbackBlock(c string) bool {
	c = strings.ToUpper(c)
	_, ok := needFastRollbackBlockCurrencies[c]
	return ok
}
