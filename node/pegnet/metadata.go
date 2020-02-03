package pegnet

import (
	"context"
	"database/sql"
	"encoding/json"
)

const createTableMetadata = `CREATE TABLE IF NOT EXISTS "pn_metadata" (
	"name" TEXT NOT NULL,
	"value" BLOB,
	
	UNIQUE("name")
);
`

type BlockSync struct {
	Synced uint32
}

func (p *Pegnet) InsertSynced(tx *sql.Tx, bs *BlockSync) error {
	data, err := json.Marshal(bs)
	if err != nil {
		return err
	}

	// Since this is called for every height, we also can mark the height
	// synced for version checking
	err = p.MarkHeightSynced(tx, bs.Synced)
	if err != nil {
		return err
	}

	_, err = tx.Exec("REPLACE INTO pn_metadata (name, value) VALUES ($1, $2)", "synced", data)
	if err != nil {
		return err
	}

	return nil
}

func (Pegnet) SelectSynced(ctx context.Context, tx QueryAble) (*BlockSync, error) {

	var data []byte
	err := tx.QueryRowContext(ctx, "SELECT value FROM pn_metadata WHERE name = $1", "synced").Scan(&data)
	if err != nil {
		return nil, err
	}

	bs := new(BlockSync)
	err = json.Unmarshal(data, bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
