package v2

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"

	"upex-wallet/wallet-base/util"

	"golang.org/x/crypto/blake2b"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	v1 "upex-wallet/wallet-tools/cmd/genwallet/wallet/v1"
)

var (
	MagicBytes = blake2b.Sum256(v1.MagicBytes)
)

const (
	magicBytesLen = 32

	encryptSaltLen = 32
	encryptIVLen   = 16
	encryptMacLen  = 32

	maxPrivKeyLen = 1024
	maxPubKeyLen  = maxPrivKeyLen
	maxDataLen    = maxPrivKeyLen
)

type Wallet struct {
	dataPath       string
	inputFileName  string
	outputFileName string
	kpGenerator    keypair.Generator
	extStorer      keypair.Storer
	// 以下数据用于加密
	salt         []byte
	iv           []byte
	mac          []byte
	class        string
	cryptography keypair.CryptoClass

	keypairs           []*EncPrivateKey
	keypairPubkeyIndex map[string]*EncPrivateKey
	HexPrivateKeys     []string
	AddressStrArray    []string
}

func New(dataPath, inputFileName, outputFileName string, generator keypair.Generator, extStorer keypair.Storer) *Wallet {
	w := &Wallet{
		dataPath:       dataPath,
		inputFileName:  filepath.Join(dataPath, inputFileName),
		outputFileName: filepath.Join(dataPath, outputFileName),
		kpGenerator:    generator,
		extStorer:      extStorer,
	}
	w.Reset()
	return w
}

func (w *Wallet) SetGenerator(generator keypair.Generator) {
	w.kpGenerator = generator
}

func (w *Wallet) SetExtStorer(extStorer keypair.Storer) {
	w.extStorer = extStorer
}

// 钱包生成函数
// password 密码
// 账户数量: 系统用户 + 正常用户
func (w *Wallet) Generate(password string, count uint) error {
	defer util.DeferLogTimeCost("[generate]")()

	// string<---> 币种 "BTC"
	if w.kpGenerator == nil {
		return fmt.Errorf("there is no generator")
	}

	err := w.kpGenerator.Init()
	if err != nil {
		return err
	}

	// 定义加密器, 加密结构体
	var encryptor *util.Encryptor
	// 判断钱包中 加密数据是否为空
	if len(w.mac) == 0 {
		// 通过密码生成新的加密器
		encryptor = util.NewEncryptor(password)
		// 将加密器生成的数据复制给钱包中对应参数用于加密
		w.salt = encryptor.Salt()
		w.iv = encryptor.IV()
		w.mac = encryptor.Mac()
	} else {
		// 如果已经有加密的数据, 直接通过密码重新生成加密数据
		encryptor = util.MakeEncryptor(password, w.salt, w.iv)
		if !bytes.Equal(encryptor.Mac(), w.mac) {
			return ErrIncorrectPassword
		}
	}

	if w.Len() == 0 {
		w.initCap(int(count))
	}
	// 协程调用
	err = util.BatchDo(
		int(count),
		func(i int) (interface{}, error) {
			tryGC(i)
			kp, err := w.kpGenerator.Generate(i)
			if err != nil {
				return nil, err
			}
			return kp, nil
		},
		func(i int, data interface{}) error {
			traceProgress("generate", i+1, int(count))
			kp := data.(keypair.KeyPair)
			err = util.Try(3, func(int) error {
				pubKeyHex := hex.EncodeToString(kp.PublicKey())

				if w.Contains(pubKeyHex) {
					newKp, _ := w.kpGenerator.Generate(i)
					if newKp != nil {
						kp = newKp
					}
					return fmt.Errorf("duplicated at %d", i)
				}
				return nil
			})
			if err != nil {
				return err
			}
			if i == 0 && len(w.class) == 0 {
				w.class = kp.Class()
				w.cryptography = kp.Cryptography()
			}
			encPrivKey, err := encryptor.Encrypt(kp.PrivateKey())
			if err != nil {
				return fmt.Errorf("encrypt private key at index %d failed, %v", i, err)
			}
			// priKeyHex := hex.EncodeToString(kp.PrivateKey())
			// fmt.Printf("PrivateKey1111==============%s\n", priKeyHex)

			// 钱包添加密钥对
			w.addKeypair(encPrivKey, kp.PublicKey())
			// 将私钥单独存放
			w.addHexPrivateKey(kp.PrivateKey())
			// 钱包地址
			w.addAddressStr(kp.AddressString())
			return nil
		})

	if err != nil {
		return err
	}
	util.TraceMemStats()
	// fmt.Printf("hexPriKeys list======%s", w.HexPrivateKeys)
	return nil
}

func (w *Wallet) addKeypair(encPrivKey, pubKey []byte) {
	if len(encPrivKey) == 0 || len(pubKey) == 0 {
		return
	}

	pubKeyHex := hex.EncodeToString(pubKey)
	if w.Contains(pubKeyHex) {
		return
	}
	encPrivateKey := NewEncPrivateKey(encPrivKey, pubKey)
	w.keypairs = append(w.keypairs, encPrivateKey)
	w.keypairPubkeyIndex[pubKeyHex] = encPrivateKey
}

func readComplete(reader io.Reader) error {
	data, _ := util.ReadBytes(reader, 1)
	if len(data) > 0 {
		return fmt.Errorf("wallet file remain data")
	}
	return nil
}

func (w *Wallet) Load() error {
	defer util.DeferLogTimeCost("[load]")()

	w.Reset()

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadVarBytes(reader, magicBytesLen, "magicbytes")
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, MagicBytes[:]) {
			return fmt.Errorf("invalid wallet magic bytes")
		}

		salt, err := util.ReadVarBytes(reader, encryptSaltLen, "salt")
		if err != nil {
			return fmt.Errorf("read wallet encrypt salt failed, %v", err)
		}

		iv, err := util.ReadVarBytes(reader, encryptIVLen, "iv")
		if err != nil {
			return fmt.Errorf("read wallet encrypt iv failed, %v", err)
		}

		mac, err := util.ReadVarBytes(reader, encryptMacLen, "mac")
		if err != nil {
			return fmt.Errorf("read wallet encrypt mac failed, %v", err)
		}

		class, err := util.ReadVarString(reader)
		if err != nil {
			return fmt.Errorf("read wallet keypair class failed, %v", err)
		}

		cryptography, err := util.ReadVarString(reader)
		if err != nil {
			return fmt.Errorf("read wallet keypair cryptography failed, %v", err)
		}

		w.salt = salt
		w.iv = iv
		w.mac = mac
		w.class = class
		w.cryptography = keypair.CryptoClass(cryptography)

		num, err := util.BinarySerializer.Uint32(reader, binary.LittleEndian)
		if err != nil {
			return fmt.Errorf("read wallet keypair len failed, %v", err)
		}

		w.initCap(int(num))

		for i := 0; i < int(num); i++ {
			pubKey, err := util.ReadVarBytes(reader, maxPubKeyLen, "pubkey")
			if err != nil {
				return fmt.Errorf("read wallet public key at index %d failed, %v", i, err)
			}

			encPrivKey, err := util.ReadVarBytes(reader, maxPrivKeyLen, "encprivkey")
			if err != nil {
				return fmt.Errorf("read wallet encrypt private key at index %d failed, %v", i, err)
			}

			w.addKeypair(encPrivKey, pubKey)

			traceProgress("load", i+1, int(num))
		}

		util.TraceMemStats()

		return readComplete(reader)
	})
}

func (w *Wallet) Find(pubKeyHex string) (*EncPrivateKey, bool) {
	encKey, ok := w.keypairPubkeyIndex[pubKeyHex]
	return encKey, ok
}

func (w *Wallet) Contains(pubKeyHex string) bool {
	_, ok := w.Find(pubKeyHex)
	return ok
}

func (w *Wallet) KeypairByPubkey(password, pubKeyHex string) (keypair.KeyPair, error) {
	kp, ok := w.Find(pubKeyHex)
	if !ok {
		return nil, fmt.Errorf("can't find keypair")
	}

	return kp.KeyPair(password, w.salt, w.iv, w.mac, w.class)
}

func (w *Wallet) KeyPairAtIndex(password string, index int) (keypair.KeyPair, error) {
	if index < 0 || w.Len() <= index {
		return nil, fmt.Errorf("out of index")
	}

	kp := w.keypairs[index]
	return kp.KeyPair(password, w.salt, w.iv, w.mac, w.class)
}

func (w *Wallet) FirstKeyPair(password string) (keypair.KeyPair, error) {
	return w.KeyPairAtIndex(password, 0)
}

func (w *Wallet) LastKeyPair(password string) (keypair.KeyPair, error) {
	return w.KeyPairAtIndex(password, w.Len()-1)
}

func (w *Wallet) Len() int {
	return len(w.keypairs)
}

func (w *Wallet) Class() string {
	return w.class
}

func (w *Wallet) Cryptography() keypair.CryptoClass {
	return w.cryptography
}

func (w *Wallet) Store(password string) error {
	defer util.DeferLogTimeCost("[store]")()

	if len(w.keypairs) == 0 {
		return nil
	}

	// 私钥写入wallet.dat文件
	err := util.WithWriteFile(w.outputFileName, func(writer *bufio.Writer) error {
		err := util.WriteVarBytes(writer, MagicBytes[:])
		if err != nil {
			return fmt.Errorf("store wallet magic bytes failed, %v", err)
		}
		err = util.WriteVarBytes(writer, w.salt)
		if err != nil {
			return fmt.Errorf("store wallet encrypt salt failed, %v", err)
		}
		err = util.WriteVarBytes(writer, w.iv)
		if err != nil {
			return fmt.Errorf("store wallet encrypt iv failed, %v", err)
		}
		err = util.WriteVarBytes(writer, w.mac)
		if err != nil {
			return fmt.Errorf("store wallet encrypt mac failed, %v", err)
		}
		err = util.WriteVarString(writer, w.class)
		if err != nil {
			return fmt.Errorf("store wallet keypair class failed, %v", err)
		}
		err = util.WriteVarString(writer, string(w.cryptography))
		if err != nil {
			return fmt.Errorf("store wallet keypair cryptography failed, %v", err)
		}
		err = util.BinarySerializer.PutUint32(writer, binary.LittleEndian, uint32(len(w.keypairs)))
		if err != nil {
			return fmt.Errorf("store wallet keypair len failed, %v", err)
		}

		for i, kp := range w.keypairs {
			// 将公钥文件写入文件钱包文件
			err = util.WriteVarBytes(writer, kp.pubKeyData)
			if err != nil {
				return fmt.Errorf("store wallet public key at index %d failed, %v", i, err)
			}
			// 将私钥文件写钱包wallet.dat文件
			err = util.WriteVarBytes(writer, kp.encPrivKey)
			if err != nil {
				return fmt.Errorf("store wallet encrypt private key at index %d failed, %v", i, err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	util.TraceMemStats()

	if w.extStorer != nil {
		w.extStorer.Open(w.dataPath)
		defer w.extStorer.Close()

		err := util.BatchDo(
			w.Len(),
			func(i int) (interface{}, error) {
				tryGC(i)
				pubKey, err := w.keypairs[i].PublicKey(password, w.salt, w.iv, w.mac, w.class)
				if err != nil {
					return nil, err
				}
				return pubKey, nil
			},
			func(i int, data interface{}) error {
				traceProgress("store", i+1, w.Len())
				return w.extStorer.Append(data.(keypair.PublicKey))
			})
		if err != nil {
			return err
		}

		util.TraceMemStats()
	}

	return nil
}

func (w *Wallet) ChangePassword(password, newPassword string) error {
	if w.Len() == 0 {
		return nil
	}

	oldEncryptor := util.MakeEncryptor(password, w.salt, w.iv)
	if !bytes.Equal(oldEncryptor.Mac(), w.mac) {
		return ErrIncorrectPassword
	}

	if newPassword == password {
		return nil
	}

	newEncryptor := util.NewEncryptor(newPassword)
	for i, kp := range w.keypairs {
		privKey, err := oldEncryptor.Decrypt(kp.encPrivKey)
		if err != nil {
			return fmt.Errorf("decrypt private key at index %d failed, %v", i, err)
		}

		kp.encPrivKey, err = newEncryptor.Encrypt(privKey)
		if err != nil {
			return fmt.Errorf("encrypt private key at index %d failed, %v", i, err)
		}
	}

	w.salt = newEncryptor.Salt()
	w.iv = newEncryptor.IV()
	w.mac = newEncryptor.Mac()

	return nil
}

func (w *Wallet) initCap(size int) {
	if size < 0 {
		size = 0
	}

	w.keypairs = make([]*EncPrivateKey, 0, size)
	w.keypairPubkeyIndex = make(map[string]*EncPrivateKey, size)
}

func (w *Wallet) Reset() {
	w.salt = nil
	w.iv = nil
	w.mac = nil
	w.class = ""
	w.cryptography = keypair.InvalidCryptoClass

	w.initCap(0)
}

// 16进制私钥添加到钱包中
func (w *Wallet) addHexPrivateKey(priKey []byte) {
	if len(priKey) == 0 {
		return
	}
	// 私钥转16进制字符串
	hexPriKey := hex.EncodeToString(priKey)
	w.HexPrivateKeys = append(w.HexPrivateKeys, hexPriKey)
}

func (w *Wallet) addAddressStr(addressStr string) {
	if addressStr == "" {
		return
	}
	w.AddressStrArray = append(w.AddressStrArray, addressStr)

}
