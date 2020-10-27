package heightgetters

import (
	"strconv"

	"github.com/buger/jsonparser"
)

func init() {
	Set("ZIL", zilGetter)
}

func zilGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("GetLatestTxBlock", nil))
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetString(data, "result", "header", "BlockNum")
	if err != nil {
		return 0, err
	}

	heightN, err := strconv.Atoi(height)
	if err != nil {
		return 0, err
	}

	return int64(heightN), nil
}
