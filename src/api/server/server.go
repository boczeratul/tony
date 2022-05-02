package main

import (
	"flag"
	"log"

	"github.com/r04922101/portto/api"
)

const defaultEndpoint = "https://data-seed-prebsc-2-s3.binance.org:8545/"

var (
	port        = flag.String("port", ":3000", "local network address for the current service to listen on")
	sqlHost     = flag.String("sqlHost", "localhost", "sql host")
	sqlDB       = flag.String("sqlDB", "portto", "sql database name")
	sqlUser     = flag.String("sqlUser", "root", "sql user")
	sqlPassword = flag.String("sqlPassword", "portto", "sql user password")
	sqlPort     = flag.String("sqlPort", "3306", "sql port")
	rpcEndpoint = flag.String("rpcEndpoint", defaultEndpoint, "rpc endpoint")
)

func init() {
	flag.Parse()
}

func main() {
	config := api.Config{
		SQLHost:     *sqlHost,
		SQLDB:       *sqlDB,
		SQLUser:     *sqlUser,
		SQLPassword: *sqlPassword,
		SQLPort:     *sqlPort,
		RPCEndpoint: *rpcEndpoint,
	}

	r, err := api.NewRouter(config)
	if err != nil {
		log.Fatalf("failed to create api router: %v", err)
	}
	if err := r.Run(*port); err != nil {
		log.Fatalf("failed to start api server: %v", err)
	}
}
