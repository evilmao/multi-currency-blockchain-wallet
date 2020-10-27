package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("ETP", etpGetter)
}

func etpGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("fetch-height", []interface{}{}))
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "result")
	if err != nil {
		return 0, err
	}

	return height, nil
}
