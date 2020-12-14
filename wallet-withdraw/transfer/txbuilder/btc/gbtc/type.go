package gbtc

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
)

var (
	binarySerializer = util.BinarySerializer
	littleEndian     = binary.LittleEndian
)

const (
	HashSize = sha256.Size

	MaxScriptSize = 1024
)

type SerializeConfig struct {
	AllowWitness bool
}

type DeserializeConfig struct {
	WithAttachment bool
	WithTime       bool
}

type Transaction struct {
	Version  uint32    `json:"version,string"`
	Time     *uint32   `json:"time,string"`
	Hash     string    `json:"hash"`
	Inputs   []*Input  `json:"inputs"`
	Outputs  []*Output `json:"outputs"`
	LockTime uint32    `json:"lock_time,string"`

	BlockHash     string `json:"blockhash"`
	Confirmations uint64 `json:"confirmations"`

	SerConf   SerializeConfig
	DeserConf DeserializeConfig
}

func (tx *Transaction) MakeHash() []byte {
	hashData := SHA256D(tx.Bytes())
	tx.Hash = HashToString(hashData)
	return hashData
}

func (tx *Transaction) Bytes() []byte {
	buffer := new(bytes.Buffer)

	// version
	binarySerializer.PutUint32(buffer, littleEndian, tx.Version)

	// time
	if tx.Time != nil {
		binarySerializer.PutUint32(buffer, littleEndian, *tx.Time)
	}

	var flags byte
	if tx.SerConf.AllowWitness && tx.HasWitness() {
		flags |= 1
	}

	if flags > 0 {
		util.WriteVarInt(buffer, 0)
		buffer.WriteByte(flags)
	}

	// inputs
	util.WriteVarInt(buffer, uint64(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		in.Serialize(buffer)
	}

	// outputs
	util.WriteVarInt(buffer, uint64(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		out.Serialize(buffer)
	}

	if flags&1 > 0 {
		for _, in := range tx.Inputs {
			in.ScriptWitness.Serialize(buffer)
		}
	}

	// locktime
	binarySerializer.PutUint32(buffer, littleEndian, tx.LockTime)

	return buffer.Bytes()
}

func (tx *Transaction) SetBytes(buf []byte) error {
	var err error
	reader := bytes.NewReader(buf)

	// version
	tx.Version, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return err
	}

	// time
	if tx.DeserConf.WithTime {
		var txTime uint32
		txTime, err = binarySerializer.Uint32(reader, littleEndian)
		if err != nil {
			return err
		}

		tx.Time = &txTime
	}

	n, err := util.ReadVarInt(reader)
	if err != nil {
		return err
	}

	var flags byte
	if n == 0 {
		flags, err = reader.ReadByte()
		if err != nil {
			return err
		}

		if flags != 0 {
			// inputs
			n, err := util.ReadVarInt(reader)
			if err != nil {
				return err
			}

			for i := 0; i < int(n); i++ {
				var in Input
				err = in.Deserialize(reader)
				if err != nil {
					return fmt.Errorf("deserialize tx input at index %d failed, %v", i, err)
				}
				tx.Inputs = append(tx.Inputs, &in)
			}

			// outputs
			n, err = util.ReadVarInt(reader)
			if err != nil {
				return err
			}

			for i := 0; i < int(n); i++ {
				var out Output
				err = out.Deserialize(reader, tx.DeserConf)
				if err != nil {
					return fmt.Errorf("deserialize tx output at index %d failed, %v", i, err)
				}

				out.Index = uint32(i)
				tx.Outputs = append(tx.Outputs, &out)
			}
		}
	} else {
		// inputs
		for i := 0; i < int(n); i++ {
			var in Input
			err = in.Deserialize(reader)
			if err != nil {
				return fmt.Errorf("deserialize tx input at index %d failed, %v", i, err)
			}
			tx.Inputs = append(tx.Inputs, &in)
		}

		// outputs
		n, err = util.ReadVarInt(reader)
		if err != nil {
			return err
		}

		for i := 0; i < int(n); i++ {
			var out Output
			err = out.Deserialize(reader, tx.DeserConf)
			if err != nil {
				return fmt.Errorf("deserialize tx output at index %d failed, %v", i, err)
			}

			out.Index = uint32(i)
			tx.Outputs = append(tx.Outputs, &out)
		}
	}

	if flags&1 > 0 {
		flags ^= 1

		for _, in := range tx.Inputs {
			err := in.ScriptWitness.Deserialize(reader)
			if err != nil {
				return err
			}
		}
	}

	if flags > 0 {
		// Unknown flag in the serialization.
		return fmt.Errorf("unknown transaction optional data")
	}

	// locktime
	tx.LockTime, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return err
	}

	_, err = util.ReadBytes(reader, 1)
	if err == nil {
		return fmt.Errorf("remain data to read")
	}

	return nil
}

func (tx *Transaction) HasWitness() bool {
	for _, in := range tx.Inputs {
		if !in.ScriptWitness.IsNull() {
			return true
		}
	}
	return false
}

func (tx *Transaction) JSON() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *Transaction) Clone() *Transaction {
	tmp := &Transaction{
		Version:       tx.Version,
		Time:          tx.Time,
		Hash:          tx.Hash,
		Inputs:        make([]*Input, len(tx.Inputs)),
		Outputs:       make([]*Output, len(tx.Outputs)),
		LockTime:      tx.LockTime,
		BlockHash:     tx.BlockHash,
		Confirmations: tx.Confirmations,

		SerConf:   tx.SerConf,
		DeserConf: tx.DeserConf,
	}

	for i, in := range tx.Inputs {
		tmp.Inputs[i] = in.Clone()
	}

	for i, out := range tx.Outputs {
		tmp.Outputs[i] = out.Clone()
	}

	return tmp
}

type Input struct {
	PreOutput  *OutputPoint `json:"previous_output"`
	Script     string       `json:"script"`
	ScriptData []byte
	Sequence   uint32 `json:"sequence"`

	ScriptWitness ScriptWitness // Only serialized through Transaction
}

func (in *Input) Serialize(writer io.Writer) {
	in.PreOutput.Serialize(writer)
	util.WriteVarBytes(writer, in.ScriptData)
	binarySerializer.PutUint32(writer, littleEndian, in.Sequence)
}

func (in *Input) Deserialize(reader io.Reader) error {
	in.PreOutput = &OutputPoint{}
	err := in.PreOutput.Deserialize(reader)
	if err != nil {
		return err
	}

	in.ScriptData, err = util.ReadVarBytes(reader, MaxScriptSize, "ScriptData")
	if err != nil {
		return err
	}

	in.Script = hex.EncodeToString(in.ScriptData)

	in.Sequence, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return err
	}

	return nil
}

func (in *Input) Clone() *Input {
	tmp := &Input{
		PreOutput:     in.PreOutput.Clone(),
		Script:        in.Script,
		ScriptData:    make([]byte, len(in.ScriptData)),
		Sequence:      in.Sequence,
		ScriptWitness: in.ScriptWitness.Clone(),
	}
	copy(tmp.ScriptData, in.ScriptData)
	return tmp
}

type OutputPoint struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`

	Address    string
	Amount     uint64
	ScriptData []byte
}

func (out *OutputPoint) Serialize(writer io.Writer) {
	data, _ := StringToHash(out.Hash)
	writer.Write(data)
	binarySerializer.PutUint32(writer, littleEndian, out.Index)
}

func (out *OutputPoint) Deserialize(reader io.Reader) error {
	data := make([]byte, HashSize)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return err
	}

	out.Hash = HashToString(data)

	out.Index, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return err
	}

	return nil
}

func (out *OutputPoint) Clone() *OutputPoint {
	tmp := &OutputPoint{
		Hash:       out.Hash,
		Index:      out.Index,
		Address:    out.Address,
		Amount:     out.Amount,
		ScriptData: make([]byte, len(out.ScriptData)),
	}
	copy(tmp.ScriptData, out.ScriptData)
	return tmp
}

type ScriptWitness [][]byte

func (wit *ScriptWitness) Serialize(writer io.Writer) {
	util.WriteVarInt(writer, uint64(len(*wit)))
	for _, w := range *wit {
		util.WriteVarBytes(writer, w)
	}
}

func (wit *ScriptWitness) Deserialize(reader io.Reader) error {
	n, err := util.ReadVarInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(n); i++ {
		data, err := util.ReadVarBytes(reader, MaxScriptSize, "ScriptWitness")
		if err != nil {
			return err
		}

		*wit = append(*wit, data)
	}
	return nil
}

func (wit *ScriptWitness) IsNull() bool {
	return len(*wit) == 0
}

func (wit *ScriptWitness) Clone() ScriptWitness {
	tmp := make([][]byte, len(*wit))
	for i, w := range *wit {
		b := make([]byte, len(w))
		copy(b, w)
		tmp[i] = b
	}
	return tmp
}

type Output struct {
	Value      uint64 `json:"value"`
	Script     string `json:"script"`
	ScriptData []byte
	Attachment *Attachment `json:"attachment"`

	Index             uint32 `json:"index"`
	Address           string `json:"address"`
	LockedHeightRange uint64 `json:"locked_height_range"`
}

func NewOutput(address string, value uint64) *Output {
	var script []byte
	if addr, err := ParseAddress(address, AddressParamBTC); err == nil {
		script = addr.Script()
	}
	return &Output{
		Value:      value,
		Script:     hex.EncodeToString(script),
		ScriptData: script,
		Address:    address,
	}
}

func (out *Output) Serialize(writer io.Writer) {
	binarySerializer.PutUint64(writer, littleEndian, out.Value)
	util.WriteVarBytes(writer, out.ScriptData)

	if out.Attachment != nil {
		out.Attachment.Serialize(writer)
	}
}

func (out *Output) Deserialize(reader io.Reader, cfg DeserializeConfig) error {
	var err error
	out.Value, err = binarySerializer.Uint64(reader, littleEndian)
	if err != nil {
		return fmt.Errorf("read output value failed, %v", err)
	}

	out.ScriptData, err = util.ReadVarBytes(reader, MaxScriptSize, "ScriptData")
	if err != nil {
		return fmt.Errorf("read output script failed, %v", err)
	}

	out.Script = hex.EncodeToString(out.ScriptData)

	if cfg.WithAttachment {
		out.Attachment = &Attachment{}
		err = out.Attachment.Deserialize(reader)
		if err != nil {
			log.Errorf("deserialize output attachment failed, %v", err)
		}
	}

	return nil
}

func (out *Output) Clone() *Output {
	tmp := &Output{
		Value:             out.Value,
		Script:            out.Script,
		ScriptData:        make([]byte, len(out.ScriptData)),
		Index:             out.Index,
		Address:           out.Address,
		LockedHeightRange: out.LockedHeightRange,
	}
	copy(tmp.ScriptData, out.ScriptData)

	if out.Attachment != nil {
		tmp.Attachment = out.Attachment.Clone()
	}
	return tmp
}

type Attachment struct {
	Version uint32
	Type    uint32

	TypeName string `json:"type"`
}

func (att *Attachment) Serialize(writer io.Writer) {
	binarySerializer.PutUint32(writer, littleEndian, att.Version)
	binarySerializer.PutUint32(writer, littleEndian, att.Type)
}

func (att *Attachment) Deserialize(reader io.Reader) error {
	var err error
	att.Version, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return fmt.Errorf("read attachment version failed, %v", err)
	}

	att.Type, err = binarySerializer.Uint32(reader, littleEndian)
	if err != nil {
		return fmt.Errorf("read attachment type failed, %v", err)
	}

	return nil
}

func (att *Attachment) Clone() *Attachment {
	return &Attachment{
		Version: att.Version,
		Type:    att.Type,
	}
}
