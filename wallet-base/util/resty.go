package util

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "upex-wallet/wallet-base/newbitx/misclib/log"

    "gopkg.in/resty.v1"
)

const (
    // StatusOK represents the api response status.
    StatusOK  = iota
    Status200 = "000"

    // RestyMaxRetryCount is the max retry times.
    RestyMaxRetryCount = 3
)

func init() {
    resty.
        SetLogger(log.GetOutput()).
        SetRetryCount(RestyMaxRetryCount).
        SetTimeout(5 * time.Second)
}

func restyStatusError(resp *resty.Response) error {
    if resp.StatusCode() > http.StatusOK {
        return fmt.Errorf("resty request failed, %s (%d), %s", resp.Status(), resp.StatusCode(), string(resp.Body()))
    }

    return nil
}

// Response represents the server response message.
type Response struct {
    Status interface{} `json:"errno"`
    Msg    string      `json:"errmsg"`
    Data   interface{} `json:"data"`
}

// RestPost wrappers a simple http restfull post request.
func RestPost(data interface{}, url string) (interface{}, int, error) {

    respData, err := RestRawPost(url, data)
    if err != nil {
        return nil, RestyMaxRetryCount, err
    }

    var (
        resp      Response
        requestOK bool
    )
    err = json.Unmarshal(respData, &resp)
    if err != nil {
        return nil, RestyMaxRetryCount, fmt.Errorf("decode response from api fail, request url:%s, detail %v", url, err)
    }

    resStatus := resp.Status
    switch resStatus.(type) {
    case int:
        requestOK = resStatus.(int) == StatusOK
    case string:
        requestOK = resStatus.(string) == Status200
    }

    if requestOK && !strings.Contains(strings.ToLower(resp.Msg), "error") {
        return resp.Data, 1, nil
    }

    return nil, RestyMaxRetryCount, fmt.Errorf("api response failed, %s", string(respData))
}

// RestGet wrappers a simple http restfull get request.
func RestGet(data map[string]string, url string) (interface{}, int, error) {
    respData, err := RestRawGet(url, data)
    if err != nil {
        return nil, RestyMaxRetryCount, err
    }

    var resp Response
    json.Unmarshal(respData, &resp)
    if resp.Status == StatusOK {
        return resp.Data, 1, nil
    }

    return nil, RestyMaxRetryCount, fmt.Errorf("api response failed, %s", string(respData))
}

// RestRawGet wrappers a simple http restfull get request.
func RestRawGet(url string, params map[string]string) ([]byte, error) {
    return RestRequest("get", url, map[string]string{"Accept": "application/json"}, params)
}

// RestRawPost wrappers a simple http restfull post request.
func RestRawPost(url string, data interface{}) ([]byte, error) {
    return RestRequest("post", url, map[string]string{"Accept": "application/json"}, data)
}

// RestRequest wrappers a simple http restfull request.
func RestRequest(method, url string, headers map[string]string, data interface{}) ([]byte, error) {
    req := resty.R()
    req.SetHeaders(headers)

    var (
        resp *resty.Response
        err  error
    )
    if strings.EqualFold(method, "get") {
        if data != nil {
            if data, ok := data.(map[string]string); !ok {
                return nil, fmt.Errorf("query params must be map[string]string type")
            } else {
                req.SetQueryParams(data)
            }
        }
        resp, err = req.Get(url)
    } else {
        req.SetBody(data)
        resp, err = req.Post(url)
    }
    if err != nil {
        return nil, err
    }

    if err = restyStatusError(resp); err != nil {
        return nil, err
    }

    return resp.Body(), err
}
