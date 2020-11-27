package gbtc

import (
	"encoding/json"
)

type Response struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

func (r *Response) UnmarshalResult(x interface{}) error {
	switch x.(type) {
	case *[]byte:
		*(x.(*[]byte)) = *r.Result
		return nil
	default:
		return json.Unmarshal(*r.Result, x)
	}
}
