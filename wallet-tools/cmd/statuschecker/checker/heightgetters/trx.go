package heightgetters

import (
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("TRX", trxGetter)
}

func trxGetter(url string) (int64, error) {
	url = strings.TrimRight(url, "/") + "/wallet/getnowblock"
	data, err := restyPost(url, nil)
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "block_header", "raw_data", "number")
	if err != nil {
		return 0, err
	}

	return height, nil
}
