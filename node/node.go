package node

import (
	"context"

	"github.com/Factom-Asset-Tokens/factom"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/node/pegnet"
	"github.com/spf13/viper"
)

type Pegnetd struct {
	FactomClient *factom.Client
	Config       *viper.Viper

	// Tracking indicates which chains we are tracking for the sync routing
	Tracking map[string]factom.Bytes32
	Network  string

	Sync BlockSync

	Pegnet *pegnet.Pegnet
}

func NewPegnetd(ctx context.Context, conf *viper.Viper) (*Pegnetd, error) {
	// TODO : Init factom clients better
	n := new(Pegnetd)
	n.FactomClient = factom.NewClient(nil, nil)
	n.FactomClient.FactomdServer = conf.GetString(config.Server)
	n.FactomClient.WalletdServer = conf.GetString(config.Wallet)
	n.Config = conf

	// TODO: Handle all casings and handle testnet -> testnet-pM2 or w/e
	n.Network = viper.GetString(config.Network)

	// Ignore the factoid chain, as that is tracked separately
	n.Tracking = map[string]factom.Bytes32{
		// OPR Chain
		"opr": ComputeChainIDFromStrings([]string{"PegNet", n.Network, "OraclePriceRecords"}),
	}

	n.Pegnet = pegnet.New(conf)
	if err := n.Pegnet.Init(); err != nil {
		return nil, err
	}

	// TODO: Check this, harcoding it high to skip the initial stuff
	n.Sync.Synced = n.Pegnet.SelectSynced(ctx)

	// TODO :Is this the spot spot to init?
	grader.InitLX()

	return n, nil
}
