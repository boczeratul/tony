package eth

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

// Block defines a data structure representing an eth block
type Block struct {
	Num          uint64   `json:"block_num"`  // hash in hex
	Hash         string   `json:"block_hash"` // hash in hex
	Time         uint64   `json:"block_time"`
	ParentHash   string   `json:"parent_hash"`
	Transactions []string `json:"transactions"` // hash in hex
}

// FromEthBlock converts go-ethereum block to block
func FromEthBlock(b *types.Block) *Block {
	block := &Block{
		Num:          b.NumberU64(),
		Hash:         b.Hash().Hex(),
		Time:         b.Time(),
		ParentHash:   b.ParentHash().Hex(),
		Transactions: make([]string, len(b.Transactions())),
	}
	for i, tx := range b.Transactions() {
		block.Transactions[i] = tx.Hash().Hex()
	}
	return block
}

// Transaction defines a data structure representing an eth transaction
type Transaction struct {
	Hash   string `json:"tx_hash"` // hash in hex
	From   string `json:"from"`    // from address in hex
	To     string `json:"to"`      // to address in hex
	Nounce uint64 `json:"nounce"`
	Data   string `json:"data"`
	Value  string `json:"value"`
	Logs   []*Log `json:"logs"`
}

// Log defines a data structure representing a transaction log
type Log struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

// FromLogs converts eth transaction logs to logs
func FromLogs(logs []*types.Log) []*Log {
	ret := make([]*Log, len(logs))
	for i, l := range logs {
		ret[i] = &Log{
			Index: l.Index,
			Data:  fmt.Sprintf("0x%s", hex.EncodeToString(l.Data)),
		}
	}
	return ret
}
