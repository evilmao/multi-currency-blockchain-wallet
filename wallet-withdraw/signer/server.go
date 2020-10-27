package signer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"upex-wallet/wallet-config/withdraw/signer/config"

	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/crypto/aes"
	"upex-wallet/wallet-base/newbitx/misclib/crypto/rsa"
	"upex-wallet/wallet-base/newbitx/misclib/log"
)

const (
	MaxSigBlockSize = 256 - 11

	// SigSep is the separator of signatures for strings.Join.
	SigSep = "|"
)

var (
	baseSalt, _ = hex.DecodeString("f6f5759b77b272ea4702d1ffb6d34c68527673c53331770f0e5a85cf25b118bd")
	baseIv, _   = hex.DecodeString("3e12b3667df43b652c02827c27f876f8")
	baseKey     = "b9b6650194171059026ea56f5a94e0d4b8114833e655e99b4f55990b1a1901db"
)

// Server represents a signatrue server.
type Server struct {
	cfg        *config.Config
	ks         *KeyStore
	passPhrase []byte
}

// NewServer returns a signature server instance.
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
		ks:  NewKeyStore(cfg.DataPath, cfg.FileNames),
	}
}

func (srv Server) getPass(auth string) (string, error) {
	derivedKey := aes.GetDerivedKey(baseKey, baseSalt)
	mac := crypto.Keccak256(derivedKey[16:32])

	passBytes, err := aes.Decrypt(derivedKey[:16], srv.passPhrase, baseIv, mac)
	if err != nil {
		return "", fmt.Errorf("aes decrypt server pass failed, %v", err)
	}

	clientPass, err := rsa.B64Decrypt(auth, srv.cfg.RSAKey)
	if err != nil {
		return "", fmt.Errorf("rsa decrypt client pass failed, %v", err)
	}

	return string(passBytes) + string(clientPass), nil
}

// SetPassPhrase sets passphrase.
func (srv *Server) SetPassPhrase(auth string) error {
	derivedKey := aes.GetDerivedKey(baseKey, baseSalt)
	encryptPass, err := aes.Encrypt(derivedKey[:16], []byte(auth), baseIv)
	if err != nil {
		return err
	}

	srv.passPhrase = encryptPass
	return nil
}

func (srv Server) auth(req Request) bool {
	return true
}

func (srv Server) encryptSig(sig []byte) (string, error) {
	var (
		blockNum = (len(sig)-1)/MaxSigBlockSize + 1
		sigs     []string
	)

	for i := 0; i < blockNum; i++ {
		l := i * MaxSigBlockSize
		r := (i + 1) * MaxSigBlockSize
		if r > len(sig) {
			r = len(sig)
		}
		encSig, err := rsa.B64Encrypt(string(sig[l:r]), srv.cfg.RSAPubKey)
		if err != nil {
			return "", fmt.Errorf("encrypt signature at block %d failed, %v", i, err)
		}

		sigs = append(sigs, encSig)
	}

	return strings.Join(sigs, SigSep), nil
}

func (srv Server) sign(req *Request) ([]string, error) {
	pass, err := srv.getPass(req.AuthToken)
	if err != nil {
		return nil, err
	}

	var sigs []string
	for i, pubKey := range req.PubKeys {
		hash, err := hex.DecodeString(req.Digests[i])
		if err != nil {
			return nil, fmt.Errorf("decode digest at index %d failed, %v", i, err)
		}

		sig, err := srv.ks.Sign(pass, pubKey, hash)
		if err != nil {
			return nil, fmt.Errorf("sign at index %d failed, %s", i, err)
		}

		encSig, err := srv.encryptSig(sig)
		if err != nil {
			return nil, fmt.Errorf("encrypt signature at index %d failed, %v", i, err)
		}

		sigs = append(sigs, encSig)
	}

	return sigs, nil
}

func responseJSON(w io.Writer, status int, signature []string, msg string) {
	if status != Success && len(msg) > 0 {
		log.Error(msg)
	}

	resp := Response{
		Status:    status,
		Signature: signature,
		Msg:       msg,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(resp)
	w.Write(buf.Bytes())
}

func (srv Server) handler(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responseJSON(w, Fail, nil, fmt.Sprintf("decode request failed, %v", err))
		return
	}

	if !srv.auth(req) {
		responseJSON(w, Fail, nil, fmt.Sprintf("server auth failed"))
		return
	}

	log.Infof("request: %v, %v", req.PubKeys, req.Digests)

	if len(req.PubKeys) != len(req.Digests) {
		responseJSON(w, Fail, nil, fmt.Sprintf("pubkey and digest count mismatch, %d vs %d", len(req.PubKeys), len(req.Digests)))
		return
	}

	hexSigs, err := srv.sign(&req)
	if err != nil {
		responseJSON(w, Fail, nil, fmt.Sprintf("sign failed, %v", err))
		return
	}

	log.Infof("response: %v", hexSigs)
	responseJSON(w, Success, hexSigs, "")
}

// Start starts a signature server.
func (srv *Server) Start() error {
	err := srv.ks.Load()
	if err != nil {
		return fmt.Errorf("load keystore failed, %v", err)
	}

	http.HandleFunc("/", srv.handler)
	// http.ListenAndServeTLS(getListenAddress, "server.crt", "server.key", nil)
	return http.ListenAndServe(srv.cfg.ListenAddr, nil)
}
