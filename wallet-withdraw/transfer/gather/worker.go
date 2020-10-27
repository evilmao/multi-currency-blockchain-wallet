package gather

import (
	"fmt"
	"math/rand"
	"strings"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/service"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"
)

type Worker struct {
	*service.SimpleWorker
	*transfer.Broadcaster
	cfg       *config.Config
	txBuilder txbuilder.Builder

	supplementaryFeeBuilder txbuilder.SupplementaryFeeBuilder

	unsupported map[string]struct{}
}

func New(cfg *config.Config, txBuilder txbuilder.Builder) *Worker {
	w := &Worker{
		Broadcaster: transfer.NewBroadcaster(cfg),
		cfg:         cfg,
		txBuilder:   txBuilder,
		unsupported: make(map[string]struct{}),
	}

	if txBuilder, ok := txBuilder.(txbuilder.SupplementaryFeeBuilder); ok {
		w.supplementaryFeeBuilder = txBuilder
	}

	return w
}

func (w *Worker) Name() string {
	return "gather"
}

func (w *Worker) Work() {
	sysAddress := bmodels.GetSystemAddress()

	if len(sysAddress) == 0 {
		log.Errorf("%s, db get system address failed", w.Name())
		return
	}

	sysAddr := sysAddress[rand.Intn(len(sysAddress))]
	w.gather(sysAddr.Address)
}

func (w *Worker) gather(address string) {
	task := &models.Tx{}
	task.Symbol = strings.ToLower(w.cfg.Currency)
	task.Address = address
	task.TxType = models.TxTypeGather
	task.UpdateLocalTransIDSequenceID()

	txInfo, err := w.txBuilder.BuildGather(task)
	if err != nil {
		switch err := err.(type) {
		case *txbuilder.ErrBalanceForFeeNotEnough:
			if w.supplementaryFeeBuilder != nil {
				w.supplementaryFee(err.Address)
				return
			}
		}

		log.Errorf("%s, build tx %s failed, %v", w.Name(), task, err)
		return
	}

	if txInfo == nil {
		return
	}

	err = transfer.CheckBalanceEnough(w.cfg, txInfo.Inputs)
	if err != nil {
		log.Errorf("%s, check balance enough failed, %v", w.Name(), err)
		return
	}

	log.Infof("%s, start gather tx, %s", w.Name(), task)

	err = w.storeAndBroadcast(txInfo, task)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s, gather tx, %s", w.Name(), task)
}

func (w *Worker) supplementaryFee(toAddress string) {
	if w.supplementaryFeeBuilder == nil {
		return
	}

	feeCode, _ := config.CC.Code(w.supplementaryFeeBuilder.FeeSymbol())
	if exist, err := existUnconfirmedSupplementaryFeeTx(int(feeCode), toAddress); err != nil {
		log.Errorf("%s, check unconfirmed supplementary-fee-tx failed, %v", err)
		return
	} else if exist {
		return
	}

	task := &models.Tx{}
	task.Address = toAddress
	task.TxType = models.TxTypeSupplementaryFee
	task.UpdateLocalTransIDSequenceID()

	txInfo, err := w.supplementaryFeeBuilder.BuildSupplementaryFee(task)
	if err != nil {
		log.Errorf("%s, build supplementaryFee tx %s failed, %v", w.Name(), task, err)
		return
	}

	err = transfer.CheckBalanceEnough(w.cfg, txInfo.Inputs)
	if err != nil {
		log.Errorf("%s, check balance enough failed, %v", w.Name(), err)
		return
	}

	log.Infof("%s, start supplementary fee, %s", w.Name(), task)

	err = w.storeAndBroadcast(txInfo, task)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s, supplementary fee, %s", w.Name(), task)
}

func (w *Worker) storeAndBroadcast(txInfo *txbuilder.TxInfo, task *models.Tx) error {
	task.Hex = txInfo.TxHex
	if txInfo.Nonce != nil {
		task.Nonce = *txInfo.Nonce
	}
	err := task.FirstOrCreate()
	if err != nil {
		return fmt.Errorf("%s, db insert tx failed, %v", w.Name(), err)
	}

	err = w.BroadcastTx(txInfo, task)
	if err != nil {
		return fmt.Errorf("%s, broadcast tx failed, %v", w.Name(), err)
	}

	err = task.Update(map[string]interface{}{
		"tx_status": models.TxStatusBroadcast,
	}, nil)
	if err != nil {
		return fmt.Errorf("%s, db update tx failed, %v", w.Name(), err)
	}

	err = transfer.SpendTxIns(int(w.cfg.Code), task.SequenceID, txInfo.Inputs, txInfo.Nonce, txInfo.DiscardAddress)
	if err != nil {
		return fmt.Errorf("%s, spend (sequenceID: %s) utxo failed, %v", w.Name(), task.SequenceID, err)
	}

	return nil
}

func existUnconfirmedSupplementaryFeeTx(code int, toAddress string) (bool, error) {
	tx, err := models.GetLastSupplementaryFeeTxByAddress(code, toAddress)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	switch tx.TxStatus {
	case models.TxStatusBroadcastFailed, models.TxStatusDiscard:
		return false, nil
	case models.TxStatusBroadcastSuccess, models.TxStatusSuccess:
		if bmodels.GetTxByHash(tx.Hash) {
			return false, nil
		}
	}

	return true, nil
}
