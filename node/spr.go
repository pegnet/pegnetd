package node

import (
	"context"
	"fmt"
	"github.com/pegnet/pegnet/modules/spr"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/graderDelegateStake"
	"github.com/pegnet/pegnet/modules/graderStake"
	"github.com/pegnet/pegnetd/config"
)

// Grade Staking Price Records
func (d *Pegnetd) GradeS(ctx context.Context, block *factom.EBlock) (graderStake.GradedBlock, error) {
	if block == nil {
		// TODO: Handle the case where there is no opr block.
		// 		Must delay conversions if this- happens
		return nil, nil
	}

	if *block.ChainID != config.SPRChain {
		return nil, fmt.Errorf("trying to grade a non-spr chain")
	}

	ver := uint8(5)
	if block.Height >= config.V20HeightActivation {
		ver = 5
	}
	if block.Height >= config.SprSignatureActivation {
		ver = 6
	}
	if block.Height >= config.V202EnhanceActivation {
		ver = 7
	}
	if block.Height >= config.PIP18DelegateStakingActivation {
		ver = 8
	}

	if ver < 8 {
		g, err := graderStake.NewGrader(ver, int32(block.Height))
		if err != nil {
			return nil, err
		}
		for _, entry := range block.Entries {
			extids := make([][]byte, len(entry.ExtIDs))
			for i := range entry.ExtIDs {
				extids[i] = entry.ExtIDs[i]
			}
			// allow only top 100 stake holders submit prices
			stakerRCD := extids[1]
			if d.Pegnet.IsIncludedTopPEGAddress(stakerRCD) {
				// ignore bad opr errors
				err = g.AddSPR(entry.Hash[:], extids, entry.Content)
				if err != nil {
					// This is a noisy debug print
					//logrus.WithError(err).WithFields(logrus.Fields{"hash": entry.Hash.String()}).Debug("failed to add spr")
				}
			}
		}
		return g.Grade(), nil
	}
	return nil, nil
}

// Grade Staking Price Records
func (d *Pegnetd) GradeDelegatedS(ctx context.Context, block *factom.EBlock) (graderDelegateStake.DelegatedGradedBlock, error) {
	if block == nil {
		// TODO: Handle the case where there is no opr block.
		// 		Must delay conversions if this- happens
		return nil, nil
	}

	if *block.ChainID != config.SPRChain {
		return nil, fmt.Errorf("trying to grade a non-spr chain")
	}

	ver := uint8(8)
	if block.Height >= config.PIP18DelegateStakingActivation {
		ver = 8
	}

	if ver == 8 {
		g, err := graderDelegateStake.NewDelegatedGrader(ver, int32(block.Height))
		if err != nil {
			return nil, err
		}
		for _, entry := range block.Entries {
			extids := make([][]byte, len(entry.ExtIDs))
			for i := range entry.ExtIDs {
				extids[i] = entry.ExtIDs[i]
			}
			o2, errP := spr.ParseS1Content(entry.Content)
			var balanceOfPEG uint64 = 0
			if errP == nil {
				balanceOfPEG, _ = d.Pegnet.GetPEGAddress([]byte(o2.Address))
			}
			if errP == nil && len(extids) == 5 && len(extids[0]) == 1 && extids[0][0] == 8 {
				listOfDelegatorsAddress, err := g.GetDelegatorsAddress(extids[3], extids[4], o2.Address)
				if err != nil {
					continue
				}
				for i := 0; i < len(listOfDelegatorsAddress); i++ {
					individualBalance, _ := d.Pegnet.GetPEGAddress([]byte(listOfDelegatorsAddress[i]))
					balanceOfPEG += individualBalance
				}
			}
			err = g.AddSPRV4(entry.Hash[:], extids, entry.Content, balanceOfPEG)
			if err != nil {
				// This is a noisy debug print
				//logrus.WithError(err).WithFields(logrus.Fields{"hash": entry.Hash.String()}).Debug("failed to add spr")
			}
		}
		return g.Grade(), nil
	}
	return nil, nil
}
