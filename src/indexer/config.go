package indexer

// Config defines the config for starting an indexer
type Config struct {
	SQLHost     string
	SQLDB       string
	SQLUser     string
	SQLPassword string
	SQLPort     string
	RPCEndpoint string
	WorkerNum   int
}
