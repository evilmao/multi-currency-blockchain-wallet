package gbtc

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"upex-wallet/wallet-base/util"
)

// hash type
const (
	SigHashAll = 1
)

// SignatureVersion is the version of signature.
type SignatureVersion int

const (
	SigVersionBase      SignatureVersion = 0
	SigVersionWitnessV0 SignatureVersion = 1
)

type Keypair interface {
	PublicKey() []byte
	Sign(hash []byte) ([]byte, error)
}

// Sign signs tx use hash type of SigHashAll.
func Sign(client RPC, tx *Transaction, sigVersion SignatureVersion, kp Keypair) error {
	var preScripts [][]byte
	var preAmounts []uint64
	for _, in := range tx.Inputs {
		preTx, err := client.GetRawTransaction(in.PreOutput.Hash)
		if err != nil {
			return fmt.Errorf("get raw transaction (hash=%s) failed, %v", in.PreOutput.Hash, err)
		}

		if len(preTx.Outputs) <= int(in.PreOutput.Index) {
			return fmt.Errorf("invalid tx input at index %d", in.PreOutput.Index)
		}

		preScripts = append(preScripts, preTx.Outputs[in.PreOutput.Index].ScriptData)
		preAmounts = append(preAmounts, preTx.Outputs[in.PreOutput.Index].Value)
	}

	if len(preScripts) != len(tx.Inputs) {
		return fmt.Errorf("invalid number of pre-script: %d (should be %d)", len(preScripts), len(tx.Inputs))
	}

	for i, in := range tx.Inputs {
		sigHash := SignatureHash(tx, i, preScripts[i], preAmounts[i], sigVersion)
		sig, err := kp.Sign(sigHash)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)

		util.WriteVarBytes(buffer, append(sig, SigHashAll))
		util.WriteVarBytes(buffer, kp.PublicKey())

		in.ScriptData = buffer.Bytes()
	}

	return nil
}

func SignatureHash(tx *Transaction, idx int, preScript []byte, amount uint64, sigVersion SignatureVersion) []byte {
	if sigVersion == SigVersionWitnessV0 {
		return signatureHashWitnessV0(tx, idx, preScript, amount)
	}

	return signatureHash(tx, idx, preScript)
}

func signatureHash(tx *Transaction, idx int, preScript []byte) []byte {
	copyTx := tx.Clone()
	for i, in := range copyTx.Inputs {
		if i == idx {
			in.ScriptData = preScript
		} else {
			in.ScriptData = nil
		}
	}

	buffer := new(bytes.Buffer)
	buffer.Write(copyTx.Bytes())
	util.BinarySerializer.PutUint32(buffer, binary.LittleEndian, SigHashAll)
	return SHA256D(buffer.Bytes())
}

func signatureHashWitnessV0(tx *Transaction, idx int, preScript []byte, amount uint64) []byte {
	txIn := tx.Inputs[idx]

	bf := new(bytes.Buffer)
	for _, in := range tx.Inputs {
		in.PreOutput.Serialize(bf)
	}
	hashPrevouts := SHA256D(bf.Bytes())

	bf = new(bytes.Buffer)
	for _, in := range tx.Inputs {
		util.BinarySerializer.PutUint32(bf, binary.LittleEndian, in.Sequence)
	}
	hashSequence := SHA256D(bf.Bytes())

	bf = new(bytes.Buffer)
	for _, out := range tx.Outputs {
		out.Serialize(bf)
	}
	hashOutputs := SHA256D(bf.Bytes())

	buffer := new(bytes.Buffer)

	// version
	util.BinarySerializer.PutUint32(buffer, binary.LittleEndian, tx.Version)
	buffer.Write(hashPrevouts)
	buffer.Write(hashSequence)

	txIn.PreOutput.Serialize(buffer)

	util.WriteVarBytes(buffer, preScript)
	util.BinarySerializer.PutUint64(buffer, binary.LittleEndian, amount)
	util.BinarySerializer.PutUint32(buffer, binary.LittleEndian, txIn.Sequence)

	// outputs
	buffer.Write(hashOutputs)

	// locktime
	util.BinarySerializer.PutUint32(buffer, binary.LittleEndian, tx.LockTime)

	util.BinarySerializer.PutUint32(buffer, binary.LittleEndian, SigHashAll)
	return SHA256D(buffer.Bytes())
}
