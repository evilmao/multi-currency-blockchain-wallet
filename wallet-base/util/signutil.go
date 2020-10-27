package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

/*JoinStringsInASCII 按照规则，参数名ASCII码从小到大排序后拼接
//data 待拼接的数据
//sep 连接符
//onlyValues 是否只包含参数值，true则不包含参数名，否则参数名和参数值均有
//includeEmpty 是否包含空值，true则包含空值，否则不包含，注意此参数不影响参数名的存在
*/
func MapSortByKeyToString(dataData map[string]interface{}, sep string, onlyValues, includeEmpty bool, exceptKeys ...string) string {
	var list []string
	var keyList []string
	m := make(map[string]int)
	// 排除不用签名的key
	if len(exceptKeys) > 0 {
		for _, except := range exceptKeys {
			m[except] = 1
		}
	}

	for k := range dataData {
		if _, ok := m[k]; ok {
			continue
		}
		value := dataData[k]
		if !includeEmpty && value == "" {
			continue
		}
		if onlyValues {
			keyList = append(keyList, k)
		} else {
			switch value.(type) {
			case int64, int32, int:
				value = fmt.Sprintf("%d", value)
			case float64:
				value = fmt.Sprintf("%f", value)
			case bool:
				value = fmt.Sprintf("%t", value)
			case uint64, uint, uint16:
				value = fmt.Sprintf("%d", value)
			}
			list = append(list, fmt.Sprintf("%s=%s", k, value))
		}
	}
	if onlyValues {
		sort.Strings(keyList)
		for _, v := range keyList {
			switch dataData[v].(type) {
			case int64, int32, int:
				list = append(list, fmt.Sprintf("%d", dataData[v]))
			case float64:
				list = append(list, fmt.Sprintf("%f", dataData[v]))
			case bool:
				list = append(list, fmt.Sprintf("%t", dataData[v]))
			default:
				list = append(list, dataData[v].(string))
			}
		}
	} else {
		sort.Strings(list)
	}
	return strings.Join(list, sep)
}

func SignSHA1(encryptText interface{}, signKey string) string {
	var (
		buffer = encryptText.(string)
	)
	signature := hmac.New(sha1.New, []byte(signKey))
	signature.Write([]byte(buffer))
	return hex.EncodeToString(signature.Sum(nil))
}

// test
func main() {
	data := make(map[string]interface{})
	data["appid"] = "wx_1234535"
	data["body"] = "test data"
	data["mch_id"] = "572836589"
	data["notify_url"] = "http://www.baidu.com"
	data["trade_type"] = "MWEB"
	data["spbill_create_ip"] = "192.169.0.1"
	data["total_fee"] = "100"
	data["out_trade_no"] = "2745890486870"
	data["nonce_str"] = "kdjskgjokghdk"
	data["sign"] = "abc11111111"
	encryptText := MapSortByKeyToString(data, "&", false, false, "sign")
	fmt.Println(encryptText)
	// 签名
	signStr := SignSHA1(encryptText, data["sign"].(string))
	fmt.Println("签名的字符串:", signStr)
	// log.Println("str :", MapSortByKeyToString(data, "&", false, false, "sign"))
	// log.Println("str2 :", MapSortByKeyToString(data, "&", true, false, "sign"))
}
