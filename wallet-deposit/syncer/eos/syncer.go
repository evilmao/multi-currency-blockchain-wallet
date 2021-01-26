package eos

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/eosio"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/syncer"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

type Syncer struct {
	cfg         *config.Config
	RPCClient   *eosio.API
	Subscribers []syncer.Subscriber
}

// New returns sync instance.
func New(cfg *config.Config, client *eosio.API) *Syncer {
	return &Syncer{
		cfg:       cfg,
		RPCClient: client,
	}
}

// AddSubscriber adds a subscriber.
func (s *Syncer) AddSubscriber(sub syncer.Subscriber) {
	s.Subscribers = append(s.Subscribers, sub)
}

func (s *Syncer) Close() {
	defer s.RPCClient.Close()
	for _, sub := range s.Subscribers {
		sub.Close()
	}
}

// FetchBlocks fetches block via rpc.
func (s *Syncer) FetchBlocks() {
	var (
		err      error
		nodeInfo []byte
	)
	for err == nil {
		nodeInfo, err = s.RPCClient.GetInfo()
		if err != nil {
			break
		}
		lastIrreversibleBlockNumber, _ := jsonparser.GetInt(nodeInfo, "last_irreversible_block_num")
		lastIrreversibleBlockID, _ := jsonparser.GetString(nodeInfo, "last_irreversible_block_id")
		log.Infof("Last IrreversibleBlock %d, %s", lastIrreversibleBlockNumber, lastIrreversibleBlockID)

		for _, addr := range models.GetAllAddresses() {
			err = s.fetchAccountActions(addr)
			if err != nil {
				log.Errorf("fetch account %s actions failed, %v", addr.Address, err)
			}
		}

		// deposit retry
		for _, s := range s.Subscribers {
			err = s.ImportBlock(models.BlockInfo{
				Height: uint64(lastIrreversibleBlockNumber),
				Hash:   lastIrreversibleBlockID,
			})
		}

		time.Sleep(10 * time.Second)
	}
	log.Errorf("FetchBlocks error %v", err)
}

func (s *Syncer) fetchAccountActions(addr *models.Address) error {
	var (
		offset  = 10
		data    []byte
		txNum   int
		actions [][]byte
		err     error
	)

	act := models.GetAccountByAddress(addr.Address, s.cfg.Currency)
	if len(act.Address) == 0 {
		act = models.DeprecatedGetAccountByAddress(addr.Address)
		if len(act.Address) == 0 || (act.Symbol != "" && act.Symbol != s.cfg.Currency) {

			act = &models.Account{
				Address: addr.Address,
				Balance: &decimal.Zero,
				Symbol:  s.cfg.Currency,
				Type:    addr.Type,
				Version: addr.Version,
			}
			_ = act.Insert()
		} else {
			_ = act.ForUpdate(models.M{
				"symbol": s.cfg.Currency,
			})
		}
	}

	log.Infof("Account %s, Account Sequence %d", act.Address, act.Nonce)

	defer syncer.DeferTraceImportBlock(0, "", 0, func() (monitor.SpanTags, error) {
		tags := monitor.SpanTags{"pos": act.Nonce, "txNum": txNum}
		return tags, err
	})()

	data, err = s.RPCClient.GetActions(act.Address, int(act.Nonce), offset)
	if err != nil {
		return fmt.Errorf("rpc get actions failed, %v", err)
	}
	log.Infof("Fetch account(%s) actions, Sequence %d", act.Address, act.Nonce)
	lastIrreversibleBlockNumber, _ := jsonparser.GetInt(data, "last_irreversible_block")
	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		actions = append(actions, value)
	}, "actions")
	if err != nil {
		return fmt.Errorf("parse actions failed, %v", err)
	}

	for _, sub := range s.Subscribers {
		for _, actionItem := range actions {
			err = sub.AddTx(&action{
				lastIrreversibleBlockNumber,
				actionItem,
			})
			if err != nil {
				_ = act.ForUpdate(models.M{
					"nonce": act.Nonce + uint64(txNum),
				})
				return fmt.Errorf("addTx error %+v", err)
			}
			txNum++
		}
	}

	_ = act.ForUpdate(models.M{
		"nonce": act.Nonce + uint64(txNum),
	})
	return err
}
