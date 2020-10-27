package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

// HmacDRBG ...
type HmacDRBG struct {
	k             []byte
	v             []byte
	reseedCounter int
}

var (
	reseedInterval = 10000
	// Key = 0x00 00...00
	initK = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// V = 0x01 01...01
	initV = []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
)

// NewHmacDRBG ...
func NewHmacDRBG(entropy, nonce, pers []byte) *HmacDRBG {
	// seed = entropy || nonce || personalization
	seed := append(append(entropy, nonce...), pers...)

	k := initK
	v := initV

	h := &HmacDRBG{
		k:             k,
		v:             v,
		reseedCounter: 0,
	}

	h.update(seed)
	h.reseedCounter = 1

	return h
}

// Reseed ...
func (h *HmacDRBG) Reseed(entropy []byte, input []byte) error {
	seed := append(entropy, input...)
	h.update(seed)
	h.reseedCounter = 1

	return nil
}

// Generate ...
func (h *HmacDRBG) Generate(byteLength int32, input []byte) ([]byte, error) {
	if h.reseedCounter > reseedInterval {
		return nil, errors.New("Reseed is reqired")
	}

	if len(input) > 0 {
		h.update(input)
	}

	temp := make([]byte, 0, byteLength)
	for i := int32(0); i < byteLength; i += int32(32) {
		mac := hmac.New(sha256.New, h.k)
		mac.Write(h.v)
		h.v = mac.Sum(nil)
		temp = append(temp, h.v...)
	}
	result := make([]byte, byteLength)
	copy(result, temp[:byteLength])

	h.update(input)
	h.reseedCounter++

	return result, nil
}

func (h *HmacDRBG) update(input []byte) {
	// K = HMAC(K, V || 0x00 || input)
	var data = make([]byte, 0, len(h.v)+len(input)+1)
	data = append(data, h.v...)
	data = append(data, 0x00)
	data = append(data, input...)
	mac := hmac.New(sha256.New, h.k)
	mac.Write(data)
	h.k = mac.Sum(nil)

	// V = HMAC(K, V)
	mac = hmac.New(sha256.New, h.k)
	mac.Write(h.v)
	h.v = mac.Sum(nil)

	if len(input) == 0 {
		return
	}

	// K = HMAC(K, V || 0x01 || input)
	data = make([]byte, 0, len(h.v)+len(input)+1)
	data = append(data, h.v...)
	data = append(data, 0x01)
	data = append(data, input...)
	mac = hmac.New(sha256.New, h.k)
	mac.Write(data)
	h.k = mac.Sum(nil)

	// V = HMAC(K, V)
	mac = hmac.New(sha256.New, h.k)
	mac.Write(h.v)
	h.v = mac.Sum(nil)

	return
}
