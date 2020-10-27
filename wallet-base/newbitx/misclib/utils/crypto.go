package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"

	"github.com/howeyc/gopass"
)

var (
	Salt = HexToBytes("f6f5759b77b272ea4702d1ffb6d34c68527673c53331770f0e5a85cf25b118bd")
	Iv   = HexToBytes("3e12b3667df43b652c02827c27f876f8")
	Key  = "b9b6650194171059026ea56f5a94e0d4b8114833e655e99b4f55990b1a1901db"
)

// GetPassword gets password you input.
func GetPassword() string {
	fmt.Printf("Password:")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}
	return string(pass)
}

// GetHmacSign calculates and returns the request sign.
func GetHmacSign(data []byte, key []byte) string {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	return BytesToHex(mac.Sum(nil))
}
