package jsonrpc

import (
	"encoding/json"
)

type Request struct {
	ID      int             `json:"id"`
	JsonRPC JsonRPCVersion  `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func NewRequest(method string, params Params, version JsonRPCVersion) *Request {
	if params == nil {
		params = Params{}
	}

	pb, _ := json.Marshal(params)
	return &Request{
		ID:      1,
		JsonRPC: version,
		Method:  method,
		Params:  pb,
	}
}

type Params []interface{}
