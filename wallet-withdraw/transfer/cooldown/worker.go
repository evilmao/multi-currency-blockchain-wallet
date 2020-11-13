package cooldown

import (
    "fmt"
    "strings"
    "time"

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
    coldAddress      string
    minAccountRemain decimal.Decimal
    maxAccountRemain decimal.Decimal
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
    now := time.Now()
    if now.Sub(w.lastCoolDownTxTime) > w.coolDownTaskInterval {
        log.Infof("cool down worker process...")
        w.lastCoolDownTxTime = now
    } else {
        return
    }

    err := w.cooldown()
    if err != nil {
        log.Errorf("cooldown start failed, %v ", err)
    }
}

func (w *Worker) cooldown() error {
    // verify cold info: address,balance
    info, err := w.verifyColdInfo()
    if err != nil {
        return fmt.Errorf("invalid cold info, %v", err)
    }

    // get balance from db
    balance := bmodels.GetSystemBalance()
    if balance.LessThanOrEqual(info.maxAccountRemain) {
        return nil
    }

    if w.txBuilder.Model() == txbuilder.AccountModel {
        fromAccount := bmodels.GetMatchedAccount(info.maxAccountRemain.String(), bmodels.AddressTypeSystem)
        if len(fromAccount.Address) == 0 {
            return nil
        }

        balance = fromAccount.Balance
    }

    // build cold down transfer task
    task := &models.Tx{}
    task.Symbol = strings.ToLower(w.cfg.Currency)
    task.TxType = models.TxTypeCold
    task.Address = info.coldAddress
    task.Amount = balance.Sub(info.maxAccountRemain)
    task.UpdateLocalTransIDSequenceID()

    if task.Amount.LessThan(decimal.NewFromFloat(w.cfg.MinFee)){
        return nil
    }

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
            e := alarm.NewWarnCoolWalletBalanceChange(info.maxAccountRemain, *balance, info.coldAddress, TxHash)
            go alarm.SendEmail(w.cfg, task, e, e.WarnDetail)
            break
        }
        time.Sleep(time.Second * 5)
    }

    return nil
}

func (w *Worker) verifyColdInfo() (*ColdWalletInfo, error) {

    var (
        coldAddress      = w.cfg.ColdAddress
        minAccountRemain = decimal.NewFromFloat(w.cfg.MinAccountRemain)
        maxAccountRemain = decimal.NewFromFloat(w.cfg.MaxAccountRemain)
        err              = "verify Cold wallet Info fail "
    )

    if coldAddress == "" {
        return nil, fmt.Errorf("%s,cold-address is empty,check if settings in config", err)
    }

    if minAccountRemain.GreaterThan(maxAccountRemain) {
        return nil, fmt.Errorf("%s,remain-balance (%s) is greater than max-balance (%s) ",
            err, minAccountRemain.String(), maxAccountRemain.String())
    }

    return &ColdWalletInfo{
        coldAddress:      coldAddress,
        maxAccountRemain: maxAccountRemain,
        minAccountRemain: minAccountRemain,
    }, nil
}
