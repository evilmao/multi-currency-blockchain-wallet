package gbtc

import (
	"encoding/json"
)

type Request struct {
	ID      int             `json:"id"`
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func NewRequest(method string, params Params) *Request {
	pb, _ := json.Marshal(params)
	return &Request{
		ID:      1,
		JsonRPC: "1.0",
		Method:  method,
		Params:  pb,
	}
}

type Params []interface{}
