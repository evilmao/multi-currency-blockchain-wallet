package syncer

import (
	"strings"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/deposit"

	"github.com/jinzhu/gorm"
)

// BaseFetcher parses blockchain data and notifies exchange.
// It's a basic implementation, and any other currency can be implemented based on it.
type BaseFetcher struct {
	ExAPI              *api.ExAPI
	Cfg                *config.Config
	TxCh               chan *models.Tx
	GetTxConfirmations func(h string) (uint64, error)
	needNotify         util.AtomicBool
	Run                bool
}

// NewFetcher returns fetcher instance.
func NewFetcher(api *api.ExAPI, cfg *config.Config) *BaseFetcher {

	return &BaseFetcher{
		ExAPI: api,
		Cfg:   cfg,
		Run:   true,
		TxCh:  make(chan *models.Tx, 2000),
	}
}

// ImportBlock imports new block.
func (f *BaseFetcher) ImportBlock(block models.BlockInfo) error {
	if _, ok := models.GetBlockInfoByHeight(block.Height, f.Cfg.Currency, f.Cfg.UseBlockTable); ok {
		return nil
	}
	monitor.MetricsGauge("block_height", float64(block.Height), monitor.MetricsTags{"currency": f.Cfg.Currency})

	block.Symbol = f.Cfg.Currency
	err := block.Insert()
	if err != nil {
		return err
	}

	f.needNotify.Set(true)
	return nil
}

// ProcessOrphanBlock processes orphan-block.
func (f *BaseFetcher) ProcessOrphanBlock(block models.BlockInfo) error {
	log.Warnf("rollback block, height: %d, hash: %s", block.Height, block.Hash)
	block.Symbol = f.Cfg.Currency
	return block.Delete()
}

// Close release resource.C
func (f *BaseFetcher) Close() {
	f.Run = false
}

// reprocessDeposit re-notify tx to exchange.
func (f *BaseFetcher) reprocessDeposit() {

	for f.Run {
		if !f.needNotify.Is() {
			time.Sleep(10 * time.Second)
			continue
		}

		symbol := strings.ToLower(f.Cfg.Currency)
		txs := models.GetUnfinishedTxs(symbol)
		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			if int(tx.Confirm) < f.Cfg.MaxConfirm {
				confirm, err := f.GetTxConfirmations(tx.Hash)
				if err != nil {
					log.Errorf("get tx %s confirmations failed, %v", tx.Hash, err)
					continue
				}
				tx.Confirm = uint16(confirm)
			}
			f.TxCh <- &tx
		}

		f.needNotify.Set(false)
		time.Sleep(10 * time.Second)
	}
	close(f.TxCh)
}

// DepositSchedule notify tx to exchange.
func (f *BaseFetcher) DepositSchedule() {
	defer util.DeferRecover("DepositSchedule", nil)()

	// just store deposit detail,not need notify to broker
	if !f.Cfg.IgnoreNotifyAudit {
		f.needNotify.Set(true)
		util.Go("reprocessDeposit", f.reprocessDeposit, nil)
	}

L:
	for f.Run {
		var (
			tx *models.Tx
			ok bool
		)

		select {
		case tx, ok = <-f.TxCh:
		case <-time.After(time.Second):
			continue L
		}

		if !ok {
			break
		}

		if f.Cfg.IgnoreNotifyAudit {
			continue
		}

		if tx.IsFinished() {
			continue
		}

		if f.Cfg.IsNeedTag && !deposit.ValidTxTag(tx.Extra, tx.Symbol) {
			log.Infof("ignore invalid-tag deposit tx, %s", deposit.TxString(tx))

			_ = tx.Update(models.M{
				"confirm":       tx.Confirm,
				"notify_status": 1,
			})
			continue
		}

		if min, _ := currency.MinAmount(tx.Symbol); tx.Amount.LessThan(min) {
			log.Infof("ignore min-amount deposit tx, %s", deposit.TxString(tx))

			_ = tx.Update(models.M{
				"confirm":       tx.Confirm,
				"notify_status": 1,
			})
			continue
		}

		log.Infof("Detect deposit tx, %s", deposit.TxString(tx))

		util.Go("audit-notify", func() {
			// audit
			var (
				notifyRetryCount int
				notifyStatus     int
				err              error
			)

			// for request broker
			txInfo := tx.DepositNotifyFormat()
			txInfo["app_id"] = f.Cfg.BrokerAccessKey
			txInfo["symbol"] = f.Cfg.Currency

			// for update db
			data := make(map[string]interface{})
			data["confirm"] = tx.Confirm

			// request broker to notify deposit
			_, notifyRetryCount, err = f.ExAPI.DepositNotify(txInfo)
			data["notify_retry_count"] = gorm.Expr("`notify_retry_count` + ?", notifyRetryCount)
			monitor.MetricsCount("deposit_notify", 1, monitor.MetricsTags{"currency": tx.Symbol})

			if err != nil {
				log.Errorf("Deposit notify request failed, %v, retry %d times, txinfo: %v", err, notifyRetryCount, txInfo)

				err = tx.Update(data)
				if err != nil {
					log.Errorf("update tx data %v failed, %v", data, err)
				}
				return
			}

			// request broker success
			if int(tx.Confirm) >= f.Cfg.MaxConfirm {
				notifyStatus = 1
				data["notify_status"] = notifyStatus
			}

			err = tx.Update(data)
			if err != nil {
				log.Errorf(",update tx data %v failed, %v", data, err)
			}
		}, nil)
	}

	log.Debugf("DepositSchedule exit")
}
