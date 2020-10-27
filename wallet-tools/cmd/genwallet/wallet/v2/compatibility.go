package v2

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"upex-wallet/wallet-tools/base/crypto"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	v1 "upex-wallet/wallet-tools/cmd/genwallet/wallet/v1"

	"upex-wallet/wallet-base/util"
)

func (w *Wallet) LoadPyBTC(password string, class string) error {
	w.Reset()

	var (
		BlockSize = 32
		Padding   = []byte{'{'}

		pad = func(data []byte) []byte {
			return append(data, bytes.Repeat(Padding, (BlockSize-len(data))%BlockSize)...)
		}

		NumLen        = 8
		H160Len       = 20
		EncPrivKeyLen = 88
	)

	encryptor := util.NewEncryptor(password)
	w.salt = encryptor.Salt()
	w.iv = encryptor.IV()
	w.mac = encryptor.Mac()

	aesCipher, err := aes.NewCipher(pad([]byte(password)))
	if err != nil {
		return fmt.Errorf("create aes cipher failed, %v", err)
	}

	ecb := crypto.NewECBDecrypter(aesCipher)

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		buf, err := util.ReadBytes(reader, NumLen)
		if err != nil {
			return fmt.Errorf("read num failed, %v", err)
		}

		n := binary.LittleEndian.Uint32(buf)
		for i := 0; i < int(n); i++ {
			h160Addr, err := util.ReadBytes(reader, H160Len)
			if err != nil {
				return fmt.Errorf("read addr at index %d failed, %v", i, err)
			}

			privKey, err := util.ReadBytes(reader, EncPrivKeyLen)
			if err != nil {
				return fmt.Errorf("read privkey at index %d failed, %v", i, err)
			}

			privKey, err = base64.StdEncoding.DecodeString(string(privKey))
			if err != nil {
				return fmt.Errorf("base64 decode privkey at index %d failed, %v", i, err)
			}

			ecb.CryptBlocks(privKey, privKey)

			privKey = bytes.TrimRight(privKey, string(Padding))
			wifKey, err := crypto.DecodeWIFKey(string(privKey), 1)
			if err != nil {
				return fmt.Errorf("decode WIF key at index %d failed, %v", i, err)
			}

			kp, err := keypair.Build(class, wifKey.PrivateKey())
			if err != nil {
				return fmt.Errorf("build keypair at index %d failed, %v", i, err)
			}

			if !bytes.Equal(kp.Address(), h160Addr) {
				return fmt.Errorf("address at index %d mismatch", i)
			}

			if i == 0 {
				w.class = class
				w.cryptography = kp.Cryptography()
			}

			encPrivKey, err := encryptor.Encrypt(kp.PrivateKey())
			if err != nil {
				return fmt.Errorf("encrypt private key at index %d failed, %v", i, err)
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
		}

		return readComplete(reader)
	})
}

func (w *Wallet) LoadCoinlibXLM(password string) error {
	w.Reset()

	const (
		NumLen        = 4
		AddrLen       = 32
		EncPrivKeyLen = 32
	)

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadBytes(reader, magicBytesLen)
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
			return fmt.Errorf("invalid wallet magic bytes")
		}

		salt, err := util.ReadBytes(reader, encryptSaltLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt salt failed, %v", err)
		}

		iv, err := util.ReadBytes(reader, encryptIVLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt iv failed, %v", err)
		}

		mac, err := util.ReadBytes(reader, encryptMacLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt mac failed, %v", err)
		}

		encryptor := util.MakeEncryptor(password, salt, iv)
		if !bytes.Equal(encryptor.Mac(), mac) {
			return ErrIncorrectPassword
		}

		w.salt = salt
		w.iv = iv
		w.mac = mac
		w.class = "XLM"

		num, err := util.BinarySerializer.Uint32(reader, binary.LittleEndian)
		if err != nil {
			return fmt.Errorf("read wallet keypair len failed, %v", err)
		}

		for i := 0; i < int(num); i++ {
			addr, err := util.ReadBytes(reader, AddrLen)
			if err != nil {
				return fmt.Errorf("read addr at index %d failed, %v", i, err)
			}

			encPrivKey, err := util.ReadBytes(reader, EncPrivKeyLen)
			if err != nil {
				return fmt.Errorf("read privkey at index %d failed, %v", i, err)
			}

			privKey, err := encryptor.Decrypt(encPrivKey)
			if err != nil {
				return fmt.Errorf("decrypt privkey at index %d failed, %v", i, err)
			}

			kp, err := keypair.Build(w.class, privKey)
			if err != nil {
				return fmt.Errorf("build keypair at index %d failed, %v", i, err)
			}

			if !bytes.Equal(kp.Address(), addr) {
				return fmt.Errorf("address at index %d mismatch", i)
			}

			if i == 0 {
				w.cryptography = kp.Cryptography()
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
		}

		return readComplete(reader)
	})
}

func (w *Wallet) LoadIotaV1(password string, num uint) error {
	defer util.DeferLogTimeCost("[load-iota-v1]")()

	if num == 0 {
		return fmt.Errorf("number can't be 0")
	}

	w.Reset()

	const (
		EncryptedSeedLen = 81
	)

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadBytes(reader, magicBytesLen)
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
			return fmt.Errorf("invalid wallet magic bytes")
		}

		salt, err := util.ReadBytes(reader, encryptSaltLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt salt failed, %v", err)
		}

		iv, err := util.ReadBytes(reader, encryptIVLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt iv failed, %v", err)
		}

		mac, err := util.ReadBytes(reader, encryptMacLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt mac failed, %v", err)
		}

		encryptor := util.MakeEncryptor(password, salt, iv)
		if !bytes.Equal(encryptor.Mac(), mac) {
			return ErrIncorrectPassword
		}

		w.salt = salt
		w.iv = iv
		w.mac = mac
		w.class = "IOTA"

		seedBytes, err := util.ReadBytes(reader, EncryptedSeedLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt seed failed, %v", err)
		}

		seedBytes, err = encryptor.Decrypt(seedBytes)
		if err != nil {
			return fmt.Errorf("decrypt seed failed, %v", err)
		}

		kp, err := keypair.Build(w.class, seedBytes)
		if err != nil {
			return fmt.Errorf("build keypair failed, %v", err)
		}

		w.cryptography = kp.Cryptography()

		originKP := kp.(keypair.DerivableKeyPair)
		err = util.BatchDo(int(num), func(i int) (interface{}, error) {
			kp, err := originKP.Derive(i)
			if err != nil {
				return nil, err
			}

			return kp, nil
		}, func(i int, data interface{}) error {
			kp := data.(keypair.KeyPair)
			encPrivKey, err := encryptor.Encrypt(kp.PrivateKey())
			if err != nil {
				return fmt.Errorf("encrypt private key at index %d failed, %v", i, err)
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
			return nil
		})
		if err != nil {
			return err
		}

		util.TraceMemStats()

		return readComplete(reader)
	})
}

func (w *Wallet) LoadPreV1(password string) error {
	w.Reset()

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadVarBytes(reader, magicBytesLen, "magicbytes")
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
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

		encryptor := util.MakeEncryptor(password, salt, iv)

		w.salt = salt
		w.iv = iv
		w.mac = encryptor.Mac()

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

			if i == 0 {
				w.class = class
				w.cryptography = kp.Cryptography()
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
		}

		return readComplete(reader)
	})
}

func (w *Wallet) LoadV1(password string) error {
	w.Reset()

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadVarBytes(reader, magicBytesLen, "magicbytes")
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
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

		encryptor := util.MakeEncryptor(password, salt, iv)

		encMagic, err := util.ReadVarBytes(reader, maxDataLen, "encmagic")
		if err != nil {
			return fmt.Errorf("read wallet encypt magic bytes failed, %v", err)
		}

		magic, err = encryptor.Decrypt(encMagic)
		if err != nil {
			return fmt.Errorf("decrypte magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
			return ErrIncorrectPassword
		}

		w.salt = salt
		w.iv = iv
		w.mac = encryptor.Mac()

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

			if i == 0 {
				w.class = class
				w.cryptography = kp.Cryptography()
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
		}

		return readComplete(reader)
	})
}

func (w *Wallet) LoadCoinlibV1(password, class string) error {
	w.Reset()

	const (
		NumLen        = 4
		AddrLen       = 20
		EncPrivKeyLen = 32
	)

	return util.WithReadFile(w.inputFileName, func(reader *bufio.Reader) error {
		magic, err := util.ReadBytes(reader, magicBytesLen)
		if err != nil {
			return fmt.Errorf("read wallet magic bytes failed, %v", err)
		}

		if !bytes.Equal(magic, v1.MagicBytes) {
			return fmt.Errorf("invalid wallet magic bytes")
		}

		salt, err := util.ReadBytes(reader, encryptSaltLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt salt failed, %v", err)
		}

		iv, err := util.ReadBytes(reader, encryptIVLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt iv failed, %v", err)
		}

		mac, err := util.ReadBytes(reader, encryptMacLen)
		if err != nil {
			return fmt.Errorf("read wallet encrypt mac failed, %v", err)
		}

		encryptor := util.MakeEncryptor(password, salt, iv)
		if !bytes.Equal(encryptor.Mac(), mac) {
			return ErrIncorrectPassword
		}

		w.salt = salt
		w.iv = iv
		w.mac = mac
		w.class = class

		num, err := util.BinarySerializer.Uint32(reader, binary.LittleEndian)
		if err != nil {
			return fmt.Errorf("read wallet keypair len failed, %v", err)
		}

		for i := 0; i < int(num); i++ {
			addr, err := util.ReadBytes(reader, AddrLen)
			if err != nil {
				return fmt.Errorf("read addr at index %d failed, %v", i, err)
			}

			encPrivKey, err := util.ReadBytes(reader, EncPrivKeyLen)
			if err != nil {
				return fmt.Errorf("read privkey at index %d failed, %v", i, err)
			}

			privKey, err := encryptor.Decrypt(encPrivKey)
			if err != nil {
				return fmt.Errorf("decrypt privkey at index %d failed, %v", i, err)
			}

			kp, err := keypair.Build(w.class, privKey)
			if err != nil {
				return fmt.Errorf("build keypair at index %d failed, %v", i, err)
			}

			if !bytes.Equal(kp.Address(), addr) && !bytes.Equal(kp.Address()[:AddrLen], addr) {
				if bytes.Equal(make([]byte, AddrLen), addr) {
					continue
				} else {
					return fmt.Errorf("address at index %d mismatch, %s vs %s",
						i, hex.EncodeToString(kp.Address()), hex.EncodeToString(addr))
				}
			}

			if i == 0 {
				w.cryptography = kp.Cryptography()
			}

			w.addKeypair(encPrivKey, kp.PublicKey())
		}

		return readComplete(reader)
	})
}
