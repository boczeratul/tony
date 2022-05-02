package eth

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionIDs is a custom type for gorm
type TransactionIDs []string

// Scan implements the Scanner interface
func (ts *TransactionIDs) Scan(src interface{}) error {
	return json.Unmarshal([]byte(src.(string)), ts)
}

// Value implements the Valuer interface
func (ts *TransactionIDs) Value() (driver.Value, error) {
	val, err := json.Marshal(ts)
	return string(val), err
}

// Block defines a data structure representing an eth block
type Block struct {
	Num            uint64         `json:"block_num" gorm:"primaryKey"` // hash in hex
	Hash           string         `json:"block_hash" gorm:"index"`     // hash in hex
	Time           uint64         `json:"block_time"`
	ParentHash     string         `json:"parent_hash"`
	Transactions   []Transaction  `json:"-" gorm:"foreignKey:BlockNum;references:Num"`
	TransactionIDs TransactionIDs `json:"transactions" gorm:"-"`
}

// Transaction defines a data structure representing an eth transaction
type Transaction struct {
	BlockNum uint64 `json:"-" gorm:"index"`
	Hash     string `json:"tx_hash" gorm:"primaryKey"` // hash in hex
	From     string `json:"from"`                      // from address in hex
	To       string `json:"to"`                        // to address in hex
	Nounce   uint64 `json:"nounce"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	Logs     []Log  `json:"logs" gorm:"foreignKey:TransactionHash;references:Hash"`
}

// Log defines a data structure representing a transaction log
type Log struct {
	TransactionHash string `json:"-" gorm:"index"` // hash in hex
	Index           uint   `json:"index"`
	Data            string `json:"data"`
}

func toLogs(logs []*types.Log) []Log {
	ret := make([]Log, len(logs))
	data := ""
	for i, l := range logs {
		if d := l.Data; len(d) > 0 {
			data = fmt.Sprintf("0x%s", hex.EncodeToString(d))
		}
		ret[i] = Log{
			TransactionHash: l.TxHash.Hex(),
			Index:           l.Index,
			Data:            data,
		}
	}
	return ret
}
