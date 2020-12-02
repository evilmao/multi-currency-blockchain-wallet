package cmd

import (
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/deposit/config"
)

// Runnable def.
type Runnable func(*config.Config, int)

type RunType struct {
	Runnable
	Type int
}

var (
	RunTypeMap = make(map[string]*RunType)
)

func NewRunType(runType int, run Runnable) *RunType {
	return &RunType{
		run,
		runType,
	}
}

func Register(currencyType string, runType *RunType) {
	currencyType = strings.ToUpper(currencyType)
	if _, ok := Find(currencyType); ok {
		log.Errorf("runnable.Register, duplicate of %s\n", currencyType)
		return
	}
	log.Infof("Register runnable success:[%s]", currencyType)
	RunTypeMap[currencyType] = runType
}

func Find(currencyType string) (*RunType, bool) {
	currencyType = strings.ToUpper(currencyType)
	c, ok := RunTypeMap[currencyType]
	return c, ok
}
