package checker

import (
    "fmt"

    bmodels "upex-wallet/wallet-base/models"
    "upex-wallet/wallet-base/newbitx/misclib/log"
    "upex-wallet/wallet-config/withdraw/transfer/config"
    "upex-wallet/wallet-withdraw/base/models"

    "github.com/shopspring/decimal"
)

type ReadjustFeeInfo struct {
    RemainFee decimal.Decimal
    FeeSymbol string
}

type Calculator func(*config.Config, string) (*ReadjustFeeInfo, error)

type FeeReadJuster struct {
    calculateReadjust Calculator
    cfg               *config.Config
}

func NewFeeReadJuster(calculate Calculator) *FeeReadJuster {
    return &FeeReadJuster{
        calculateReadjust: calculate,
    }
}

func (a *FeeReadJuster) Name() string {
    return "FeeReadjuster"
}

func (a *FeeReadJuster) Init(cfg *config.Config) {
    a.cfg = cfg
}

func (a *FeeReadJuster) Check() error {
    txs := models.GetUnReadjustedFeeTxs()
    if len(txs) == 0 {
        return nil
    }

    for _, tx := range txs {
        err := a.readjustFee(tx)
        if err != nil {
            return err
        }
    }

    return nil
}

func (a *FeeReadJuster) readjustFee(tx *models.Tx) error {
    if a.calculateReadjust == nil {
        return fmt.Errorf("readjust calculator is nil")
    }

    info, err := a.calculateReadjust(a.cfg, tx.Hash)
    if err != nil {
        return err
    }

    if info == nil {
        return nil
    }

    if info.RemainFee.GreaterThan(decimal.Zero) {
        txIns, err := models.GetTxInsBySequenceID(tx.SequenceID)
        if err != nil {
            return fmt.Errorf("db get txins by sequence_id (%s) failed, %v", tx.SequenceID, err)
        }

        for _, in := range txIns {
            if in.Symbol == info.FeeSymbol {
                acc := &bmodels.Account{
                    Address: in.Address,
                    Symbol:  in.Symbol,
                }
                err = acc.ForUpdate(bmodels.M{
                    "op":      "add",
                    "balance": info.RemainFee,
                })
                if err != nil {
                    return fmt.Errorf("db update account (%s, %s) balance failed, %v", acc.Address, acc.Symbol, err)
                }

                log.Infof("checker, readjust tx fee, hash: %s, remainFee: %s, fromAddress: %s",
                    tx.Hash, info.RemainFee, in.Address)
                break
            }
        }
    }

    err = tx.Update(models.M{
        "readjusted_fee": true,
    }, nil)
    if err != nil {
        return fmt.Errorf("db update tx readjusted_fee failed, %v", err)
    }

    return nil
}
