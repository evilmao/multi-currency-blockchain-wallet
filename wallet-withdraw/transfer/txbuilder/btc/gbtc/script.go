package gbtc

import (
	"bytes"

	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-tools/base/crypto"
)

const (
	OpDup          = 0x76
	OPEqual        = 0x87
	OpEqualCVerify = 0x88
	OpHash160      = 0xa9
	OpCheckSig     = 0xac
)

func CreateP2PKHScript(address string, prefixLen uint8) []byte {
	if len(address) == 0 {
		return nil
	}

	_, h := crypto.DeBase58Check(address, prefixLen, false)
	return P2PKHScript(h)
}

func P2PKHScript(pkHash []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(OpDup)
	buf.WriteByte(OpHash160)
	_ = util.WriteVarBytes(buf, pkHash)
	buf.WriteByte(OpEqualCVerify)
	buf.WriteByte(OpCheckSig)
	return buf.Bytes()
}

func P2SHScript(sHash []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(OpHash160)
	_ = util.WriteVarBytes(buf, sHash)
	buf.WriteByte(OPEqual)
	return buf.Bytes()
}

func P2Bech32(version byte, data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(version)
	_ = util.WriteVarBytes(buf, data)
	return buf.Bytes()
}
