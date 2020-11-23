package checker

import (
	"strings"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/checker/checker"
	_ "upex-wallet/wallet-withdraw/transfer/checker/checker/calculator"
)

type Worker struct {
	*service.SimpleWorker
	cfg      *config.Config
	checkers []checker.Checker
}

func New(cfg *config.Config) *Worker {
	return &Worker{
		cfg: cfg,
	}
}

func (w *Worker) Name() string {
	return "checker"
}

func (w *Worker) Init() error {
	if len(w.cfg.ScheduleChecker) == 0 {
		return nil
	}

	checkers := checker.Find(w.cfg.Currency)
	if len(checkers) == 0 {
		return nil
	}

	var names []string
	if strings.EqualFold(w.cfg.ScheduleChecker[0], "all") {
		for _, c := range checkers {
			c.Init(w.cfg)
			w.checkers = append(w.checkers, c)
			names = append(names, c.Name())
		}
	} else {
		for _, name := range w.cfg.ScheduleChecker {
			c, ok := checkers[name]
			if ok {
				c.Init(w.cfg)
				w.checkers = append(w.checkers, c)
				names = append(names, c.Name())
			}
		}
	}

	if len(names) > 0 {
		log.Infof("%s, active %v", w.Name(), names)
	}

	return nil
}

func (w *Worker) Work() {
	for _, c := range w.checkers {
		log.Warnf("check of %s start", c.Name())
		err := c.Check()
		if err != nil {
			log.Errorf("%s, %s, %v", w.Name(), c.Name(), err)
		}
	}
}
