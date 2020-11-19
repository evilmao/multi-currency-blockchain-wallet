package txbuilder

import (
	"fmt"

	bmodels "upex-wallet/wallet-base/models"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/shopspring/decimal"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
)

// BuildExtInfo def.
type BuildExtInfo struct {
	Inputs       []*TxIn
	TotalInput   decimal.Decimal
	MaxOutAmount decimal.Decimal
}

var errEmptyInputs = fmt.Errorf("build extinfo failed, inputs is empty")

type utxoSelector func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error)

// fromAccounts
// func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {}
// maxTxInLen = metaData.MaxTxInLen = 30,
// maxOutAmount = decimal.Zero= 0
func createBuildExtInfo(fromAccounts []*bmodels.Account, selectUTXO utxoSelector, maxTxInLen int, maxOutAmount decimal.Decimal) (*BuildExtInfo, error) {
	var (
		extInfo = &BuildExtInfo{
			MaxOutAmount: maxOutAmount, // 0
		}
		utxoLen int // 0
	)
	for _, acc := range fromAccounts {
		// acc- Account
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
		// 更新输入交易信息
		extInfo.Inputs = append(extInfo.Inputs, &TxIn{
			Account:   acc,
			Cost:      totalIn,
			CostUTXOs: utxos,
		})
		// 更新总的交易输入资金
		extInfo.TotalInput = extInfo.TotalInput.Add(totalIn)
		// 每一笔不能多于三十笔输入
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

	// maxFee := decimal.NewFromFloat(cfg.MaxFee)
	// don't need to maxFee 同builder
	// if maxFee.GreaterThan(metaData.Fee) {
	// 	metaData.Fee = maxFee
	// }

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
		if err, ok := err.(*ErrFeeNotEnough); ok {
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

//BuildWithdraw UtXO like , build withdraw.
// func (b *UTXOModelBuilder) BuildWithdraw(task *models.Tx) (*TxInfo, error) {
// 	if !b.builder.Support(task.Symbol) {
// 		return nil, NewErrUnsupportedCurrency(task.Symbol)
// 	}
// 	log.Warnf("-----------step2---------n")
//
// 	// cost need add transaction fee
// 	return BuildByMetaData(b.metaData,
// 		func(metaData *MetaData) (*TxInfo, error) {
// 			// total cost = amount +  transaction fee
// 			cost := task.Amount.Add(metaData.Fee)
// 			log.Warnf("-----------step3----------%+v", b.metaData)
// 			fromAccounts := bmodels.GetAllMatchedAccounts(metaData.Fee.String(), uint(task.Code), bmodels.AddressTypeSystem)
// 			fromAccounts, ok := models.SelectAccount(fromAccounts, cost)
// 			log.Warnf("-----------step4----------%v", fromAccounts)
// 			if !ok {
// 				return nil, fmt.Errorf("wallet balance not enough")
// 			}
//
// 			extInfo, err := createBuildExtInfo(
// 				fromAccounts,
// 				func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
// 					// withdraw less or equal zero
// 					if cost.LessThanOrEqual(decimal.Zero) {
// 						return nil, decimal.Zero, false, nil
// 					}
// 					// withdraw amount
// 					amount := cost
// 					// 交易费用 大于 账户余额, 将账户下的全部币作为交易金额
// 					if amount.GreaterThan(*acc.Balance) {
// 						amount = *acc.Balance
// 					}
// 					// 根据 account表中筛选的address
// 					utxos, totalIn, ok := models.SelectUTXO(uint(task.Code), acc.Address, amount, limitLen)
// 					if !ok {
// 						return nil, decimal.Zero, false, fmt.Errorf("balance of %s mismatch to utxo, greater", acc.Address)
// 					}
// 					// 交易花费减去当前满足的交易
// 					cost = cost.Sub(totalIn)
// 					return utxos, totalIn, ok, nil
// 				},
// 				metaData.MaxTxInLen,
// 				decimal.Zero)
//
// 			if err != nil {
// 				return nil, err
// 			}
// 			// src\github.com\fb996de\wallet-withdraw\transfer\txbuilder\btc\btc.go
// 			return b.builder.DoBuild(metaData, task, extInfo)
// 		})
// }

// func (b *UTXOModelBuilder) BuildGather(task *models.Tx) (*TxInfo, error) {
// 	if !b.builder.Support(task.Symbol) {
// 		return nil, NewErrUnsupportedCurrency(task.Symbol)
// 	}
//
// 	maxWithdrawAmount, ok := currency.MaxWithdrawAmount(task.Symbol)
// 	if !ok {
// 		return nil, fmt.Errorf("can't find max withdraw amount of %s", task.Symbol)
// 	}
//
// 	// Set maxOutAmount = KYC单次最大提现值 * 5%.
// 	maxOutAmount := maxWithdrawAmount.Mul(decimal.NewFromFloat(0.05))
//
// 	buildExt := func(metaData *MetaData) (*BuildExtInfo, error) {
// 		// Build from normal address.
// 		fromAccounts := bmodels.GetAllMatchedAccounts(metaData.Fee.String(), uint(task.Code), bmodels.AddressTypeNormal)
//
// 		if len(fromAccounts) > 0 {
// 			return createBuildExtInfo(
// 				fromAccounts,
// 				func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
// 					utxos, totalIn, ok := models.SelectUTXO(uint(task.Code), acc.Address, decimal.Zero, limitLen)
// 					return utxos, totalIn, ok, nil
// 				},
// 				metaData.MaxTxInLen,
// 				maxOutAmount)
// 		}
//
// 		// Build from system address.
// 		fromAccounts = bmodels.GetAllMatchedAccounts(metaData.Fee.String(), uint(task.Code), bmodels.AddressTypeSystem)
// 		if len(fromAccounts) > 0 {
// 			return createBuildExtInfo(
// 				fromAccounts,
// 				func(acc *bmodels.Account, limitLen int) ([]*bmodels.UTXO, decimal.Decimal, bool, error) {
// 					// Set maxSmallUTXOAmount = maxOutAmount * 70%.
// 					maxSmallUTXOAmount := maxOutAmount.Mul(decimal.NewFromFloat(0.7))
// 					utxos, totalIn, ok := models.SelectSmallUTXO(uint(task.Code), acc.Address, maxSmallUTXOAmount, limitLen)
// 					return utxos, totalIn, ok, nil
// 				},
// 				metaData.MaxTxInLen,
// 				maxOutAmount)
// 		}
//
// 		return nil, nil
// 	}
//
// 	return BuildByMetaData(b.metaData, func(metaData *MetaData) (*TxInfo, error) {
// 		extInfo, err := buildExt(metaData)
// 		if err != nil {
// 			if err == errEmptyInputs {
// 				return nil, nil
// 			}
// 			return nil, err
// 		}
//
// 		if extInfo == nil {
// 			return nil, nil
// 		}
//
// 		task.Amount = extInfo.TotalInput.Sub(metaData.Fee)
// 		return b.builder.DoBuild(metaData, task, extInfo)
// 	})
// }

// OutputsAdder def.
type OutputsAdder func(string, uint64)

// MakeOutputs totalIn = extInfo.TotalInput
// mainOut =  task.Amount
// maxOutAmount = extInfo.MaxOutAmount = 0  //最大提现金额, 0表示无限制
// outAddress = task.Address,
// changeAddress = extInfo.Input.Account.Address
// metaData =  *txbuilder.MetaData
// addOutput OutputsAdder func type ====> sub func
func MakeOutputs(
	totalIn, mainOut, maxOutAmount decimal.Decimal,
	outAddress, changeAddress string,
	metaData *MetaData,
	addOutput OutputsAdder) {

	// 提现金额大于0
	if mainOut.GreaterThan(decimal.Zero) {
		if maxOutAmount.GreaterThan(decimal.Zero) {
			amount := mainOut
			v := maxOutAmount.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
			for amount.GreaterThan(maxOutAmount) {
				addOutput(outAddress, uint64(v))
				// amount - maxOutAmount
				amount = amount.Sub(maxOutAmount)
			}
			if amount.GreaterThan(decimal.Zero) {
				v := amount.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
				addOutput(outAddress, uint64(v))
			}
		} else {
			v := mainOut.Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
			addOutput(outAddress, uint64(v))
		}
	}

	cost := mainOut.Add(metaData.Fee)
	// 支付账户总的 UTXO 大于总的转账金额 也需要一笔 交易输出
	if totalIn.GreaterThan(cost) {
		changeValue := totalIn.Sub(cost).Mul(decimal.New(1, int32(metaData.Precision))).IntPart()
		addOutput(changeAddress, uint64(changeValue))
	}
}
