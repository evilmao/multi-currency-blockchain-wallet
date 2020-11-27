package transfer

import (
	"fmt"
	"time"

	bmodels "upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
	"upex-wallet/wallet-withdraw/broadcast/types"
	"upex-wallet/wallet-withdraw/signer"
	"upex-wallet/wallet-withdraw/transfer/txbuilder"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

type Broadcaster struct {
	cfg        *config.Config
	signClient *signer.Client
}

func NewBroadcaster(cfg *config.Config) *Broadcaster {
	return &Broadcaster{
		cfg:        cfg,
		signClient: signer.NewClient(cfg.SignURL, cfg.SignPass, cfg.SignTimeout),
	}
}

func (b *Broadcaster) BroadcastTx(txInfo *txbuilder.TxInfo, task *models.Tx) error {
	sigResp, err := b.signClient.Request(&signer.Request{
		PubKeys: txInfo.SigPubKeys,
		Digests: txInfo.SigDigests,
	})
	if err != nil {
		return fmt.Errorf("sign tx failed, %v", err)
	}

	if len(sigResp.Signature) == 0 {
		return fmt.Errorf("sign tx failed, empty response")
	}

	task.Hex = txInfo.TxHex

	_, err = util.RestPostToBroadCast(&types.QueryArgs{
		Task:       *task,
		Signatures: sigResp.Signature,
		PubKeys:    txInfo.SigPubKeys,
	}, b.cfg.BroadcastURL)
	return err
}

// CheckBalanceEnough checks whether the balance of tx_ account is enough
func CheckBalanceEnough(txIns []*txbuilder.TxIn) error {
	for _, in := range txIns {
		if in.Account.Balance.LessThan(in.Cost) {
			return fmt.Errorf("balance of %s address %s not enought, need: %s, got: %s",
				in.Account.Symbol, in.Account.Address, in.Cost, in.Account.Balance)
		}
	}
	return nil
}

// SpendTxIns updates balance of the account, and the utxos.
func SpendTxIns(cfgCode int, sequenceID string, txIns []*txbuilder.TxIn, txNonce *uint64, discardAddress bool) error {
	for i, in := range txIns {
		dbIn := models.TxIn{
			TxSequenceID: sequenceID,
			Address:      in.Account.Address,
			Symbol:       in.Account.Symbol,
			Amount:       in.Cost,
		}

		err := dbIn.FirstOrCreate()
		if err != nil {
			log.Errorf("insert or create tx_in data fail, %v", err)
		}

		// Spend account.
		err = in.Account.ForUpdate(bmodels.M{
			"balance": in.Cost,
			"op":      "sub",
		})
		if err != nil {
			return fmt.Errorf("db update account (address: %s, symbol: %s) balance failed, %v",
				in.Account.Address, in.Account.Symbol, err)
		}

		// Update nonce.
		if i == 0 && txNonce != nil {
			err = models.SetBlockchainNonceIfGreater(in.Account.Address, cfgCode, *txNonce)
			if err != nil {
				return fmt.Errorf("db set %s blockchain nonce failed, %v", in.Account.Address, err)
			}
		}

		// Spend utxo.
		for _, u := range in.CostUTXOs {
			err = util.TryWithInterval(3, time.Second, func(int) error {
				return u.Spend(sequenceID)
			})
			if err != nil {
				log.Errorf("db spend utxo (symbolID: %d, hash: %s, index: %d) failed, %v",
					u.SymbolID, u.TxHash, u.OutIndex, err)
			}
		}

		// Discard address.
		if discardAddress {
			// update address status to 100
			addr := bmodels.Address{Address: in.Account.Address}
			err = util.TryWithInterval(3, time.Second, func(int) error {
				return addr.Discard()
			})
			if err != nil {
				log.Errorf("db discard address (address: %s, symbol: %s) failed, %v",
					in.Account.Address, in.Account.Symbol, err)
			}
		}
	}
	return nil
}
