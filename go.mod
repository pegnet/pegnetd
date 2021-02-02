module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/AdamSLevy/jsonrpc2/v13 v13.0.1
	github.com/Factom-Asset-Tokens/base58 v0.0.0-20191118025050-4fa02e92ec20 // indirect
	github.com/Factom-Asset-Tokens/factom v0.0.0-20200222022020-d06cbcfe6ece
	github.com/FactomProject/factomd v6.11.0+incompatible // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/ethereum/go-ethereum v1.9.25 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pegnet/LXRHash v0.0.0-20200611040256-b33327b51c91 // indirect
	github.com/pegnet/pegnet v0.5.1-0.20210202190654-9e83e202e2e4
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.4.0
	github.com/ugorji/go v1.1.4 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20200113153851-17d98c31e1bd

replace crawshaw.io/sqlite => github.com/AdamSLevy/sqlite v0.1.3-0.20191014215059-b98bb18889de

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4
