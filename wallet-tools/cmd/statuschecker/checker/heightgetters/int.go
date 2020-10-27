package heightgetters

import (
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("INT", intGetter)
}

type request struct {
	Method string                 `json:"funName"`
	Params map[string]interface{} `json:"args"`
}

func intGetter(url string) (int64, error) {
	data, err := restyPost(strings.TrimRight(url, "/")+"/rpc", map[string]interface{}{
		"funName": "getBlock",
		"args": map[string]interface{}{
			"which":        "latest",
			"transactions": false,
		},
	})
	if err != nil {
		return 0, err
	}

	errCode, err := jsonparser.GetInt(data, "err")
	if err != nil {
		return 0, err
	}

	if errCode != 0 {
		return 0, fmt.Errorf("errcode: %d, %s", errCode, string(data))
	}

	height, err := jsonparser.GetInt(data, "block", "number")
	if err != nil {
		return 0, err
	}

	return height, nil
}
