package gbtc

import (
	"encoding/hex"
	"testing"
)

func assert(t *testing.T, ok bool, format string, a ...interface{}) {
	if !ok {
		t.Fatalf(format, a...)
	}
}

func TestAddress(t *testing.T) {
	for a, s := range map[string]string{
		"1H4V33BQsPNx2eQEAaCyF2oBXN5JCUmuTa":                             "76a914b02aa9b8af95d296f89c5508bd4db7515fce94e188ac",
		"3DdBZ734DjaqBf29PvjxWv9mvU6VaQEs8H":                             "a91482e7e273ec4e0dff2590e50d5fec2ebc665cd1f787",
		"bc1qzhayf65p2j4h3pfw22aujgr5w42xfqzx5uvddt":                     "001415fa44ea8154ab78852e52bbc920747554648046",
		"bc1qwgwu5ndyceq3q2e2022naseszfzxmqxt4p05v494qqqktc9q59nsrrwt9g": "0020721dca4da4c641102b2a7a953ec33012446d80cba85f4654b5000165e0a0a167",
	} {
		addr, err := ParseAddress(a, AddressParamBTC)
		assert(t, err == nil, "parse address %s failed, %v", a, err)
		assert(t, addr.String() == a, "address mismatch, %s vs %s", addr.String(), a)
		assert(t, hex.EncodeToString(addr.Script()) == s, "script mismatch, %x vs %s", addr.Script(), s)
	}
}
