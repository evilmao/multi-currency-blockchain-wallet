package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type JsonRPCVersion string

const (
	JsonRPCV1 JsonRPCVersion = "1.0"
	JsonRPCV2 JsonRPCVersion = "2.0"
)

type Client struct {
	url     string
	version JsonRPCVersion
}

func NewClient(url string, version JsonRPCVersion) *Client {
	return &Client{
		url:     url,
		version: version,
	}
}

func (c *Client) request(method string, params Params) ([]byte, error) {
	reqData := NewRequest(method, params, c.version)
	rawData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("json marshal request data failed, %v", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(rawData))
	if err != nil {
		return nil, fmt.Errorf("create http request failed, %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request failed, %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed, %v", err)
	}

	return body, nil
}

func (rpc *Client) Call(method string, params Params, result interface{}) error {
	rawData, err := rpc.request(method, params)
	if err != nil {
		return err
	}

	var respData Response
	err = json.Unmarshal(rawData, &respData)
	if err != nil {
		return fmt.Errorf("json unmarshal response body failed, %s", err)
	}

	if respData.Error != nil {
		return fmt.Errorf("%+v", respData.Error)
	}

	return respData.UnmarshalResult(result)
}
