package models

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"upex-wallet/wallet-base/db"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"upex-wallet/wallet-config/withdraw/transfer/config"
)

var (
	NoExistCurrencyError       = errors.New("can't update suggest Fee with no currency")
	getTransactionFeeZero      = errors.New("get current transaction fee error,check api")
	InvalidTxTypeError         = errors.New("invalid tx type ")
	InvalidTxTypeOrSymbolError = errors.New("no data in db for tx type and symbol type")
)

// SuggestFee represents the exchange that wallet create
// also it represents a wallet withdraw task.
type SuggestFee struct {
	gorm.Model
	Symbol     string `gorm:"size:32;index" json:"symbol"`
	UpdateFlag string `gorm:"size:32;index;default:'fee1'" json:"update_flag"`
	FeeType    string `gorm:"size:32;default:'regular'" json:"fee_type"`
	Fee1       int16  `gorm:"default:0" json:"fee1"`
	Fee2       int16  `gorm:"default:0" json:"fee2"`
	Fee3       int16  `gorm:"default:0" json:"fee3"`
	Fee4       int16  `gorm:"default:0" json:"fee4"`
	Fee5       int16  `gorm:"default:0" json:"fee5"`
}

// table name
func (sf SuggestFee) TableName() string { return "suggest_fee" }

// InitCurrencyFee find first matched record or create a new one.
func (sf SuggestFee) InitCurrencyFee() (err error) {
	// for every utxo-like currency create two pieces of db
	var (
		feeType = [2]string{"regular", "priority"}
	)

	for _, f := range feeType {
		suggestFee := sf
		suggestFee.FeeType = f
		err = db.Default().FirstOrCreate(&suggestFee, "symbol = ? and fee_type = ?", suggestFee.Symbol, f).Error
		if err != nil {
			return
		}
	}
	return
}

// Update updates the tx status.
func (sf *SuggestFee) Update(values M, dbInst *gorm.DB) error {
	if len(sf.Symbol) == 0 {
		return NoExistCurrencyError
	}

	if dbInst == nil {
		dbInst = db.Default()
	}
	return dbInst.Model(sf).Updates(values).Error
}

func FindFeesBySymbol(symbol string) []SuggestFee {
	var sfs []SuggestFee
	db.Default().Where("symbol = ? ", symbol).Find(&sfs)
	return sfs
}

func FindFeeBySymbolAndFeeType(symbol, feeType string) *SuggestFee {
	var sf SuggestFee
	db.Default().Where("symbol = ? and fee_type = ?", symbol, feeType).Find(&sf)
	return &sf
}

// UpdateCurrentFee updates the tx status.
func (sf SuggestFee) UpdateCurrentFee(suggestFee float64) error {

	if suggestFee == 0 {
		return getTransactionFeeZero
	}

	var (
		values       = make(map[string]interface{})
		sfType       = reflect.TypeOf(sf) // use reflect to get sf tag name
		sfVal        = reflect.ValueOf(sf)
		sfFieldNum   = sfVal.NumField()
		updateFlag   = sf.UpdateFlag // get update current flag
		updateTime   = sf.UpdatedAt
		intervalTime = time.Second * 60
		checkTime    = time.Now()
	)

	// interval time over 60s, reset db
	if checkTime.Sub(updateTime) > intervalTime {
		values["fee1"] = suggestFee
		values["fee2"] = 0
		values["fee3"] = 0
		values["fee4"] = 0
		values["fee5"] = 0
		values["update_flag"] = updateFlag
		return db.Default().Model(&sf).Updates(values).Error
	}

	for i := 0; i < sfFieldNum; i++ {
		// get json tag name
		fieldName := sfType.Field(i).Tag.Get("json")
		nextUpdateFlag := ""
		if updateFlag == fieldName {
			if updateFlag != "fee5" {
				nextUpdateFlag = sfType.Field(i + 1).Tag.Get("json")
			} else {
				nextUpdateFlag = "fee1"
			}
			values["update_flag"] = nextUpdateFlag
			values[updateFlag] = int16(suggestFee)
		}
	}
	return db.Default().Model(&sf).Updates(values).Error
}

// CalculateTransactionFee calculate average transaction fee
// transactionFee from db or cfg map
// txType , three types, `withdraw` and `gather` by from task.Name
func CalculateTransactionFee(txType string, cfg *config.Config) (transactionFee float64, err error) {

	var (
		feeType            = ""
		totalFee     int16 = 0
		feeNums            = 0
		checkTime          = time.Now()
		intervalTime       = time.Second * 60
		symbol             = strings.ToLower(cfg.Currency)
	)

	switch strings.ToLower(txType) {
	case "withdraw":
		feeType = "priority"
	case "gather", "cold":
		feeType = "regular"
	default:
		return 0, InvalidTxTypeError
	}

	sf := FindFeeBySymbolAndFeeType(symbol, feeType)
	if sf == nil {
		return 0, fmt.Errorf("%v, symbol:%s, txType:%s", InvalidTxTypeOrSymbolError, symbol, feeType)
	}

	sfValue := reflect.ValueOf(*sf)
	sfFieldNum := sfValue.NumField()
	updateTime := sf.UpdatedAt

	// get fee from db
	for i := 0; i < sfFieldNum; i++ {
		fee, ok := sfValue.Field(i).Interface().(int16)
		if ok && fee != 0 {
			totalFee += fee
			feeNums++
		}
	}

	// if db without transaction fees or update time interval greater than 60s
	if feeNums == 0 || checkTime.Sub(updateTime) > intervalTime {
		transactionFee, ok := cfg.SuggestTransactionFees[symbol][feeType]
		if !ok {
			err = fmt.Errorf("Get current transaction fee fail ")
		}
		return transactionFee, err
	}

	// if db store available fees, use average fee
	transactionFee = math.Ceil(float64(totalFee) / float64(feeNums))

	feeRange, ok := cfg.FeeLimitMap[symbol]

	if !ok {
		log.Warnf("May transaction fee for %s is not best ", symbol)
	}

	maxTxFee := feeRange.MaxTxFee
	minTxFee := feeRange.MinTxFee

	if minTxFee > transactionFee {
		transactionFee = minTxFee
	}

	if maxTxFee < transactionFee {
		transactionFee = maxTxFee
	}

	log.Warnf("txType is %s ,use feeType:%s, transactionFee is %f", txType, feeType, transactionFee)
	return transactionFee, nil
}
