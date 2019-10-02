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

	_, err = tx.Exec("REPLACE INTO pn_metadata (name, value) VALUES ($1, $2)", "synced", data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pegnet) SelectSynced(ctx context.Context) (*BlockSync, error) {

	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT value FROM pn_metadata WHERE name = $1", "synced").Scan(&data)
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
