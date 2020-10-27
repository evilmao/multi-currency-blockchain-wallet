package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
)

const (
	privKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApVt3CUFtZPmq8f3kTNESefICQt8akrxYUsfAFkjDtHuSh8k/
Jy7Y86rnYQV9db/SmSpdfAL9YshIKR9xJWc48W8rrtVmmQpPmo4GwNDYB0J/Jv+Y
N6ZyRqTD32D7hHPwYrzJ+QzbiG5Yo0vj0JE12EPZWHW1ozprcufIu57Fqux05peL
BC4Tq8/6uHLPv57Nu2E3vG/Z+DDxbOYcu29xh7CI/japHH+F9V5o2RupjD1vbEka
+ebZQWbPdbZLf5y5pcNqATkW9Ay1dSYj4TG5RjSl6BshRNgxm97W+fkeibAuNgDv
pC08GVctFz2F//52jyrvsFkIDPT/fbsbMwq8oQIDAQABAoIBAESbIVVGvxOQ/tru
QyWX7PmJbmS+WjEdUevukoTsZb5hMteBqOEh78ORWfSIKZiMIN2JlUXZm7W7cS0W
rYQeQqLxRAeC5NGEVKjEWorvW2IPtd/BVi9osKscu2PXwiMfU1I0D/xz2DXPypjd
6MRlKxjydZ/dHqJ9n40KG9+0LUEPtRoKt53DGobtLrqOmOeA7u5jAOFdF25yyV32
5kh6+W1SJUnv9GdW51NeNa395RE380+6LKILQDFMI72tCAhMx0e5AFOz4QWlpKpF
Tg39aN2D8Vvs0DHTGvd4KoX4rTFSoCNXJ628feNbaRuNH6L1miPBECLGQIC0t/28
XTg6Rp0CgYEAwaY5iK1TMg4dZV3fXO1DFS7kdNK7UmQwKBvBCDGQdutNVLTo9Dgm
cregzUn9QdTpFK8aHXvEtEln8svxne/Tdsmg01nNYbQ7c+SfpIaTBHSTI4yxxVwg
YIokU7Mtx2R121zbd55PwGAprj/igilecMYBeO5zUyBfRYDi7Y1VR68CgYEA2pk8
aSBdQ+g1rfBClGwm4WgGLa8A8VSVF2DKNJogyjJhRyJXrHtSc0LPa49i6qqoK439
2b4K7NKYtczefKXzWYHkDeyqTPvkZy2EzyRu1UD7hEKXocyS+mxkgghQtcY+KRpc
SDcVpYbfBwmI34PWejKKcpaQ/TFz35rx91FgBK8CgYBgfsUynzPSwIfTaCiSdMQ1
vQ0oTY38a2I3ykSxIYmcSHpbWF6wu34lMe2V/mWNtVuD7BE2WeNV9zIuIYQ/sC8O
hUB3sMsQAbCSen02jbyavsBHOaen8dVMZeneL24DasLz0VynSaLx+LksVDc5pwWh
anl3WlLrPDldN/FccE3rjwKBgFcTl2bhB4XXaBqjjEIHWu2LPHrSLXP0l3c6jRGr
G8ivjOSDH52LemqDgJB+C48i7955r1cfRsbTlRVGSJIqoOdUwH1zetszs+YN0cuZ
3bSBMC+dPz2qehnbN6Y8nbnrADPrVjtGBg9rzEfHWoh+wd7nZxMCOztAicHWvPAh
sDftAoGBAL9GZeTyC241vQNBagBIfVDofEEbsImUXusQSDnKFs8cXkhDfIL9rs1d
YD8tA9SRRM2x67x+/HAty8+Y7Rz+nPhVtYEvR6R1+dg8iHOAl9QGSeXr+P++9Phb
GCP0lVD2+P9cTa5qZdtpssR0sDY36oQAswtHPwat+xdczDNZHQy+
-----END RSA PRIVATE KEY-----`

	pubKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApVt3CUFtZPmq8f3kTNES
efICQt8akrxYUsfAFkjDtHuSh8k/Jy7Y86rnYQV9db/SmSpdfAL9YshIKR9xJWc4
8W8rrtVmmQpPmo4GwNDYB0J/Jv+YN6ZyRqTD32D7hHPwYrzJ+QzbiG5Yo0vj0JE1
2EPZWHW1ozprcufIu57Fqux05peLBC4Tq8/6uHLPv57Nu2E3vG/Z+DDxbOYcu29x
h7CI/japHH+F9V5o2RupjD1vbEka+ebZQWbPdbZLf5y5pcNqATkW9Ay1dSYj4TG5
RjSl6BshRNgxm97W+fkeibAuNgDvpC08GVctFz2F//52jyrvsFkIDPT/fbsbMwq8
oQIDAQAB
-----END PUBLIC KEY-----`
)

func TestRSA(t *testing.T) {
	const data = "456"
	encData, err := rsa.B64Encrypt(data, pubKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(encData)

	raw, err := rsa.B64Decrypt(encData, privKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(raw), string(raw) == data)
}

func TestPriKeyToPubKey(t *testing.T) {
	block, _ := pem.Decode([]byte(privKey))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := rsa.ExportPubkey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(pub))
}
