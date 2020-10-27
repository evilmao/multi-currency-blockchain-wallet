package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func _httpRequest(client *http.Client, reqType string, reqUrl string, postData string, requstHeaders map[string]string) ([]byte, error) {
	req, _ := http.NewRequest(reqType, reqUrl, strings.NewReader(postData))

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36")

	if requstHeaders != nil {
		for k, v := range requstHeaders {
			req.Header.Add(k, v)
		}
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("HttpStatusCode:%d ,Desc:%s", resp.StatusCode, resp.Status))
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//var bodyDataMap map[string]interface{};
	//err = json.Unmarshal(bodyData, &bodyDataMap);
	//if err != nil {
	//	println(string(bodyData));
	//	return nil, err;
	//}

	return bodyData, nil
}

func HttpGet(client *http.Client, reqUrl string) (map[string]interface{}, error) {
	respData, err := _httpRequest(client, "GET", reqUrl, "", nil)
	if err != nil {
		return nil, err
	}

	var bodyDataMap map[string]interface{}
	//fmt.Printf("\n%s\n", respData);
	err = json.Unmarshal(respData, &bodyDataMap)
	if err != nil {
		return nil, err
	}
	return bodyDataMap, nil
}

func HttpPost(client *http.Client, reqUrl string, postData string) ([]byte, error) {
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8"}
	return _httpRequest(client, "POST", reqUrl, postData, headers)
}

func HttpPostForm(client *http.Client, reqUrl string, postData Value) ([]byte, error) {
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8"}
	params, err := json.Marshal(postData)
	if err != nil {
		return nil, err
	}
	return _httpRequest(client, "POST", reqUrl, string(params), headers)
}

func HttpPostForm2(client *http.Client, reqUrl string, postData Value, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json;charset=UTF-8"
	return _httpRequest(client, "POST", reqUrl, postData.Encode(), headers)
}

func HttpPostForm3(client *http.Client, reqUrl string, postData string, headers map[string]string) ([]byte, error) {
	return _httpRequest(client, "POST", reqUrl, postData, headers)
}

// Value alters from url.Values.
type Value map[string]interface{}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (v Value) Get(key string) interface{} {
	if v == nil {
		return ""
	}
	vs := v[key]

	return vs
}

// Set sets the key to value. It replaces any existing
// values.
func (v Value) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Value) Del(key string) {
	delete(v, key)
}

// Encode encodes the values into ``URL encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func (v Value) Encode() string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := url.QueryEscape(k) + "="
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(prefix)
		buf.WriteString(url.QueryEscape(fmt.Sprintf("%v", vs)))
	}
	return buf.String()
}
