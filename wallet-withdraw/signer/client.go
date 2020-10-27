package signer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gopkg.in/resty.v1"
)

const (
	minTimeout = time.Second * 5
)

const (
	Success = iota
	Fail
)

// Request represents trade request.
type Request struct {
	PubKeys   []string `json:"pubkeys"`
	Digests   []string `json:"digests"`
	AuthToken string   `json:"authToken"`
}

// Response represents signing server response.
type Response struct {
	Status    int      `json:"status"`
	Signature []string `json:"signature"`
	Msg       string   `json:"msg"`
}

// Client represents signature client.
type Client struct {
	url      string
	password string
}

// NewClient returns a client instance.
func NewClient(url, password string, timeout time.Duration) *Client {
	if timeout < minTimeout {
		timeout = minTimeout
	}

	restyClient.SetTimeout(timeout)

	return &Client{
		url:      url,
		password: password,
	}
}

// Request requests signature server.
func (c *Client) Request(req *Request) (*Response, error) {
	req.AuthToken = c.password
	jsonReq, _ := json.Marshal(req)
	return httpPost(jsonReq, c.url)
}

var (
	restyClient = resty.New()
)

func httpPost(data interface{}, url string) (*Response, error) {
	var (
		resp Response
	)

	_resp, err := restyClient.R().
		SetHeader("Accept", "application/json").
		SetBody(data).
		Post(url)

	if err != nil || _resp.StatusCode() != 200 {
		return &resp, errors.New(_resp.Status())
	}

	err = json.Unmarshal(_resp.Body(), &resp)
	if err != nil {
		return nil, fmt.Errorf("httpPost unmarshal resp body failed, %v", err)
	}

	if resp.Status != Success {
		return nil, fmt.Errorf("%s", resp.Msg)
	}

	return &resp, nil
}
