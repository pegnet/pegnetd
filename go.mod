module github.com/pegnet/pegnetd

go 1.13

require (
	github.com/Factom-Asset-Tokens/factom v0.0.0-20190911201853-7b283996f02a
	github.com/Factom-Asset-Tokens/fatd v0.6.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/pegnet/pegnet v0.1.0-rc4.0.20190924093136-5a53cdfd85af
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/Factom-Asset-Tokens/factom => github.com/Emyrk/factom v0.0.0-20190930214432-0b837ff2681e
