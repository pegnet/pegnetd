package node

import (
	"context"
	"database/sql"

	"github.com/Factom-Asset-Tokens/factom"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/node/pegnet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var OPRChain = *factom.NewBytes32FromString("a642a8674f46696cc47fdb6b65f9c87b2a19c5ea8123b3d2f0c13b6f33a9d5ef")

type Pegnetd struct {
	FactomClient *factom.Client
	Config       *viper.Viper

	// Tracking indicates which chains we are tracking for the sync routing
	Tracking map[string]factom.Bytes32

	Sync *pegnet.BlockSync

	Pegnet *pegnet.Pegnet
}

func NewPegnetd(ctx context.Context, conf *viper.Viper) (*Pegnetd, error) {
	// TODO : Update emyrk's factom library
	n := new(Pegnetd)
	n.FactomClient = factom.NewClient()
	n.FactomClient.FactomdServer = conf.GetString(config.Server)
	n.FactomClient.WalletdServer = conf.GetString(config.Wallet)
	n.Config = conf

	// Ignore the factoid chain, as that is tracked separately
	n.Tracking = map[string]factom.Bytes32{
		// OPR Chain
		"opr": OPRChain,
	}

	n.Pegnet = pegnet.New(conf)
	if err := n.Pegnet.Init(); err != nil {
		return nil, err
	}

	if sync, err := n.Pegnet.SelectSynced(ctx); err != nil {
		if err == sql.ErrNoRows {
			n.Sync = new(pegnet.BlockSync)
			n.Sync.Synced = 206421
			log.Debug("connected to a fresh database")
		} else {
			return nil, err
		}
	} else {
		n.Sync = sync
	}

	grader.InitLX()
	return n, nil
}
