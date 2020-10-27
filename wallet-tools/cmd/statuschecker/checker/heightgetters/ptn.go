package heightgetters

import (
	"github.com/buger/jsonparser"
)

func init() {
	Set("PTN", ptnGetter)
}

func ptnGetter(url string) (int64, error) {
	data, err := restyPost(url, jsonrpcRequest2("dag_getStableUnit", nil))
	if err != nil {
		return 0, err
	}

	return jsonparser.GetInt(data, "result", "number")
}
