package heightgetters

import (
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("NAS", nasGetter)
}

func nasGetter(url string) (int64, error) {
	data, err := restyGet(strings.TrimRight(url, "/") + "/v1/user/nebstate")
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetString(data, "result", "height")
	if err != nil {
		return 0, err
	}

	heightN, err := strconv.Atoi(height)
	if err != nil {
		return 0, err
	}

	return int64(heightN), nil
}
