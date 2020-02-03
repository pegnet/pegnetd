package pegnet

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// This file can be used for node administration related functions

var (
	// PegnetdSyncVersion is an indicator of the version of pegnetd
	// at each height synced. This version number can differ from the tagged
	// version, and is likely only to be updated at hard forks. It is used to
	// detect if a pegnetd was updated late, and therefore has an invalid state.
	//
	// Each fork should increment this number by at least 1
	PegnetdSyncVersion = 1
)

type ForkEvent struct {
	ActivationHeight uint32
	MinimumVersion   int
}

var (
	Hardforks = []ForkEvent{
		// This is the most basic check. All versions are valid for 0
		{0, -1}, // {0, -1}, means at height 0 any version >= -1 is sufficient

		// Future hardforks go here
		// If the pegnet node syncs a hardfork height with any height less than
		// the minimum version, the node will not start.
		//
		// V4 OPR Update
		// Estimated to be  Feb 12, 2020, 18:00 UTC
		{ActivationHeight: 231620, MinimumVersion: 1},
	}
)

// createTableSyncVersion is a SQL string that creates the
// "pn_sync_version" table. This table tracks which heights are synced
// with what version of pegnetd. This will allow pegnetd to detect if
// it was updated before or after a hardfork and respond appropriately.
const createTableSyncVersion = `CREATE TABLE IF NOT EXISTS "pn_sync_version" (
        "height"    		INTEGER NOT NULL,
        "version"       	INTEGER NOT NULL,
        "unix_timestamp"	INTEGER NOT NULL,

        PRIMARY KEY("height")
);
`

// CreateTableSyncVersion is used to expose this table for unit tests
func (p *Pegnet) CreateTableSyncVersion() error {
	_, err := p.DB.Exec(createTableSyncVersion)
	if err != nil {
		return err
	}
	return nil
}

func (p Pegnet) MarkHeightSynced(tx QueryAble, height uint32) error {
	return p.markHeightSyncedVersion(tx, height, PegnetdSyncVersion)
}

func (Pegnet) markHeightSyncedVersion(tx QueryAble, height uint32, version int) error {
	stmtStringFmt := `INSERT INTO "pn_sync_version" 
			("height", "version", "unix_timestamp")
			VALUES (?, ?, ?);`

	stmt, err := tx.Prepare(stmtStringFmt)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(height, version, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

func (p Pegnet) HighestSynced(tx QueryAble) (uint32, error) {
	return p.synced("max", tx)
}

func (p Pegnet) LowestSynced(tx QueryAble) (uint32, error) {
	return p.synced("min", tx)
}

func (Pegnet) synced(adj string, tx QueryAble) (uint32, error) {
	var height uint32
	err := tx.QueryRow(fmt.Sprintf(`SELECT COALESCE(%s(height), 0) FROM pn_sync_version;`, adj)).Scan(&height)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return height, err
}

// FetchMinSyncedVersion returns -1, nil if the height was not found
func (Pegnet) FetchMinSyncedVersion(tx QueryAble, height uint32) (int, error) {
	var version int
	err := tx.QueryRow(`SELECT COALESCE(MIN(version), -1) FROM pn_sync_version WHERE height >= ?;`, height).Scan(&version)
	if err != nil {
		return -1, err
	}
	return version, nil
}

// FetchMaxSyncedVersion returns -1, nil if the height was not found
func (Pegnet) FetchMaxSyncedVersion(tx QueryAble, height uint32) (int, error) {
	var version int
	err := tx.QueryRow(`SELECT COALESCE(MAX(version), -1) FROM pn_sync_version WHERE height >= ?;`, height).Scan(&version)
	if err != nil {
		return -1, err
	}
	return version, nil
}

// CheckHardForks will iterate over all the hardforks post the version_lock
// update, and verify the version that was used to sync was appropriate.
func (p Pegnet) CheckHardForks(tx QueryAble) error {
	// The lowest synced height is the lowest synced height after
	// version_lock feature is included.
	minSynced, err := p.LowestSynced(tx)
	if err != nil {
		return err
	}

	// bs is the node's synced height.
	bs, _ := p.SelectSynced(context.Background(), tx)
	// If bs > minSynced, then this node was just upgraded to version_lock
	// We will check if they are past any hardforks.
	if bs != nil && bs.Synced > minSynced {
		// This means we already have some heights synced. So we need to insert
		// -1s for all hardfork heights we synced prior to the version tracking.
		for _, event := range Hardforks {
			// If we are past the hardfork, put in a -1
			if bs.Synced > event.ActivationHeight {
				_ = p.markHeightSyncedVersion(tx, event.ActivationHeight, -1)
			}
		}
	}

	top, err := p.HighestSynced(tx)
	if err != nil {
		return err
	}

	for _, event := range Hardforks {
		// If the event is not synced past, then we do not need to check
		if event.ActivationHeight <= top {
			version, err := p.FetchMinSyncedVersion(tx, event.ActivationHeight)
			if err != nil {
				return err
			}
			if version < event.MinimumVersion {
				return fmt.Errorf("a hardfork occurred at height %d. This node was not updated prior to the hardfork (or was downgraded at some point after the fork), and synced these blocks with the incorrect version number. The found sync version was %d, and it required %d. The only way to fix this error is to ensure your node is updated, delete your database, and resync", event.ActivationHeight, version, event.MinimumVersion)
			}
		}
	}

	// Catch downgrade with the hardfork check code in it
	// If our PegnetdSyncVersion is less than the highest version we have in
	// our db, then we downgraded
	max, err := p.FetchMaxSyncedVersion(tx, 0)
	if err != nil {
		return err
	}
	if PegnetdSyncVersion < max {
		return fmt.Errorf("pegnetd downgrade was detected. The current pegnetd sync version is %d, but the database shows it was previously at %d. Update pegnetd to prevent a downgrade", PegnetdSyncVersion, max)
	}

	return nil
}
