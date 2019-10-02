package pegnet

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/opr"
)

const createTableGrade = `CREATE TABLE IF NOT EXISTS "pn_grade" (
	"height" INTEGER PRIMARY KEY,
	"keymr" BLOB,
	"prevkeymr" BLOB,
	"eb_seq" INTEGER,
	"shorthashes" BLOB,
	"version" INTEGER,
	"cutoff" INTEGER,
	"count" INTEGER,
	
	UNIQUE("height")
);
`

const createTableWinners = `CREATE TABLE IF NOT EXISTS "pn_winners" (
	"height" INTEGER NOT NULL,
	"entryhash" BLOB,
	"oprhash" BLOB,
	"payout" INTEGER,
	"grade" REAL,
	"nonce" BLOB,
	"difficulty" BLOB, -- sqlite can't do uint64, stored as bigendian 8 bytes
	"position" INTEGER,
	"minerid" TEXT,
	"address" BLOB,
	UNIQUE("height", "position")
);`

const createTableRate = `CREATE TABLE IF NOT EXISTS "pn_rate" (
	"height" INTEGER NOT NULL,
	"token" TEXT,
	"value" INTEGER,
	
	UNIQUE("height", "token")
);
`

func (p *Pegnet) InsertRate(tx *sql.Tx, height uint32, rates []opr.AssetUint) error {
	for _, r := range rates {
		_, err := tx.Exec("INSERT INTO pn_rate (height, token, value) VALUES ($1, $2, $3)", height, r.Name, r.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pegnet) InsertGradeBlock(tx *sql.Tx, eblock *factom.EBlock, graded grader.GradedBlock) error {

	data, err := json.Marshal(graded.WinnersShortHashes())
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO pn_grade (height, keymr, prevkeymr, eb_seq, shorthashes, version, cutoff, count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		eblock.Height, eblock.KeyMR, eblock.PrevKeyMR, eblock.Sequence, data, graded.Version(), graded.Cutoff(), graded.Count())
	if err != nil {
		return err
	}

	for _, o := range graded.Graded() {
		diff := make([]byte, 8)
		binary.BigEndian.PutUint64(diff, o.SelfReportedDifficulty)
		_, err = tx.Exec(`INSERT INTO pn_winners (height, entryhash, oprhash, payout, grade, nonce, difficulty, position, minerid, address) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			eblock.Height, o.EntryHash, o.OPRHash, o.Payout(), o.Grade, o.Nonce, diff, o.Position(), o.OPR.GetID(), o.OPR.GetAddress())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pegnet) SelectPreviousWinners(ctx context.Context, height uint32) ([]string, error) {
	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT shorthashes FROM pn_grade WHERE height < $1 ORDER BY height DESC LIMIT 1", height).Scan(&data)
	if err != nil {
		return nil, err
	}

	var winners []string
	err = json.Unmarshal(data, &winners)
	if err != nil {
		return nil, err
	}

	return winners, nil
}
