package gbtc

type RPC interface {
	GetBestblockhash() (string, error)
	GetBlockByHash(hash string) ([]byte, error)
	GetBlockByHeight(height uint64) ([]byte, error)
	GetRawTransaction(txhash string) (*Transaction, error)
	GetTransactionDetail(txhash string) ([]byte, error)
	CreateRawTransaction(version uint32, preOuts []*OutputPoint, outs []*Output) (*Transaction, error)
	SendRawTransaction(tx *Transaction) (string, error)
	EstimateSmartFee(confirmNum int) (float64, error)
}
