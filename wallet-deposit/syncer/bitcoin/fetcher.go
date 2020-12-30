package bitcoin

import (
	"fmt"
	"strings"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/newbitx/misclib/crypto"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/newbitx/misclib/utils"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/deposit"
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
		account  = make(map[string]decimal.Decimal)
		outIndex = -1
		utxos    []*models.UTXO
		rawTx    *gbtc.Transaction
	)

	txid, err := jsonparser.GetString(data, "txid")
	if err != nil {
		return fmt.Errorf("json parser txid failed, %v", err)
	}

	err = util.JSONParserArrayEach(data, func(vo []byte, _ jsonparser.ValueType) error {
		outIndex++

		var addresses []string
		err := util.JSONParserArrayEach(vo, func(addr []byte, _ jsonparser.ValueType) error {
			address, err := jsonparser.ParseString(addr)
			if err != nil {
				return fmt.Errorf("json parser address failed, %v", err)
			}

			addresses = append(addresses, address)
			return nil
		}, "scriptPubKey", "addresses")
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			return err
		}

		if len(addresses) != 1 {
			return nil
		}

		ok, err := models.HasAddress(addresses[0])
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		amount, err := jsonparser.GetFloat(vo, "value")
		if err != nil {
			return fmt.Errorf("json parser amount failed, %v", err)
		}

		var script string
		switch {
		case strings.EqualFold(f.Cfg.Currency, "ABBC"):
			if rawTx == nil {
				deserCfg := gbtc.DeserializeConfig{
					WithTime: true,
				}

				rawTx, err = f.rpcClient.GetRawTransactionData(txid, deserCfg)
				if err != nil {
					return fmt.Errorf("get raw tx hash = %s failed, %v", txid, err)
				}
			}

			if outIndex >= len(rawTx.Outputs) {
				return fmt.Errorf("tx hash = %s outputs not match, %d vs %d",
					txid, len(rawTx.Outputs), outIndex+1)
			}

			script = rawTx.Outputs[outIndex].Script
		default:
			script, err = jsonparser.GetString(vo, "scriptPubKey", "hex")
			if err != nil {
				return fmt.Errorf("json parser output script failed, %v", err)
			}
		}

		account[addresses[0]] = account[addresses[0]].Add(decimal.NewFromFloat(amount))
		utxos = append(utxos, &models.UTXO{
			Symbol:     f.Cfg.Currency,
			TxHash:     txid,
			Amount:     decimal.NewFromFloat(amount),
			OutIndex:   uint(outIndex),
			Address:    addresses[0],
			ScriptData: script,
		})

		return nil
	}, "vout")
	if err != nil {
		return err
	}

	err = deposit.StoreUTXOs(utxos)
	if err != nil {
		return err
	}

	confirm := 1
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
