package main

import (
	"context"
	"flag"
	"log"
	"runtime"

	"github.com/r04922101/portto/indexer"
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

func main() {
	config := indexer.Config{
		SQLHost:     *sqlHost,
		SQLDB:       *sqlDB,
		SQLUser:     *sqlUser,
		SQLPassword: *sqlPassword,
		SQLPort:     *sqlPort,
		RPCEndpoint: *rpcEndpoint,
		WorkerNum:   *workerNum,
	}

	indexer, err := indexer.NewIndexer(config)
	if err != nil {
		log.Fatalf("failed to new indexer: %v", err)
	}

	if err := indexer.CheckTables(); err != nil {
		log.Fatalf("failed to check required tables exist: %v", err)
	}

	ret, err := indexer.IndexRecentBlocks(context.Background(), *blockNumber)
	if err != nil {
		log.Fatalf("failed to index recent blocks from block #%d: %v", *blockNumber, err)
	}

	log.Printf("finish indexing blocks starting from block %d until %d", *blockNumber, ret)
}
