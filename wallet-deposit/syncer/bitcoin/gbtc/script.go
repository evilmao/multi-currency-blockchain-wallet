package gbtc

import (
	"bytes"

	"github.com/fb996de/wallet-base/util"
	"github.com/fb996de/wallet-tools/base/crypto"
)

const (
	OP_DUP = 0x76

	OP_EQUALVERIFY = 0x88

	OP_HASH160 = 0xa9

	OP_CHECKSIG = 0xac
)

func CreateP2PKHScript(address string, prefixLen uint8) []byte {
	if len(address) == 0 {
		return nil
	}

	_, h := crypto.DeBase58Check(address, prefixLen, false)

	buf := new(bytes.Buffer)
	buf.WriteByte(OP_DUP)
	buf.WriteByte(OP_HASH160)
	util.WriteVarBytes(buf, h)
	buf.WriteByte(OP_EQUALVERIFY)
	buf.WriteByte(OP_CHECKSIG)
	return buf.Bytes()
}
