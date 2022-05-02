package indexer

import (
	"context"
	"fmt"
	"log"

	"github.com/r04922101/portto/db"
	"github.com/r04922101/portto/eth"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Indexer defines an interface, which can index a block into DB
type Indexer interface {
	IndexRecentBlocks(ctx context.Context, blockNum uint64) (uint64, error)
	IndexBlockByNum(ctx context.Context, blockNum uint64) error
	IndexBlock(block *eth.Block) error
	CheckTables() error
	Cron(cronExp string)
}

type impl struct {
	db        *gorm.DB
	ethClient eth.Client
	workerNum int
}

func (i *impl) IndexRecentBlocks(ctx context.Context, startingBlockNum uint64) (uint64, error) {
	targetNum, err := i.ethClient.GetCurrentNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %v", err)
	}

	// index blocks until the most recent one
	startingBlockNum = targetNum - 5
	for startingBlockNum <= targetNum {
		eg, gctx := errgroup.WithContext(context.Background())
		// dispatch workload to workers
		for w := 0; w < i.workerNum; w++ {
			n := startingBlockNum
			startingBlockNum++
			eg.Go(func() error {
				if err := i.IndexBlockByNum(gctx, n); err != nil {
					return fmt.Errorf("failed to index block %d to database: %v", n, err)
				}
				return nil
			})

			if startingBlockNum > targetNum {
				break
			}
		}
		if err := eg.Wait(); err != nil {
			return 0, err
		}

		// update remote current block number
		if startingBlockNum > targetNum {
			targetNum, err = i.ethClient.GetCurrentNumber(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to get current block number: %v", err)
			}
		}
	}
	return targetNum, nil
}

// Cron starts a cronjob to index recent blocks every time period
func (i *impl) Cron(cronExp string) {
	c := cron.New()
	fmt.Println("cron")
	c.AddFunc(cronExp, func() {
		n, err := db.GetLatestNumFromDB(i.db)
		if err != nil {
			log.Printf("[cronjob] failed to get latest num from db: %v", err)
		}
		if nn, err := i.IndexRecentBlocks(context.Background(), n+1); err != nil {
			log.Printf("[cronjob] failed to index recent blocks to db: %v", err)
		} else {
			log.Printf("[cronjob] indexed blocks %d-%d to db", n, nn)
		}
	})
	c.Start()
}

// IndexBlockByNum inserts a block with blockNum to DB
func (i *impl) IndexBlockByNum(ctx context.Context, blockNum uint64) error {
	block, err := i.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %v", blockNum, err)
	}
	return i.IndexBlock(block)
}

// IndexBlock inserts a block to DB
func (i *impl) IndexBlock(block *eth.Block) error {
	if err := i.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(block).Error; err != nil {
		return fmt.Errorf("failed to insert block to DB: %v", err)
	}

	return nil
}

func (i *impl) CheckTables() error {
	if err := i.db.AutoMigrate(&eth.Block{}); err != nil {
		return fmt.Errorf("failed to check `blocks` table exists: %v", err)
	}
	if err := i.db.AutoMigrate(&eth.Transaction{}); err != nil {
		return fmt.Errorf("failed to check `transactions` table exists: %v", err)
	}
	if err := i.db.AutoMigrate(&eth.Log{}); err != nil {
		return fmt.Errorf("failed to check `logs` table exists: %v", err)
	}
	return nil
}

// NewIndexer creates an indexer
func NewIndexer(config Config) (Indexer, error) {
	gdb, err := db.InitDB(config.SQLHost, config.SQLDB, config.SQLPort, config.SQLUser, config.SQLPassword)
	if err != nil {
		log.Fatalf("failed to connect to sql DB: %v", err)
	}

	ethClient, err := eth.NewClient(config.RPCEndpoint)
	if err != nil {
		log.Fatalf("failed to new eth client with endpoint %s: %v", config.RPCEndpoint, err)
	}

	return &impl{
		db:        gdb,
		ethClient: ethClient,
		workerNum: config.WorkerNum,
	}, nil
}
