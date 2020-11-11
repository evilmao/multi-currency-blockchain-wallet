package checker

import (
	"strings"

	"upex-wallet/wallet-config/withdraw/transfer/config"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

var (
	checkers = make(map[string]map[string]Checker)
)

type Checker interface {
	Name() string
	Init(*config.Config)
	Check() error
}

func Add(currencyType string, checker Checker) {
	currencyType = strings.ToUpper(currencyType)
	cs, ok := checkers[currencyType]
	if !ok {
		cs = make(map[string]Checker)
		checkers[currencyType] = cs
	}

	if _, ok := cs[checker.Name()]; ok {
		log.Errorf("checker.Add, duplicate of %s checker %s\n", currencyType, checker.Name())
		return
	}

	cs[checker.Name()] = checker
}

func Find(currencyType string) map[string]Checker {
	currencyType = strings.ToUpper(currencyType)
	return checkers[currencyType]
}
