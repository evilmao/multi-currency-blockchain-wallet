package handler

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/signer"

	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jinzhu/gorm"
)

var (
	ErrTxHexForamt            = errors.New("invalid tx format")
	ErrConnectionFail         = errors.New("connect node failed")
	ErrDecryptSignatureFail   = errors.New("decrypt signature failed")
	ErrInvalidSignatureFormat = errors.New("invalid signature format")
	ErrDecodePubKeyFail       = errors.New("decode pubkey failed")
	ErrPubKeyCountMismatch    = errors.New("pubkey count mismatch")
	ErrSignatureCountMismatch = errors.New("signature count mismatch")
	ErrBuildTxBusy            = errors.New("build tx is busy")
	ErrBroadcastFail          = errors.New("broadcast tx failed")
)

type Tx interface{}

type Handler interface {
	Init() error

	DB() *gorm.DB
	DSN() string
	Ctrler() *Controller
	BuildTx(txHex string, signatures []string, pubKeys []string) (Tx, string, error)
	BroadcastTransaction(tx Tx, txHash string) (string, error)
	VerifyTxBroadCasted(txHash string) bool
}

type BaseHandler struct {
	Controller
	dsn string
	db  *gorm.DB
}

func (h *BaseHandler) InitDB(dsn string) error {
	if len(dsn) == 0 {
		return nil
	}

	dbInst, err := db.New(dsn, "")
	if err != nil {
		return err
	}

	models.Init(dbInst)
	h.dsn = dsn
	h.db = dbInst
	return nil
}

func (h *BaseHandler) DB() *gorm.DB {
	return h.db
}

func (h *BaseHandler) DSN() string {
	return h.dsn
}

func (h *BaseHandler) Ctrler() *Controller {
	return &h.Controller
}

var (
	handlers = make(map[string]Handler)
)

func Init() error {
	for c, h := range handlers {
		log.Infof("init %s handler", c)
		if err := h.Init(); err != nil {
			return fmt.Errorf("failed to init %s handler, %v", c, err)
		}
	}
	return nil
}

func Register(currency string, h Handler) {
	currency = strings.ToUpper(currency)
	if _, ok := Find(currency); ok {
		log.Errorf("handler.Register, duplicate of %s\n", currency)
		return
	}

	handlers[currency] = h
}

func Find(currency string) (Handler, bool) {
	currency = strings.ToUpper(currency)
	h, ok := handlers[currency]
	return h, ok
}

func Foreach(f func(string, Handler) error) error {
	if f == nil {
		return nil
	}

	for c, h := range handlers {
		err := f(c, h)
		if err != nil {
			return err
		}
	}
	return nil
}

func DecryptSignatures(rsaKey string, encSignatures []string) [][]byte {
	signatures := make([][]byte, 0, len(encSignatures))
	for i := 0; i < len(encSignatures); i++ {
		encSigBlocks := strings.Split(encSignatures[i], signer.SigSep)
		if len(encSigBlocks) == 0 {
			log.Errorf("empty encrypted signature at index %d", i)
			return nil
		}

		sig := make([]byte, 0, signer.MaxSigBlockSize*len(encSigBlocks))
		for j, block := range encSigBlocks {
			b, err := rsa.B64Decrypt(block, rsaKey)
			if err != nil {
				log.Errorf("decrypt signature block %d at index %d failed, %v", j, i, err)
				return nil
			}

			frag, err := hex.DecodeString(string(b))
			if err != nil {
				frag = b
			}

			sig = append(sig, frag...)
		}
		signatures = append(signatures, sig)
	}
	return signatures
}
