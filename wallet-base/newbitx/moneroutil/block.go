package moneroutil

import (
	"bytes"
	"fmt"
	"io"
)

type ErrOutOfBounds struct {
	message string
}

func (e *ErrOutOfBounds) Error() string {
	if len(e.message) != 0 {
		return fmt.Sprintf("varint out of bounds: %s", e.message)
	}

	return "varint out of bounds"
}

type BlockHeader struct {
	MajorVersion uint8
	MinorVersion uint8
	TimeStamp    uint64
	PreviousHash Hash
	Nonce        uint32
}

type Block struct {
	BlockHeader
	MinerTx  Transaction
	TxHashes []Hash
}

var (
	correct202612hash  = Hash{0x42, 0x6d, 0x16, 0xcf, 0xf0, 0x4c, 0x71, 0xf8, 0xb1, 0x63, 0x40, 0xb7, 0x22, 0xdc, 0x40, 0x10, 0xa2, 0xdd, 0x38, 0x31, 0xc2, 0x20, 0x41, 0x43, 0x1f, 0x77, 0x25, 0x47, 0xba, 0x6e, 0x33, 0x1a}
	existing202612hash = Hash{0xbb, 0xd6, 0x04, 0xd2, 0xba, 0x11, 0xba, 0x27, 0x93, 0x5e, 0x00, 0x6e, 0xd3, 0x9c, 0x9b, 0xfd, 0xd9, 0x9b, 0x76, 0xbf, 0x4a, 0x50, 0x65, 0x4b, 0xc1, 0xe1, 0xe6, 0x12, 0x17, 0x96, 0x26, 0x98}
)

func ParseBlockHeader(buf io.Reader) (*BlockHeader, error) {
	major, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}

	if major > uint64(0xFF) {
		return nil, &ErrOutOfBounds{fmt.Sprintf("BlockHeader.MajorVersion has value %d", major)}
	}

	minor, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}

	if minor > uint64(0xFF) {
		return nil, &ErrOutOfBounds{fmt.Sprintf("BlockHeader.MinorVersion has value %d", minor)}
	}

	timestamp, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}

	h, err := ParseHash(buf)
	if err != nil {
		return nil, err
	}

	nonceBuf := make([]byte, 4)
	_, err = io.ReadFull(buf, nonceBuf)
	if err != nil {
		return nil, err
	}

	res := &BlockHeader{
		MajorVersion: uint8(major),
		MinorVersion: uint8(minor),
		TimeStamp:    timestamp,
		Nonce:        bytesToUint32(nonceBuf),
		PreviousHash: h,
	}

	return res, nil
}

func (b *BlockHeader) SerializeBlockHeader() []byte {
	res := Uint64ToBytes(uint64(b.MajorVersion))
	res = append(res, Uint64ToBytes(uint64(b.MinorVersion))...)
	res = append(res, Uint64ToBytes(b.TimeStamp)...)
	res = append(res, b.PreviousHash.Serialize()...)
	res = append(res, uint32ToBytes(b.Nonce)...)

	return res
}

func (b *Block) GetTxTreeHash() Hash {
	hashes := []Hash{b.MinerTx.GetHash()}
	hashes = append(hashes, b.TxHashes...)
	return TreeHash(hashes)
}

func (b *Block) GetHashingBlob() []byte {
	blob := b.SerializeBlockHeader()
	root := b.GetTxTreeHash()
	blob = append(blob, root[:]...)
	blob = append(blob, Uint64ToBytes(uint64(len(b.TxHashes)+1))...)
	return blob
}

func (b *Block) GetHash() Hash {
	bhb := b.GetHashingBlob()
	toHash := Uint64ToBytes(uint64(len(bhb)))
	toHash = append(toHash, bhb...)
	hash := Keccak256(toHash)

	if HashesEqual(hash, correct202612hash) {
		return existing202612hash
	}

	return hash
}

func ParseBlockBytes(buf []byte) (*Block, error) {
	reader := bytes.NewReader(buf)
	return ParseBlock(reader)
}

func ParseBlock(buf io.Reader) (*Block, error) {
	header, err := ParseBlockHeader(buf)
	if err != nil {
		return nil, err
	}

	minerTx, err := ParseTransaction(buf)
	if err != nil {
		return nil, err
	}

	numHashes, err := ReadVarInt(buf)
	if err != nil {
		return nil, err
	}

	hashes := make([]Hash, 0, numHashes)
	for i := uint64(0); i < numHashes; i++ {
		h, err := ParseHash(buf)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, h)
	}

	return &Block{
		*header,
		*minerTx,
		hashes,
	}, nil
}
