package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client defines an interface wrapping eth client
type Client interface {
	GetBlockByNumber(ctx context.Context, n uint64) (*Block, error)
	GetCurrentNumber(ctx context.Context) (uint64, error)
	GetBlockByHash(ctx context.Context, h string) (*Block, error)
	GetTransactionByHash(ctx context.Context, h string) (*Transaction, error)
}

type impl struct {
	delegate *ethclient.Client
}

func (i *impl) GetBlockByNumber(ctx context.Context, n uint64) (*Block, error) {
	b, err := i.delegate.BlockByNumber(ctx, big.NewInt(int64(n)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number %v: %v", n, err)
	}
	return FromEthBlock(b), nil
}

func (i *impl) GetCurrentNumber(ctx context.Context) (uint64, error) {
	n, err := i.delegate.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %v", err)
	}
	return n, nil
}

func (i *impl) GetBlockByHash(ctx context.Context, h string) (*Block, error) {
	b, err := i.delegate.BlockByHash(ctx, common.HexToHash(h))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %v", h, err)
	}
	return FromEthBlock(b), nil
}

func (i *impl) GetTransactionByHash(ctx context.Context, h string) (*Transaction, error) {
	chainID, err := i.delegate.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %v", err)
	}

	txHash := common.HexToHash(h)
	tx, _, err := i.delegate.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash %s: %v", h, err)
	}

	// get sender address
	msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction message: %v", err)
	}

	// get logs
	receipt, err := i.delegate.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	return &Transaction{
		Hash:   txHash.Hex(),
		From:   msg.From().Hex(),
		To:     msg.To().Hex(),
		Nounce: tx.Nonce(),
		Data:   fmt.Sprintf("0x%s", hex.EncodeToString(tx.Data())),
		Value:  tx.Value().String(),
		Logs:   FromLogs(receipt.Logs),
	}, nil
}

// NewClient creates a EthClient connecting to endpoint
func NewClient(endpoint string) (Client, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to endpoint %s: %v", endpoint, err)
	}

	return &impl{
		delegate: client,
	}, nil
}
