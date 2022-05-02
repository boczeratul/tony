package api

import (
	"github.com/r04922101/portto/eth"
)

func toRepsonseBlock(b *eth.Block) {
	b.TransactionIDs = make(eth.TransactionIDs, len(b.Transactions))
	for i, t := range b.Transactions {
		b.TransactionIDs[i] = t.Hash
	}
}
