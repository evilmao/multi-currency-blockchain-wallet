package checker

import (
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"
)

var (
	checkers = make(map[string]map[string]Checker)
)

type Checker interface {
	Name() string
	Init(*config.Config)
	Check() error
}

func Add(checkerType string, checker Checker) {
	checkerType = strings.ToUpper(checkerType)
	cs, ok := checkers[checkerType]
	if !ok {
		cs = make(map[string]Checker)
		checkers[checkerType] = cs
	}

	if _, ok := cs[checker.Name()]; ok {
		log.Errorf("checker.Add, duplicate of %s checker %s\n", checkerType, checker.Name())
		return
	}

	cs[checker.Name()] = checker
}

func Find(checkerType string) map[string]Checker {
	checkerType = strings.ToUpper(checkerType)
	return checkers[checkerType]
}
