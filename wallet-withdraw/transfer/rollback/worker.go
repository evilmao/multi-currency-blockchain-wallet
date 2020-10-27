package rollback

import (
	"upex-wallet/wallet-base/db"
	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
)

type Worker struct {
	*service.SimpleWorker
	cfg *config.Config
}

func New(cfg *config.Config) *Worker {
	return &Worker{
		cfg: cfg,
	}
}

func (w *Worker) Name() string {
	return "rollback"
}

func (w *Worker) Work() {
	txs := models.GetTxsByStatus(models.TxStatusBroadcastFailed)
	for _, tx := range txs {
		w.rollback(tx)
	}
}

func (w *Worker) rollback(tx *models.Tx) {
	var (
		t   = db.Default().Begin()
		err error
	)

	defer util.DeferRecover(w.Name(), func(error) {
		t.Rollback()
	})()

	defer func() {
		if err != nil {
			log.Errorf("%s, rollback tx sequenceID = %s failed, %v", w.Name(), tx.SequenceID, err)
			t.Rollback()
		}
	}()

	// 1. Rollback the from account's balance, blockchain_nonce and address status.
	txIns, err := models.GetTxInsBySequenceID(tx.SequenceID)
	if err != nil {
		return
	}

	for _, in := range txIns {
		err = t.Model(bmodels.Account{}).Where("address = ?", in.Address).
			Update(map[string]interface{}{
				"balance": gorm.Expr("`balance` + " + in.Amount.String()),
			}).Error
		if err != nil {
			return
		}

		err = t.Model(bmodels.Address{}).Where("address = ?", in.Address).
			Update(map[string]interface{}{
				"status": bmodels.AddressStatusRecord,
			}).Error
		if err != nil {
			return
		}
	}

	// 2. Rollback the utxos if necessary.
	utxos := bmodels.GetSpentUTXOs(tx.SequenceID)
	for _, u := range utxos {
		err = t.Model(u).Updates(map[string]interface{}{
			"status":   bmodels.UTXOStatusRecord,
			"spent_id": "",
		}).Error
		if err != nil {
			return
		}
	}

	// 3. Update the tx status to TxStatusDiscard.
	err = t.Model(tx).Update(map[string]interface{}{
		"tx_status": models.TxStatusDiscard,
	}).Error
	if err != nil {
		return
	}

	// 4. Insert a new copy of the tx if it is TxTypeWithdraw.
	if tx.TxType == models.TxTypeWithdraw {
		txCopy := tx.CloneCore()
		txCopy.SequenceID = util.HashString32([]byte(tx.SequenceID))
		txCopy.TxStatus = models.TxStatusRecord

		err = t.FirstOrCreate(txCopy, "sequence_id = ? and tx_type = ?", txCopy.SequenceID, txCopy.TxType).Error
		if err != nil {
			return
		}
	}

	err = t.Commit().Error
	if err != nil {
		return
	}

	log.Warnf("%s, success rollback tx, %s", w.Name(), tx)
}
