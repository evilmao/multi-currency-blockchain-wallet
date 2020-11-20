package cooldown

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/currency"
	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer"
	"upex-wallet/wallet-withdraw/transfer/alarm"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"

	"github.com/shopspring/decimal"
)

type Worker struct {
	*service.SimpleWorker
	*transfer.Broadcaster
	cfg                  *config.Config
	txBuilder            txbuilder.Builder
	lastCoolDownTxTime   time.Time
	coolDownTaskInterval time.Duration
}

type ColdWalletInfo struct {
	ColdAddress   string
	RemainBalance decimal.Decimal
	MaxBalance    decimal.Decimal
}

func New(cfg *config.Config, txBuilder txbuilder.Builder) *Worker {

	return &Worker{
		Broadcaster: transfer.NewBroadcaster(cfg),
		cfg:         cfg,
		txBuilder:   txBuilder,

		lastCoolDownTxTime:   time.Now(),
		coolDownTaskInterval: cfg.CoolDownTaskInterval,
	}
}

func (w *Worker) Name() string {
	return "cooldown"
}

func (w *Worker) Work() {
	// per 15 minutes, scan system address balance,then do cool down task
	var (
		now     = time.Now()
		err     = fmt.Errorf("")
		symbols = currency.Symbols(w.cfg.Currency)
	)
	if now.Sub(w.lastCoolDownTxTime) < w.coolDownTaskInterval {
		return
	}

	log.Infof("cool down worker process...")
	w.lastCoolDownTxTime = now

	for _, symbol := range symbols {
		err = w.cooldown(symbol)
		if err != nil {
			log.Errorf("cooldown %s start failed, %v ", err)
		}
	}

	if err != nil {
		log.Errorf("%s, %v", w.Name(), err)
	}
}

func (w *Worker) cooldown(symbol string) error {
	// verify cold info: address,balance
	info, err := w.verifyColdInfo(symbol)
	if err != nil {
		return fmt.Errorf("invalid cold info, %v", err)
	}

	// get balance from db
	balance := bmodels.GetSystemBalance(symbol)
	if balance.LessThanOrEqual(info.MaxBalance) {
		return nil
	}

	if w.txBuilder.Model() == txbuilder.AccountModel {
		fromAccount := bmodels.GetMatchedAccount(info.MaxBalance.String(), symbol, bmodels.AddressTypeSystem)
		if len(fromAccount.Address) == 0 {
			return nil
		}

		balance = fromAccount.Balance
	}

	// build cold down transfer task
	task := &models.Tx{}
	task.Symbol = symbol
	task.TxType = models.TxTypeCold
	task.Address = info.ColdAddress
	task.Amount = balance.Sub(info.MaxBalance)
	task.UpdateLocalTransIDSequenceID()

	// if task.Amount.LessThan(decimal.NewFromFloat(w.cfg.MinFee)) {
	// 	return nil
	// }

	txInfo, err := w.txBuilder.BuildWithdraw(task)
	if err != nil {
		return fmt.Errorf("build tx failed, %v", err)
	}

	if txInfo == nil {
		return nil
	}

	err = transfer.CheckBalanceEnough(txInfo.Inputs)
	if err != nil {
		return fmt.Errorf("check balance enough failed, %v", err)
	}

	log.Warnf("%s, start cool-down tx, %s", w.Name(), task)

	task.Hex = txInfo.TxHex
	if txInfo.Nonce != nil {
		task.Nonce = *txInfo.Nonce
	}
	err = task.FirstOrCreate()
	if err != nil {
		return fmt.Errorf("db insert tx failed, %v", err)
	}

	err = w.BroadcastTx(txInfo, task)
	if err != nil {
		return fmt.Errorf("broadcast tx failed, %v", err)
	}

	err = task.Update(map[string]interface{}{
		"tx_status": models.TxStatusBroadcast,
		"fees":      txInfo.Fee,
	}, nil)
	if err != nil {
		return fmt.Errorf("db update tx failed, %v", err)
	}

	err = transfer.SpendTxIns(w.cfg.Code, task.SequenceID, txInfo.Inputs, txInfo.Nonce, txInfo.DiscardAddress)
	if err != nil {
		return fmt.Errorf("spend (sequenceID: %s) utxo failed, %v", task.SequenceID, err)
	}

	log.Warnf("%s, cool-down tx, %s", w.Name(), task)

	// if cool_down tx success, will send email
	for {
		TxHash := models.GetTxHashBySequenceID(task.SequenceID)
		if TxHash != "" {
			e := alarm.NewWarnCoolWalletBalanceChange(info.MaxBalance, *balance, info.ColdAddress, TxHash)
			go alarm.SendEmail(w.cfg, task, e, e.WarnDetail)
			break
		}
		time.Sleep(time.Second * 1)
	}

	return nil
}

func (w *Worker) verifyColdInfo(symbol string) (*ColdWalletInfo, error) {

	var (
		remainBalance decimal.Decimal
		maxBalance    decimal.Decimal
		coldAddress   = w.cfg.ColdAddress
		err           = "verify Cold wallet Info fail "
	)

	if coldAddress == "" {
		return nil, fmt.Errorf("%s,cold-address is empty,check if settings in config", err)
	}

	if symbol == w.cfg.Currency {
		remainBalance = decimal.NewFromFloat(w.cfg.MinAccountRemain)
		maxBalance = decimal.NewFromFloat(w.cfg.MaxAccountRemain)
	} else {
		s := bmodels.GetCurrency(w.cfg.Currency, symbol)
		remainBalance, _ = decimal.NewFromString(s.MinBalance)
		maxBalance, _ = decimal.NewFromString(s.MaxBalance)
	}

	if remainBalance.GreaterThan(maxBalance) {
		return nil, fmt.Errorf("%s,remain-balance (%s) is greater than max-balance (%s) ",
			err, remainBalance.String(), maxBalance.String())
	}

	return &ColdWalletInfo{
		ColdAddress:   coldAddress,
		MaxBalance:    maxBalance,
		RemainBalance: remainBalance,
	}, nil
}
