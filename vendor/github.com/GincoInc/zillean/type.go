package zillean

import "math/big"

// Balance describes the balance for an account.
type Balance struct {
	Balance string `json:"balance"`
	Nonce   int64  `json:"nonce"`
}

// DsBlock describes a DS-Block.
type DsBlock struct {
	Header struct {
		BlockNum     string   `json:"blockNum"`
		Difficulty   int64    `json:"difficulty"`
		DifficultyDS int64    `json:"difficultyDS"`
		GasPrice     string   `json:"gasPrice"`
		LeaderPubKey string   `json:"leaderPubKey"`
		PoWWinners   []string `json:"powWinners"`
		Prevhash     string   `json:"prevhash"`
		Timestamp    string   `json:"timestamp"`
	} `json:"header"`
	Signature string `json:"signature"`
}

// TxBlock describes a TX-Block.
type TxBlock struct {
	Body struct {
		HeaderSign      string `json:"HeaderSign"`
		MicroBlockInfos []struct {
			MicroBlockHash        string `json:"MicroBlockHash"`
			MicroBlockShardID     int64  `json:"MicroBlockShardId"`
			MicroBlockTxnRootHash string `json:"MicroBlockTxnRootHash"`
		} `json:"MicroBlockInfos"`
	} `json:"body"`
	Header struct {
		BlockNum       string `json:"BlockNum"`
		DsBlockNum     string `json:"DSBlockNum"`
		GasLimit       string `json:"GasLimit"`
		GasUsed        string `json:"GasUsed"`
		MbInfoHash     string `json:"MbInfoHash"`
		MinerPubKey    string `json:"MinerPubKey"`
		NumMicroBlocks int64  `json:"NumMicroBlocks"`
		NumTxns        int64  `json:"NumTxns"`
		PrevBlockHash  string `json:"PrevBlockHash"`
		Rewards        string `json:"Rewards"`
		StateDeltaHash string `json:"StateDeltaHash"`
		StateRootHash  string `json:"StateRootHash"`
		Timestamp      string `json:"Timestamp"`
		TxnHash        string `json:"TxnHash"`
		Version        int64  `json:"version"`
	} `json:"header"`
}

// Transaction describes a transaction object.
type Transaction struct {
	ID       string `json:"ID"`
	Amount   string `json:"amount"`
	GasLimit string `json:"gasLimit"`
	GasPrice string `json:"gasPrice"`
	Nonce    string `json:"nonce"`
	Receipt  struct {
		CumulativeGas string `json:"cumulative_gas"`
		EpochNum      string `json:"epoch_num"`
		Success       bool   `json:"success"`
	} `json:"receipt"`
	SenderPubKey string `json:"senderPubKey"`
	Signature    string `json:"signature"`
	ToAddr       string `json:"toAddr"`
	Version      string `json:"version"`
}

// RawTransaction describes a raw transaction object, which can be used in creating a new transaction.
type RawTransaction struct {
	Version   uint32   `json:"version"`
	Nonce     uint64   `json:"nonce"`
	To        string   `json:"toAddr"`
	Amount    string   `json:"amount"`
	PubKey    string   `json:"pubKey"`
	GasPrice  *big.Int `json:"gasPrice"`
	GasLimit  uint64   `json:"gasLimit"`
	Code      string   `json:"code"`
	Data      string   `json:"data"`
	Signature string   `json:"signature"`
}

// BlockchainInfo describes the information about Zilliqa blockchain.
type BlockchainInfo struct {
	CurrentDSEpoch    string            `json:"CurrentDSEpoch"`
	CurrentMiniEpoch  string            `json:"CurrentMiniEpoch"`
	DSBlockRate       float64           `json:"DSBlockRate"`
	NumDSBlocks       string            `json:"NumDSBlocks"`
	NumPeers          int64             `json:"NumPeers"`
	NumTransactions   string            `json:"NumTransactions"`
	NumTxBlocks       string            `json:"NumTxBlocks"`
	NumTxnsDSEpoch    string            `json:"NumTxnsDSEpoch"`
	NumTxnsTxEpoch    int64             `json:"NumTxnsTxEpoch"`
	ShardingStructure ShardingStructure `json:"ShardingStructure"`
	TransactionRate   int64             `json:"TransactionRate"`
	TxBlockRate       float64           `json:"TxBlockRate"`
}

// ShardingStructure contains the number of peers in each shard.
type ShardingStructure struct {
	NumPeers []int64 `json:"NumPeers"`
}

// RecentTransactions contains the most recent transactions (up to 100).
type RecentTransactions struct {
	TxnHashes []string `json:"TxnHashes"`
	Number    int64    `json:"number"`
}

// ListedBlocks contains the paginated list of Blocks. This can be used for both DS-Blocks and TX-Blocks.
type ListedBlocks struct {
	Data []struct {
		BlockNum int64  `json:"BlockNum"`
		Hash     string `json:"Hash"`
	} `json:"data"`
	MaxPages int64 `json:"maxPages"`
}

// SmartContract describes the smart contracts created by an address.
type SmartContract struct {
	Address string               `json:"address"`
	State   []SmartContractState `json:"state"`
}

// SmartContractState describes the state of a smart contract.
type SmartContractState struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Vname string `json:"vname"`
}
