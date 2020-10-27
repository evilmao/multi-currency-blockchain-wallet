package checker

import (
	"fmt"
	"testing"
)

func TestUpgradeChecker(t *testing.T) {
	warn = func(body string) {
		fmt.Println(body)
	}

	c := NewUpgradeChecker(map[string]string{
		"https://huobiglobal.zendesk.com/hc/zh-cn/sections/360000039481": "关于.+?[暂停,升级].+?公告",
	})
	c.Work()
}
