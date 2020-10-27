package moneroutil

import (
	"errors"
	"io"
)

const (
	TxExtraTagPadding           = byte(0)
	TxExtraTagPubkey            = byte(1)
	TxExtraNonce                = byte(2)
	TxExtraMergeMiningTag       = byte(3)
	TxExtraTagAdditionalPubKeys = byte(4)
	TxExtraMinergate            = byte(0xde)
)

type TransactionExtra struct {
	PubKeys []Key
}

func ParseTransactionExtra(reader io.Reader) (*TransactionExtra, error) {
	extra := TransactionExtra{
		PubKeys: make([]Key, 1),
	}
	additionalKeys := []Key{}

	for {
		tag := make([]byte, 1)
		_, err := reader.Read(tag)
		if err == io.EOF {
			return mergeKeys(&extra, additionalKeys), nil
		}

		if err != nil {
			return mergeKeys(&extra, additionalKeys), err
		}

		switch tag[0] {
		case TxExtraTagPadding:
			return nil, errors.New("TxExtraTagPadding is not implemented yet")

		case TxExtraTagPubkey:
			key, err := ParseKey(reader)
			if err == io.EOF {
				extra.PubKeys[0] = key
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

			extra.PubKeys[0] = key

		case TxExtraNonce:
			size, err := ReadVarInt(reader)
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

			temp := make([]byte, size)
			_, err = io.ReadFull(reader, temp)

			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

		case TxExtraMergeMiningTag:
			_, err := ReadVarInt(reader)
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

			h := Hash{}
			_, err = io.ReadFull(reader, h[:])
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

		case TxExtraTagAdditionalPubKeys:
			count, err := ReadVarInt(reader)
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

			for i := uint64(0); i < count; i++ {
				key, err := ParseKey(reader)
				if err == io.EOF {
					return mergeKeys(&extra, additionalKeys), nil
				}

				if err != nil {
					return mergeKeys(&extra, additionalKeys), err
				}

				additionalKeys = append(additionalKeys, key)
			}
		case TxExtraMinergate:
			size, err := ReadVarInt(reader)
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}

			temp := make([]byte, size)
			_, err = io.ReadFull(reader, temp)
			if err == io.EOF {
				return mergeKeys(&extra, additionalKeys), nil
			}

			if err != nil {
				return mergeKeys(&extra, additionalKeys), err
			}
		default:
			return mergeKeys(&extra, additionalKeys), errors.New("unknown transaction extra tag")
		}
	}
}

func mergeKeys(extra *TransactionExtra, keys []Key) *TransactionExtra {
	if extra == nil {
		return extra
	}

	extra.PubKeys = append(extra.PubKeys, keys...)
	return extra
}