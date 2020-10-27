package jsonrpc

import (
	"encoding/json"
)

type Response struct {
	Id     int64            `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

func (r *Response) UnmarshalResult(x interface{}) error {
	switch x.(type) {
	case *[]byte:
		if r.Result != nil {
			*(x.(*[]byte)) = *r.Result
		}
		return nil
	default:
		return json.Unmarshal(*r.Result, x)
	}
}
