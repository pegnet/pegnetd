module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/AdamSLevy/jsonrpc2 v2.0.0+incompatible
	github.com/AdamSLevy/jsonrpc2/v11 v11.3.2
	github.com/AdamSLevy/sqlitechangeset v0.0.0-20191006235841-dce5d9b996f1 // indirect
	github.com/Factom-Asset-Tokens/factom v0.0.0-20191114224337-71de98ff5b3e
	github.com/Factom-Asset-Tokens/fatd v1.0.1-0.20191115033315-aa22fa985791
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pegnet/pegnet v0.1.0-rc4.0.20191105153926-e82140e1ce44
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20191115203952-1520e3ca4f6a

replace crawshaw.io/sqlite => github.com/AdamSLevy/sqlite v0.1.3-0.20191014215059-b98bb18889de

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4
