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

var (
	OPRChain         = factom.NewBytes32("a642a8674f46696cc47fdb6b65f9c87b2a19c5ea8123b3d2f0c13b6f33a9d5ef")
	SPRChain         = factom.NewBytes32("d5e395125335a21cef0ceca528168e87fe929fdac1f156870c1b1be6502448b4")
	TransactionChain = factom.NewBytes32("cffce0f409ebba4ed236d49d89c70e4bd1f1367d86402a3363366683265a242d")

	// Acivation Heights

	PegnetActivation    uint32 = 206421
	GradingV2Activation uint32 = 210330

	// TransactionConversionActivation indicates when tx/conversions go live on mainnet.
	// Target Activation Height is Oct 7, 2019 15 UTC
	TransactionConversionActivation uint32 = 213237

	// This is when PEG is priced by the market cap equation
	// Estimated to be Oct 14 2019, 15:00:00 UTC
	PEGPricingActivation uint32 = 214287

	// OneWaypFCTConversions makes pFCT a 1 way conversion. This means pFCT->pXXX,
	// but no asset can go into pFCT. AKA pXXX -/> pFCT.
	// The only way to aquire pFCT is to burn FCT. The burn command will remain.
	// Estimated to be Nov 25, 2019 17:47:00 UTC
	OneWaypFCTConversions uint32 = 220346

	// Once this is activated, a maximum amount of PEG of 5,000 can be
	// converted per block. At a future height, a dynamic bank should be used.
	// Estimated to be  Dec 9, 2019, 17:00 UTC
	PegnetConversionLimitActivation uint32 = 222270

	// This is when PEG price is determined by the exchange price
	// Estimated to be  Dec 9, 2019, 17:00 UTC
	PEGFreeFloatingPriceActivation uint32 = 222270

	// V4OPRUpdate indicates the activation of additional currencies and ecdsa keys.
	// Estimated to be  Feb 12, 2020, 18:00 UTC
	V4OPRUpdate uint32 = 231620

	// V20HeightActivation indicates the activation of PegNet 2.0.
	// Estimated to be  Aug 19th 2020 14:00 UTC
	V20HeightActivation uint32 = 258796

	// Activation height for developer rewards
	V20DevRewardsHeightActivation uint32 = 295000
)

func SetAllActivations(act uint32) {
	PegnetActivation = act
	GradingV2Activation = act
	TransactionConversionActivation = act
	PEGPricingActivation = act
	OneWaypFCTConversions = act
	PegnetConversionLimitActivation = act
	PEGFreeFloatingPriceActivation = act
	fat2.Fat2RCDEActivation = act
	V4OPRUpdate = act
	V20HeightActivation = act
	V20DevRewardsHeightActivation = act
}

type Pegnetd struct {
	FactomClient *factom.Client
	Config       *viper.Viper

	Sync   *pegnet.BlockSync
	Pegnet *pegnet.Pegnet
}

func NewPegnetd(ctx context.Context, conf *viper.Viper) (*Pegnetd, error) {
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
			n.Sync.Synced = PegnetActivation
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

	// init burn address
	FAGlobalBurnAddress, err := factom.NewFAAddress(GlobalBurnAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Info("error getting burn address")
	}

	log.WithFields(log.Fields{
		"addr": FAGlobalBurnAddress,
	}).Info("burn address loaded")

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
