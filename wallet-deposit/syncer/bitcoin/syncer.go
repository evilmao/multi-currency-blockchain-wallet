package bitcoin

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/syncer"
	"upex-wallet/wallet-deposit/syncer/bitcoin/gbtc"

	"github.com/buger/jsonparser"
)

// Syncer represents a syncer.
type Syncer struct {
	cfg          *config.Config
	RPCClient    *gbtc.Client
	CurrentBlock *models.BlockInfo
	Subscribers  []syncer.Subscriber
}

// New returns syncer instance.
func New(cfg *config.Config, client *gbtc.Client, block *models.BlockInfo) *Syncer {
	return &Syncer{
		cfg:          cfg,
		RPCClient:    client,
		CurrentBlock: block,
	}
}

// AddSubscriber adds a subscriber.
func (s *Syncer) AddSubscriber(sub syncer.Subscriber) {
	s.Subscribers = append(s.Subscribers, sub)
}

// FetchBlocks fetches block via rpc.
func (s *Syncer) FetchBlocks() {
	var (
		err               error
		bestBlockHash     string
		blockData         []byte
		previousBlockHash string
	)
	err = s.initCurrentBlock()
	if len(s.cfg.ForceTxs) > 0 {
		if err = s.tryProcessForceTxs(); err == nil {
			s.cfg.ForceTxs = nil
		} else {
			log.Errorf("force process tx failed, %v", err)
		}
	}
	for err == nil {
		bestBlockHash, err = s.RPCClient.GetBestBlockHash()
		if err != nil {
			break
		}

		if len(bestBlockHash) == 0 {
			err = fmt.Errorf("bestblockhash is empty")
			break
		}

		if time.Now().Unix()%100 == 0 {
			log.Infof("bestBlockHash %s, %+v", bestBlockHash, err)
		}

		for s.CurrentBlock.Hash != bestBlockHash && err == nil {
			if supportFullDataRPC(s.cfg.Currency) {
				blockData, err = s.RPCClient.GetFullBlockByHeight(s.CurrentBlock.Height + 1)
			} else {
				blockData, err = s.RPCClient.GetBlockByHeight(s.CurrentBlock.Height + 1)
			}

			if err != nil {
				break
			}

			previousBlockHash, err = jsonparser.GetString(blockData, "previousblockhash")
			if err != nil {
				err = fmt.Errorf("get block at height %d failed", s.CurrentBlock.Height+1)
				break
			}
			if previousBlockHash == s.CurrentBlock.Hash {
				err = s.importBlock(blockData)
			} else {
				err = s.processOrphanBlock(s.CurrentBlock.Hash)
			}
		}

		time.Sleep(2 * time.Second)
	}
	log.Errorf("FetchBlocks error %v", err)
}

// Close releases sync resource
func (s *Syncer) Close() {
	for _, sub := range s.Subscribers {
		sub.Close()
	}
}

func (s *Syncer) initCurrentBlock() error {
	if s.CurrentBlock.Height == 0 {
		bestBlockHash, err := s.RPCClient.GetBestBlockHash()
		if err != nil {
			return err
		}
		blockData, err := s.RPCClient.GetBlockByHash(bestBlockHash)
		if err != nil {
			return err
		}
		height, err := jsonparser.GetInt(blockData, "height")
		s.CurrentBlock.Height = uint64(height)
		s.CurrentBlock.Hash = bestBlockHash
	}

	if s.CurrentBlock.Hash == "" {
		s.CurrentBlock.Hash, _ = s.RPCClient.GetBlockHash(s.CurrentBlock.Height)
	}

	log.Infof("Syncer Current Block height: %d, hash: %s",
		s.CurrentBlock.Height,
		s.CurrentBlock.Hash)
	return nil
}

func (s *Syncer) processOrphanBlock(h string) error {
	var (
		data []byte
		err  error
	)
	if supportFullDataRPC(s.cfg.Currency) {
		data, err = s.RPCClient.GetFullBlockByHash(h)
	} else {
		data, err = s.RPCClient.GetBlockByHash(h)
	}
	if err != nil {
		return err
	}

	height, hash, err := s.parseBlock(data, true)
	if err != nil {
		return err
	}

	block := models.BlockInfo{
		Height: uint64(height),
		Hash:   hash,
	}

	for _, s := range s.Subscribers {
		err = s.ProcessOrphanBlock(block)
	}

	previousBlockHash, _ := jsonparser.GetString(data, "previousblockhash")
	s.CurrentBlock.Height = block.Height - 1
	s.CurrentBlock.Hash = previousBlockHash

	return err
}

func (s *Syncer) importBlock(data []byte) error {
	height, hash, err := s.parseBlock(data, false)
	if err != nil {
		return err
	}

	block := models.BlockInfo{
		Height: uint64(height),
		Hash:   hash,
	}

	for _, s := range s.Subscribers {
		err = s.ImportBlock(block)
	}

	s.CurrentBlock = &block

	return err
}

func (s *Syncer) parseBlock(data []byte, isOrphan bool) (int64, string, error) {
	var (
		height int64
		hash   string
		txs    [][]byte
		err    error
	)

	height, err = jsonparser.GetInt(data, "height")
	if err != nil {
		return height, hash, err
	}

	hash, err = jsonparser.GetString(data, "hash")
	if err != nil {
		return height, hash, err
	}

	if isOrphan && needFastRollbackBlock(s.cfg.Currency) {
		return height, hash, nil
	}

	log.Infof("Parse block, hash: %s height: %d", hash, height)

	var innerErr error
	_, err = jsonparser.ArrayEach(data, func(value []byte, _ jsonparser.ValueType, _ int, err error) {
		if innerErr != nil {
			return
		}

		if err != nil {
			innerErr = err
			return
		}

		if supportFullDataRPC(s.cfg.Currency) {
			txs = append(txs, value)
		} else {
			txid, err := jsonparser.ParseString(value)
			if err != nil {
				innerErr = fmt.Errorf("parse tx id failed, %v", err)
				return
			}

			rawTx, err := s.RPCClient.GetRawTransaction(txid)
			if err != nil {
				innerErr = fmt.Errorf("rpc get tx hash = %s failed, %v", txid, err)
				return
			}

			txs = append(txs, rawTx)
		}
	}, "tx")
	if err != nil {
		return height, hash, fmt.Errorf("parse block tx array failed, %v", err)
	}

	if innerErr != nil {
		return height, hash, innerErr
	}

	numOfTxs := len(txs)

	defer syncer.DeferTraceImportBlock(int(height), hash, numOfTxs, func() (monitor.SpanTags, error) {
		return nil, err
	})()

	for i, rawTx := range txs {
		for _, sub := range s.Subscribers {
			txid, _ := jsonparser.GetString(rawTx, "txid")
			if !models.GetTxByHash(txid) {
				if isOrphan {
					err = sub.AddOrphanTx(rawTx)
				} else {
					err = sub.AddTx(rawTx)
				}

				if err != nil {
					log.Errorf("addTx failed, %v", err)
					return height, hash, err
				}

				if i%100 == 0 {
					log.Infof("Processing progress	%d/%d", i, numOfTxs)
				}
			}
		}
	}

	log.Infof("Import new block, height: %d, hash: %s, txs: %d, orphan: %t",
		height, hash, numOfTxs, isOrphan)

	models.DeleteTxs()

	return height, hash, err
}

func (s *Syncer) tryProcessForceTxs() error {
	if len(s.cfg.ForceTxs) == 0 {
		return nil
	}

	for _, tx := range s.cfg.ForceTxs {
		log.Infof("start to process tx %s", tx)
		rawTx, err := s.RPCClient.GetRawTransaction(tx)
		if err != nil {
			return fmt.Errorf("GetRawTransaction %s failed, %v", tx, err)
		}
		for _, sub := range s.Subscribers {
			err = sub.AddTx(rawTx)
			if err != nil {
				return fmt.Errorf("process tx %s failed, %v", tx, err)
			}
		}
	}
	return nil
}
