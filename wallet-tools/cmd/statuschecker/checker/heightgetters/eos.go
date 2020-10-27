package heightgetters

import (
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("EOS", eosGetter)
}

func eosGetter(url string) (int64, error) {
	data, err := restyGet(strings.TrimRight(url, "/") + "/v1/chain/get_info")
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "last_irreversible_block_num")
	if err != nil {
		return 0, err
	}

	return height, nil
}
