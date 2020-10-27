package util

import (
	"fmt"
	"math/big"

	"github.com/buger/jsonparser"
)

func JsonParserGetBigInt(data []byte, keys ...string) (val *big.Int, err error) {
	v, t, _, e := jsonparser.Get(data, keys...)

	if e != nil {
		return nil, e
	}

	if t != jsonparser.Number {
		return nil, fmt.Errorf("Value is not a number: %s", string(v))
	}

	return parseBigInt(v)
}

func parseBigInt(bytes []byte) (*big.Int, error) {
	if len(bytes) == 0 {
		return nil, jsonparser.MalformedValueError
	}

	val := new(big.Int)
	n, ok := val.SetString(string(bytes), 10)
	if !ok {
		return nil, jsonparser.MalformedValueError
	}

	return n, nil
}

// JSONParserArrayEach wrappers jsonparser.ArrayEach.
func JSONParserArrayEach(data []byte, f func([]byte, jsonparser.ValueType) error, keys ...string) (err error) {
	var valueType jsonparser.ValueType
	data, valueType, _, err = jsonparser.Get(data, keys...)
	if err != nil {
		return
	}

	if valueType == jsonparser.Null {
		return
	}

	if valueType != jsonparser.Array {
		err = fmt.Errorf("value type is %s, not array", valueType)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	idx := -1
	_, err = jsonparser.ArrayEach(data, func(v []byte, dttype jsonparser.ValueType, _ int, e error) {
		idx++

		if e != nil {
			panic(fmt.Sprintf("jsonparser at index %d failed, %v", idx, e))
		}

		e = f(v, dttype)
		if e != nil {
			panic(fmt.Sprintf("jsonparser at index %d failed, %v", idx, e))
		}
	})
	return
}
