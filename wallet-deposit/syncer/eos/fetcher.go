package eos

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/eosio"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/deposit"
	"upex-wallet/wallet-deposit/syncer"
	"upex-wallet/wallet-deposit/syncer/eos/geos"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

type action struct {
	lastIrreversibleBlockNumber int64
	data                        []byte
}

// Fetcher parses blockchain data.
type Fetcher struct {
	*syncer.BaseFetcher
	api *eosio.API
}

// NewFetcher returns fetcher instance.
func NewFetcher(api *api.ExAPI, cfg *config.Config, client *eosio.API) *Fetcher {
	f := &Fetcher{
		syncer.NewFetcher(api, cfg),
		client,
	}
	f.GetTxConfirmations = f.getTxConfirmations
	return f
}

// AddTx fetches and adds tx in local.
func (f *Fetcher) AddTx(data interface{}) error {
	return f.parseAction(data.(*action))
}

// AddOrphanTx processes orphan tx.
func (f *Fetcher) AddOrphanTx(data interface{}) error {
	return f.parseAction(data.(*action))
}

func (f *Fetcher) GenSequenceID(data interface{}, addr string) string {
	var sb strings.Builder
	txData := data.([]byte)

	txid, _ := jsonparser.GetString(txData, "action_trace", "trx_id")
	receiver, _ := jsonparser.GetString(txData, "action_trace", "act", "data", "to")
	globalSequence := 0
	recvSequence, _ := jsonparser.GetInt(txData, "action_trace", "receipt", "recv_sequence")

	fmt.Fprint(&sb, txid)
	fmt.Fprint(&sb, receiver)
	fmt.Fprint(&sb, globalSequence)
	fmt.Fprint(&sb, recvSequence)

	return utils.BytesToHex(crypto.Hash160([]byte(sb.String() + addr + f.Cfg.Currency)))[:32]
}

func (f *Fetcher) parseAction(a *action) error {
	data := a.data
	actionAccount, _ := jsonparser.GetString(data, "action_trace", "act", "account")
	actionName, _ := jsonparser.GetString(data, "action_trace", "act", "name")
	if !geos.CheckActionName(actionName, f.Cfg.Currency) {
		return nil
	}

	traceData, _, _, err := jsonparser.Get(data, "action_trace")
	if err != nil {
		return err
	}
	amountParts, contractAddress, err := geos.GetTransferInfo(traceData, actionName)
	if err != nil {
		return err
	}

	// if neither main chain coin nor token, return
	if actionAccount == "eosio.token" {
		if actionName == "extransfer" && contractAddress != "eosio" {
			return nil
		}
		if amountParts[1] != strings.ToUpper(f.Cfg.Currency) {
			return nil
		}
	} else {
		c, ok := currency.CurrencyDetailByAddress(actionAccount)
		if !ok {
			return nil
		}

		if !c.ChainBelongTo(f.Cfg.Currency) {
			return nil
		}
		if amountParts[1] != strings.ToUpper(c.Symbol) {
			return nil
		}
	}

	txid, _ := jsonparser.GetString(data, "action_trace", "trx_id")
	from, _ := jsonparser.GetString(data, "action_trace", "act", "data", "from")
	to, _ := jsonparser.GetString(data, "action_trace", "act", "data", "to")
	receiver, _ := jsonparser.GetString(data, "action_trace", "receipt", "receiver")
	blockNumber, _ := jsonparser.GetInt(data, "block_num")
	log.Infof("parse action step 1, %s -> %s (%s), txid: %s, lastIrreversibleBlockNumber: %d, blockNumber: %d",
		from, to, receiver, txid, a.lastIrreversibleBlockNumber, blockNumber)

	confirm := int64(0)
	if a.lastIrreversibleBlockNumber > blockNumber {
		confirm = a.lastIrreversibleBlockNumber - blockNumber + 1
	} else {
		return fmt.Errorf("unmature transation txid: %s", txid)
	}

	ok, err := models.HasAddress(to)
	if err != nil {
		return err
	}

	if ok && to == receiver {
		txData, err := f.api.GetTransaction(txid)
		if err != nil {
			return fmt.Errorf("get transaction failed, error: %v", err)
		}

		status, _ := jsonparser.GetString(txData, "trx", "receipt", "status")
		if status != "executed" {
			return fmt.Errorf("status of transation is not executed,txid: %s status: %s", txid, status)
		}

		memo, _ := jsonparser.GetString(data, "action_trace", "act", "data", "memo")
		memo = deposit.TruncateTxTag(memo)

		globalSequence := 0
		recvSequence, _ := jsonparser.GetInt(data, "action_trace", "receipt", "recv_sequence")
		sequence, _ := jsonparser.GetInt(data, "account_action_seq")

		var sb strings.Builder
		fmt.Fprint(&sb, txid)
		fmt.Fprint(&sb, "_")
		fmt.Fprint(&sb, globalSequence)
		fmt.Fprint(&sb, "_")
		fmt.Fprint(&sb, recvSequence)
		txid = sb.String()

		log.Infof("parse action step 2, account: %s, name: %s, txid: %s, memo: %s, %s -> %s %s %s, sequnce: %d",
			actionAccount, actionName, txid, memo, from, to, amountParts[0], amountParts[1], sequence)

		decAmount, _ := decimal.NewFromString(amountParts[0])

		if !models.GetTxByHash(txid) {
			tx := &models.Tx{
				Address:    to,
				SequenceID: f.GenSequenceID(data, to),
				Amount:     decAmount,
				Hash:       txid,
				Type:       models.TxDeposit,
				Confirm:    uint16(confirm),
				Symbol:     amountParts[1],
				Extra:      memo,
			}

			err = tx.Insert()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *Fetcher) getTxConfirmations(h string) (uint64, error) {
	txids := strings.Split(h, "_")
	txData, err := f.api.GetTransaction(txids[0])
	if err != nil {
		log.Errorf("Getrawtransaction error %v", err)
		return 0, err
	}

	lastIrreversibleBlockNumber, _ := jsonparser.GetInt(txData, "last_irreversible_block")
	txBlockNumber, _ := jsonparser.GetInt(txData, "block_num")

	confirm := int64(0)
	if lastIrreversibleBlockNumber >= txBlockNumber {
		confirm = lastIrreversibleBlockNumber - txBlockNumber
	}

	return uint64(confirm), err
}
