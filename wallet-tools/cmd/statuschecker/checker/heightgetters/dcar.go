package heightgetters

import (
	"strconv"

	"github.com/buger/jsonparser"
)

func init() {
	Set("DCAR", dcarGetter)
}

func dcarGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("dcar_blockNumber", nil))
	if err != nil {
		return 0, err
	}

	str, err := jsonparser.GetString(data, "result")
	if err != nil {
		return 0, err
	}

	height, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return 0, err
	}

	return height, nil
}
