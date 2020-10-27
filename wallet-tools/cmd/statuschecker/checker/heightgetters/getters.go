package heightgetters

import (
	"math/rand"

	"gopkg.in/resty.v1"
)

type Getter func(url string) (int64, error)

var (
	heightGetters = map[string]Getter{}
)

func Set(name string, g Getter) {
	heightGetters[name] = g
}

func Get(name string) (Getter, bool) {
	g, ok := heightGetters[name]
	return g, ok
}

func init() {
	resty.SetDebug(false)
}

type httpHeader struct {
	K, V string
}

func restyGet(url string, headers ...*httpHeader) ([]byte, error) {
	req := resty.R()
	req.SetHeader("Content-Type", "application/json")
	for _, h := range headers {
		req.SetHeader(h.K, h.V)
	}

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func restyPost(url string, data interface{}, headers ...*httpHeader) ([]byte, error) {
	req := resty.R()
	req.SetBody(data)
	req.SetHeader("Content-Type", "application/json")
	for _, h := range headers {
		req.SetHeader(h.K, h.V)
	}

	resp, err := req.Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func jsonrpcRequest1(method string, args interface{}) map[string]interface{} {
	return jsonrpcRequest("1.0", method, args)
}

func jsonrpcRequest2(method string, args interface{}) map[string]interface{} {
	return jsonrpcRequest("2.0", method, args)
}

func jsonrpcRequest(version, method string, args interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": version,
		"method":  method,
		"params":  args,
		"id":      rand.Int63(),
	}
}
