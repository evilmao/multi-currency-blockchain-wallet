package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("XLM", xlmGetter)
}

func xlmGetter(url string) (int64, error) {
	data, err := restyGet(url)
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "history_latest_ledger")
	if err != nil {
		return 0, err
	}

	return height, nil
}
