package utils

import (
	"time"

	"gopkg.in/resty.v1"
)

func init() {
	resty.
		SetRetryCount(3).
		SetTimeout(30 * time.Second)
	//SetRetryWaitTime(3 * time.Second).
	//SetRetryMaxWaitTime(30 * time.Second).
}

// RestyPost defines a simple http request wrapper of resty.
func RestyPost(req, url string) ([]byte, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(req)).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

// RestyGet defines a simple http get request wrapper of resty.
func RestyGet(req, url string) ([]byte, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(req)).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}
