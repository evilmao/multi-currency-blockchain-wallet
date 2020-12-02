package cmd

import (
	"fmt"
	"strings"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
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

func ChoseRunnable(cfg *config.Config) error {

	runType, ok := Find(cfg.Currency)
	if !ok {
		return fmt.Errorf("chose runnable fail, currency %s is not exits", cfg.Currency)
	}

	runType0 := func(run Runnable) {
		restartTimes := 0
		for {
			util.WithRecover("deposit-run", func() {
				run(cfg, restartTimes)
			}, nil)

			time.Sleep(2 * time.Second)
			restartTimes++
			log.Errorf("%s deposit Service Restart %d Times", strings.ToUpper(cfg.Currency), restartTimes)
		}
	}

	switch runType.Type {
	case 0:
		runType0(runType.Runnable)
	case 1:
		runType.Runnable(cfg, 1)
	default:
		return fmt.Errorf("runnable type is wrong,should 1 or 0")
	}

	return nil
}
