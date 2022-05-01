package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
)

// Client defines an interface wrapping eth client
type Client interface {
	GetBlockByNumber(ctx context.Context, n uint64) (*Block, error)
	GetCurrentNumber(ctx context.Context) (uint64, error)
	GetBlockByHash(ctx context.Context, h string) (*Block, error)
	GetTransactionByHash(ctx context.Context, h string) (*Transaction, error)
}

type serviceImpl struct {
	delegate *ethclient.Client
	chainID  *big.Int
}

func (s *serviceImpl) toTransaction(ctx context.Context, blockNum uint64, tx *types.Transaction) (*Transaction, error) {
	// get sender address
	msg, err := tx.AsMessage(types.NewEIP155Signer(s.chainID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction message: %v", err)
	}

	// get logs
	receipt, err := s.delegate.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	toAddress := ""
	if to := msg.To(); to != nil {
		toAddress = to.Hex()
	}
	data := ""
	if d := tx.Data(); len(d) > 0 {
		data = fmt.Sprintf("0x%s", hex.EncodeToString(d))
	}
	return &Transaction{
		BlockNum: blockNum,
		Hash:     tx.Hash().Hex(),
		From:     msg.From().Hex(),
		To:       toAddress,
		Nounce:   tx.Nonce(),
		Data:     data,
		Value:    tx.Value().String(),
		Logs:     toLogs(receipt.Logs),
	}, nil
}

func (s *serviceImpl) toTransactions(ctx context.Context, blockNum uint64, transactions types.Transactions) ([]Transaction, error) {
	ret := make([]Transaction, len(transactions))

	eg, gctx := errgroup.WithContext(ctx)
	for i, t := range transactions {
		i, t := i, t
		eg.Go(func() error {
			tx, err := s.toTransaction(gctx, blockNum, t)
			if err != nil {
				return fmt.Errorf("failed to convert transaction: %v", err)
			}
			ret[i] = *tx
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

// fromEthBlock converts go-ethereum block to block
func (s *serviceImpl) toBlock(ctx context.Context, b *types.Block) (*Block, error) {
	transactions, err := s.toTransactions(ctx, b.NumberU64(), b.Transactions())
	if err != nil {
		return nil, fmt.Errorf("failed to convert transactions: %v", err)
	}
	block := &Block{
		Num:          b.NumberU64(),
		Hash:         b.Hash().Hex(),
		Time:         b.Time(),
		ParentHash:   b.ParentHash().Hex(),
		Transactions: transactions,
	}
	return block, nil
}

func (s *serviceImpl) GetBlockByNumber(ctx context.Context, n uint64) (*Block, error) {
	b, err := s.delegate.BlockByNumber(ctx, big.NewInt(int64(n)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number %d: %v", n, err)
	}

	block, err := s.toBlock(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to convert block: %v", err)
	}

	return block, nil
}

func (s *serviceImpl) GetCurrentNumber(ctx context.Context) (uint64, error) {
	n, err := s.delegate.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %v", err)
	}
	return n, nil
}

func (s *serviceImpl) GetBlockByHash(ctx context.Context, h string) (*Block, error) {
	b, err := s.delegate.BlockByHash(ctx, common.HexToHash(h))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash %s: %v", h, err)
	}

	block, err := s.toBlock(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to conver block: %v", err)
	}

	return block, nil
}

func (s *serviceImpl) GetTransactionByHash(ctx context.Context, h string) (*Transaction, error) {
	t, _, err := s.delegate.TransactionByHash(ctx, common.HexToHash(h))
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash %s: %v", h, err)
	}

	tx, err := s.toTransaction(ctx, 0, t)
	if err != nil {
		return nil, fmt.Errorf("failed to construct transaction: %v", err)
	}

	return tx, nil
}

// NewClient creates a EthClient connecting to endpoint
func NewClient(endpoint string) (Client, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to endpoint %s: %v", endpoint, err)
	}
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %v", err)
	}

	return &serviceImpl{
		delegate: client,
		chainID:  chainID,
	}, nil
}
