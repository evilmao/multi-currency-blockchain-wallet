package types

import "upex-wallet/wallet-withdraw/base/models"

type QueryArgs struct {
	Task       models.Tx `json:"task"`
	Signatures []string  `json:"signatures"`
	PubKeys    []string  `json:"pubkeys"`
}
