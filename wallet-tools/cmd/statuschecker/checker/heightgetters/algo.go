package heightgetters

import (
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("ALGO", algoGetter)
}

func algoGetter(url string) (int64, error) {
	urls := strings.Split(url, "#")
	if len(urls) != 2 {
		return 0, fmt.Errorf("invalid url format")
	}

	data, err := restyGet(strings.TrimRight(urls[0], "/")+"/v1/status", &httpHeader{"X-Algo-API-Token", urls[1]})
	if err != nil {
		return 0, err
	}

	return jsonparser.GetInt(data, "lastRound")
}
