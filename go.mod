module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/AdamSLevy/jsonrpc2/v12 v12.0.2-0.20191015223217-9181d6ac9347 // indirect
	github.com/AdamSLevy/jsonrpc2/v13 v13.0.1
	github.com/Factom-Asset-Tokens/factom v0.0.0-20191114224337-71de98ff5b3e
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pegnet/pegnet v0.3.0
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20191206212506-739cb32ffc9d

replace crawshaw.io/sqlite => github.com/AdamSLevy/sqlite v0.1.3-0.20191014215059-b98bb18889de

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4
