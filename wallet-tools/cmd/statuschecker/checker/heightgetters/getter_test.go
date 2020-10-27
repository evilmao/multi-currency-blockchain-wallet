package heightgetters

import (
	"fmt"
	"testing"
)

var (
	params = map[string]string{
		"ADA": "http://127.0.0.1:5517",
		"TRX": "http://127.0.0.1:5516",
		"ETH": "http://127.0.0.1:5500",
		//"DCAR": "http://127.0.0.1:6015",
		//"ETP":  "http://127.0.0.1:5514/rpc/v3/",
		"PTN":  "http://127.0.0.1:6016",
		"KMD":  "http://kmd:75RWa8vkjmNv36oE@127.0.0.1:6017",
		"EOS":  "http://127.0.0.1:8888",
		"XLM":  "http://127.0.0.1:5506",
		"XEM":  "http://127.0.0.1:5510",
		"FT":   "http://127.0.0.1:6026",
		"INT":  "http://127.0.0.1:6021",
		"ZIL":  "http://127.0.0.1:xxxx",
		"NAS":  "http://127.0.0.1:6035",
		"ALGO": "http://127.0.0.1:6027#fa3cc7ee04419d2ff380c7741169fe55d2ebd0ef6efaf4d6b984f19a72dce08f",
		"FAB":  "http://TbO8rI8wfYvKTY0D:TbO8rI8wfYvKTY0Dc1@127.0.0.1:6019",
	}
)

func TestGetter(t *testing.T) {
	var failedN int
	for name, url := range params {
		getter, ok := Get(name)
		if !ok {
			fmt.Println(name, "can't find getter")
			failedN++
			continue
		}

		height, err := getter(url)
		if err != nil {
			fmt.Println(name, err)
			failedN++
			continue
		}
		fmt.Println(name, height)
	}

	if failedN > 0 {
		t.Fatal("failed")
	}
}
