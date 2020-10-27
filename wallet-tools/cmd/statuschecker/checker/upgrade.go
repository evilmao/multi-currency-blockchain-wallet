package checker

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"gopkg.in/resty.v1"
)

func init() {
	resty.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))
}

type UpgradeChecker struct {
	service.SimpleWorker
	items []*UpgradeCheckerItem
}

func NewUpgradeChecker(urls map[string]string) *UpgradeChecker {
	c := &UpgradeChecker{}

	for url, pattern := range urls {
		item, err := NewUpgradeCheckerItem(url, pattern)
		if err != nil {
			log.Errorf("%s, %v", c.Name(), err)
			continue
		}

		c.items = append(c.items, item)
	}

	return c
}

func (c *UpgradeChecker) Name() string {
	return "UpgradeChecker"
}

func (c *UpgradeChecker) Work() {
	for _, item := range c.items {
		err := item.check()
		if err != nil {
			log.Errorf("%s, %v", c.Name(), err)
		}
	}
}

var (
	upgradePostPathRegStr = `<a href="(.+?)".+?`
	upgradePostTimeReg    = regexp.MustCompile(`time datetime="(.+?)"`)
)

type UpgradeCheckerItem struct {
	url  string
	host string
	reg  *regexp.Regexp

	lastCheckTime time.Time
}

func NewUpgradeCheckerItem(urlStr, pattern string) (*UpgradeCheckerItem, error) {
	if len(urlStr) == 0 {
		return nil, fmt.Errorf("url can't be empty")
	}

	if len(pattern) == 0 {
		return nil, fmt.Errorf("pattern can't be emtpy")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse url %s failed, %v", urlStr, err)
	}

	pattern = upgradePostPathRegStr + `(` + pattern + `)`
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile regexp %s failed, %v", pattern, err)
	}

	return &UpgradeCheckerItem{
		url:  urlStr,
		host: u.Scheme + "://" + u.Host,
		reg:  reg,

		lastCheckTime: time.Now(),
	}, nil
}

func (item *UpgradeCheckerItem) check() error {
	data, err := util.RestRequest("get", item.url, nil, nil)
	if err != nil {
		return fmt.Errorf("get url %s failed, %v", item.url, err)
	}

	results := item.reg.FindAllStringSubmatch(string(data), -1)
	for _, matches := range results {
		if len(matches) < 3 {
			continue
		}

		postURL := item.host + matches[1]
		postTime, err := item.getPostTime(postURL)
		if err != nil {
			return fmt.Errorf("get %s post time failed, %v", postURL, err)
		}

		if postTime.Before(item.lastCheckTime) {
			continue
		}

		postTitle := matches[2]
		body := fmt.Sprintf(`find currency upgrade post "%s" from %s at %s.`,
			postTitle, postURL, postTime.Format("2006-01-02 15:04:05"))
		warn(body)
	}

	item.lastCheckTime = time.Now()
	return nil
}

func (item *UpgradeCheckerItem) getPostTime(url string) (tm time.Time, err error) {
	tm = time.Unix(0, 0)
	data, err := util.RestRequest("get", url, nil, nil)
	if err != nil {
		return
	}

	results := upgradePostTimeReg.FindStringSubmatch(string(data))
	if len(results) == 0 {
		err = fmt.Errorf("can't find time data")
		return
	}

	t, err := time.Parse(time.RFC3339, results[1])
	if err != nil {
		return
	}

	return t, nil
}
