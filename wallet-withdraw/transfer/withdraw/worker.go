package withdraw

import (
	"fmt"

	bapi "upex-wallet/wallet-base/api"
	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/service"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/shopspring/decimal"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer"
	"upex-wallet/wallet-withdraw/transfer/alarm"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
)

type Worker struct {
	*transfer.Broadcaster
	cfg             *config.Config
	exAPI           *bapi.ExAPI
	txBuilder       txbuilder.Builder
	taskProducer    *taskProducer
	taskProducerSrv *service.Service
}

func New(cfg *config.Config, txBuilder txbuilder.Builder) *Worker {
	// init broker api
	exAPI := bapi.NewExAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey)

	producer := newTaskProducer(cfg, exAPI)
	return &Worker{
		Broadcaster:     transfer.NewBroadcaster(cfg),
		cfg:             cfg,
		exAPI:           exAPI,
		txBuilder:       txBuilder,
		taskProducer:    producer,
		taskProducerSrv: service.New(producer),
	}
}

func (w *Worker) Name() string {
	return "withdraw"
}

func (w *Worker) Init() error {
	w.updateTxHashIfNeeded()
	w.rejectWithdrawIfNeeded()

	go w.taskProducerSrv.Start()
	return nil
}

func (w *Worker) updateTxHashIfNeeded() {

	txSequenceID := w.cfg.TxSequenceIDForUpdateTxHashToExchange
	if len(txSequenceID) == 0 {
		return
	}
	const desc = "try to update tx hash"

	tx, err := models.GetTxBySequenceID(nil, txSequenceID)
	if err != nil {
		log.Errorf("%s, %s, load tx of sequence id %s failed, %v",
			w.Name(), desc, txSequenceID, err)
		return
	}

	if len(tx.Hash) == 0 {
		log.Errorf("%s, %s, sequence id %s, hash is empty",
			w.Name(), desc, txSequenceID)
		return
	}

	data := tx.WithdrawNotifyFormat()
	data["app_id"] = w.cfg.BrokerAccessKey

	err = w.exAPI.DangerousAPIForUpdateTxHash(tx.TransID, tx.Hash, data)
	if err != nil {
		log.Errorf("%s, %s failed, %v", w.Name(), desc, err)
		return
	}

	log.Warnf("%s, update tx hash to exchange, %s", w.Name(), tx)
}

func (w *Worker) rejectWithdrawIfNeeded() {
	txSequenceID := w.cfg.TxSequenceIDForRejectToExchange
	if len(txSequenceID) == 0 {
		return
	}

	const desc = "try to reject withdraw tx"

	tx, err := models.GetTxBySequenceID(nil, txSequenceID)
	if err != nil {
		log.Errorf("%s, %s, load tx of sequence id %s failed, %v",
			w.Name(), desc, txSequenceID, err)
		return
	}

	err = tx.Update(map[string]interface{}{
		"tx_status": models.TxStatusReject,
	}, nil)
	if err != nil {
		log.Errorf("%s, %s, db update tx status failed, %v",
			w.Name(), desc, err)
		return
	}

	log.Warnf("%s, reject withdraw tx to exchange, %s", w.Name(), tx)
}

func (w *Worker) Work() {

	task, ok := w.taskProducer.next()
	if !ok {
		return
	}

	err := w.processTask(task)
	if err != nil {
		log.Errorf("%s, process task failed, %v", w.Name(), err)
	}
}

// Destroy ,stop a work
func (w *Worker) Destroy() {
	w.taskProducerSrv.Stop()
}

// process a withdraw task
func (w *Worker) processTask(task *models.Tx) error {
	if tx, err := models.GetWithdrawTxByTransID(task.TransID); err == nil && tx.TxStatus > models.TxStatusRecord {
		return nil
	}

	log.Infof("%s, start process task %+v", w.Name(), task)

	// request broker api
	data := task.WithdrawNotifyFormat()
	_, _, err := w.exAPI.WithdrawNotify(data)
	if err != nil {
		return fmt.Errorf("withdraw notify failed, %v", err)
	}

	// withdraw lessThan 0 is not available
	if task.Amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// check if a new trans
	if task.TxStatus == models.TxStatusNotRecord {
		err = task.FirstOrCreate()
		if err != nil {
			return fmt.Errorf("db insert tx failed, %v", err)
		}
	}

	// balance verify
	total := task.Amount
	// sum balance from all system accounts.
	balance := bmodels.GetSystemBalance()

	if balance == nil || balance.LessThan(total) {
		// send email
		e := alarm.NewErrorBalanceNotEnough(balance, total)
		go alarm.SendEmail(w.cfg, task, e, e.ErrorDetail)

		return fmt.Errorf("wallet balance not enough, balance %v need %v", balance, total)
	}

	// will submit a transaction
	txInfo, err := w.txBuilder.BuildWithdraw(task)
	if err != nil {
		return fmt.Errorf("build tx failed, %v", err)
	}

	if txInfo == nil {
		return fmt.Errorf("build tx failed, txInfo is nil")
	}
	// check balance
	err = transfer.CheckBalanceEnough(txInfo.Inputs)
	if err != nil {
		return fmt.Errorf("check balance enough failed, %v", err)
	}

	if txInfo.Nonce != nil {
		err = task.Update(map[string]interface{}{
			"hex":   txInfo.TxHex,
			"nonce": *txInfo.Nonce,
		}, nil)
	} else {
		err = task.Update(map[string]interface{}{
			"hex": txInfo.TxHex,
		}, nil)
	}
	if err != nil {
		return fmt.Errorf("db update tx hex failed, %v", err)
	}

	// signer and broadcast process
	err = w.BroadcastTx(txInfo, task)
	if err != nil {
		return fmt.Errorf("broadcast tx failed, %v", err)
	}

	// update task process status
	err = task.Update(map[string]interface{}{
		"tx_status": models.TxStatusBroadcast,
	}, nil)
	if err != nil {
		return fmt.Errorf("db update tx status failed, %v", err)
	}

	err = transfer.SpendTxIns(int(w.cfg.Code), task.SequenceID, txInfo.Inputs, txInfo.Nonce, txInfo.DiscardAddress)
	if err != nil {
		return err
	}

	log.Infof("%s, withdraw tx, %s", w.Name(), task)
	return nil
}
