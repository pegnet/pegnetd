package node

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node/pegnet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Pegnetd struct {
	FactomClient *factom.Client
	Config       *viper.Viper

	Sync   *pegnet.BlockSync
	Pegnet *pegnet.Pegnet

	LastAveragesData   map[fat2.PTicker][]uint64 // The last set of data used to create averages
	LastAverages       map[fat2.PTicker]uint64   // Cache for averages when requested for the same height
	LastAveragesHeight uint32                    // Height of the current cache
}

func NewPegnetd(ctx context.Context, conf *viper.Viper) (*Pegnetd, error) {
	// init chainIds
	InitChainsFromConfig(conf)

	// TODO : Update emyrk's factom library
	n := new(Pegnetd)
	n.FactomClient = FactomClientFromConfig(conf)
	n.Config = conf

	n.Pegnet = pegnet.New(conf)
	if err := n.Pegnet.Init(); err != nil {
		return nil, err
	}

	if sync, err := n.Pegnet.SelectSynced(ctx, n.Pegnet.DB); err != nil {
		if err == sql.ErrNoRows {
			n.Sync = new(pegnet.BlockSync)
			n.Sync.Synced = config.PegnetActivation
			log.Debug("connected to a fresh database")
		} else {
			return nil, err
		}
	} else {
		n.Sync = sync
	}

	err := n.Pegnet.CheckHardForks(n.Pegnet.DB)
	if err != nil {
		err = fmt.Errorf("pegnetd database hardfork check failed: %s", err.Error())
		if conf.GetBool(config.DisableHardForkCheck) {
			log.Warnf(err.Error())
		} else {
			return nil, err
		}
	}

	grader.InitLX()
	return n, nil
}

func FactomClientFromConfig(conf *viper.Viper) *factom.Client {
	cl := factom.NewClient()
	cl.FactomdServer = conf.GetString(config.Server)
	cl.WalletdServer = conf.GetString(config.Wallet)
	if config.WalletUser != "" {
		cl.Walletd.BasicAuth = true
		cl.Walletd.User = conf.GetString(config.WalletUser)
		cl.Walletd.Password = conf.GetString(config.WalletPass)
	}

	return cl
}

func InitChainsFromConfig(conf *viper.Viper) {
	network := conf.GetString(config.Network)
	if network == "MainNet" {
		config.OPRChain = factom.NewBytes32("a642a8674f46696cc47fdb6b65f9c87b2a19c5ea8123b3d2f0c13b6f33a9d5ef")
		config.SPRChain = factom.NewBytes32("d5e395125335a21cef0ceca528168e87fe929fdac1f156870c1b1be6502448b4")
		config.TransactionChain = factom.NewBytes32("cffce0f409ebba4ed236d49d89c70e4bd1f1367d86402a3363366683265a242d")
	} else if network == "TestNet" {
		config.OPRChain = factom.NewBytes32("ad98d39f002d4cae9ed07a8f5689cb029a83ad3b4bd8d23c49345d4ca7ca4393")
		config.SPRChain = factom.NewBytes32("e3b1668158026b2450d123ba993aca5367a8b96c6018f63640101a28b8ab5bc7")
		config.TransactionChain = factom.NewBytes32("2ac925fe946543a83d4c232d788dd589177611c0dbe970172c21b42039682a8a")
	}
}
