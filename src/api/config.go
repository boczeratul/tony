package api

// Config defines the config for starting a api server
type Config struct {
	SQLHost     string
	SQLDB       string
	SQLUser     string
	SQLPassword string
	SQLPort     string
	RPCEndpoint string
}
