package txbuilder

import (
    "encoding/hex"
    "fmt"
    "strings"
    "sync"
    "time"

    bmodels "upex-wallet/wallet-base/models"
    "upex-wallet/wallet-base/newbitx/misclib/log"
    "upex-wallet/wallet-config/withdraw/transfer/config"
    "upex-wallet/wallet-withdraw/base/models"
    "upex-wallet/wallet-withdraw/transfer/alarm"

    "github.com/jinzhu/gorm"
    "github.com/shopspring/decimal"
)

type FeeMeta struct {
    Symbol   string
    Fee      decimal.Decimal
    GasLimit decimal.Decimal
    GasPrice decimal.Decimal
}

func (f FeeMeta) Clone() FeeMeta {
    return FeeMeta{
        Symbol:   f.Symbol,
        Fee:      f.Fee,
        GasLimit: f.GasLimit,
        GasPrice: f.GasPrice,
    }
}

func (f *FeeMeta) AdjustFee(min, max decimal.Decimal) {
    if min.GreaterThan(max) {
        min, max = max, min
    }

    if f.Fee.LessThan(min) {
        f.Fee = min
    } else if f.Fee.GreaterThan(max) {
        f.Fee = max
    }
}

type AccountModelBuildInfo struct {
    FromAccount *bmodels.Account
    FromPubKey  []byte
    Task        *models.Tx
    FeeMeta     FeeMeta
}

type AccountModelTxBuilder interface {
    DefaultFeeMeta() FeeMeta
    EstimateFeeMeta(string, int8) *FeeMeta
    DoBuild(*AccountModelBuildInfo) (*TxInfo, error)
}

type AccountModelBuilder struct {
    cfg     *config.Config
    builder AccountModelTxBuilder
    feeMeta FeeMeta
}

func NewAccountModelBuilder(cfg *config.Config, builder AccountModelTxBuilder) Builder {
    feeMeta := builder.DefaultFeeMeta()
    if len(feeMeta.Symbol) == 0 {
        feeMeta.Symbol = cfg.Currency
    }

    feeMeta.AdjustFee(decimal.NewFromFloat(cfg.MinFee), decimal.NewFromFloat(cfg.MaxFee))

    cfgGasLimit := decimal.NewFromFloat(cfg.MaxGasLimit)
    if cfgGasLimit.GreaterThan(feeMeta.GasLimit) {
        feeMeta.GasLimit = cfgGasLimit
    }

    cfgGasPrice := decimal.NewFromFloat(cfg.MaxGasPrice)
    if cfgGasPrice.GreaterThan(feeMeta.GasPrice) {
        feeMeta.GasPrice = cfgGasPrice
    }

    return &AccountModelBuilder{
        cfg:     cfg,
        builder: builder,
        feeMeta: feeMeta,
    }

}

type BuildByFeeMetaFunc func(FeeMeta, *models.Tx) (*TxInfo, error)

// BuildByFeeMeta build TxInfo by feeMeta, handle ErrFeeNotEnough.
func BuildByFeeMeta(cfg *config.Config, feeMeta FeeMeta, estimateFeeMeta *FeeMeta, task *models.Tx, doBuild BuildByFeeMetaFunc) (*TxInfo, error) {

    if estimateFeeMeta != nil && estimateFeeMeta.Fee.GreaterThan(feeMeta.Fee) {
        meta := estimateFeeMeta.Clone()
        meta.Symbol = feeMeta.Symbol
        meta.AdjustFee(decimal.NewFromFloat(cfg.MinFee), decimal.NewFromFloat(cfg.MaxFee))
        feeMeta = meta
    }

    txInfo, err := doBuild(feeMeta, task)
    if err != nil {
        if err, ok := err.(*alarm.ErrFeeNotEnough); ok {
            log.Warnf("%v, try to rebuild by new fee", err)
            feeMeta := feeMeta.Clone()
            feeMeta.Fee = err.NeedFee
            return doBuild(feeMeta, task)
        }

        return nil, err
    }

    return txInfo, nil
}

func (b *AccountModelBuilder) FeeMeta() FeeMeta {
    return b.feeMeta
}

func (b *AccountModelBuilder) Model() Model {
    return AccountModel
}

func (b *AccountModelBuilder) BuildWithdraw(task *models.Tx) (*TxInfo, error) {

    txInfo, err := BuildByFeeMeta(b.cfg, b.feeMeta, b.builder.EstimateFeeMeta(task.Symbol, task.TxType), task, b.buildWithdraw)

    go alarm.AlarmWhenBuildTaskFail(b.cfg, task, err)

    return txInfo, err
}

func (b *AccountModelBuilder) BuildGather(task *models.Tx) (*TxInfo, error) {

    txInfo, err := BuildByFeeMeta(b.cfg, b.feeMeta, b.builder.EstimateFeeMeta(task.Symbol, task.TxType), task, b.buildGather)

    go alarm.AlarmWhenBuildTaskFail(b.cfg, task, err)

    return txInfo, err
}

func (b *AccountModelBuilder) BuildSupplementaryFee(task *models.Tx) (*TxInfo, error) {
    return BuildByFeeMeta(b.cfg, b.feeMeta, b.builder.EstimateFeeMeta(task.Symbol, task.TxType), task, b.buildSupplementaryFee)
}

func (b *AccountModelBuilder) buildWithdraw(feeMeta FeeMeta, task *models.Tx) (*TxInfo, error) {

    var (
        fromAccount *bmodels.Account
        feeAccount  *bmodels.Account
    )
    if b.isFeeSymbol(task.Symbol) {
        fromAccount = bmodels.GetMatchedAccount(task.Amount.Add(feeMeta.Fee).String(), task.Symbol, bmodels.AddressTypeSystem)
    } else {
        fromAccount = bmodels.GetMatchedAccount(task.Amount.String(), task.Symbol, bmodels.AddressTypeSystem)
        // main currency
        feeAccount = bmodels.GetMatchedAccount(feeMeta.Fee.String(), b.cfg.Currency, bmodels.AddressTypeSystem)
    }

    if len(fromAccount.Address) == 0 {
        account := bmodels.GetMaxBalanceAccount(task.Symbol, bmodels.AddressTypeSystem)
        return nil, alarm.NewNotMatchAccount(feeMeta.Fee, task.Amount.Add(feeMeta.Fee), *account.Balance, account.Address)
    }

    if feeAccount != nil && len(feeAccount.Address) == 0 {
        return nil, alarm.NewErrorAccountBalanceNotEnough(fromAccount.Address, b.cfg.Currency, feeMeta.Fee)
    }

    return b.doBuild(fromAccount, feeMeta, task, feeAccount)
}

func (b *AccountModelBuilder) buildGather(feeMeta FeeMeta, task *models.Tx) (*TxInfo, error) {
    // if !b.builder.Support(task.Symbol) {
    // 	return nil, NewErrUnsupportedCurrency(task.Symbol)
    // }

    var (
        fromAccount *bmodels.Account
        feeAccount  *bmodels.Account
    )

    if b.isFeeSymbol(task.Symbol) {
        maxRemain := decimal.NewFromFloat(b.cfg.MaxAccountRemain)
        wideRemainWithFee := maxRemain.Mul(decimal.NewFromFloat(1.5)).Add(feeMeta.Fee)
        fromAccount = bmodels.GetMatchedAccount(wideRemainWithFee.String(), task.Symbol, bmodels.AddressTypeNormal)

        if fromAccount.Address == "" {
            return nil, nil
        }
        // fee := feeMeta.Fee
        maxRemainWithFee := maxRemain.Add(feeMeta.Fee)
        task.Amount = fromAccount.Balance.Sub(maxRemainWithFee)
    } else {
        fromAccount = bmodels.GetMatchedAccount("0", task.Symbol, bmodels.AddressTypeNormal)
        if fromAccount.Address == "" {
            return nil, nil
        }
        // main blockChain balance -- for pay transaction fees
        feeAccount = bmodels.GetAccountByAddress(fromAccount.Address, b.cfg.Currency)
        if feeAccount.Address == "" || feeAccount.Balance == nil || feeAccount.Balance.LessThan(feeMeta.Fee) {
            return nil, alarm.NewErrorAccountBalanceNotEnough(fromAccount.Address, b.cfg.Currency, feeMeta.Fee)
        }

        task.Amount = *fromAccount.Balance
    }

    return b.doBuild(fromAccount, feeMeta, task, feeAccount)
}

func (b *AccountModelBuilder) buildSupplementaryFee(feeMeta FeeMeta, task *models.Tx) (*TxInfo, error) {
    task.Symbol = feeMeta.Symbol

    minRemain := decimal.NewFromFloat(b.cfg.MaxAccountRemain)
    if feeMeta.Fee.GreaterThan(minRemain) {
        return nil, fmt.Errorf("tx fee (%s) is greater than min-account-remain balance (%s)",
            feeMeta.Fee, minRemain)
    }

    task.Amount = minRemain

    fromAccount := bmodels.GetMatchedAccount(task.Amount.Add(feeMeta.Fee).String(), task.Symbol, bmodels.AddressTypeSystem)
    if len(fromAccount.Address) == 0 {
        return nil, fmt.Errorf("wallet balance not enough")
    }

    return b.doBuild(fromAccount, feeMeta, task, nil)
}

func (b *AccountModelBuilder) FeeSymbol() string {
    return b.feeMeta.Symbol
}

func (b *AccountModelBuilder) isFeeSymbol(symbol string) bool {
    return strings.EqualFold(symbol, b.FeeSymbol())
}

func (b *AccountModelBuilder) doBuild(fromAccount *bmodels.Account, feeMeta FeeMeta, task *models.Tx, feeAccount *bmodels.Account) (*TxInfo, error) {
    if task.Amount.LessThanOrEqual(decimal.Zero) {
        return nil, fmt.Errorf("can't build tx with 0 amount")
    }

    pubKey, ok := bmodels.GetPubKey(nil, fromAccount.Address)
    if !ok || len(pubKey) == 0 {
        return nil, fmt.Errorf("db get pubkey failed")
    }

    fromPubKey, err := hex.DecodeString(pubKey)
    if err != nil {
        return nil, fmt.Errorf("decode sender public key failed, %v", err)
    }

    err = LockAddr(b.cfg, fromAccount.Address, task.SequenceID)
    if err != nil {
        return nil, fmt.Errorf("lock address failed, %v", err)
    }

    txInfo, err := b.builder.DoBuild(&AccountModelBuildInfo{
        FromAccount: fromAccount,
        FromPubKey:  fromPubKey,
        Task:        task,
        FeeMeta:     feeMeta,
    })
    if err != nil {
        return nil, err
    }

    if feeAccount != nil {
        costFee := feeMeta.Fee
        if txInfo.Fee.GreaterThan(decimal.Zero) {
            costFee = txInfo.Fee
        }

        feeInput := &TxIn{
            Account: feeAccount,
            Cost:    costFee,
        }
        txInfo.Inputs = append([]*TxIn{feeInput}, txInfo.Inputs...)
    }

    return txInfo, nil
}

func CalculateNextNonce(txType int8, txNonce, localNextNonce, remoteNextNonce uint64) uint64 {
    if txType != models.TxTypeWithdraw {
        return remoteNextNonce
    }

    if txNonce > 0 {
        return txNonce
    }

    // For the case that the initial nonce of account in blockchain is -1 (not 0),
    // so remoteNextNonce is 0.
    if remoteNextNonce == 0 || remoteNextNonce > localNextNonce {
        return remoteNextNonce
    }

    return localNextNonce
}

var (
    _addrLockStatus = &addressLockStatus{
        status: make(map[string]string),
    }
)

func LockAddr(cfg *config.Config, address, txSequenceID string) error {
    return _addrLockStatus.lockAddr(cfg, address, txSequenceID)
}

type addressLockStatus struct {
    sync.Mutex
    status map[string]string
}

func (s *addressLockStatus) lockAddr(cfg *config.Config, address, txSequenceID string) error {
    for {
        select {
        case <-cfg.ExitSignal:
            return fmt.Errorf("receive exit signal")
        default:
            ok, err := s.tryLockAddr(address, txSequenceID)
            if err != nil {
                return err
            }

            if ok {
                return nil
            }
        }
    }
}

func (s *addressLockStatus) tryLockAddr(address, txSequenceID string) (bool, error) {
    s.Lock()
    defer s.Unlock()

    ok, err := s.addressIdle(address)
    if err != nil {
        return false, err
    }

    if ok {
        s.status[address] = txSequenceID
        return true, nil
    }

    return false, nil
}

func (s *addressLockStatus) addressIdle(address string) (bool, error) {
    id := s.status[address]
    if len(id) == 0 {
        return true, nil
    }

    // Wait for build, store and broadcast tx.
    time.Sleep(time.Second * 2)

    tx, err := models.GetTxBySequenceID(nil, id)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            s.status[address] = ""
            return true, nil
        }

        return false, fmt.Errorf("db get tx by sequenceID %s failed, %v", id, err)
    }

    if tx.TxStatus == models.TxStatusBroadcast {
        return false, nil
    }

    s.status[address] = ""
    return true, nil
}
