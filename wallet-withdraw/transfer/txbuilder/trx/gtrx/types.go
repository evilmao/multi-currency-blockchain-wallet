package gtrx

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sasaxie/go-client-api/core"
)

const (
	// TxDefaultExpiration is the default tx expiration time in milliseconds.
	TxDefaultExpiration = 2 * 60 * 1000
)

type Transaction struct {
	TxID    string          `json:"txID,omitempty"`
	RawData *TransactionRaw `json:"raw_data,omitempty"`
	// only support size = 1,  repeated list here for muti-sig extension
	Signature []string `json:"signature,omitempty"`
}

func (t *Transaction) UpdateTimestamp() ([]byte, error) {
	rawData := t.RawData
	if rawData == nil || len(rawData.Contract) == 0 {
		return nil, fmt.Errorf("invalid tx raw data")
	}

	rawData.Timestamp = time.Now().UnixNano() / 1000000
	rawData.Expiration = rawData.Timestamp + TxDefaultExpiration

	rawValue, err := proto.Marshal(rawData.Contract[0].Parameter.Value.ProtoValue())
	if err != nil {
		return nil, fmt.Errorf("proto marshal Any.Value failed, %v", err)
	}

	contract := &core.Transaction_Contract{
		Type: core.Transaction_Contract_ContractType(core.Transaction_Contract_ContractType_value[rawData.Contract[0].Type]),
		Parameter: &any.Any{
			TypeUrl: rawData.Contract[0].Parameter.TypeUrl,
			Value:   rawValue,
		},
	}

	refBlockBytes, _ := hex.DecodeString(rawData.RefBlockBytes)
	refBlockHash, _ := hex.DecodeString(rawData.RefBlockHash)
	txRaw := &core.TransactionRaw{
		RefBlockBytes: refBlockBytes,
		RefBlockNum:   rawData.RefBlockNum,
		RefBlockHash:  refBlockHash,
		Expiration:    rawData.Expiration,
		Timestamp:     rawData.Timestamp,
		Contract:      []*core.Transaction_Contract{contract},
	}

	if rawData.FeeLimit != 0 {
		txRaw.FeeLimit = rawData.FeeLimit
	}

	protoRaw, err := proto.Marshal(txRaw)
	if err != nil {
		return nil, fmt.Errorf("marshal transaction failed, %v", err)
	}

	hash := sha256.Sum256(protoRaw)
	t.TxID = hex.EncodeToString(hash[:])
	return hash[:], nil
}

type TransactionRaw struct {
	RefBlockBytes string                  `json:"ref_block_bytes,omitempty"`
	RefBlockNum   int64                   `json:"ref_block_num,omitempty"`
	RefBlockHash  string                  `json:"ref_block_hash,omitempty"`
	Expiration    int64                   `json:"expiration,omitempty"`
	Contract      []*Transaction_Contract `json:"contract,omitempty"`
	Timestamp     int64                   `json:"timestamp,omitempty"`
	FeeLimit      int64                   `json:"fee_limit,omitempty"`
}

type Transaction_Contract struct {
	Type      string `json:"type,omitempty"`
	Parameter *Any   `json:"parameter,omitempty"`
}

type ProtoValue interface {
	ProtoValue() proto.Message
}

type Any struct {
	TypeUrl string     `json:"type_url,omitempty"`
	Value   ProtoValue `json:"value,omitempty"`
}

const (
	TransferContractTypeURL      = "type.googleapis.com/protocol.TransferContract"
	TransferAssetContractTypeURL = "type.googleapis.com/protocol.TransferAssetContract"
	TriggerSmartContractTypeURL  = "type.googleapis.com/protocol.TriggerSmartContract"
)

func (a *Any) UnmarshalJSON(data []byte) error {
	a.TypeUrl, _ = jsonparser.GetString(data, "type_url")
	if len(a.TypeUrl) == 0 {
		return fmt.Errorf("can't find type_url")
	}

	data, _, _, _ = jsonparser.Get(data, "value")

	var err error
	switch a.TypeUrl {
	case TransferContractTypeURL:
		v := &TransferContract{}
		err = json.Unmarshal(data, v)
		a.Value = v
	case TransferAssetContractTypeURL:
		v := &TransferAssetContract{}
		err = json.Unmarshal(data, v)
		a.Value = v
	case TriggerSmartContractTypeURL:
		v := &TriggerSmartContract{}
		err = json.Unmarshal(data, v)
		a.Value = v
	default:
		err = fmt.Errorf("unsupport transfer type url: %s", a.TypeUrl)
	}

	if err != nil {
		return err
	}

	return nil
}

type TransferContract struct {
	OwnerAddress string `json:"owner_address,omitempty"`
	ToAddress    string `json:"to_address,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
}

func (c *TransferContract) ProtoValue() proto.Message {
	ownerAddress, _ := hex.DecodeString(c.OwnerAddress)
	toAddress, _ := hex.DecodeString(c.ToAddress)
	return &core.TransferContract{
		OwnerAddress: ownerAddress,
		ToAddress:    toAddress,
		Amount:       c.Amount,
	}
}

type TransferAssetContract struct {
	AssetName    string `json:"asset_name,omitempty"`
	OwnerAddress string `json:"owner_address,omitempty"`
	ToAddress    string `json:"to_address,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
}

func (c *TransferAssetContract) ProtoValue() proto.Message {
	assetName, _ := hex.DecodeString(c.AssetName)
	ownerAddress, _ := hex.DecodeString(c.OwnerAddress)
	toAddress, _ := hex.DecodeString(c.ToAddress)
	return &core.TransferAssetContract{
		AssetName:    assetName,
		OwnerAddress: ownerAddress,
		ToAddress:    toAddress,
		Amount:       c.Amount,
	}
}

type TriggerSmartContract struct {
	OwnerAddress    string `json:"owner_address,omitempty"`
	Data            string `json:"data"`
	ContractAddress string `json:"contract_address"`
}

func (c *TriggerSmartContract) ProtoValue() proto.Message {
	ownerAddress, _ := hex.DecodeString(c.OwnerAddress)
	contractAddress, _ := hex.DecodeString(c.ContractAddress)
	data, _ := hex.DecodeString(c.Data)
	return &core.TriggerSmartContract{
		OwnerAddress:    ownerAddress,
		ContractAddress: contractAddress,
		Data:            data,
	}
}

func JSONUnmarshalTx(jsonData []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(jsonData, &tx)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal tx failed, %v", err)
	}

	return &tx, nil
}
