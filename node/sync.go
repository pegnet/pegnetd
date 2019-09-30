package node

import (
	"context"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/config"
	log "github.com/sirupsen/logrus"
)

type BlockSync struct {
	Synced uint32
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

		for d.Sync.Synced < heights.DirectoryBlock {
			if isDone(ctx) {
				return
			}

			// We are not synced, so we need to iterate through the dblocks and sync them
			// one by one. We can only sync our current synced height +1
			// TODO: This skips the genesis block. I'm sure that is fine
			if err := d.SyncBlock(ctx, d.Sync.Synced+1); err != nil {
				log.WithError(err).WithFields(log.Fields{"height": d.Sync.Synced + 1}).Errorf("failed to sync height")
				time.Sleep(retryPeriod)
				// If we fail, we backout to the outer loop. This allows error handling on factomd state to be a bit
				// cleaner, such as a rebooted node with a different db. That node would have a new heights response.
				continue OuterSyncLoop
			}

			// Bump our sync, and march forward
			d.Sync.Synced++
		}

	}

}

// If SyncBlock returns no error, than that height was synced and saved. If any part of the sync fails,
// the whole sync should be rolled back and not applied. An error should then be returned.
// The context should be respected if it is cancelled
func (d *Pegnetd) SyncBlock(ctx context.Context, height uint32) error {
	if isDone(ctx) { // Just an example about how to handle it being cancelled
		return context.Canceled
	}

	fLog := log.WithFields(log.Fields{"height": height})
	fLog.Debug("syncing...")
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
