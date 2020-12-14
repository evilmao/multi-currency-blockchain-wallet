package txbuilder

import (
	"fmt"

	"github.com/shopspring/decimal"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/transfer/alarm"
)

var (
	feeFloatUp     = 0.30 // 交易费浮动百分比 50%
	errEmptyInputs = fmt.Errorf("build extinfo failed, inputs is empty")
)

// BuildExtInfo def.
type BuildExtInfo struct {
	Inputs       []*TxIn
	TotalInput   decimal.Decimal
	MaxOutAmount decimal.Decimal
}

type utxoSelector func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error)

func createBuildExtInfo(fromAccounts []*bmodels.Account, selectUTXO utxoSelector, maxTxInLen int, maxOutAmount decimal.Decimal) (*BuildExtInfo, error) {
	var (
		extInfo = &BuildExtInfo{
			MaxOutAmount: maxOutAmount,
		}
		utxoLen int
	)
	for _, acc := range fromAccounts {
		utxos, totalIn, ok, err := selectUTXO(acc, maxTxInLen-utxoLen)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		if acc.Balance.LessThan(totalIn) {
			return nil, fmt.Errorf("balance of %s mismatch to utxo, less", acc.Address)
		}
		extInfo.Inputs = append(extInfo.Inputs, &TxIn{
			Account:   acc,
			Cost:      totalIn,
			CostUTXOs: utxos,
		})
		extInfo.TotalInput = extInfo.TotalInput.Add(totalIn)
		utxoLen += len(utxos)
		if utxoLen >= maxTxInLen {
			break
		}
	}

	if len(extInfo.Inputs) == 0 {
		return nil, errEmptyInputs
	}

	return extInfo, nil
}

// UTXOModelTxBuilder def.
type UTXOModelTxBuilder interface {
	Support(string) bool
	DoBuild(*MetaData, *models.Tx, *BuildExtInfo) (*TxInfo, error)
	SupportFeeRate() (feeRate float64, ok bool)
	CalculateFee(nIn, nOut int, feeRate float64, precision int32) (fee decimal.Decimal)
}

type UTXOModelBuilder struct {
	cfg      *config.Config
	metaData *MetaData
	builder  UTXOModelTxBuilder
}

// NewUTXOModelBuilder factory func to instance a UTXO Builder
func NewUTXOModelBuilder(cfg *config.Config, builder UTXOModelTxBuilder) Builder {

	metaData, ok := FindMetaData(cfg.Currency)
	if !ok {
		panic(fmt.Errorf("can't get meta data of currency %s", cfg.Currency))
	}

	minFee := decimal.NewFromFloat(cfg.MinFee)
	if minFee.GreaterThan(metaData.Fee) {
		metaData.Fee = minFee
	}

	return &UTXOModelBuilder{
		cfg:      cfg,
		metaData: metaData,
		builder:  builder,
	}
}

// BuildByMetaData build TxInfo by metaData, handle ErrFeeNotEnough.
func (b *UTXOModelBuilder) BuildByMetaData(doBuild func(*MetaData) (*TxInfo, error)) (*TxInfo, error) {
	txInfo, err := doBuild(b.metaData)

	if err != nil {
		if err, ok := err.(*alarm.ErrFeeNotEnough); ok {
			log.Warnf("%v, try to rebuild by new fee", err)
			feeFloatUp = b.cfg.FeeFloatUp
			return doBuild(b.metaData)
		}
		return nil, err
	}
	return txInfo, nil
}

// Model to instance a UTXO model
func (b *UTXOModelBuilder) Model() Model {
	return UTXOModel
}

// BuildWithdraw , use suggest transaction fee to calculate inputsNums and nOut
func (b *UTXOModelBuilder) BuildWithdraw(task *models.Tx) (txInfo *TxInfo, err error) {
	var (
		MaxOutAmount = decimal.Zero
		txType       = models.TxTypeName(task.TxType)
	)

	if !b.builder.Support(task.Symbol) {
		return nil, NewErrUnsupportedCurrency(task.Symbol)
	}

	feeRate, ok := b.supportFeeRate(txType)

	if !ok {
		txInfo, err = BuildByMetaData(b.metaData, func(metaData *MetaData) (*TxInfo, error) {
			extInfo, err := createWithdrawExtByMeta(metaData, task)
			if err != nil {
				return nil, err
			}

			return b.builder.DoBuild(metaData, task, extInfo)
		})
	} else {
		txInfo, err = BuildByFeeRate(b.metaData, func(metaData *MetaData) (*TxInfo, error) {
			extInfo, err := createWithdrawExtByFeeRate(b, task, feeRate, MaxOutAmount)
			if err != nil {
				return nil, err
			}
			return b.builder.DoBuild(metaData, task, extInfo)
		})
	}

	go alarm.AlarmWhenBuildTaskFail(b.cfg, task, err)

	return
}

// new BuildGather
func (b *UTXOModelBuilder) BuildGather(task *models.Tx) (*TxInfo, error) {

	var (
		txType            = models.TxTypeName(task.TxType)
		maxWithdrawAmount = decimal.NewFromFloat(b.cfg.MaxWithdrawAmount)
	)

	if !b.builder.Support(task.Symbol) {
		return nil, NewErrUnsupportedCurrency(task.Symbol)
	}

	if maxWithdrawAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("err for max withdraw amount of %s", task.Symbol)
	}
	maxOutAmount := maxWithdrawAmount.Mul(decimal.NewFromFloat(0.05))

	feeRate, _ := b.supportFeeRate(txType)

	txInfo, err := BuildByMetaData(
		b.metaData,
		func(metaData *MetaData) (*TxInfo, error) {

			extInfo, err := b.buildExtGather(metaData, feeRate, task, maxOutAmount)
			if err != nil {
				if err == errEmptyInputs {
					return nil, nil
				}
				return nil, err
			}

			if extInfo == nil {
				return nil, nil
			}

			// update metaFee
			if feeRate > 0 {
				nIn, nOut := len(extInfo.Inputs), 1
				fee := b.builder.CalculateFee(nIn, nOut, feeRate, int32(metaData.Precision))
				b.metaData.UpdateFee(fee)
			}

			// update task Amount
			task.Amount = extInfo.TotalInput.Sub(metaData.Fee)
			if task.Amount.LessThan(decimal.Zero) {
				return nil, alarm.NewErrorBalanceLessCost(metaData.Fee, extInfo.TotalInput, task.Amount)
			}

			return b.builder.DoBuild(metaData, task, extInfo)
		})

	go alarm.AlarmWhenBuildTaskFail(b.cfg, task, err)

	return txInfo, err
}

// OutputsAdder def.
type OutputsAdder func(string, uint64)

// supportFeeRate according 3rd-path api or rpc api to calculate transfer fee.
func (b *UTXOModelBuilder) supportFeeRate(txType string) (feeRate float64, ok bool) {

	feeRate, ok = b.builder.SupportFeeRate()
	if !ok {
		return
	}

	switch txType {
	case "withdraw":
		feeRate = feeRate * (1 + feeFloatUp)
	}

	return
}

func (b *UTXOModelBuilder) buildExtGather(metaData *MetaData, feeRate float64, task *models.Tx, maxOutAmount decimal.Decimal) (extInfo *BuildExtInfo, err error) {

	symbol := task.Symbol
	precision := int32(metaData.Precision)
	feeFilter := func() string {
		if feeRate > 0 {
			return b.builder.CalculateFee(1, 0, feeRate, precision).String()
		} else {
			return b.metaData.Fee.String()
		}
	}()

	// Build from normal address.
	fromAccounts := bmodels.GetAllMatchedAccounts(feeFilter, symbol, bmodels.AddressTypeNormal)
	if len(fromAccounts) > 0 {
		return createBuildExtInfo(
			fromAccounts,
			func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
				utxos, totalIn, ok := models.SelectUTXO(acc.Address, symbol, decimal.Zero, limitLen)
				return utxos, totalIn, ok, nil
			},
			metaData.MaxTxInLen,
			maxOutAmount)
	}

	// Build from system address.
	fromAccounts = bmodels.GetAllMatchedAccounts(feeFilter, symbol, bmodels.AddressTypeSystem)
	if len(fromAccounts) > 0 {
		return createBuildExtInfo(
			fromAccounts,
			func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
				maxSmallUTXOAmount := maxOutAmount.Mul(decimal.NewFromFloat(0.7))
				utxos, totalIn, ok := models.SelectSmallUTXO(symbol, acc.Address, maxSmallUTXOAmount, limitLen)
				return utxos, totalIn, ok, nil
			},
			metaData.MaxTxInLen,
			maxOutAmount)
	}

	return nil, nil
}

// MakeOutputs totalIn = extInfo.TotalInput
func MakeOutputs(totalIn, mainOut, maxOutAmount decimal.Decimal, outAddress, changeAddress string, metaData *MetaData, addOutput OutputsAdder) {

	if mainOut.GreaterThan(decimal.Zero) {
		// gather tx
		if maxOutAmount.GreaterThan(decimal.Zero) {
			amount := mainOut
			v := maxOutAmount.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
			for amount.GreaterThan(maxOutAmount) {
				addOutput(outAddress, uint64(v))
				amount = amount.Sub(maxOutAmount)
			}
			if amount.GreaterThan(decimal.Zero) {
				v := amount.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
				addOutput(outAddress, uint64(v))
			}
		} else {
			// withdraw
			v := mainOut.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
			addOutput(outAddress, uint64(v))
		}
	}

	cost := mainOut.Add(metaData.Fee)
	if totalIn.GreaterThan(cost) {
		changeValue := totalIn.Sub(cost).Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
		addOutput(changeAddress, uint64(changeValue))
	}
}

func BuildByMetaData(metaData *MetaData, doBuild func(*MetaData) (*TxInfo, error)) (*TxInfo, error) {

	txInfo, err := doBuild(metaData)

	if err != nil {
		if err, ok := err.(*alarm.ErrFeeNotEnough); ok {
			log.Warnf("%v, try to rebuild by new fee", err)

			metaData := metaData.Clone()
			metaData.Fee = err.NeedFee
			return doBuild(metaData)
		}

		return nil, err
	}

	return txInfo, nil
}

func BuildByFeeRate(metaData *MetaData, doBuild func(*MetaData) (*TxInfo, error)) (*TxInfo, error) {
	txInfo, err := doBuild(metaData)

	if err != nil {
		if err, ok := err.(*alarm.ErrFeeNotEnough); ok {
			log.Warnf("%v, try to rebuild by new fee", err)
			metaData.Fee = err.NeedFee
			return doBuild(metaData)
		}

		return nil, err
	}

	return txInfo, nil
}

func createWithdrawExtByFeeRate(b *UTXOModelBuilder, task *models.Tx, feeRate float64, maxOutAmount decimal.Decimal) (*BuildExtInfo, error) {

	var (
		cost      = task.Amount
		symbol    = task.Symbol
		utxoLen   int
		nIn       = 0 // init inPuts  number
		nOut      = 1 // init OutPuts number
		precision = int32(b.metaData.Precision)
		oneOutFee = b.builder.CalculateFee(0, nOut, feeRate, precision)
		oneInFee  = b.builder.CalculateFee(nIn, 0, feeRate, precision)
		filterFee = b.builder.CalculateFee(1, 1, feeRate, precision)
		extInfo   = &BuildExtInfo{
			MaxOutAmount: maxOutAmount,
		}
	)

	if task.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("task amount must be greater than zero")
	}

	fromAccounts := bmodels.GetAllMatchedAccounts(filterFee.String(), symbol, bmodels.AddressTypeSystem)
	fromAccounts, ok := models.SortAccountsByBalance(fromAccounts)

	if !ok {
		return nil, alarm.NewErrorBalanceLessThanFee(filterFee)
	}

	cost = cost.Add(oneOutFee)
	for _, account := range fromAccounts {
		var (
			totalIn      decimal.Decimal
			AccountUTXOS []*bmodels.UTXO
		)

		if cost.Equal(decimal.Zero) || cost.LessThan(oneOutFee) {
			break
		}

		// according the address in account table select lowest required utxos
		utxos, ok := models.SelectUTXOWithTransFee(account.Address, task.Symbol, b.metaData.MaxTxInLen-utxoLen, true)
		if !ok {
			return nil, fmt.Errorf("Balance of %s mismatch to utxo, greater ", account.Address)
		}

		for _, u := range utxos {
			// only a just output || if change fee and meet one output transaction fee
			if cost.Equal(decimal.Zero) || cost.LessThanOrEqual(oneOutFee) {
				break
			}

			AccountUTXOS = append(AccountUTXOS, u) // update utxos
			cost = cost.Add(oneInFee)              // add a input transaction fee
			nIn++                                  // input number add one
			cost = cost.Sub(u.Amount)              // cost subtract current uxto amount
			totalIn = totalIn.Add(u.Amount)        // sum all utxo amount

			// If the cost is less than 0, it means that after consuming the current UTXO, there is a balance, and change is needed
			// outPut number at most 2
			if cost.LessThan(decimal.Zero) && nOut < 2 {
				nOut++                     // if need change, need add  one outputs
				cost = cost.Add(oneOutFee) // cost need add one outPut transaction fee
			}
		}

		// update inputs
		extInfo.Inputs = append(extInfo.Inputs, &TxIn{
			Account:   account,
			Cost:      totalIn,
			CostUTXOs: AccountUTXOS,
		})

		// calculate all of inputs of using utxos' amount
		extInfo.TotalInput = extInfo.TotalInput.Add(totalIn)

		// every account with utxos must less than MaxTxInLen--will case to many inputs
		utxoLen += len(AccountUTXOS)
		if utxoLen >= b.metaData.MaxTxInLen {
			break
		}
	}

	if len(extInfo.Inputs) == 0 {
		return nil, errEmptyInputs
	}

	//  update metaFee
	fee := b.builder.CalculateFee(nIn, nOut, feeRate, int32(b.metaData.Precision))
	b.metaData.UpdateFee(fee)

	// check total cost is less than totalInput
	totalCost := task.Amount.Add(fee)
	if extInfo.TotalInput.LessThan(totalCost) {
		return nil, alarm.NewErrorBalanceLessCost(b.metaData.Fee, extInfo.TotalInput, task.Amount)
	}

	return extInfo, nil
}

func createWithdrawExtByMeta(metaData *MetaData, task *models.Tx) (*BuildExtInfo, error) {
	cost := task.Amount.Add(metaData.Fee)
	filter := metaData.Fee.String()
	symbol := task.Symbol

	fromAccounts := bmodels.GetAllMatchedAccounts(filter, symbol, bmodels.AddressTypeSystem)
	fromAccounts, ok := models.SelectAccount(fromAccounts, cost)
	if !ok {
		return nil, fmt.Errorf("wallet balance not enough")
	}

	extInfo, err := createBuildExtInfo(
		fromAccounts,
		func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
			if cost.LessThanOrEqual(decimal.Zero) {
				return nil, decimal.Zero, false, nil
			}

			amount := cost
			if amount.GreaterThan(*acc.Balance) {
				amount = *acc.Balance
			}

			utxos, totalIn, ok := models.SelectUTXO(task.Symbol, acc.Address, amount, limitLen)
			if !ok {
				return nil, decimal.Zero, false, fmt.Errorf("balance of %s mismatch to utxo, greater", acc.Address)
			}

			cost = cost.Sub(totalIn)
			return utxos, totalIn, ok, nil
		},
		metaData.MaxTxInLen,
		decimal.Zero)

	return extInfo, err
}
