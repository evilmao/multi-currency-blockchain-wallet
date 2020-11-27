package bitcoin

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/syncer"
	"upex-wallet/wallet-deposit/syncer/bitcoin/gbtc"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

// Fetcher parses blockchain data.
type Fetcher struct {
	*syncer.BaseFetcher
	rpcClient *gbtc.Client
}

// NewFetcher returns fetcher instance.
func NewFetcher(api *api.ExAPI, cfg *config.Config, client *gbtc.Client) *Fetcher {
	f := &Fetcher{
		syncer.NewFetcher(api, cfg),
		client,
	}
	f.GetTxConfirmations = f.getTxConfirmations
	return f
}

// AddTx fetch and adds tx in local.
func (f *Fetcher) AddTx(data interface{}) error {
	return f.parseTx(data.([]byte))
}

// AddOrphanTx processes orphan tx.
func (f *Fetcher) AddOrphanTx(data interface{}) error {
	return f.parseTx(data.([]byte))
}

// GenSequenceID generates sequence id for each transaction.
func (f *Fetcher) GenSequenceID(data interface{}, addr string) string {
	var sb strings.Builder
	jsonparser.ArrayEach(data.([]byte),
		func(vi []byte,
			dataType jsonparser.ValueType,
			offset int, err error) {

			if err != nil {
				return
			}
			txid, _ := jsonparser.GetString(vi, "txid")
			n, _ := jsonparser.GetInt(vi, "vout")
			fmt.Fprint(&sb, txid)
			fmt.Fprint(&sb, string(n))
		}, "vin")

	// hack zec
	if f.Cfg.Currency == "ZEC" && sb.String() == "" {
		txid, _ := jsonparser.GetString(data.([]byte), "txid")
		fmt.Fprint(&sb, txid)
	}

	return utils.BytesToHex(crypto.Hash160([]byte(sb.String() + addr + f.Cfg.Currency)))[:32]
}

func (f *Fetcher) parseTx(data []byte) error {
	var (
		account = make(map[string]decimal.Decimal)
	)
	_, err := jsonparser.ArrayEach(data,
		func(vo []byte,
			dataType jsonparser.ValueType,
			offset int, err error) {

			if err != nil {
				return
			}

			var addresses []string
			jsonparser.ArrayEach(vo,
				func(addr []byte,
					dataType jsonparser.ValueType,
					offset int, err error) {
					if err != nil {
						return
					}
					address, _ := jsonparser.ParseString(addr)
					addresses = append(addresses, address)
				}, "scriptPubKey", "addresses")

			if len(addresses) == 1 {
				amount, _ := jsonparser.GetFloat(vo, "value")
				var ok bool
				ok, err = models.HasAddress(addresses[0])
				if ok && err == nil {
					account[addresses[0]] = account[addresses[0]].Add(decimal.NewFromFloat(amount))
				}
			}
		}, "vout")

	if err != nil {
		return err
	}
	txid, _ := jsonparser.GetString(data, "txid")
	confirm, _ := jsonparser.GetInt(data, "confirmations")

	if len(account) > 0 {
		for k, v := range account {
			tx := &models.Tx{
				Address:    k,
				SequenceID: f.GenSequenceID(data, k),
				Amount:     v,
				Hash:       txid,
				Type:       models.TxDeposit,
				Confirm:    uint16(confirm),
				Symbol:     f.Cfg.Currency,
			}
			err = tx.Insert()
			if err != nil {
				return err
			}
		}
	} else {
		tx := &models.Tx{
			Address:      "",
			SequenceID:   utils.BytesToHex(crypto.Hash160(data))[:32],
			Hash:         txid,
			Type:         100,
			Confirm:      uint16(confirm),
			Symbol:       f.Cfg.Currency,
			NotifyStatus: 1,
		}
		err = tx.Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Fetcher) getTxConfirmations(h string) (uint64, error) {
	var (
		txData  []byte
		err     error
		confirm int64
	)
	txData, err = f.rpcClient.GetRawTransaction(h)
	if err != nil {
		log.Errorf("Getrawtransaction error %v", err)
		return 0, err
	}
	confirm, err = jsonparser.GetInt(txData, "confirmations")
	return uint64(confirm), err
}
