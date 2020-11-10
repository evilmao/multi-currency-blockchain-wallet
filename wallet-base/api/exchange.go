package api

import (
    "fmt"
    "strings"
    "time"

    "upex-wallet/wallet-base/util"
)

// ExAPI represents an exchange api client.
// Detail doc at https://upex-wallet/document/blob/master/dev/assets-api/assets-dw-schedule-wallet-api.md.
// BrokerAPI represents a broker api client.
type ExAPI struct {
    url        string
    accessKey  string
    privateKey string
}

// NewExAPI returns broker api client.
func NewExAPI(brokerUrl, brokerAccessKey, brokerPrivateKey string) *ExAPI {
    brokerUrl = strings.TrimRight(brokerUrl, "/") + "/"
    return &ExAPI{
        brokerUrl,
        brokerAccessKey,
        brokerPrivateKey,
    }
}

// Deposit apis.
func (api *ExAPI) CommonRequestData(sign string) map[string]interface{} {
    return map[string]interface{}{
        "app_id":    api.accessKey,
        "symbol":    "fm",
        "timestamp": time.Now().Unix(),
    }
}

func (api *ExAPI) Sign(data map[string]interface{}) string {
    mapStr := SortMapToString(data)
    return util.SignSHA1(mapStr, api.privateKey)
}

func (api *ExAPI) UpdateRequestSign(data map[string]interface{}, sign string) map[string]interface{} {
    data["sign"] = sign
    return data
}

func (api *ExAPI) GetBrokerAppID() string {
    return api.accessKey
}

// DepositNotify notify deposit.
func (api *ExAPI) DepositNotify(data map[string]interface{}) (interface{}, int, error) {
    // 签名
    signStr := api.Sign(data)
    //  更新请求
    data = api.UpdateRequestSign(data, signStr)
    return util.RestPost(data, api.url+"depositNotify")
}

// BalanceChangeNotify, response deposit result
func (api *ExAPI) DepositBalanceChangeNotify(data map[string]interface{}) (interface{}, int, error) {
    signStr := api.Sign(data)
    data = api.UpdateRequestSign(data, signStr)
    return util.RestPost(data, api.url+"balanceChangeNotify")
}

// WithdrawNotify updates the withdraw task status.
func (api *ExAPI) WithdrawNotify(data map[string]interface{}) (interface{}, int, error) {
    signStr := api.Sign(data)
    data = api.UpdateRequestSign(data, signStr)
    return util.RestPost(data, api.url+"withdrawNotify/")
}

// BalanceChangeNotify, response  withdraw result
func (api *ExAPI) WithdrawBalanceChangeNotify(data map[string]interface{}) (interface{}, int, error) {
    signStr := api.Sign(data)
    data = api.UpdateRequestSign(data, signStr)
    return util.RestPost(data, api.url+"balanceChangeNotify")
}

// GetWithdraws gets withdraw list by currency.
// request args: {"symbol":"eth","count":100}
func (api *ExAPI) GetWithdraws(data map[string]interface{}) (interface{}, int, error) {
    signStr := api.Sign(data)
    data = api.UpdateRequestSign(data, signStr)
    return util.RestPost(data, api.url+"withdrawConsume")
}

func SortMapToString(data map[string]interface{}) string {
    return util.MapSortByKeyToString(data, "&", false, false)
}

// DangerousAPIForUpdateTxHash updates withdraw tx hash, use it carefully.
func (api *ExAPI) DangerousAPIForUpdateTxHash(transID, newTxHash string, data map[string]interface{}) error {
    if len(transID) == 0 {
        return fmt.Errorf("invalid transID")
    }

    if len(newTxHash) == 0 {
        return fmt.Errorf("invalid new tx hash")
    }

    data["trans_id"] = transID
    data["txid"] = newTxHash

    _, _, err := api.WithdrawNotify(data)

    return err
}
