package v1

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

var (
	MagicBytes, _ = hex.DecodeString("0901419396d7679bf46d7c0c28a7a8eb2d793bea3c9bea222e7eedc77dc7e174")
)

const (
	magicBytesLen = 32

	encryptSaltLen = 32
	encryptIVLen   = 16

	maxPrivKeyLen = 1024
	maxDataLen    = maxPrivKeyLen
)

type Wallet struct {
	password    string
	dataPath    string
	fileName    string
	kpGenerator keypair.Generator
	extStorer   keypair.Storer

	keypairs           []keypair.KeyPair
	keypairAddrIndex   map[string]keypair.KeyPair
	keypairPubkeyIndex map[string]keypair.KeyPair
}

func New(password, dataPath, fileName string, generator keypair.Generator, extStorer keypair.Storer) *Wallet {
	w := &Wallet{
		password:    password,
		dataPath:    dataPath,
		fileName:    filepath.Join(dataPath, fileName),
		extStorer:   extStorer,
		kpGenerator: generator,
	}
	w.Reset()
	return w
}

func (w *Wallet) Generate(count uint) error {
	if w.kpGenerator == nil {
		return fmt.Errorf("there is no generator")
	}

	err := w.kpGenerator.Init()
	if err != nil {
		return err
	}

	for i := 0; i < int(count); i++ {
		var (
			kp  keypair.KeyPair
			err error
		)
		err = util.Try(3, func(int) error {
			kp, err = w.kpGenerator.Generate(i)
			if err != nil {
				return err
			}

			if _, existed := w.keypairAddrIndex[kp.AddressString()]; existed {
				return fmt.Errorf("duplicated at %d", i)
			}

			return nil
		})
		if err != nil {
			return err
		}

		w.keypairs = append(w.keypairs, kp)
		w.keypairAddrIndex[kp.AddressString()] = kp
		w.keypairPubkeyIndex[hex.EncodeToString(kp.PublicKey())] = kp
	}
	return nil
}

func (w *Wallet) Load() error {
	w.Reset()

	return util.WithReadFile(w.fileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadVarBytes(reader, magicBytesLen, "magicbytes")
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, MagicBytes) {
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

		encryptor := util.MakeEncryptor(w.password, salt, iv)

		encMagic, err := util.ReadVarBytes(reader, maxDataLen, "encmagic")
		if err != nil {
			return fmt.Errorf("read wallet encypt magic bytes failed, %v", err)
		}

		magic, err = encryptor.Decrypt(encMagic)
		if err != nil {
			return fmt.Errorf("decrypte magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, MagicBytes) {
			return fmt.Errorf("incorrect wallet password")
		}

		num, err := util.BinarySerializer.Uint32(reader, binary.LittleEndian)
		if err != nil {
			return fmt.Errorf("read wallet keypair len failed, %v", err)
		}

		var class string
		for i := 0; i < int(num); i++ {
			if i == 0 {
				class, err = util.ReadVarString(reader)
				if err != nil {
					return fmt.Errorf("read wallet keypair class failed, %v", err)
				}
			}

			encPrivKey, err := util.ReadVarBytes(reader, maxPrivKeyLen, "encprivkey")
			if err != nil {
				return fmt.Errorf("read wallet encrypt private key at index %d failed, %v", i, err)
			}

			privKey, err := encryptor.Decrypt(encPrivKey)
			if err != nil {
				return fmt.Errorf("decrypte private key at index %d failed, %v", i, err)
			}

			kp, err := keypair.Build(class, privKey)
			if err != nil {
				return fmt.Errorf("build keypair at index %d failed, %v", i, err)
			}

			if _, existed := w.keypairAddrIndex[kp.AddressString()]; existed {
				continue
			}

			w.keypairs = append(w.keypairs, kp)
			w.keypairAddrIndex[kp.AddressString()] = kp
			w.keypairPubkeyIndex[hex.EncodeToString(kp.PublicKey())] = kp
		}
		return nil
	})
}

func (w *Wallet) KeyPair(address string) (keypair.KeyPair, bool) {
	kp, ok := w.keypairAddrIndex[address]
	return kp, ok
}

func (w *Wallet) KeypairByPubkey(pubkey string) (keypair.KeyPair, bool) {
	kp, ok := w.keypairPubkeyIndex[pubkey]
	return kp, ok
}

func (w *Wallet) KeyPairAtIndex(index int) (keypair.KeyPair, bool) {
	if index < 0 || w.Len() <= index {
		return nil, false
	}

	return w.keypairs[index], true
}

func (w *Wallet) FirstKeyPair() (keypair.KeyPair, bool) {
	return w.KeyPairAtIndex(0)
}

func (w *Wallet) LastKeyPair() (keypair.KeyPair, bool) {
	return w.KeyPairAtIndex(w.Len() - 1)
}

func (w *Wallet) Foreach(fun func(int, keypair.KeyPair) (bool, error)) error {
	for i, kp := range w.keypairs {
		ok, err := fun(i, kp)
		if err != nil {
			return err
		}

		if !ok {
			break
		}
	}
	return nil
}

func (w *Wallet) Len() int {
	return len(w.keypairs)
}

func (w *Wallet) Class() string {
	if kp, ok := w.FirstKeyPair(); ok {
		return kp.Class()
	}

	return ""
}

func (w *Wallet) Cryptography() keypair.CryptoClass {
	if kp, ok := w.FirstKeyPair(); ok {
		return kp.Cryptography()
	}

	return keypair.InvalidCryptoClass
}

func (w *Wallet) Store() error {
	if len(w.keypairs) == 0 {
		return nil
	}

	err := util.WithWriteFile(w.fileName, func(writer *bufio.Writer) error {
		err := util.WriteVarBytes(writer, MagicBytes)
		if err != nil {
			return fmt.Errorf("store wallet magic bytes failed, %v", err)
		}

		encryptor := util.NewEncryptor(w.password)
		err = util.WriteVarBytes(writer, encryptor.Salt())
		if err != nil {
			return fmt.Errorf("store wallet encrypt salt failed, %v", err)
		}

		err = util.WriteVarBytes(writer, encryptor.IV())
		if err != nil {
			return fmt.Errorf("store wallet encrypt iv failed, %v", err)
		}

		encMagic, err := encryptor.Encrypt(MagicBytes)
		if err != nil {
			return fmt.Errorf("encrypt magic bytes failed, %v", err)
		}

		err = util.WriteVarBytes(writer, encMagic)
		if err != nil {
			return fmt.Errorf("store wallet encrypt magic bytes failed, %v", err)
		}

		err = util.BinarySerializer.PutUint32(writer, binary.LittleEndian, uint32(len(w.keypairs)))
		if err != nil {
			return fmt.Errorf("store wallet keypair len failed, %v", err)
		}

		return w.Foreach(func(i int, kp keypair.KeyPair) (bool, error) {
			if i == 0 {
				err := util.WriteVarString(writer, kp.Class())
				if err != nil {
					return false, fmt.Errorf("store wallet keypair class failed, %v", err)
				}
			}

			encPrivKey, err := encryptor.Encrypt(kp.PrivateKey())
			if err != nil {
				return false, fmt.Errorf("encrypt private key at index %d failed, %v", i, err)
			}

			err = util.WriteVarBytes(writer, encPrivKey)
			if err != nil {
				return false, fmt.Errorf("store wallet encrypt private key at index %d failed, %v", i, err)
			}

			return true, nil
		})
	})

	if err != nil {
		return err
	}

	if w.extStorer != nil {
		w.extStorer.Open(w.dataPath)
		defer w.extStorer.Close()

		for _, kp := range w.keypairs {
			err := w.extStorer.Append(kp.(keypair.PublicKey))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Wallet) Reset() {
	w.keypairs = nil
	w.keypairAddrIndex = make(map[string]keypair.KeyPair)
	w.keypairPubkeyIndex = make(map[string]keypair.KeyPair)
}
