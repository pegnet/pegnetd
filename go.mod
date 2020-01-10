module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/AdamSLevy/jsonrpc2 v2.0.0+incompatible // indirect
	github.com/AdamSLevy/jsonrpc2/v13 v13.0.1
	github.com/Factom-Asset-Tokens/factom v0.0.0-20190911201853-7b283996f02a
	github.com/Factom-Asset-Tokens/fatd v0.6.1-0.20190927200133-81408234a2b5
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pegnet/pegnet v0.3.0
	github.com/rs/cors v1.7.0
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20191126143921-b5fc57ecd146
