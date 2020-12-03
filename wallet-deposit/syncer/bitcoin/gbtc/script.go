package gbtc

import (
	"bytes"

	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-tools/base/crypto"
)

const (
	OpDup          = 0x76
	OpEqualCVerify = 0x88
	OpHash160      = 0xa9
	OpCheckSig     = 0xac
)

func CreateP2PKHScript(address string, prefixLen uint8) []byte {
	if len(address) == 0 {
		return nil
	}

	_, h := crypto.DeBase58Check(address, prefixLen, false)

	buf := new(bytes.Buffer)
	buf.WriteByte(OpDup)
	buf.WriteByte(OpHash160)
	_ = util.WriteVarBytes(buf, h)
	buf.WriteByte(OpEqualCVerify)
	buf.WriteByte(OpCheckSig)
	return buf.Bytes()
}
