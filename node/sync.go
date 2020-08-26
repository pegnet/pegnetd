package node

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/conversions"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/graderStake"
	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/transactionid"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node/pegnet"
	log "github.com/sirupsen/logrus"
)

func (d *Pegnetd) GetCurrentSync() uint32 {
	// Should be thread safe since we only have 1 routine writing to it
	return d.Sync.Synced
}

// DBlockSync iterates through dblocks and syncs the various chains
func (d *Pegnetd) DBlockSync(ctx context.Context) {
	retryPeriod := d.Config.GetDuration(config.DBlockSyncRetryPeriod)
	isFirstSync := true

OuterSyncLoop:
	for {
		if isDone(ctx) {
			return // If the user does ctl+c or something
		}

		// Fetch the current highest height
		heights := new(factom.Heights)
		err := heights.Get(nil, d.FactomClient)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{}).Errorf("failed to fetch heights")
			time.Sleep(retryPeriod)
			continue // Loop will just keep retrying until factomd is reached
		}

		if d.Sync.Synced >= heights.DirectoryBlock {
			// We are currently synced, nothing to do. If we are above it, the factomd could
			// be rebooted
			if d.Sync.Synced > heights.DirectoryBlock {
				log.Debugf("Factom node behind. database height = %d, factom height = %d", d.Sync.Synced, heights.DirectoryBlock)
			}

			if isFirstSync {
				isFirstSync = false
				log.WithField("height", d.Sync.Synced).Info("Node is up to date")
			}

			time.Sleep(retryPeriod) // TODO: Should we have a separate polling period?
			continue
		}

		var totalDur time.Duration
		var iterations int

		var longSync bool
		if isFirstSync || heights.DirectoryBlock-d.Sync.Synced > 1 {
			log.WithFields(log.Fields{
				"height":     d.Sync.Synced,
				"syncing-to": heights.DirectoryBlock,
			}).Infof("Starting sync job of %d blocks", heights.DirectoryBlock-d.Sync.Synced)
			longSync = true
		}

		begin := time.Now()
		lastReport := begin
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

			////////////////////////
			// Zeroing funds at Global Burn Address

			// One time operation, Inserts negative balance for the burn address that used during the attack
			// We need to do this before main logic because sqlite db will be locked
			if d.Sync.Synced+1 == V20DevRewardsHeightActivation {
				d.NullifyBurnAddress(ctx, tx, d.Sync.Synced+1)
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
			// update every 15 seconds
			if time.Since(lastReport) > time.Second*15 {
				lastReport = time.Now()
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

		isFirstSync = false
		if longSync {
			longSync = false
			log.WithField("height", d.Sync.Synced).WithField("blocks-synced", iterations).Infof("Finished sync job")
		} else if d.Sync.Synced%6 == 0 {
			log.WithField("height", d.Sync.Synced).Infof("status report")
		}
	}
}

func (d *Pegnetd) NullifyBurnAddress(ctx context.Context, tx *sql.Tx, height uint32) error {
	fLog := log.WithFields(log.Fields{"height": height})

	dblock := new(factom.DBlock)
	dblock.Height = height
	if err := dblock.Get(nil, d.FactomClient); err != nil {
		return err
	}
	heightTimestamp := dblock.Timestamp
	// We need to mock a TXID to record zeroing
	txid := fmt.Sprintf("%064d", height)

	// 1. check current balance
	// 2. substract amounts for every ticker

	// Get all balances for the address
	balances, err := d.Pegnet.SelectBalances(&FAGlobalBurnAddress)

	if err != nil {
		fLog.WithFields(log.Fields{
			"err": err,
		}).Info("zeroing burn | balances retrieval failed")
	}

	i := 0 // value to keep witin 0-9 range for mock tx
	j := 0 //
	for ticker := fat2.PTickerInvalid + 1; ticker < fat2.PTickerMax; ticker++ {
		// Substract from every issuance
		value, _ := balances[ticker]
		_, _, err := d.Pegnet.SubFromBalance(tx, &FAGlobalBurnAddress, ticker, value) // lastInd, txErr, err
		if err != nil {
			fLog.WithFields(log.Fields{
				"time":    heightTimestamp,
				"ticker":  ticker,
				"balance": value,
			}).Info("zeroing burn | substract from balance failed")
		}

		// We need to mock a TXID to record zeroing and it should be unique
		txid = fmt.Sprintf("%064d", height-(uint32(j)))

		// Mock entry hash value
		addTxid := fmt.Sprintf("%d-%s", i, txid)
		j++ // iterate all the time
		i++ // drop to zero to be within 0-9 range
		if i > 9 {
			i = 0
		}

		fLog.WithFields(log.Fields{
			"txid":    txid,
			"addtxid": addTxid,
		}).Info("burn nullify | prep")

		err = d.Pegnet.InsertZeroingCoinbase(tx, txid, addTxid, height, heightTimestamp, value, ticker.String(), FAGlobalBurnAddress)
		if err != nil {
			fLog.WithFields(log.Fields{
				"error": err,
			}).Info("zeroing burn | coinbase tx failed")
			return err
		}

	}

	return nil
}

// If SyncBlock returns no error, than that height was synced and saved. If any part of the sync fails,
// the whole sync should be rolled back and not applied. An error should then be returned.
// The context should be respected if it is cancelled
func (d *Pegnetd) SyncBlock(ctx context.Context, tx *sql.Tx, height uint32) error {
	fLog := log.WithFields(log.Fields{"height": height})
	if isDone(ctx) { // Just an example about how to handle it being cancelled
		return context.Canceled
	}

	dblock := new(factom.DBlock)
	dblock.Height = height
	if err := dblock.Get(nil, d.FactomClient); err != nil {
		return err
	}

	// First, gather all entries we need from factomd
	oprEBlock := dblock.EBlock(OPRChain)
	if oprEBlock != nil {
		if err := multiFetch(oprEBlock, d.FactomClient); err != nil {
			return err
		}
	}
	transactionsEBlock := dblock.EBlock(TransactionChain)
	if transactionsEBlock != nil {
		if err := multiFetch(transactionsEBlock, d.FactomClient); err != nil {
			return err
		}
	}
	sprEBlock := dblock.EBlock(SPRChain)
	if sprEBlock != nil {
		if err := multiFetch(sprEBlock, d.FactomClient); err != nil {
			return err
		}
	}

	// Then, grade the new OPR Block. The results of this will be used
	// to execute conversions that are in holding.
	gradedBlock, err := d.Grade(ctx, oprEBlock)
	gradedSPRBlock, err_s := d.GradeS(ctx, sprEBlock)
	isRatesAvailable := false
	if height < V20HeightActivation {
		if err != nil {
			return err
		}
		isRatesAvailable = gradedBlock != nil && 0 < len(gradedBlock.Winners())
		if gradedBlock != nil {
			err = d.Pegnet.InsertGradeBlock(tx, oprEBlock, gradedBlock)
			if err != nil {
				return err
			}
			winners := gradedBlock.Winners()
			if 0 < len(winners) {
				// PEG has 3 current pricing phases
				// 1: Price is 0
				// 2: Price is determined by equation
				// 3: Price is determine by miners
				var phase pegnet.PEGPricingPhase
				if height < PEGPricingActivation {
					phase = pegnet.PEGPriceIsZero
				}
				if height >= PEGPricingActivation {
					phase = pegnet.PEGPriceIsEquation
				}
				if height >= PEGFreeFloatingPriceActivation {
					phase = pegnet.PEGPriceIsFloating
				}

				err = d.Pegnet.InsertRates(tx, height, winners[0].OPR.GetOrderedAssetsUint(), phase)
				if err != nil {
					return err
				}
			} else {
				fLog.WithFields(log.Fields{"section": "grading", "reason": "no winners"}).Tracef("block not graded")
			}
		} else {
			fLog.WithFields(log.Fields{"section": "grading", "reason": "no graded block"}).Tracef("block not graded")
		}
	} else {
		if err != nil {
			return err
		}
		if err_s != nil {
			return err_s
		}
		// 1. Determine the rates from 2 OPRs (OPR, SPR)
		// 2. Insert rates to DB
		var oprWinners []opr.AssetUint
		var sprWinners []opr.AssetUint

		if gradedBlock != nil {
			err = d.Pegnet.InsertGradeBlock(tx, oprEBlock, gradedBlock)
			if err != nil {
				return err
			}
			winnersOpr := gradedBlock.Winners()
			if 0 < len(winnersOpr) {
				oprWinners = winnersOpr[0].OPR.GetOrderedAssetsUint()
			}
		}
		if gradedSPRBlock != nil {
			winnersSpr := gradedSPRBlock.Winners()
			if 0 < len(winnersSpr) {
				sprWinners = winnersSpr[0].SPR.GetOrderedAssetsUint()
			}
		}
		if 0 < len(oprWinners) || 0 < len(sprWinners) {
			filteredRates, errRate := d.GetAssetRates(oprWinners, sprWinners)
			if errRate != nil {
				return err
			}
			isRatesAvailable = true
			var phase pegnet.PEGPricingPhase
			phase = pegnet.PEGPriceIsFloating
			err = d.Pegnet.InsertRates(tx, height, filteredRates, phase)
			if err != nil {
				return err
			}
		} else {
			fLog.WithFields(log.Fields{"section": "grading", "reason": "no winners from OPR & SPR"}).Tracef("block not graded")
			fmt.Println("no winners from OPR & SPR", ": block not graded")
		}
	}

	// Only apply transactions if we crossed the activation
	if height >= TransactionConversionActivation {
		rates, err := d.Pegnet.SelectPendingRates(ctx, tx, height)
		if err != nil {
			return err
		}

		// Before we apply any balance changes, we will snapshot the balances at the START of the block.
		// This means the balances are the same as the end of the block n-1.
		// This activation is nested in the activation that has the rates
		if height >= V20HeightActivation && height%pegnet.SnapshotRate == 0 {

			// check if the height has no rates, what do we do?
			// check rates from previous height
			if rates == nil {
				// We need to handle the no rates case. Miners could avoid mining this last block.
				// use the last valid rates from last block
				rates, err = d.Pegnet.SelectPendingRates(ctx, tx, height-1)
			}

			// If no rates for second time, skip Snapshot logic
			// otherwise proceed with payout
			if rates != nil {
				err := d.SnapshotPayouts(tx, fLog, rates, height, dblock.Timestamp)
				if err != nil {
					// something wrong happend during payout execution
					return err
				}
			} else {
				// We don't return error as it will stop synchronisation
				// we continue execution but skiping payout for this time
				fLog.WithFields(log.Fields{"section": "staking", "reason": "no rates"}).Tracef("2 last blocks does not contains rates")
			}
		}

		// At this point, we start making updates to the database in a specific order:
		// TODO: ensure we rollback the tx when needed
		// 1) Apply transaction batches that are in holding (conversions are always applied here)
		if isRatesAvailable {
			// Before conversions can be run, we have to adjust and discover the bank's value.
			// We also only sync the bank if the block is a pegnet block
			if err := d.SyncBank(ctx, tx, height); err != nil {
				return err
			}

			if err = d.ApplyTransactionBatchesInHolding(ctx, tx, height, rates); err != nil {
				return err
			}
		}

		//2) Sync transactions in current height and apply transactions
		if transactionsEBlock != nil {
			if err = d.ApplyTransactionBlock(tx, transactionsEBlock); err != nil {
				return err
			}
		}
	}

	// Only apply burn transaction if height does not cross the activation
	if height < V20HeightActivation {
		// 3) Apply FCT --> pFCT burns that happened in this block
		//    These funds will be available for transactions and conversions executed in the next block
		// TODO: Check the order of operations on this and what block to add burns from.
		if err := d.ApplyFactoidBlock(ctx, tx, dblock); err != nil {
			return err
		}
	}

	// 4) Apply effects of graded OPR Block (PEG rewards, if any)
	//    These funds will be available for transactions and conversions executed in the next block
	if gradedBlock != nil {
		if err := d.ApplyGradedOPRBlock(tx, gradedBlock, dblock.Timestamp); err != nil {
			return err
		}
	}

	if height >= V20HeightActivation {
		// 5) Apply effects of graded SPR Block (PEG rewards, if any)
		//    These funds will be available for transactions and conversions executed in the next block
		if gradedSPRBlock != nil {
			if err := d.ApplyGradedSPRBlock(tx, gradedSPRBlock, dblock.Timestamp); err != nil {
				return err
			}
		}
	}

	// 6) Apply Developers Rewards
	if height >= V20DevRewardsHeightActivation && height%pegnet.SnapshotRate == 0 {

		// init developers list explicitely
		// and forward to function
		// we use hardcoded list of dev payouts
		developersList := DeveloperRewardAddreses

		// we want function to accepts dev list as parameter, so different corner cases
		// can be assigned
		err := d.DevelopersPayouts(tx, fLog, height, dblock.Timestamp, developersList)
		if err != nil {
			fLog.WithFields(log.Fields{"section": "devReward", "reason": "developer reward"}).Tracef("something wrong happend during dev payout execution")
		}
	}

	return nil
}

// SnapshotPayouts moves the current shapshot to the "past", and updates the current snapshot. Then
// it proceeds to do the snapshot staking payouts.
func (d *Pegnetd) SnapshotPayouts(tx *sql.Tx, fLog *log.Entry, rates map[fat2.PTicker]uint64, height uint32, heightTimestamp time.Time) error {
	// Snapshot
	snapStart := time.Now()
	err := d.Pegnet.SnapshotCurrent(tx)
	if err != nil {
		return err // Snapshot fails stop all progress and block syncing
	}

	// Payout snapshot
	balances, err := d.Pegnet.SelectSnapshotBalances(tx)
	if err != nil {
		return err // Need to do staking payouts
	}
	staked := make(map[factom.FAAddress]*big.Int)
	for _, bal := range balances {
		// We want all balances in pUSD
		total := new(big.Int)
		for i := fat2.PTicker(1); i < fat2.PTickerMax; i++ {
			if i == fat2.PTickerPEG {
				continue // PEG does not count towards stake total
			}
			if bal.Balances[i] == 0 { // Ignore 0 balances
				continue
			}

			// Convert from pXXX -> pUSD
			c, err := conversions.Convert(int64(bal.Balances[i]), rates[i], rates[fat2.PTickerUSD])
			if err != nil {
				return err
			}

			// add c to running sum
			total = total.Add(total, big.NewInt(c))
		}

		staked[*bal.Address] = total
	}
	// We need to mock a TXID for the staked payouts
	txid := fmt.Sprintf("%064d", height)

	// Sort the staked by highest PUSD
	type StakedAmount struct {
		Address factom.FAAddress
		PUSD    uint64
	}

	var list []StakedAmount
	for add, amt := range staked {
		if !amt.IsUint64() {
			return fmt.Errorf("%s has balance that is not uint64: %s", add, amt)
		}

		uAmt := amt.Uint64()
		if uAmt <= 0 { // Apply a minimum required amount in pUSD
			continue
		}
		// TODO: Check uint64 is safe
		list = append(list, StakedAmount{Address: add, PUSD: uAmt})
	}

	if len(list) == 0 {
		// Abort early since there is no one to pay out
		fLog.WithFields(log.Fields{
			"duration": time.Since(snapStart),
			"eligible": len(list),
		}).Info("staking | balances snapshotted | not paid, there none eligible")
		return nil
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].PUSD < list[j].PUSD
	})

	// Calculate payouts
	payoutindex := make(map[string]int)
	addressMap := make(map[string]factom.FAAddress)

	// 4.5K per block allowed
	// as described in conversions
	totalPayout := uint64(conversions.PerBlockAssetHolders) * pegnet.SnapshotRate
	set := conversions.NewConversionSupply(totalPayout)
	for i, stake := range list {
		addTxid := fmt.Sprintf("%d-%s", i, txid)
		payoutindex[addTxid] = i
		addressMap[addTxid] = stake.Address
		err := set.AddConversion(addTxid, stake.PUSD)
		if err != nil {
			return err
		}
	}

	// ---- Database Payouts ----
	// Inserts tx into the db
	err = d.Pegnet.InsertStakingCoinbase(tx, txid, height, heightTimestamp, set.Payouts(), addressMap)
	if err != nil {
		return err
	}

	// Increase balances
	for addTxid, payout := range set.Payouts() {
		add := addressMap[addTxid] // The address to pay

		_, err = d.Pegnet.AddToBalance(tx, &add, fat2.PTickerPEG, payout)
		if err != nil {
			return err
		}
	}

	// -- End staking calculations
	fLog.WithFields(log.Fields{
		"duration": time.Since(snapStart),
		"eligible": len(list),
		"PEG":      float64(totalPayout) / 1e8, // Float is good enough here,
		"txid":     txid,
	}).Info("staking | balances snapshotted | paid to eligible")
	return nil
}

// Developers Reward Payouts
// implementation of PIP16 - distributed rewards collected for developers every 24h
func (d *Pegnetd) DevelopersPayouts(tx *sql.Tx, fLog *log.Entry, height uint32, heightTimestamp time.Time, developers []DevReward) error {

	totalPayout := uint64(conversions.PerBlockDevelopers) * pegnet.SnapshotRate // every day
	payoutStart := time.Now()

	txid := fmt.Sprintf("%064d", height)
	log.Info("--------------------------------------------------")

	i := 0
	// we need more iterating values to construct unique mock hash
	// should start from 1, because 0-hash reserved for staking mock tx
	j := 1
	for _, dev := range developers {

		// We need to mock a TXID to record dev rewards
		// add more uniqness into hash value by reusing iterating j value in addtion to current height
		// so it doesn't repeat in 25+ blocks
		txid = fmt.Sprintf("%02d%062d", j, height)

		// we calculate developers reward from % pre-defined
		rewardPayout := uint64((conversions.PerBlockDevelopers / 100) * dev.DevRewardPct)
		addr, err := factom.NewFAAddress(dev.DevAddress)

		_, err = d.Pegnet.AddToBalance(tx, &addr, fat2.PTickerPEG, rewardPayout)
		if err != nil {
			return err
		}

		// Mock entry hash value
		addTxid := fmt.Sprintf("%d-%s", i, txid)
		j++ // iterate all the time to build unique hash
		i++
		if i > 9 {
			i = 0
		}

		fLog.WithFields(log.Fields{
			//"txid":    txid,
			"addtxid": addTxid,
			"prct":    dev.DevRewardPct,
		}).Info("developer reward | prep")

		// Get dev address as FAAdress
		FADevAddress, err := factom.NewFAAddress(dev.DevAddress)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"addr":  dev.DevAddress,
			}).Info("error getting developer address")
			return err
		}

		// ---- Database Payouts ----
		// Inserts tx into the db
		err = d.Pegnet.InsertDeveloperRewardCoinbase(tx, txid, addTxid, height, heightTimestamp, rewardPayout, FADevAddress)
		if err != nil {
			log.Info("dev insertion error")
			return err
		}

		fLog.WithFields(log.Fields{
			"total": float64(totalPayout) / 1e8,
			"PEG":   float64(rewardPayout) / 1e8, // Float is good enough here
			"pct":   dev.DevRewardPct,
			"addr":  FADevAddress,
		}).Info("developer reward | paid out to")

		fLog.Info("developer reward | for ", dev.DevGroup)

	}

	fLog.WithFields(log.Fields{
		"total":   float64(totalPayout) / 1e8,
		"elapsed": time.Since(payoutStart),
	}).Info("developer reward | paid out")

	return nil
}

func multiFetch(eblock *factom.EBlock, c *factom.Client) error {
	err := eblock.Get(nil, c)
	if err != nil {
		return err
	}

	work := make(chan int, len(eblock.Entries))
	defer close(work)
	errs := make(chan error)
	defer close(errs)

	for i := 0; i < 8; i++ {
		go func() {
			// TODO: Fix the channels such that a write on a closed channel never happens.
			//		For now, just kill the worker go routine
			defer func() {
				recover()
			}()

			for j := range work {
				errs <- eblock.Entries[j].Get(nil, c)
			}
		}()
	}

	for i := range eblock.Entries {
		work <- i
	}

	count := 0
	for e := range errs {
		count++
		if e != nil {
			// If we return, we close the errs channel, and the working go routine will
			// still try to write to it.
			return e
		}
		if count == len(eblock.Entries) {
			break
		}
	}

	return nil
}

// SyncBank will input the bank value for all heights >= V4OPRUpdate
// The bank table helps track the demand for peg at a given height.
// The bank is the total amount of PEG allowed to be issued for any given height.
func (d *Pegnetd) SyncBank(ctx context.Context, sqlTx *sql.Tx, currentHeight uint32) error {
	if (currentHeight >= V4OPRUpdate) && (currentHeight < V20HeightActivation) { // V4 forward tracks this
		err := d.Pegnet.InsertBankAmount(sqlTx, int32(currentHeight), int64(pegnet.BankBaseAmount))
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyTransactionBatchesInHolding attempts to apply the transaction batches from previous
// blocks that were put into holding because they contained conversions.
// If an error is returned, the sql.Tx should be rolled back by the caller.
func (d *Pegnetd) ApplyTransactionBatchesInHolding(ctx context.Context, sqlTx *sql.Tx, currentHeight uint32, rates map[fat2.PTicker]uint64) error {
	_, height, err := d.Pegnet.SelectMostRecentRatesBeforeHeight(ctx, sqlTx, currentHeight)
	if err != nil {
		return err
	}

	// All batches with a PEG conversion
	var pegConversions []*fat2.TransactionBatch

	// Usually height is just currentHeight-1, but it can be farther back
	// if the miners have skipped a block
	for i := height; i < currentHeight; i++ {
		txBatches, err := d.Pegnet.SelectTransactionBatchesInHoldingAtHeight(uint64(i))
		if err != nil {
			return err
		}

		// For all conversions, we need to apply the PEG conversion limit.
		// This means, we need to find all valid conversions and pay them out
		// on a proportional basis.
		for i, txBatch := range txBatches {
			// Re-validate transaction batch because timestamp might not be valid anymore

			if currentHeight >= V20HeightActivation {
				if err := txBatch.ValidatePegTx(int32(currentHeight)); err != nil {
					d.Pegnet.SetTransactionHistoryExecuted(sqlTx, txBatch, -2)
					continue
				}
			}

			if err := txBatch.Validate(int32(currentHeight)); err != nil {
				d.Pegnet.SetTransactionHistoryExecuted(sqlTx, txBatch, -2)
				continue
			}
			isReplay, err := d.Pegnet.IsReplayTransaction(sqlTx, txBatch.Entry.Hash)
			if err != nil {
				return err
			} else if isReplay {
				continue
			}

			// This will apply all batche inputs, and all batch outputs except
			// conversions to PEG if we are above the PegnetConversionLimit Act
			err = d.applyTransactionBatch(sqlTx, txBatch, rates, currentHeight)
			// The err needs to be converted to a code. If the err is still
			// not nil, then the code is 0 and the error is probably db related.
			// If the code is < 0, the tx is rejected.
			// If the code is > 0 and the err is nil, the tx is accepted.
			rejectCode, err := pegnet.IsRejectedTx(err)
			if err != nil { // Likely a db error
				return err
			} else if rejectCode < 0 { // Tx rejected
				d.Pegnet.SetTransactionHistoryExecuted(sqlTx, txBatch, rejectCode)
			} else if err == nil { // Tx accepted
				if currentHeight < V20HeightActivation {
					// If PegnetConversion limits are on, we process conversions to
					// peg in a second pass.
					if currentHeight >= PegnetConversionLimitActivation && txBatch.HasPEGRequest() {
						// Batch applied, we need to do the PEG conversions at the end
						pegConversions = append(pegConversions, txBatches[i])
					}
				}
			}
		}

		// Apply all PEG Requests
		// The `conversions.PerBlock` is the allowed amount of PEG to be
		// converted. So when the bank is implemented, it should be passed in
		// here.
		//
		// This is processing each height of conversions as its own block
		// of conversions. After the v4 update, all pending conversions get
		// processed together for peg conversions
		if currentHeight >= PegnetConversionLimitActivation && currentHeight < V4OPRUpdate {
			// All heights before v4 use the currentHeight-1 with a 5K PEG bank
			bank := pegnet.BankBaseAmount
			err = d.recordPegnetRequests(sqlTx, pegConversions, rates, currentHeight, bank, int32(currentHeight-1))
			if err != nil {
				return err
			}
			pegConversions = []*fat2.TransactionBatch{}
		}
	}

	// Process all pending using the same bank
	if (currentHeight >= V4OPRUpdate) && (currentHeight < V20HeightActivation) {
		// The bank entry should be here from the sync banks called before this function.
		bentry, err := d.Pegnet.SelectBankEntry(sqlTx, int32(currentHeight))
		if err != nil {
			return err
		}
		err = d.recordPegnetRequests(sqlTx, pegConversions, rates, currentHeight, uint64(bentry.BankAmount), int32(currentHeight))
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyTransactionBlock puts conversion-containing transaction batches into holding,
// and applys the balance updates for all transaction batches able to be executed
// immediately. If an error is returned, the sql.Tx should be rolled back by the caller.
func (d *Pegnetd) ApplyTransactionBlock(sqlTx *sql.Tx, eblock *factom.EBlock) error {
	for blockorder, entry := range eblock.Entries {
		txBatch, err := fat2.NewTransactionBatch(entry, int32(eblock.Height))
		if err != nil {
			continue // Bad formatted entry
		}

		log.WithFields(log.Fields{
			"height":      eblock.Height,
			"entryhash":   entry.Hash.String(),
			"conversions": txBatch.HasConversions(),
			"txs":         len(txBatch.Transactions)}).Tracef("tx found")

		isReplay, err := d.Pegnet.IsReplayTransaction(sqlTx, txBatch.Entry.Hash)
		if err != nil {
			return err
		} else if isReplay {
			continue
		}
		// At this point, we know that the transaction batch is valid and able to be executed.

		if err = d.Pegnet.InsertTransactionHistoryTxBatch(sqlTx, blockorder, txBatch, eblock.Height); err != nil {
			return err
		}

		// A transaction batch that contains conversions must be put into holding to be executed
		// in a future block. This prevents gaming of conversions where an actor
		// can know the exchange rates of the future ahead of time.
		if txBatch.HasConversions() {
			_, err = d.Pegnet.InsertTransactionBatchHolding(sqlTx, txBatch, uint64(eblock.Height), eblock.KeyMR)
			if err != nil {
				return err
			}
			continue
		}

		// No conversions in the batch, it can be applied immediately
		if err = d.applyTransactionBatch(sqlTx, txBatch, nil, eblock.Height); err != nil &&
			err != pegnet.InsufficientBalanceErr { // Allowed Exception
			return err
		} else if err == pegnet.InsufficientBalanceErr {
			d.Pegnet.SetTransactionHistoryExecuted(sqlTx, txBatch, -1)
		}
	}
	return nil
}

// applyTransactionBatch
//	currentHeight is just for tracing
func (d *Pegnetd) applyTransactionBatch(sqlTx *sql.Tx, txBatch *fat2.TransactionBatch, rates map[fat2.PTicker]uint64, currentHeight uint32) error {
	balances := make(map[factom.FAAddress]map[fat2.PTicker]uint64)

	// We need to do all checks up front, then apply the tx
	for _, tx := range txBatch.Transactions {
		// First check the input address has the funds
		bals, err := d.Pegnet.SelectPendingBalances(sqlTx, &tx.Input.Address)
		if err != nil {
			return err
		}

		balances[tx.Input.Address] = bals
		bal := bals[tx.Input.Type]

		if tx.Input.Amount > bal {
			return pegnet.InsufficientBalanceErr // This error is safe to pass, and it is handled to skip this batch
		}

		// conversion checks
		if tx.IsConversion() {
			if rates == nil || len(rates) == 0 {
				// This error will fail the block
				return fmt.Errorf("rates must exist if TransactionBatch contains conversions")
			}
			if rates[tx.Input.Type] == 0 || rates[tx.Conversion] == 0 {
				// This error will not fail the block, skip the tx
				return pegnet.ZeroRatesError // 0 rates result in an invalid tx. So we drop it
			}

			// pXXX -> pFCT conversions are disabled at the activation height
			if currentHeight >= OneWaypFCTConversions && tx.Conversion == fat2.PTickerFCT {
				return pegnet.PFCTOneWayError
			}

			// TODO: For now any bogus amounts will be tossed. Someone can fake an overflow for example,
			// 		and hold us up forever.
			_, err := conversions.Convert(int64(tx.Input.Amount), rates[tx.Input.Type], rates[tx.Conversion])
			if err != nil {
				return nil
			}
		} else {
			// There are no additional transfer checks
		}
	}

	// Now check the batch does not drive an input negative
	// TODO: A nested tx would be much easier, since we have to literally implement the same loop twice
	for _, tx := range txBatch.Transactions {
		if balances[tx.Input.Address][tx.Input.Type] < tx.Input.Amount {
			return pegnet.InsufficientBalanceErr
		}
		balances[tx.Input.Address][tx.Input.Type] -= tx.Input.Amount

		if tx.IsConversion() {
			outputAmount, err := conversions.Convert(int64(tx.Input.Amount), rates[tx.Input.Type], rates[tx.Conversion])
			if err != nil {
				return err
			}
			balances[tx.Input.Address][tx.Conversion] += uint64(outputAmount)
		} else {
			for _, transfer := range tx.Transfers {
				// If it is one of our inputs
				if _, ok := balances[transfer.Address]; ok {
					balances[transfer.Address][tx.Input.Type] += transfer.Amount
				}
			}
		}
	}

	// The tx batch should be 100% valid to apply
	err := d.recordBatch(sqlTx, txBatch, rates, currentHeight)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"height":     currentHeight, // Just for log traces
		"entryhash":  txBatch.Entry.Hash.String(),
		"conversion": txBatch.HasConversions(),
		"txs":        len(txBatch.Transactions)}).Tracef("tx applied")

	return nil
}

// recordBatch will submit the batch to the database. We assume the tx is 100%
// valid at this point.
func (d *Pegnetd) recordBatch(sqlTx *sql.Tx, txBatch *fat2.TransactionBatch, rates map[fat2.PTicker]uint64, currentHeight uint32) error {
	for txIndex, tx := range txBatch.Transactions {
		_, txErr, err := d.Pegnet.SubFromBalance(sqlTx, &tx.Input.Address, tx.Input.Type, tx.Input.Amount)
		if err != nil {
			return err
		} else if txErr != nil {
			// This should fail the block
			return fmt.Errorf("uncaught: %s", txErr.Error())
		}
		_, err = d.Pegnet.InsertTransactionRelation(sqlTx, tx.Input.Address, txBatch.Entry.Hash, uint64(txIndex), false, tx.IsConversion())
		if err != nil {
			return err
		}

		if err = d.Pegnet.SetTransactionHistoryExecuted(sqlTx, txBatch, int64(currentHeight)); err != nil {
			return err
		}

		// All conversions to PEG after the activation height have their
		// outputs processed later. We only subtract their inputs right now.
		if currentHeight >= PegnetConversionLimitActivation && tx.IsPEGRequest() {
			// Ensure the output is valid, as we will process it later
			_, err := conversions.Convert(int64(tx.Input.Amount), rates[tx.Input.Type], rates[tx.Conversion])
			if err != nil {
				return err
			}
			continue // PEG Outputs are handled elsewhere
		}

		// Outputs
		if tx.IsConversion() {
			// Conversions Output
			outputAmount, err := conversions.Convert(int64(tx.Input.Amount), rates[tx.Input.Type], rates[tx.Conversion])
			if err != nil {
				return err
			}

			if err = d.Pegnet.SetTransactionHistoryConvertedAmount(sqlTx, txBatch, txIndex, outputAmount); err != nil {
				return err
			}

			_, err = d.Pegnet.AddToBalance(sqlTx, &tx.Input.Address, tx.Conversion, uint64(outputAmount))
			if err != nil {
				return err
			}

		} else {
			// Transfer Outputs

			for _, transfer := range tx.Transfers {
				// if transfer to Burn address do nothing, otherwise add balances
				if transfer.Address != FAGlobalBurnAddress {
					_, err = d.Pegnet.AddToBalance(sqlTx, &transfer.Address, tx.Input.Type, transfer.Amount)
					if err != nil {
						return err
					}
					_, err = d.Pegnet.InsertTransactionRelation(sqlTx, transfer.Address, txBatch.Entry.Hash, uint64(txIndex), true, false)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type pegRequest struct {
	TxID               string
	Batch              *fat2.TransactionBatch
	PegAmountRequested uint64
	TxIndex            int
}

func (d *Pegnetd) recordPegnetRequests(sqlTx *sql.Tx, txBatchs []*fat2.TransactionBatch, rates map[fat2.PTicker]uint64, currentHeight uint32, bank uint64, bankHeight int32) error {
	limit := conversions.NewConversionSupply(bank)
	txData := make(map[string]pegRequest)

	// First we need to extract all the txs that are pegnet requests
	// The batches we are given might contain 0 or more pegnet requests.
	for i := range txBatchs {
		for j := range txBatchs[i].Transactions {
			// Retrieve each tx individually.
			tx := txBatchs[i].Transactions[j]
			// The txid helps determine the order when deciding who
			// gets the dust
			txid := transactionid.FormatTxID(j, txBatchs[i].Entry.Hash.String())

			// We caught the error earlier, so we can ignore it here.
			pegAmt, _ := conversions.Convert(int64(tx.Input.Amount), rates[tx.Input.Type], rates[tx.Conversion])
			txData[txid] = pegRequest{
				TxID:               txid,
				Batch:              txBatchs[i],
				PegAmountRequested: uint64(pegAmt),
				TxIndex:            j,
			}

			// limit calculates how much each PEG each tx is allocted
			err := limit.AddConversion(txid, txData[txid].PegAmountRequested)
			if err != nil {
				return err // No recovery
			}
		}
	}

	// Now we have all the PEG amounts and requests in, time to do the payouts.
	pegPayouts := limit.Payouts()
	var totalPaid int64
	for txid, pegYield := range pegPayouts {
		totalPaid += int64(pegYield)
		tx := txData[txid].Batch.Transactions[txData[txid].TxIndex]

		refundAmt := conversions.Refund(int64(tx.Input.Amount), int64(pegYield), rates[tx.Input.Type], rates[tx.Conversion])

		log.WithFields(log.Fields{
			"batch-entryhash": txData[txid].Batch.Entry.Hash.String(),
			"height":          currentHeight,
			"txid":            txid,
			"txindex":         txData[txid].TxIndex,
			"refund":          refundAmt,
			"pegyield":        pegYield,
			"inputtype":       tx.Input.Type.String(),
		}).Tracef("refund set")

		if err := d.Pegnet.SetTransactionHistoryPEGConvertedRequestAmount(sqlTx, txData[txid].Batch, txData[txid].TxIndex, int64(pegYield), refundAmt); err != nil {
			return err
		}

		// PEG addition
		if _, err := d.Pegnet.AddToBalance(sqlTx, &tx.Input.Address, tx.Conversion, pegYield); err != nil {
			return err
		}

		// Refund
		if _, err := d.Pegnet.AddToBalance(sqlTx, &tx.Input.Address, tx.Input.Type, uint64(refundAmt)); err != nil {
			return err
		}
	}

	// The bankheight == currentheight after V4Update fork
	if bankHeight >= int32(V4OPRUpdate) {
		err := d.Pegnet.UpdateBankEntry(sqlTx, bankHeight, totalPaid, int64(limit.TotalRequested()))
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyFactoidBlock applies the FCT burns that occurred within the given
// DBlock. If an error is returned, the sql.Tx should be rolled back by the caller.
func (d *Pegnetd) ApplyFactoidBlock(ctx context.Context, tx *sql.Tx, dblock *factom.DBlock) error {
	fblock := new(factom.FBlock)
	fblock.Height = dblock.Height
	if err := fblock.Get(nil, d.FactomClient); err != nil {
		return err
	}

	var totalBurned uint64
	var burns []factom.FactoidTransaction

	// Register all burns. Burns have a few requirements
	// - Only 1 output, and that output must be the EC burn address
	// - The output amount must be 0
	// - Must only have 1 input
	for i := range fblock.Transactions {
		if isDone(ctx) {
			return context.Canceled
		}

		if err := fblock.Transactions[i].Get(nil, d.FactomClient); err != nil {
			return err
		}

		tx := fblock.Transactions[i]
		// Check number of inputs/outputs
		if len(tx.ECOutputs) != 1 || len(tx.FCTInputs) != 1 || len(tx.FCTOutputs) > 0 {
			continue // Wrong number of ins/outs for a burn
		}

		// Check correct output
		out := tx.ECOutputs[0]
		if BurnRCD != out.Address {
			continue // Wrong EC output for a burn
		}

		// Check right output amount
		if out.Amount != 0 {
			continue // You cannot buy EC and burn
		}

		in := tx.FCTInputs[0]
		totalBurned += in.Amount
		burns = append(burns, tx)
	}

	var _ = burns
	if totalBurned > 0 { // Just some debugging
		log.WithFields(log.Fields{"height": dblock.Height, "amount": totalBurned, "quantity": len(burns)}).Debug("fct burned")
	}

	// All burns are FCT inputs
	for i := range burns {
		var add factom.FAAddress
		copy(add[:], burns[i].FCTInputs[0].Address[:])
		if _, err := d.Pegnet.AddToBalance(tx, &add, fat2.PTickerFCT, burns[i].FCTInputs[0].Amount); err != nil {
			return err
		}

		if err := d.Pegnet.InsertFCTBurn(tx, fblock.KeyMR, burns[i], dblock.Height); err != nil {
			return err
		}
	}

	return nil
}

// ApplyGradedOPRBlock pays out PEG to the winners of the given GradedBlock.
// If an error is returned, the sql.Tx should be rolled back by the caller.
func (d *Pegnetd) ApplyGradedOPRBlock(tx *sql.Tx, gradedBlock grader.GradedBlock, timestamp time.Time) error {
	winners := gradedBlock.Winners()
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

		if _, err := d.Pegnet.AddToBalance(tx, &addr, fat2.PTickerPEG, uint64(winners[i].Payout())); err != nil {
			return err
		}

		if err := d.Pegnet.InsertCoinbase(tx, winners[i], addr[:], timestamp); err != nil {
			return err
		}
	}
	return nil
}

// ApplyGradedOPRBlock pays out PEG to the winners of the given GradedBlock.
// If an error is returned, the sql.Tx should be rolled back by the caller.
func (d *Pegnetd) ApplyGradedSPRBlock(tx *sql.Tx, gradedSPRBlock graderStake.GradedBlock, timestamp time.Time) error {
	winners := gradedSPRBlock.Winners()
	for i := range winners {
		addr, err := factom.NewFAAddress(winners[i].SPR.GetAddress())
		if err != nil {
			// TODO: This is kinda an odd case. I think we should just drop the rewards
			// 		for an invalid address. We can always add back the rewards and they will have
			//		a higher balance after a change.
			log.WithError(err).WithFields(log.Fields{
				"height": winners[i].SPR.GetHeight(),
				"ehash":  fmt.Sprintf("%x", winners[i].EntryHash),
			}).Warnf("failed to reward")
			continue
		}

		if _, err := d.Pegnet.AddToBalance(tx, &addr, fat2.PTickerPEG, uint64(winners[i].Payout())); err != nil {
			return err
		}

		if err := d.Pegnet.InsertStaking100Coinbase(tx, winners[i], addr[:], timestamp); err != nil {
			return err
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

func (d *Pegnetd) GetAssetRates(oprWinners []opr.AssetUint, sprWinners []opr.AssetUint) ([]opr.AssetUint, error) {
	if oprWinners != nil && sprWinners == nil {
		return oprWinners, nil
	}
	if oprWinners == nil && sprWinners != nil {
		return sprWinners, nil
	}
	if oprWinners != nil && sprWinners != nil {
		// 1) sprWinners determine tolerance range
		// 2) oprWinners set the real rates
		if len(oprWinners) != len(sprWinners) {
			return nil, fmt.Errorf("SPR & OPR use different assets version")
		}

		var filteredRates []opr.AssetUint
		for i := range oprWinners {
			if oprWinners[i].Name == sprWinners[i].Name {
				sprRate := sprWinners[i].Value
				oprRate := oprWinners[i].Value
				toleranceRate := 0.1 // 10% band
				toleranceBandHigh := float64(sprRate) * (1 + toleranceRate)
				toleranceBandLow := float64(sprRate) * (1 - toleranceRate)
				if (float64(oprRate) >= toleranceBandLow) && (float64(oprRate) <= toleranceBandHigh) {
					filteredRates = append(filteredRates, oprWinners[i])
				} else {
					fmt.Println("opr is out side of spr's tolerance band")
					fmt.Println("=== asset name:", oprWinners[i].Name)
					fmt.Println("<<<<=== opr rate:", oprWinners[i].Value)
					fmt.Println("<<<<=== spr rate:", sprWinners[i].Value)
					return nil, fmt.Errorf("opr is out side of tolerance band")
				}
			}
		}
		return filteredRates, nil
	}
	return nil, fmt.Errorf("no winners")
}
