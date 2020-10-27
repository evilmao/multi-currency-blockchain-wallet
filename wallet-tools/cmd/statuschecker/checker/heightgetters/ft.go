package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("FT", ftGetter)
}

func ftGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("ft_getCurrentBlock", []interface{}{true}))
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "result", "number")
	if err != nil {
		return 0, err
	}

	return height, nil
}
