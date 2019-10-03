module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/AdamSLevy/jsonrpc2 v2.0.0+incompatible
	github.com/AdamSLevy/jsonrpc2/v11 v11.3.2
	github.com/Factom-Asset-Tokens/factom v0.0.0-20190911201853-7b283996f02a
	github.com/Factom-Asset-Tokens/fatd v0.6.1-0.20190927200133-81408234a2b5
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pegnet/pegnet v0.1.0-rc4.0.20191002204629-5a6fd621ca60
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20191001194233-40c0cdc2f2a0
