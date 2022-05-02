package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	ginerror "github.com/r04922101/gin-error"
	"github.com/r04922101/portto/db"
	"github.com/r04922101/portto/eth"
	"github.com/r04922101/portto/indexer"
)

// NewRouter creates a router for api svc
func NewRouter(config Config) (*gin.Engine, error) {
	gdb, err := db.InitDB(config.SQLHost, config.SQLDB, config.SQLPort, config.SQLUser, config.SQLPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sql DB: %v", err)
	}

	ethClient, err := eth.NewClient(config.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to new eth client with endpoint %s: %v", config.RPCEndpoint, err)
	}

	indexer, err := indexer.NewIndexer(indexer.Config{
		SQLHost:     config.SQLHost,
		SQLDB:       config.SQLDB,
		SQLUser:     config.SQLUser,
		SQLPassword: config.SQLPassword,
		SQLPort:     config.SQLPort,
		RPCEndpoint: config.RPCEndpoint,
		WorkerNum:   10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new indexer: %v", err)
	}
	if err := indexer.CheckTables(); err != nil {
		return nil, fmt.Errorf("failed to check required tables exist: %v", err)
	}
	// index newest blocks every minute
	indexer.Cron("@every 1m")

	s := &serviceImpl{
		db:        gdb,
		ethClient: ethClient,
		indexer:   indexer,
	}

	r := gin.Default()
	r.Use(ginerror.RespondError)

	// block group
	blockGroup := r.Group("/blocks")
	{
		blockGroup.GET("/", s.getBlocks)
		blockGroup.GET("/:id", s.getBlockByHash)
	}
	// transaction group
	transactionGroup := r.Group("/transaction")
	{
		transactionGroup.GET("/:txHash", s.getTransactionByHash)
	}

	return r, nil
}
