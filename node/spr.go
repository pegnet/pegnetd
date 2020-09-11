package node

import (
	"context"
	"fmt"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/graderStake"
	"math/rand"
	"time"
)

func ShuffleEntries(entries []factom.Entry) ([]factom.Entry, error) {
	rand.Seed(time.Now().UnixNano())
	for i := 1; i < len(entries); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			entries[r], entries[i] = entries[i], entries[r]
		}
	}
	return entries, nil
}

// Grade Staking Price Records
func (d *Pegnetd) GradeS(ctx context.Context, block *factom.EBlock) (graderStake.GradedBlock, error) {
	if block == nil {
		// TODO: Handle the case where there is no opr block.
		// 		Must delay conversions if this- happens
		return nil, nil
	}

	if *block.ChainID != SPRChain {
		return nil, fmt.Errorf("trying to grade a non-spr chain")
	}

	ver := uint8(5)
	if block.Height >= V20HeightActivation {
		ver = 5
	}
	if block.Height >= SprSignatureActivation {
		ver = 6
	}

	g, err := graderStake.NewGrader(ver, int32(block.Height))
	if err != nil {
		return nil, err
	}

	sprEntries := block.Entries
	if block.Height >= V202EnhanceActivation {
		ver = 7
		sprEntries, _ = ShuffleEntries(block.Entries)
	}

	for _, entry := range sprEntries {
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
