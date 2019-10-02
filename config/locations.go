package config

// A list of config locations
const (
	LoggingLevel = "app.loglevel"
	SqliteDBPath = "app.dbpath"
	APIListen    = "app.APIListen"

	// DBlockSync Stuff
	DBlockSyncRetryPeriod = "dblocksync.retry"

	Server  = "app.Server"
	Wallet  = "app.Wallet"
	Pegnetd = "app.Pegnetd"
)
