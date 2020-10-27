package moneroutil

import (
	"bytes"
	"fmt"
	"io"
)

const (
	txInGenMarker    = 0xff
	txInToKeyMarker  = 2
	txOutToKeyMarker = 2
)

var UnimplementedError = fmt.Errorf("Unimplemented")

type TxInGen struct {
	Height uint64
}

type TxInToKey struct {
	Amount     uint64
	KeyOffsets []uint64
	KeyImage   Key
}

type TxInSerializer interface {
	TxInSerialize() []byte
	MixinLen() int
}

type TxOut struct {
	Amount uint64
	Key    Key
}

type TransactionPrefix struct {
	Version    uint32
	UnlockTime uint64
	Vin        []TxInSerializer
	Vout       []*TxOut
	Extra      []byte
}

type Transaction struct {
	TransactionPrefix
	Signatures   []RingSignature
	RctSignature *RctSig
	Expanded     bool
}

func (h *Hash) Serialize() (result []byte) {
	result = h[:]
	return
}

func (p *Key) Serialize() (result []byte) {
	result = p[:]
	return
}

func (t *TxOut) Serialize() (result []byte) {
	result = append(Uint64ToBytes(t.Amount), txOutToKeyMarker)
	result = append(result, t.Key[:]...)
	return
}

func (t *TxOut) String() (result string) {
	result = fmt.Sprintf("key: %x", t.Key)
	return
}

func (t *TxInGen) TxInSerialize() (result []byte) {
	result = append([]byte{txInGenMarker}, Uint64ToBytes(t.Height)...)
	return
}

func (t *TxInGen) MixinLen() int {
	return 0
}

func (t *TxInToKey) TxInSerialize() (result []byte) {
	result = append([]byte{txInToKeyMarker}, Uint64ToBytes(t.Amount)...)
	result = append(result, Uint64ToBytes(uint64(len(t.KeyOffsets)))...)
	for _, keyOffset := range t.KeyOffsets {
		result = append(result, Uint64ToBytes(keyOffset)...)
	}
	result = append(result, t.KeyImage[:]...)
	return
}

func (t *TxInToKey) MixinLen() int {
	return len(t.KeyOffsets)
}

func (t *TransactionPrefix) SerializePrefix() (result []byte) {
	result = append(Uint64ToBytes(uint64(t.Version)), Uint64ToBytes(t.UnlockTime)...)
	result = append(result, Uint64ToBytes(uint64(len(t.Vin)))...)
	for _, txIn := range t.Vin {
		result = append(result, txIn.TxInSerialize()...)
	}
	result = append(result, Uint64ToBytes(uint64(len(t.Vout)))...)
	for _, txOut := range t.Vout {
		result = append(result, txOut.Serialize()...)
	}
	result = append(result, Uint64ToBytes(uint64(len(t.Extra)))...)
	result = append(result, t.Extra...)
	return
}

func (t *TransactionPrefix) PrefixHash() (hash Hash) {
	hash = Keccak256(t.SerializePrefix())
	return
}

func (t *TransactionPrefix) OutputSum() (sum uint64) {
	for _, output := range t.Vout {
		sum += output.Amount
	}
	return
}

func (t *Transaction) Serialize() (result []byte) {
	result = t.SerializePrefix()
	if t.Version == 1 {
		for i := 0; i < len(t.Signatures); i++ {
			result = append(result, t.Signatures[i].Serialize()...)
		}
	} else {
		result = append(result, t.RctSignature.SerializeBase()...)
		result = append(result, t.RctSignature.SerializePrunable()...)
	}
	return
}

func (t *Transaction) SerializeBase() (result []byte) {
	if t.Version == 1 {
		result = t.Serialize()
	} else {
		result = append(t.SerializePrefix(), t.RctSignature.SerializeBase()...)
	}
	return
}

// ExpandTransaction does nothing for version 1 transactions, but for version 2
// derives all the implied elements of the ring signature
func (t *Transaction) ExpandTransaction(outputKeys [][]CtKey) {
	if t.Version == 1 {
		return
	}
	r := t.RctSignature
	if r.sigType == RCTTypeNull {
		return
	}

	// fill in the outPk property of the ring signature
	for i, ctKey := range r.outPk {
		ctKey.destination = t.Vout[i].Key
	}

	r.message = Key(t.PrefixHash())
	if r.sigType == RCTTypeFull {
		r.mixRing = make([][]CtKey, len(outputKeys[0]))
		for i := 0; i < len(outputKeys); i++ {
			r.mixRing[i] = make([]CtKey, len(outputKeys))
			for j := 0; j < len(outputKeys[0]); j++ {
				r.mixRing[j][i] = outputKeys[i][j]
			}
		}
		r.mlsagSigs = make([]MlsagSig, 1)
		r.mlsagSigs[0].ii = make([]Key, len(t.Vin))
		for i, txIn := range t.Vin {
			txInWithKey, _ := txIn.(*TxInToKey)
			r.mlsagSigs[0].ii[i] = txInWithKey.KeyImage
		}
	} else if r.sigType == RCTTypeSimple {
		r.mixRing = outputKeys
		r.mlsagSigs = make([]MlsagSig, len(t.Vin))
		for i, txIn := range t.Vin {
			txInWithKey, _ := txIn.(*TxInToKey)
			r.mlsagSigs[i].ii = make([]Key, 1)
			r.mlsagSigs[i].ii[0] = txInWithKey.KeyImage
		}
	}
	t.Expanded = true
}

func (t *Transaction) GetHash() (result Hash) {
	if t.Version == 1 {
		result = Keccak256(t.Serialize())
	} else {
		// version 2 requires first computing 3 separate hashes
		// prefix, rctBase and rctPrunable
		// and then hashing the hashes together to get the final hash
		prefixHash := t.PrefixHash()
		rctBaseHash := t.RctSignature.BaseHash()
		rctPrunableHash := t.RctSignature.PrunableHash()
		result = Keccak256(prefixHash[:], rctBaseHash[:], rctPrunableHash[:])
	}
	return
}

func ParseTxInGen(buf io.Reader) (txIn *TxInGen, err error) {
	t := new(TxInGen)
	t.Height, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	txIn = t
	return
}

func ParseTxInToKey(buf io.Reader) (txIn *TxInToKey, err error) {
	t := new(TxInToKey)
	t.Amount, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	keyOffsetLen, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	t.KeyOffsets = make([]uint64, keyOffsetLen, keyOffsetLen)
	for i := 0; i < int(keyOffsetLen); i++ {
		t.KeyOffsets[i], err = ReadVarInt(buf)
		if err != nil {
			return
		}
	}
	pubKey := make([]byte, KeyLength)
	n, err := buf.Read(pubKey)
	if err != nil {
		return
	}
	if n != KeyLength {
		err = fmt.Errorf("Buffer not long enough for public key")
		return
	}
	copy(t.KeyImage[:], pubKey)
	txIn = t
	return
}

func ParseTxIn(buf io.Reader) (txIn TxInSerializer, err error) {
	marker := make([]byte, 1)
	n, err := buf.Read(marker)
	if n != 1 {
		err = fmt.Errorf("Buffer not enough for TxIn")
		return
	}
	if err != nil {
		return
	}
	switch {
	case marker[0] == txInGenMarker:
		txIn, err = ParseTxInGen(buf)
	case marker[0] == txInToKeyMarker:
		txIn, err = ParseTxInToKey(buf)
	}
	return
}

func ParseTxOut(buf io.Reader) (txOut *TxOut, err error) {
	t := new(TxOut)
	t.Amount, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	marker := make([]byte, 1)
	n, err := buf.Read(marker)
	if err != nil {
		return
	}
	if n != 1 {
		err = fmt.Errorf("Buffer not long enough for TxOut")
		return
	}
	switch {
	case marker[0] == txOutToKeyMarker:
		t.Key, err = ParseKey(buf)
	default:
		err = fmt.Errorf("Bad Marker")
		return
	}
	if err != nil {
		return
	}
	txOut = t
	return
}

func ParseExtra(buf io.Reader) (extra []byte, err error) {
	length, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	e := make([]byte, int(length))
	n, err := buf.Read(e)
	if err != nil {
		return
	}
	if n != int(length) {
		err = fmt.Errorf("Not enough bytes for extra")
		return
	}
	extra = e
	return
}

func ParseTransactionPrefixBytes(buf []byte) (*TransactionPrefix, error) {
	reader := bytes.NewReader(buf)
	return ParseTransactionPrefix(reader)
}

func ParseTransactionPrefix(buf io.Reader) (*TransactionPrefix, error) {
	t := new(TransactionPrefix)
	version, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}
	t.Version = uint32(version)
	t.UnlockTime, err = ReadVarInt(buf)
	if err != nil {
		return nil, err
	}
	numInputs, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}
	var mixinLengths []int
	t.Vin = make([]TxInSerializer, int(numInputs), int(numInputs))
	for i := 0; i < int(numInputs); i++ {
		t.Vin[i], err = ParseTxIn(buf)
		if err != nil {
			return nil, err
		}
		mixinLen := t.Vin[i].MixinLen()
		if mixinLen > 0 {
			mixinLengths = append(mixinLengths, mixinLen)
		}
	}
	numOutputs, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}
	t.Vout = make([]*TxOut, int(numOutputs), int(numOutputs))
	for i := 0; i < int(numOutputs); i++ {
		t.Vout[i], err = ParseTxOut(buf)
		if err != nil {
			return nil, err
		}
	}
	t.Extra, err = ParseExtra(buf)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func ParseTransactionBytes(buf []byte) (*Transaction, error) {
	reader := bytes.NewReader(buf)
	return ParseTransaction(reader)
}

func ParseTransaction(buf io.Reader) (transaction *Transaction, err error) {
	t := new(Transaction)
	version, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	t.Version = uint32(version)
	t.UnlockTime, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	numInputs, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	var mixinLengths []int
	t.Vin = make([]TxInSerializer, int(numInputs), int(numInputs))
	for i := 0; i < int(numInputs); i++ {
		t.Vin[i], err = ParseTxIn(buf)
		if err != nil {
			return
		}
		mixinLen := t.Vin[i].MixinLen()
		if mixinLen > 0 {
			mixinLengths = append(mixinLengths, mixinLen)
		}
	}
	numOutputs, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	t.Vout = make([]*TxOut, int(numOutputs), int(numOutputs))
	for i := 0; i < int(numOutputs); i++ {
		t.Vout[i], err = ParseTxOut(buf)
		if err != nil {
			return
		}
	}
	t.Extra, err = ParseExtra(buf)
	if err != nil {
		return
	}
	if t.Version == 1 {
		t.Signatures, err = ParseSignatures(mixinLengths, buf)
		if err != nil {
			return
		}
	} else {
		var nMixins int
		if len(mixinLengths) > 0 {
			nMixins = mixinLengths[0] - 1
		}
		t.RctSignature, err = ParseRingCtSignature(buf, int(numInputs), int(numOutputs), nMixins)
		if err != nil {
			return
		}
	}
	transaction = t
	return
}
