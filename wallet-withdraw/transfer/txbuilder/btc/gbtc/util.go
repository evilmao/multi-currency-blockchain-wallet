package gbtc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/antonholmquist/jason"
)

func childUnmarshal(data []byte, result interface{}, key ...string) error {
	object, err := jason.NewObjectFromBytes(data)
	if err != nil {
		return err
	}
	child, err := object.GetObject(key...)
	if err != nil {
		return err
	}
	childData, err := child.Marshal()
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader([]byte(childData))).Decode(&result)
}

func SHA256D(data []byte) []byte {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])
	return h[:]
}

func ReverseBytes(data []byte) []byte {
	tmp := make([]byte, len(data))
	copy(tmp, data)
	for i := 0; i < len(tmp)/2; i++ {
		j := len(tmp) - i - 1
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp
}

func HashToString(hash []byte) string {
	return hex.EncodeToString(ReverseBytes(hash))
}

func StringToHash(s string) ([]byte, error) {
	hashData, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return ReverseBytes(hashData), nil
}
