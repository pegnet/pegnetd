package node

import (
	"database/sql"
	"os"

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

	// This is the sqlite db to store state
	DB *sql.DB
}

func NewPegnetd(conf *viper.Viper) (*Pegnetd, error) {
	// TODO : Init factom clients better
	n := new(Pegnetd)
	n.FactomClient = factom.NewClient(nil, nil)
	n.Config = conf

	// TODO: Handle all casings and handle testnet -> testnet-pM2 or w/e
	n.Network = viper.GetString(config.Network)

	// Ignore the factoid chain, as that is tracked separately
	n.Tracking = map[string]factom.Bytes32{
		// OPR Chain
		"opr": ComputeChainIDFromStrings([]string{"PegNet", n.Network, "OraclePriceRecords"}),
	}

	n.Pegnet = pegnet.New(conf)

	// TODO: Check this, harcoding it high to skip the initial stuff
	n.Sync.Synced = 206421

	// TODO :Is this the spot spot to init?
	grader.InitLX()

	// Load the sqldb (or create it)
	path := os.ExpandEnv(viper.GetString(config.SqliteDBPath))
	// TODO: Idc which sqlite to use. Change this if you want.
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	n.DB = db

	return n, nil
}
