package rpc

import (
	"fmt"

	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/deposit/config"
)

// BlockScanRPCImp implements rpc detail.
type BlockScanRPCImp interface {
	// Block ops.
	GetLastBlockHeight() (uint64, error)
	GetBlockHashByHeight(height uint64) (string, error)
	GetBlockByHeight(height uint64) (interface{}, error)

	ParsePreviousBlockHash(block interface{}) (string, error)
	ParseBlockHash(block interface{}) (string, error)
	ParseBlockTxs(height uint64, hash string, block interface{}, handleTx func(tx interface{}) error) error

	// Tx ops.
	GetTx(hash string) (interface{}, error)
	GetTxConfirmations(hash string) (uint64, error)

	ParseTx(tx interface{}) ([]*models.Tx, []*models.UTXO, error)
}

// BlockScanRPC processes deposit by scanning blocks.
type BlockScanRPC struct {
	BlockScanRPCImp
	cfg          *config.Config
	reuseAddress bool

	lastBlockHeight uint64
	currentBlock    *models.BlockInfo
	blockCache      *BlockCache
}

func NewBlockScanRPC(cfg *config.Config, imp BlockScanRPCImp, reuseAddress bool) *BlockScanRPC {
	return &BlockScanRPC{
		BlockScanRPCImp: imp,
		cfg:             cfg,
		reuseAddress:    reuseAddress,
		blockCache:      NewBlockCache(imp.GetLastBlockHeight, imp.GetBlockByHeight),
	}
}

func (r *BlockScanRPC) NextBlock(handleRollback HandleRollbackBlock) (*Block, error) {
	if r.currentBlock == nil {
		err := r.initCurrentBlock()
		if err != nil {
			return nil, err
		}
	}

	if r.currentBlock.Height >= r.lastBlockHeight {
		lastBlockHeight, err := r.GetLastBlockHeight()
		if err != nil {
			return nil, err
		}

		if lastBlockHeight <= r.lastBlockHeight {
			return nil, nil
		}

		r.lastBlockHeight = lastBlockHeight
	}

	blockData, err := r.blockCache.Get(r.currentBlock.Height + 1)
	if err != nil {
		return nil, fmt.Errorf("get block at height %d failed, %v", r.currentBlock.Height+1, err)
	}

	previousBlockHash, err := r.ParsePreviousBlockHash(blockData)
	if err != nil || len(previousBlockHash) == 0 {
		r.blockCache.Reset()
		return nil, fmt.Errorf("parse block previous hash failed, %v", err)
	}

	if previousBlockHash == r.currentBlock.Hash {
		block, err := r.parseBlock(r.currentBlock.Height+1, blockData)
		if err != nil {
			r.blockCache.Reset()
			return nil, err
		}

		r.currentBlock.Height = block.Height
		r.currentBlock.Hash = block.Hash
		return block, nil
	}

	r.currentBlock, err = handleRollback(r.currentBlock)
	if err != nil {
		r.blockCache.Reset()
		return nil, err
	}

	r.blockCache.Reset()
	return r.NextBlock(handleRollback)
}

func (r *BlockScanRPC) initCurrentBlock() (err error) {
	r.currentBlock, err = InitCurrentBlock(r.cfg, func() (uint64, string, error) {
		height, err := r.GetLastBlockHeight()
		if err != nil {
			return 0, "", err
		}

		hash, err := r.GetBlockHashByHeight(height)
		if err != nil {
			return 0, "", err
		}

		return height, hash, nil
	}, r.GetBlockHashByHeight)

	r.lastBlockHeight = r.currentBlock.Height
	return
}

func (r *BlockScanRPC) parseBlock(height uint64, block interface{}) (*Block, error) {
	hash, err := r.ParseBlockHash(block)
	if err != nil {
		return nil, fmt.Errorf("parse block hash failed, %v", err)
	}

	var (
		dbTxs   []*models.Tx
		dbUTXOs []*models.UTXO
	)

	err = r.ParseBlockTxs(height, hash, block, func(tx interface{}) error {
		txs, utxos, err := r.ParseTx(tx)

		if err != nil {
			return fmt.Errorf("parse tx failed, %v", err)
		}

		if len(txs) > 0 {
			dbTxs = append(dbTxs, txs...)
		}

		if len(utxos) > 0 {
			dbUTXOs = append(dbUTXOs, utxos...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse block txs failed, %v", err)
	}

	return &Block{
		Height: height,
		Hash:   hash,
		Txs:    dbTxs,
		UTXOs:  dbUTXOs,
	}, nil
}

func (r *BlockScanRPC) GetTxs(hashes []string) ([]*models.Tx, []*models.UTXO, error) {
	if len(hashes) == 0 {
		return nil, nil, nil
	}

	var (
		dbTxs   = make([]*models.Tx, 0, len(hashes))
		dbUTXOs = make([]*models.UTXO, 0, len(hashes))
	)
	for _, hash := range hashes {
		tx, err := r.GetTx(hash)
		if err != nil {
			return nil, nil, fmt.Errorf("get transaction %s failed, %v", hash, err)
		}

		txs, utxos, err := r.ParseTx(tx)
		if err != nil {
			return nil, nil, fmt.Errorf("parse tx %s failed, %v", hash, err)
		}

		if len(txs) > 0 {
			dbTxs = append(dbTxs, txs...)
		}
		if len(utxos) > 0 {
			dbUTXOs = append(dbUTXOs, utxos...)
		}
	}
	return dbTxs, dbUTXOs, nil
}

func (r *BlockScanRPC) ReuseAddress() bool {
	return r.reuseAddress
}
