package node

import (
	"context"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
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

	g, err := graderStake.NewGrader(ver, int32(block.Height))
	gv2, err2 := graderStake.NewGraderV4(ver, int32(block.Height))
	if err != nil && ver < 8 {
		return nil, err
	}
	if err2 != nil && ver >= 8 {
		return nil, err2
	}
	for _, entry := range block.Entries {
		extids := make([][]byte, len(entry.ExtIDs))
		for i := range entry.ExtIDs {
			extids[i] = entry.ExtIDs[i]
		}
		// allow only top 100 stake holders submit prices
		stakerRCD := extids[1]
		if block.Height < config.PIP18DelegateStakingActivation {
			if d.Pegnet.IsIncludedTopPEGAddress(stakerRCD) {
				// ignore bad opr errors
				err = g.AddSPR(entry.Hash[:], extids, entry.Content)
				if err != nil {
					// This is a noisy debug print
					//logrus.WithError(err).WithFields(logrus.Fields{"hash": entry.Hash.String()}).Debug("failed to add spr")
				}
			}
		} else if pegBalance, _ := d.Pegnet.GetPEGAddress(stakerRCD); pegBalance > 0 {
			err = gv2.AddSPRV4(entry.Hash[:], extids, entry.Content, pegBalance)
			if err != nil {
				// This is a noisy debug print
				//logrus.WithError(err).WithFields(logrus.Fields{"hash": entry.Hash.String()}).Debug("failed to add spr")
			}
		}
	}

	return g.Grade(), nil
}
