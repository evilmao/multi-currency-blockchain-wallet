package zillean

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	crypto "github.com/GincoInc/go-crypto"
	zillean "github.com/GincoInc/zillean/proto"
	"github.com/golang/protobuf/proto"
)

func publicKeyToAddress(publicKey []byte) string {
	return fmt.Sprintf("%x", crypto.Sha256(publicKey)[12:])
}

func encodeTransaction(rawTx RawTransaction) []byte {
	toAddr, _ := hex.DecodeString(rawTx.To)
	_pubKey, _ := hex.DecodeString(rawTx.PubKey)
	pubKey := zillean.ByteArray{Data: _pubKey}
	_amount := &big.Int{}
	_amount.SetString(rawTx.Amount, 10)
	amount := zillean.ByteArray{Data: bigIntToPaddedBytes(_amount, 32)}
	gasPrice := zillean.ByteArray{Data: bigIntToPaddedBytes(rawTx.GasPrice, 32)}

	protoTxCoreInfo := zillean.ProtoTransactionCoreInfo{
		Version:      &rawTx.Version,
		Nonce:        &rawTx.Nonce,
		Toaddr:       toAddr,
		Senderpubkey: &pubKey,
		Amount:       &amount,
		Gasprice:     &gasPrice,
		Gaslimit:     &rawTx.GasLimit,
	}
	if rawTx.Code != "" {
		protoTxCoreInfo.Code = []byte(rawTx.Code)
	}
	if rawTx.Data != "" {
		protoTxCoreInfo.Data = []byte(rawTx.Data)
	}
	encodedTx, _ := proto.Marshal(&protoTxCoreInfo)

	return encodedTx
}

func bigIntToPaddedBytes(i *big.Int, paddedSize int32) []byte {
	bytes := i.Bytes()
	padded, _ := hex.DecodeString(fmt.Sprintf("%0*x", paddedSize, bytes))
	return padded
}

func hash(Q []byte, pubKey []byte, msg []byte) []byte {
	var buffer bytes.Buffer
	buffer.Write(Q)
	buffer.Write(pubKey[:33])
	buffer.Write(msg)

	return crypto.Sha256(buffer.Bytes())
}

func generateDRN(nonce []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(generateRandomBytes(32))
	buffer.WriteString("Schnorr+SHA256  ")
	hmacDRBG := crypto.NewHmacDRBG(generateRandomBytes(32), nonce, buffer.Bytes())

	return hmacDRBG.Generate(int32(32), []byte{})
}

func generateRandomBytes(size int32) []byte {
	randomBytes := make([]byte, size)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}
