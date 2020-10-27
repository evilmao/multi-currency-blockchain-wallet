// Get Best transaction fee through third-path api

package utxofee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
)

// request third-path api timeout
var requestTimout = time.Second * 5

var (
	ErrorGetCurrentFee = errors.New("fee api service error")
	ErrorUnmarshal     = errors.New("Unmarshal response data fail ")
	EmptyDB            = errors.New("empty for suggest fee db")
)

type UpdateFee struct {
	Cfg    *config.Config
	client *http.Client
}

// APIResponse receive Deserialization data from api response data
// 1.FeeApiResponseData map[string]interface{} `json:"data"`
type Fees struct {
	Regular  float64 `json:"regular"`
	Priority float64 `json:"priority"`
}

func NewUpdateFee(cfg *config.Config) *UpdateFee {
	return &UpdateFee{
		Cfg: cfg,
	}
}

// getFee get request the current transfer fee
// update maxFee in the config
func (uf *UpdateFee) getFee() (fees *Fees, err error) {
	// 从初始化配置文件中获取到第三方api
	// 实例化APIResponse
	fees = &Fees{}

	var (
		SuggestTransactionFees = uf.Cfg.SuggestTransactionFees
		FeeAPI, ok             = uf.Cfg.GetFeeAPI[uf.Cfg.Currency]
		method                 = "GET"
		client                 = &http.Client{Timeout: requestTimout}
		apiUrl                 = FeeAPI.ApiFeeURL
		retryTimes             = 3
		req                    = &http.Request{}
		res                    = &http.Response{}
		body                   = make([]byte, 1024)
	)

	if !ok {
		return nil, fmt.Errorf("currency not support")
	}

	for i := 1; i < retryTimes; i++ {
		// request
		req, err = http.NewRequest(method, apiUrl, nil)
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}

		// response
		res, err = client.Do(req)
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}

		if res.StatusCode != 200 {
			err = ErrorGetCurrentFee
		}
		break

	}

	// three times require all failed, return
	if err != nil {
		log.Errorf("%v, api url:%s", err, apiUrl)
		return nil, err
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &fees)
	if err != nil {
		log.Errorf("%v", ErrorUnmarshal)
		return nil, ErrorUnmarshal
	}

	SuggestTransactionFees[uf.Cfg.Currency] = make(map[string]float64)
	SuggestTransactionFees[uf.Cfg.Currency]["regular"] = fees.Regular   // 一般交易手续费 用于归集
	SuggestTransactionFees[uf.Cfg.Currency]["priority"] = fees.Priority // 优先手续费 用于用户提现

	log.Infof("The current of %s's regular transaction fee is %v sat/byte; priority transaction fess is %v sat/byte ", uf.Cfg.Currency, fees.Regular, fees.Priority)
	return
}

func updateFees(sfs []models.SuggestFee, fees *Fees) (err error) {

	if len(sfs) == 0 {
		return EmptyDB
	}

	var fee float64
	for _, sf := range sfs {
		if sf.FeeType == "regular" {
			fee = fees.Regular
		} else {
			fee = fees.Priority
		}

		err = sf.UpdateCurrentFee(fee)
		if err != nil {
			return fmt.Errorf("update %s fee fail, %v", sf.FeeType, err)
		}
	}
	return
}

func (uf *UpdateFee) storeFee(symbol string) (err error) {

	// 1.get fee from third-path api
	fees, err := uf.getFee()
	if err != nil {
		return
	}

	// 2.find suggestFees from db
	sfs := models.FindFeesBySymbol(symbol)
	// 3. update fees in db
	err = updateFees(sfs, fees)
	if err != nil {
		return
	}

	return nil
}

// UpdateTransferFee set timer to update fee
func (uf *UpdateFee) FeeService(symbol string) {
	// timer work

	for {
		// _, err := uf.getFee()
		err := uf.storeFee(symbol)
		if err != nil {
			log.Errorf("fee service for %s panic,%v", symbol, err)
			if err == ErrorGetCurrentFee {
				panic(err)
			}
		}
		time.Sleep(uf.Cfg.UpdateFeeInterval)
	}

}

// func main() {
// 	// 测试调用 api接口
//
// 	testCfg := config.DefaultConfig()
// 	uf := &UpdateFee{
// 		Cfg: testCfg,
// 	}
//
// 	go uf.FeeService("btc")
// 	fmt.Println("")
//
// 	for {
// 		time.Sleep(time.Second * 10)
// 		fmt.Println("Main() func run...")
// 	}
// }
