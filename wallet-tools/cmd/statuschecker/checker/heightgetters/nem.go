package heightgetters

import "github.com/buger/jsonparser"

func init() {
	Set("XEM", xemGetter)
}
func xemGetter(url string) (int64, error) {
	data, err := restyGet(url + "/chain/height")
	if err != nil {
		return 0, err
	}

	height, err := jsonparser.GetInt(data, "height")
	if err != nil {
		return 0, err
	}

	return height, nil
}
