package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("FAB", btcGetter)
}

func btcGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest1("getbestblockhash", []interface{}{}))
	if err != nil {
		return 0, err
	}

	hash, err := jsonparser.GetString(data, "result")
	if err != nil {
		return 0, err
	}

	data, err = restyPost(url, jsonrpcRequest1("getblock", []interface{}{hash}))
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "result", "height")
	if err != nil {
		return 0, err
	}

	return height, nil
}
