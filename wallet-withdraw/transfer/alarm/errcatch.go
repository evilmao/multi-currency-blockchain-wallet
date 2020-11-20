package alarm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
)

/* errorCatch format
errorCatch = {
    "btc":[{"txType":"withdraw","errMsg":"balance not enough","updateTime":"15910013231"},
            ]
    }
*/

var (
	lock = sync.RWMutex{}
)

func Update(cfg *config.Config, task *models.Tx, err error) bool {

	if err == nil {
		return false
	}

	var (
		ErrorAlarmInterval = cfg.ErrorAlarmInterval
		errorCatch         = cfg.ErrorCatch
		currency           = strings.ToLower(cfg.Currency)
		txType             = models.TxTypeName(task.TxType)
		updateTime         = time.Now()
		errMsg             = fmt.Sprintf("%v", err)
	)

	lock.Lock()
	defer lock.Unlock()

	// find by txType, currency , error type, updateTime
	errorDetails, ok := errorCatch[currency]

	if !ok {
		errorCatch[currency] = make([]*config.TxError, 0)
		errorDetail := &config.TxError{
			TxType:     txType,
			Error:      err,
			UpdateTime: updateTime,
		}
		// update a new error catch
		errorCatch[currency] = append(errorCatch[currency], errorDetail)
		return true
	}

	flag := false
	for i, e := range errorDetails {
		eMsg := fmt.Sprintf("%v", e.Error)
		if txType == e.TxType && errMsg == eMsg {
			// warning time over 15 minutes: update error
			if updateTime.Sub(e.UpdateTime) > ErrorAlarmInterval {
				errorDetails[i].UpdateTime = updateTime
				flag = true
			} else {
				return false
			}
		}
	}

	// add new  error
	if !flag {
		errorDetail := &config.TxError{
			TxType:     txType,
			Error:      err,
			UpdateTime: updateTime,
		}
		// add new error type to catch
		errorDetails = append(errorDetails, errorDetail)
		errorCatch[currency] = errorDetails
		flag = true
	}

	return flag
}
