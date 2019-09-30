package node

import (
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/spf13/viper"
)

type Pegnetd struct {
	FactomClient *factom.Client
	Config       *viper.Viper

	Sync BlockSync
}

func NewPegnetd(config *viper.Viper) *Pegnetd {
	// TODO : Init factom clients better
	n := new(Pegnetd)
	n.FactomClient = factom.NewClient(nil, nil)
	n.Config = config

	return n
}
