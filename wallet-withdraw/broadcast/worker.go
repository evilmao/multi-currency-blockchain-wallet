package broadcast

import (
	"fmt"
	"strings"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/currency"
	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/broadcast/handler"
	"upex-wallet/wallet-withdraw/broadcast/types"

	"github.com/jinzhu/gorm"
)

const (
	taskLen       = 20000
	maxRetryTimes = 300
)

var (
	ErrWithdrawAudit  = fmt.Errorf("withdraw audit failed")
	ErrWithdrawNotify = fmt.Errorf("withdraw notify failed")
)

var errWhiteList = []error{
	ErrWithdrawAudit,
	ErrWithdrawNotify,
	handler.ErrBuildTxBusy,
}

type Task struct {
	record *models.BroadcastTask
	args   *types.QueryArgs
	h      handler.Handler

	tx               handler.Tx
	txID             string
	broadcasted      bool
	broadcastSuccess bool
	retry            uint16
}

type Worker struct {
	exAPI       *api.ExAPI
	taskCh      chan *Task
	retryTaskCh chan *Task

	lastUpdateTime time.Time
}

func New(exAPI *api.ExAPI) *Worker {
	return &Worker{
		exAPI:       exAPI,
		taskCh:      make(chan *Task, taskLen),
		retryTaskCh: make(chan *Task, taskLen),
	}
}

func (w *Worker) Name() string {
	return "broadcast"
}

func (w *Worker) Init() error {
	util.Go("loadUnfinishedTasks", w.loadUnfinishedTasks, nil)
	return nil
}

func (w *Worker) loadUnfinishedTasks() {
	loaded := map[string]struct{}{}
	_ = handler.Foreach(func(currency string, h handler.Handler) error {
		if h.DB() == nil {
			log.Warnf("currency %s has no db", currency)
			return nil
		}

		if _, exist := loaded[h.DSN()]; exist {
			return nil
		}
		loaded[h.DSN()] = struct{}{}

		tasks := models.GetBroadcastTasksByStatus(h.DB(), models.BroadcastTaskStatusRecord)
		for _, task := range tasks {
			tx, err := models.GetTxBySequenceID(h.DB(), task.TxSequenceID)
			if err != nil {
				log.Errorf("%s, db load tx sequenceID = %s failed, %v", w.Name(), task.TxSequenceID, err)
				continue
			}

			_ = w.Add(&types.QueryArgs{
				Task:       *tx,
				Signatures: strings.Split(task.TxSignatures, models.TaskTxSigPubKeySep),
				PubKeys:    strings.Split(task.TxPubKeys, models.TaskTxSigPubKeySep),
			}, task)
		}
		return nil
	})
}

func (w *Worker) Add(args *types.QueryArgs, task *models.BroadcastTask) error {
	if len(w.taskCh) == taskLen {
		return fmt.Errorf("the task queue is full, please retry after a while")
	}

	h, err := findHandler(&args.Task)
	if err != nil {
		return err
	}

	if h.DB() == nil {
		return fmt.Errorf("can't find db of handler %s", args.Task.Symbol)
	}

	switch args.Task.TxType {
	case models.TxTypeGather:
		if !bmodels.IsSystemAddress(h.DB(), args.Task.Address) {
			return fmt.Errorf("gather transfer to %s is not allowed", args.Task.Address)
		}
	case models.TxTypeSupplementaryFee:
		if !bmodels.IsNormalAddress(h.DB(), args.Task.Address) {
			return fmt.Errorf("supplementary-fee transfer to %s is not allowed", args.Task.Address)
		}
	case models.TxTypeCold:
		if args.Task.Address == "" {
			return fmt.Errorf("cold transfer address can not be empty ")
		}
	}

	if task == nil {
		if exist, _ := models.IsBroadcastTaskExist(h.DB(), args.Task.SequenceID); exist {
			return nil
		}

		task = &models.BroadcastTask{
			TxSequenceID: args.Task.SequenceID,
			TxSignatures: strings.Join(args.Signatures, models.TaskTxSigPubKeySep),
			TxPubKeys:    strings.Join(args.PubKeys, models.TaskTxSigPubKeySep),
		}

		err := util.TryWithInterval(3, time.Second, func(int) error {
			return task.FirstOrCreate(h.DB())
		})
		if err != nil {
			return fmt.Errorf("db insert task failed, %v", err)
		}
	}

	w.taskCh <- &Task{
		record:           task,
		args:             args,
		h:                h,
		txID:             args.Task.Hash,
		broadcasted:      args.Task.TxStatus == models.TxStatusBroadcastSuccess,
		broadcastSuccess: args.Task.TxStatus == models.TxStatusBroadcastSuccess,
		retry:            args.Task.BroadcastTryCount,
	}
	return nil
}

func (w *Worker) Work() {
	var t *Task
	select {
	case t = <-w.taskCh:
	case t = <-w.retryTaskCh:
	default:
		if time.Now().Sub(w.lastUpdateTime) >= time.Minute {
			log.Infof("%s, wait for tasks...", w.Name())
			w.lastUpdateTime = time.Now()
		}
	}

	if t == nil {
		return
	}

	w.lastUpdateTime = time.Now()

	util.Go("broadcast.process", func() {
		w.process(t)
	}, nil)
}

func (w *Worker) process(t *Task) {
	err := w.doProcess(t)
	if err != nil {
		// log.Errorf("process task (%s) failed (retry:%d/%d), %v", &t.args.Task, t.retry, maxRetryTimes, err)
		w.retry(t, costRetryTimes(err))
	}
}

func (w *Worker) doProcess(t *Task) error {
	var (
		args = t.args
		err  error
	)

	switch args.Task.TxType {
	case models.TxTypeWithdraw:
		exAPI := w.exAPI
		if exAPI == nil {
			return fmt.Errorf("invalid exchange ")
		}

		if !t.broadcastSuccess {
			// send tx
			err = w.tryBroadcast(t)
			if err != nil {
				return fmt.Errorf("broadcast transaction failed, %v", err)
			}

			// notify
			data := args.Task.WithdrawNotifyFormat()
			data["app_id"] = exAPI.GetBrokerAppID()
			_, _, err := exAPI.WithdrawNotify(data)

			if err != nil {
				return fmt.Errorf("%s, %v", ErrWithdrawNotify, err)
			}
		}
	case models.TxTypeGather, models.TxTypeSupplementaryFee, models.TxTypeCold:
		err = w.tryBroadcast(t)
		if err != nil {
			return fmt.Errorf("broadcast transaction failed, %v", err)
		}
	default:
		t.retry = maxRetryTimes
		return fmt.Errorf("unsupport tx type: %d", args.Task.TxType)
	}

	err = util.TryWithInterval(3, time.Second, func(int) error {
		err := t.record.Done(t.h.DB())
		if err != nil {
			return err
		}

		return args.Task.Update(map[string]interface{}{
			"tx_status": models.TxStatusSuccess,
			"confirm":   t.h.Ctrler().Confirms(),
		}, t.h.DB())
	})
	if err != nil {
		log.Errorf("db update tx status to success failed, %v", err)
	}

	log.Infof("process tx success, %s", &args.Task)
	return nil
}

func (w *Worker) tryBroadcast(t *Task) error {
	// 1. check task is completed
	if t.broadcastSuccess {
		return nil
	}

	args := t.args
	err := args.Task.Update(map[string]interface{}{
		"broadcast_try_count": gorm.Expr("`broadcast_try_count` + 1"),
	}, t.h.DB())
	if err != nil {
		log.Errorf("broadcast update count failed,%v", err)
	}

	updateTxID := func(txID string) {
		if len(txID) > 0 {
			t.txID = txID

			err := util.TryWithInterval(3, time.Second, func(int) error {
				return args.Task.Update(map[string]interface{}{
					"txid": txID,
				}, t.h.DB())
			})
			if err != nil {
				log.Errorf("db update tx id failed, %v", err)
			}
		}
	}

	// build a new tx by broadTask
	if t.tx == nil {
		tx, txID, err := t.h.BuildTx(args.Task.Hex, args.Signatures, args.PubKeys)
		if err != nil {
			if costRetryTimes(err) {
				t.retry = maxRetryTimes
			}
			return err
		}

		t.tx = tx
		if len(txID) > 0 {
			updateTxID(txID)
		}
	}

	// 2. check task is broadcast
	if !t.broadcasted {
		txID, err := t.h.BroadcastTransaction(t.tx, t.txID)
		if err != nil {
			return err
		}

		t.broadcasted = true
		if len(txID) > 0 {
			updateTxID(txID)
		}

		// wait for online confirm
		time.Sleep(t.h.Ctrler().VerifyInterval())
	}

	// 3. verify broadcast is success
	if !t.h.VerifyTxBroadCasted(t.txID) {
		_, _ = t.h.BroadcastTransaction(t.tx, t.txID)
		return fmt.Errorf("can't find %s tx %s in blockchain yet", args.Task.Symbol, t.txID)
	}

	// 4. update tx status and confirm_times
	err = util.TryWithInterval(3, time.Second, func(int) error {
		return args.Task.Update(map[string]interface{}{
			"tx_status": models.TxStatusBroadcastSuccess,
			"confirm":   t.h.Ctrler().Confirms(),
		}, t.h.DB())
	})
	if err != nil {
		log.Errorf("db update tx status to broadcast failed, %v", err)
	}

	log.Infof("broadcast tx success, %s", &args.Task)

	t.broadcastSuccess = true
	return nil
}

func (w *Worker) retry(t *Task, costRetry bool) {
	if t.retry >= maxRetryTimes {
		err := util.TryWithInterval(3, time.Second, func(int) error {
			// update broadcast task status, stop broadcast
			err := t.record.Done(t.h.DB())
			if err != nil {
				return err
			}

			// update tx task status, record task is "fail"
			return t.args.Task.Update(map[string]interface{}{
				"tx_status": models.TxStatusBroadcastFailed,
			}, t.h.DB())
		})
		if err != nil {
			log.Errorf("db update tx status to broadcast_failed failed, %v", err)
		}
		return
	}

	//
	time.Sleep(t.h.Ctrler().VerifyInterval())

	if costRetry {
		t.retry++
	}
	w.retryTaskCh <- t
}

func (w *Worker) Destroy() {
	//
}

func costRetryTimes(err error) bool {
	for _, e := range errWhiteList {
		if strings.Contains(err.Error(), e.Error()) {
			return false
		}
	}
	return true
}

func findHandler(task *models.Tx) (handler.Handler, error) {
	h, ok := handler.Find(task.Symbol)
	if ok {
		return h, nil
	}

	// Tokens use main-chain's handler.
	currencyDetail := currency.CurrencyDetail(task.Symbol)

	h, ok = handler.Find(currencyDetail.BlockchainName)

	if ok {
		return h, nil
	}

	return nil, fmt.Errorf("can't find handler of %s", task.Symbol)

}
