package checker

import (
	"fmt"
	"time"

	"upex-wallet/wallet-tools/cmd/statuschecker/checker/heightgetters"

	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

type StuckChecker struct {
	service.SimpleWorker
	name          string
	url           string
	getter        heightgetters.Getter
	updateMonitor *util.UpdateMonitor
}

func NewStuckChecker(name, url string) *StuckChecker {
	return &StuckChecker{
		name: name,
		url:  url,
	}
}

func (c *StuckChecker) Init() error {
	getter, ok := heightgetters.Get(c.name)
	if !ok {
		return fmt.Errorf("can't find height getter of %s", c.name)
	}
	c.getter = getter

	c.updateMonitor = util.NewAutoAdjustUpdateMonitor(time.Minute*30, func(v int64, d time.Duration) {
		body := fmt.Sprintf("%s block stucks at height %d for %v", c.name, v, d)
		warn(body)
	})
	return nil
}

func (c *StuckChecker) Name() string {
	return "StuckChecker of " + c.name
}

func (c *StuckChecker) Work() {
	height, err := c.getter(c.url)
	if err != nil {
		log.Errorf("get %s height failed, %v", c.name, err)
		c.updateMonitor.Check()
		return
	}

	c.updateMonitor.Update(height)
}
