package heightgetters

import (
	"strings"

	"github.com/buger/jsonparser"
)

func init() {
	Set("ADA", adaGetter)
}

func adaGetter(url string) (int64, error) {
	const SLOT_MASK = 100000
	url = strings.TrimRight(url, "/") + "/api/blocks/pages/"
	data, err := restyGet(url)
	if err != nil {
		return 0, err
	}

	var (
		maxEpocp int64
		maxSlot  int64
		innerErr error
	)
	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, _ int, err error) {
		if innerErr != nil {
			return
		}

		if err != nil {
			innerErr = err
			return
		}

		if dataType != jsonparser.Array {
			return
		}

		jsonparser.ArrayEach(value, func(block []byte, _ jsonparser.ValueType, _ int, err error) {
			if innerErr != nil {
				return
			}

			if err != nil {
				innerErr = err
				return
			}

			epoch, _ := jsonparser.GetInt(block, "cbeEpoch")
			slot, _ := jsonparser.GetInt(block, "cbeSlot")
			if epoch > maxEpocp || (epoch == maxEpocp && slot > maxSlot) {
				maxEpocp = epoch
				maxSlot = slot
			}
		})
	}, "Right")
	if err != nil {
		return 0, err
	}

	if innerErr != nil {
		return 0, innerErr
	}

	height := int64(maxEpocp)*SLOT_MASK + int64(maxSlot)
	return height, nil
}
