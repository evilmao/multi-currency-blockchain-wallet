package rpc

import (
	"strings"

	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/deposit/config"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

type Block struct {
	Height uint64
	Hash   string
	Txs    []*models.Tx
}

type HandleRollbackBlock func(*models.BlockInfo) (*models.BlockInfo, error)

type RPC interface {
	NextBlock(handleRollback HandleRollbackBlock) (*Block, error)
	GetTxs(hashes []string) ([]*models.Tx, error)
	GetTxConfirmations(hash string) (uint64, error)
	ReuseAddress() bool
}

// RPCCreator def.
type RPCCreator func(*config.Config) RPC

var (
	rpcCreators = make(map[string]RPCCreator)
)

func Register(currencyType string, creator RPCCreator) {
	currencyType = strings.ToUpper(currencyType)
	if _, ok := Find(currencyType); ok {
		log.Errorf("rpc.Register, duplicate of %s\n", currencyType)
		return
	}

	rpcCreators[currencyType] = creator
}

func Find(currencyType string) (RPCCreator, bool) {
	currencyType = strings.ToUpper(currencyType)
	c, ok := rpcCreators[currencyType]
	return c, ok
}

// InitCurrentBlock inits current top block from db, config, or blockchain.
func InitCurrentBlock(cfg *config.Config,
	getCurrentBlock func() (uint64, string, error),
	getBlockHashByHeight func(uint64) (string, error)) (*models.BlockInfo, error) {

	currentBlock := models.GetLastBlockInfo(cfg.Currency, cfg.UseBlockTable)
	if currentBlock == nil {
		currentBlock = &models.BlockInfo{}
	}

	log.Infof("init current block, last block, height: %d, hash: %s",
		currentBlock.Height, currentBlock.Hash)

	if cfg != nil {
		log.Infof("init current block, config start height: %d", cfg.StartHeight)
		if cfg.StartHeight > int64(currentBlock.Height) {
			currentBlock.Height = uint64(cfg.StartHeight)
			currentBlock.Hash = ""
		}
	}

	if currentBlock.Height == 0 && getCurrentBlock != nil {
		height, hash, err := getCurrentBlock()
		if err != nil {
			return nil, err
		}

		currentBlock.Height = height
		currentBlock.Hash = hash
	} else if len(currentBlock.Hash) == 0 && getBlockHashByHeight != nil {
		hash, err := getBlockHashByHeight(currentBlock.Height)
		if err != nil {
			return nil, err
		}

		currentBlock.Hash = hash
	}

	log.Infof("init current block success, height: %d, hash: %s",
		currentBlock.Height, currentBlock.Hash)
	return currentBlock, nil
}
