package util

import (
    "fmt"

    "upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
)

// transfer服务中sign--> pass签名脚本
func signPassToEncrypt(suffixPass, signerPicKey string) string {
    signStr, _ := rsa.B64Encrypt(suffixPass, signerPicKey)
    fmt.Println(signStr)
    return signStr
}

func main() {
    var (
        suffixPass   = "456"
        signerPicKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApRQ/IgIgNZe/toQAmC7C\ncRgNqYSM979cjF2Fazo3GhdnQEV4AT/UbGUxCu6AQLKWvTRoWDD2elzafuM08dES\navj0+rNzaZWnJaFZy49XzaCIuGq6hrc4rQjy8VSv/2eK3L2EG9dRQe+NtBzlFIGn\nzZsLP36jL0dbsVFz3Cqtltsk/DLfMwyo1de/wGvyAV7llvbRyLrI7YcuoaJ4PVW9\n5q6fPsnBapfEjIjiYwLl/QayKv9OLrDyZVI9EybFeBNDf+2ZfCdel/pc8BmUImvE\nf8l9d3CjUjtOGkZoLT+Ah2sQ+eNRjiQxKWbSDpDok4xwYJGk/WfQzwggmuZpif3H\nbQIDAQAB\n-----END PUBLIC KEY-----\n"
    )
    signStr := signPassToEncrypt(suffixPass, signerPicKey)
    fmt.Printf("transfers密码:%s", signStr)
}
