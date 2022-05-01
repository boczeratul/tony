package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/r04922101/portto/eth"
)

const (
	defaultEndpoint          = "https://data-seed-prebsc-2-s3.binance.org:8545/"
	defaultBlockNumber int64 = 20
)

var (
	sqlHost     = flag.String("sqlHost", "localhost", "sql host")
	sqlDB       = flag.String("sqlDB", "portto", "sql database name")
	sqlUser     = flag.String("sqlUser", "root", "sql user")
	sqlPassword = flag.String("sqlPassword", "portto", "sql user password")
	sqlPort     = flag.String("sqlPort", "3306", "sql port")
	rpcEndpoint = flag.String("rpcEndpoint", defaultEndpoint, "rpc endpoint")
	n           = flag.Int64("block n", defaultBlockNumber, "block number")
)

func main() {
	client, err := eth.NewClient(*rpcEndpoint)
	ctx := context.Background()
	n, err := client.GetCurrentNumber(ctx)
	if err != nil {
		log.Fatal(err)
	}

	b, err := client.GetBlockByNumber(ctx, n)
	if err != nil {
		log.Fatal(err)
	}
	bs, _ := json.Marshal(&b)
	fmt.Printf("%+s\n", string(bs))

	parentB, err := client.GetBlockByHash(ctx, b.ParentHash)
	if err != nil {
		log.Fatal(err)
	}
	bs, _ = json.Marshal(parentB)
	fmt.Printf("%+s\n", string(bs))

	for _, tHash := range b.Transactions {
		t, err := client.GetTransactionByHash(ctx, tHash)
		if err != nil {
			log.Fatal(err)
		}
		bs, _ = json.Marshal(t)
		fmt.Printf("%+s\n", string(bs))
	}
}
