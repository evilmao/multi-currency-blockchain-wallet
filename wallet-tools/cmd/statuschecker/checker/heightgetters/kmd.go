package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("KMD", kmdGetter)
}

func kmdGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("getblockcount", nil))
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "result")
	if err != nil {
		return 0, err
	}

	return height, nil
}
