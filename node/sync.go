package node

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/fat/fat2"
	log "github.com/sirupsen/logrus"
)

func (d *Pegnetd) GetCurrentSync() uint32 {
	// Should be thread safe since we only have 1 routine writing to it
	return d.Sync.Synced
}

// DBlockSync iterates through dblocks and syncs the various chains
func (d *Pegnetd) DBlockSync(ctx context.Context) {
	retryPeriod := d.Config.GetDuration(config.DBlockSyncRetryPeriod)
OuterSyncLoop:
	for {
		if isDone(ctx) {
			return // If the user does ctl+c or something
		}

		// Fetch the current highest height
		heights := new(factom.Heights)
		err := heights.Get(d.FactomClient)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{}).Errorf("failed to fetch heights")
			time.Sleep(retryPeriod)
			continue // Loop will just keep retrying until factomd is reached
		}

		if d.Sync.Synced >= heights.DirectoryBlock {
			// We are currently synced, nothing to do. If we are above it, the factomd could
			// be rebooted
			time.Sleep(retryPeriod) // TODO: Should we have a separate polling period?
			continue
		}

		var totalDur time.Duration
		var iterations int

		begin := time.Now()
		for d.Sync.Synced < heights.DirectoryBlock {
			start := time.Now()
			hLog := log.WithFields(log.Fields{"height": d.Sync.Synced + 1})
			if isDone(ctx) {
				return
			}

			// start transaction for all block actions
			tx, err := d.Pegnet.DB.BeginTx(ctx, nil)
			if err != nil {
				hLog.WithError(err).Errorf("failed to start transaction")
				continue
			}
			// We are not synced, so we need to iterate through the dblocks and sync them
			// one by one. We can only sync our current synced height +1
			// TODO: This skips the genesis block. I'm sure that is fine
			if err := d.SyncBlock(ctx, tx, d.Sync.Synced+1); err != nil {
				hLog.WithError(err).Errorf("failed to sync height")
				time.Sleep(retryPeriod)
				// If we fail, we backout to the outer loop. This allows error handling on factomd state to be a bit
				// cleaner, such as a rebooted node with a different db. That node would have a new heights response.
				err = tx.Rollback()
				if err != nil {
					// TODO evaluate if we can recover from this point or not
					hLog.WithError(err).Fatal("unable to roll back transaction")
				}
				continue OuterSyncLoop
			}

			// Bump our sync, and march forward

			d.Sync.Synced++
			err = d.Pegnet.InsertSynced(tx, d.Sync)
			if err != nil {
				d.Sync.Synced--
				hLog.WithError(err).Errorf("unable to update synced metadata")
				err = tx.Rollback()
				if err != nil {
					// TODO evaluate if we can recover from this point or not
					hLog.WithError(err).Fatal("unable to roll back transaction")
				}
				continue OuterSyncLoop
			}

			err = tx.Commit()
			if err != nil {
				d.Sync.Synced--
				hLog.WithError(err).Errorf("unable to commit transaction")
				err = tx.Rollback()
				if err != nil {
					// TODO evaluate if we can recover from this point or not
					hLog.WithError(err).Fatal("unable to roll back transaction")
				}
			}

			elapsed := time.Since(start)
			hLog.WithFields(log.Fields{"took": elapsed}).Debugf("synced")

			iterations++
			totalDur += elapsed
			// Only print if we are > 50 behind and every 50
			if iterations%50 == 0 {
				toGo := heights.DirectoryBlock - d.Sync.Synced
				avg := totalDur / time.Duration(iterations)
				hLog.WithFields(log.Fields{
					"avg":        avg,
					"left":       time.Duration(toGo) * avg,
					"syncing-to": heights.DirectoryBlock,
					"elapsed":    time.Since(begin),
				}).Infof("sync stats")
			}
		}

	}

}

// If SyncBlock returns no error, than that height was synced and saved. If any part of the sync fails,
// the whole sync should be rolled back and not applied. An error should then be returned.
// The context should be respected if it is cancelled
func (d *Pegnetd) SyncBlock(ctx context.Context, tx *sql.Tx, height uint32) error {
	if isDone(ctx) { // Just an example about how to handle it being cancelled
		return context.Canceled
	}

	dblock := new(factom.DBlock)
	dblock.Height = height
	if err := dblock.Get(d.FactomClient); err != nil {
		return err
	}

	// Look for the eblocks we care about, and sync them in a transactional way.
	// We should be able to rollback any one of these eblock syncs.
	var err error
	eblocks := make(map[string]*factom.EBlock)
	for k, v := range d.Tracking {
		if eblock := dblock.EBlock(v); eblock != nil {
			// TODO: Multithread this to speed up performance. This is the slowest part
			if err = eblock.GetEntries(d.FactomClient); err != nil {
				return err
			}
			eblocks[k] = eblock
		}
	}

	// Entries are gathered at this point
	// TODO: I think it might be easier just to hardcode a function for each chain we care about
	// 		currently just the opr chain, then the tx chain

	graded, err := d.Grade(ctx, eblocks["opr"])
	if err != nil {
		return err // We can still just exit at this point with no rollback
	}

	// Sync the factoid chain in a transactional way. We should be able to rollback
	// the burn sync if we need too. We can first populate the eblocks that we care about
	// TODO: Check the order of operations on this and what block to add burns from.
	if err := d.SyncFactoidBlock(ctx, tx, dblock); err != nil {
		return err
	}

	// Apply all the effects
	if graded != nil { // If graded was nil, then there was no oprs this eblock
		d.Pegnet.InsertGradedBlock(graded)
		err = d.Pegnet.InsertGradeBlock(tx, eblocks["opr"], graded)
		if err != nil {
			return err
		}
		winners := graded.Winners()
		if len(winners) > 0 {
			err = d.Pegnet.InsertRate(tx, height, winners[0].OPR.GetOrderedAssetsUint())
			if err != nil {
				return err
			}
		}

		if err := d.PayWinners(tx, winners); err != nil {
			return err
		}

	}

	// TODO: Handle converts/txs
	return nil
}

func (d *Pegnetd) PayWinners(tx *sql.Tx, winners []*grader.GradingOPR) error {
	// Batch up the winner payouts to make less sql tx calls. If 1 FA address wins 5 for example,
	// that is 1 call vs 5
	totalRewards := make(map[factom.FAAddress]uint64)

	// Reward the winners
	for i := range winners {
		addr, err := factom.NewFAAddress(winners[i].OPR.GetAddress())
		if err != nil {
			// TODO: This is kinda an odd case. I think we should just drop the rewards
			// 		for an invalid address. We can always add back the rewards and they will have
			//		a higher balance after a change.
			log.WithError(err).WithFields(log.Fields{
				"height": winners[i].OPR.GetHeight(),
				"ehash":  fmt.Sprintf("%x", winners[i].EntryHash),
			}).Warnf("failed to reward")
			continue
		}

		totalRewards[addr] += uint64(winners[i].Payout())
	}

	for addr, reward := range totalRewards {
		if _, err := d.Pegnet.AddToBalance(tx, &addr, fat2.PTickerPEG, reward); err != nil {
			return err // The tx should be rolled back by the caller if we return an error during this.
		}
	}
	return nil
}

// SyncFactoidBlock tracks the burns for a specific dblock
func (d *Pegnetd) SyncFactoidBlock(ctx context.Context, tx *sql.Tx, dblock *factom.DBlock) error {
	fblock := new(factom.FBlock)
	fblock.Header.Height = dblock.Height
	if err := fblock.Get(d.FactomClient); err != nil {
		return err
	}

	var totalBurned uint64
	var burns []factom.FactoidTransactionIO

	// Register all burns. Burns have a few requirements
	// - Only 1 output, and that output must be the EC burn address
	// - The output amount must be 0
	// - Must only have 1 input
	for i := range fblock.Transactions {
		if isDone(ctx) {
			return context.Canceled
		}

		if err := fblock.Transactions[i].Get(d.FactomClient); err != nil {
			return err
		}

		tx := fblock.Transactions[i]
		// Check number of inputs/outputs
		if len(tx.ECOutputs) != 1 || len(tx.FCTInputs) != 1 || len(tx.FCTOutputs) > 0 {
			continue // Wrong number of ins/outs for a burn
		}

		// Check correct output
		out := tx.ECOutputs[0]
		if BurnRCD != *out.Address {
			continue // Wrong EC output for a burn
		}

		// Check right output amount
		if out.Amount != 0 {
			continue // You cannot buy EC and burn
		}

		in := tx.FCTInputs[0]
		totalBurned += in.Amount
		burns = append(burns, in)
	}

	var _ = burns
	if totalBurned > 0 { // Just some debugging
		log.WithFields(log.Fields{"height": dblock.Height, "amount": totalBurned, "quantity": len(burns)}).Debug("fct burned")
	}

	// All burns are FCT inputs
	for i := range burns {
		var add factom.FAAddress
		copy(add[:], burns[i].Address[:])
		if _, err := d.Pegnet.AddToBalance(tx, &add, fat2.PTickerFCT, burns[i].Amount); err != nil {
			return err // The tx should be rolled back by the caller if we return an error during this.
		}
	}

	return nil
}

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
