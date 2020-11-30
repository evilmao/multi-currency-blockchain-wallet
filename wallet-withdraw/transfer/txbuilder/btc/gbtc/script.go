package gbtc

import (
	"bytes"

	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-tools/base/crypto"
)

const (
	OP_DUP = 0x76

	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88

	OP_HASH160 = 0xa9

	OP_CHECKSIG = 0xac
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
	buf.WriteByte(OP_DUP)
	buf.WriteByte(OP_HASH160)
	util.WriteVarBytes(buf, pkHash)
	buf.WriteByte(OP_EQUALVERIFY)
	buf.WriteByte(OP_CHECKSIG)
	return buf.Bytes()
}

func P2SHScript(sHash []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(OP_HASH160)
	util.WriteVarBytes(buf, sHash)
	buf.WriteByte(OP_EQUAL)
	return buf.Bytes()
}

func P2Bech32(version byte, data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(version)
	util.WriteVarBytes(buf, data)
	return buf.Bytes()
}
