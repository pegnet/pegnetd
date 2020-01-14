package node

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/grader"
)

func (d *Pegnetd) Grade(ctx context.Context, block *factom.EBlock) (grader.GradedBlock, error) {
	if block == nil {
		// TODO: Handle the case where there is no opr block.
		// 		Must delay conversions if this happens
		return nil, nil
	}

	if *block.ChainID != OPRChain {
		return nil, fmt.Errorf("trying to grade a non-opr chain")
	}

	ver := uint8(1)
	if block.Height >= GradingV2Activation {
		ver = 2
	}
	if block.Height >= PEGFreeFloatingPriceActivation {
		ver = 3
	}

	var prevWinners []string = nil
	prev, err := d.Pegnet.SelectPreviousWinners(ctx, block.Height)
	// assume that error means it's below genesis for now
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	} else {
		prevWinners = prev
	}

	g, err := grader.NewGrader(ver, int32(block.Height), prevWinners)
	if err != nil {
		return nil, err
	}

	for _, entry := range block.Entries {
		extids := make([][]byte, len(entry.ExtIDs))
		for i := range entry.ExtIDs {
			extids[i] = entry.ExtIDs[i]
		}
		// ignore bad opr errors
		err = g.AddOPR(entry.Hash[:], extids, entry.Content)
		if err != nil {
			// This is a noisy debug print
			//logrus.WithError(err).WithFields(logrus.Fields{"hash": entry.Hash.String()}).Debug("failed to add opr")
		}
	}

	return g.Grade(), nil
}
