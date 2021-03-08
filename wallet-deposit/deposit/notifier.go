package deposit

import (
	"strings"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/rpc"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"
)

const (
	notifierTag = "deposit.notifier"

	autoNotifyInterval = time.Minute * 10
)

var (
	// notifier must implement service.Worker.
	_ service.Worker = &notifier{}
)

type notifier struct {
	cfg            *config.Config
	rpcClient      rpc.RPC
	exAPI          *api.ExAPI
	needNotify     util.AtomicBool
	lastNotifyTime time.Time
}

func newNotifier(cfg *config.Config, rpcClient rpc.RPC) *notifier {
	return &notifier{
		cfg:            cfg,
		rpcClient:      rpcClient,
		exAPI:          api.NewExAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey),
		lastNotifyTime: time.Now(),
	}
}

func (w *notifier) Name() string {
	return "notifier"
}

func (w *notifier) Init() error {
	w.tryNotify()
	return nil
}

func (w *notifier) Work() {
	if !w.needNotify.Is() {
		if time.Now().Sub(w.lastNotifyTime) < autoNotifyInterval {
			return
		}
	}

	symbol := strings.ToLower(w.cfg.Currency)
	txs := models.GetUnfinishedTxs(symbol)

	for i := range txs {
		tx := &txs[i]
		if tx.IsFinished() {
			continue
		}

		if int(tx.Confirm) < w.cfg.MaxConfirm {
			confirm, err := w.rpcClient.GetTxConfirmations(tx.Hash)
			if err != nil {
				log.Errorf("%s, rpc get tx %s confirmations failed, %v", notifierTag, tx.Hash, err)
				continue
			}
			tx.Confirm = uint16(confirm)
		}

		w.notifyAndAudit(tx)
	}

	w.notifyDone()
}

func (w *notifier) tryNotify() {
	w.needNotify.Set(true)
}

func (w *notifier) notifyDone() {
	w.needNotify.Set(false)
	w.lastNotifyTime = time.Now()
}

// update deposit_tx after request broker api
func (w *notifier) notifyAndAudit(tx *models.Tx) {

	util.Go("notify-audit", func() {
		var (
			notifyRetryCount int
			notifyStatus     int
			err              error
		)

		// for request broker
		txInfo := tx.DepositNotifyFormat()

		// txInfo["app_id"] = w.cfg.BrokerAccessKey
		txInfo["coinName"] = models.TaskSymbolCover(w.cfg.Currency, tx)

		// for update db
		data := make(map[string]interface{})

		// request broker to notify deposit
		_, notifyRetryCount, err = w.exAPI.DepositNotify(txInfo)
		data["notify_retry_count"] = gorm.Expr("`notify_retry_count` + ?", notifyRetryCount)
		monitor.MetricsCount("deposit_notify", 1, monitor.MetricsTags{"currency": tx.Symbol})

		if err != nil {
			log.Errorf("%s, deposit notify failed, %v, retry %d times, txinfo: %v", notifierTag, err, notifyRetryCount, txInfo)

			err = tx.Update(data)
			if err != nil {
				log.Errorf("%s, update tx data %v failed, %v", notifierTag, data, err)
			}
			return
		}

		// request broker success
		if int(tx.Confirm) >= w.cfg.MaxConfirm {
			notifyStatus = 1
			data["notify_status"] = notifyStatus
			data["confirm"] = tx.Confirm
		}

		err = tx.Update(data)
		if err != nil {
			log.Errorf("%s, update tx data %v failed, %v", notifierTag, data, err)
		}
	}, nil)
}

func (w *notifier) Destroy() {
	//
}
