package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/r04922101/portto/db"
	"github.com/r04922101/portto/eth"
	"gorm.io/gorm"
)

const (
	defaultEndpoint           = "https://data-seed-prebsc-2-s3.binance.org:8545/"
	defaultBlockNumber uint64 = 18921657
)

var (
	sqlHost     = flag.String("sqlHost", "localhost", "sql host")
	sqlDB       = flag.String("sqlDB", "portto", "sql database name")
	sqlUser     = flag.String("sqlUser", "root", "sql user")
	sqlPassword = flag.String("sqlPassword", "portto", "sql user password")
	sqlPort     = flag.String("sqlPort", "3306", "sql port")
	rpcEndpoint = flag.String("rpcEndpoint", defaultEndpoint, "rpc endpoint")
	blockNumber = flag.Uint64("blockNumber", defaultBlockNumber, "starting block number")
	workerNum   = flag.Int("worker", runtime.NumCPU(), "# of worker")
)

func init() {
	flag.Parse()
}

type indexer struct {
	ethClient   eth.Client
	db          *gorm.DB
	blockNumber uint64
	workerNum   int
}

func (i *indexer) run() error {
	ctx := context.Background()
	curNum, err := i.ethClient.GetCurrentNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}

	// index blocks until the most recent one
	for i.blockNumber <= curNum {
		var wg sync.WaitGroup
		// dispatch workload to workers
		for w := 0; w < i.workerNum; w++ {
			n := i.blockNumber
			i.blockNumber++
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := i.indexBlock(context.Background(), n); err != nil {
					log.Printf("failed to index block %d to database: %v", n, err)
				}
			}()

			if i.blockNumber > curNum {
				break
			}
		}
		wg.Wait()

		// update remote current block number
		if i.blockNumber > curNum {
			curNum, err = i.ethClient.GetCurrentNumber(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current block number: %v", err)
			}
		}
	}

	return nil
}

func (i *indexer) indexBlock(ctx context.Context, blockNum uint64) error {
	block, err := i.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %v", blockNum, err)
	}
	if err := i.db.Create(block).Error; err != nil {
		return fmt.Errorf("failed to create block: %v", err)
	}

	return nil
}

func main() {
	ethClient, err := eth.NewClient(*rpcEndpoint)
	if err != nil {
		log.Fatalf("failed to new eth client with endpoint %s: %v", *rpcEndpoint, err)
	}

	db, err := db.InitDB(*sqlHost, *sqlDB, *sqlPort, *sqlUser, *sqlPassword)
	if err != nil {
		log.Fatalf("failed to connect to database on host %s: %v", *sqlHost, err)
	}

	i := &indexer{
		ethClient:   ethClient,
		db:          db,
		blockNumber: *blockNumber,
		workerNum:   *workerNum,
	}

	// check all table exist
	if err := i.db.AutoMigrate(&eth.Block{}); err != nil {
		log.Fatal(err)
	}
	if err := i.db.AutoMigrate(&eth.Transaction{}); err != nil {
		log.Fatal(err)
	}
	if err := i.db.AutoMigrate(&eth.Log{}); err != nil {
		log.Fatal(err)
	}

	if err := i.run(); err != nil {
		log.Fatalf("failed to index blocks starting from block %d: %v", *blockNumber, err)
	}
	log.Printf("finish indexing blocks starting from block %d", *blockNumber)
}
