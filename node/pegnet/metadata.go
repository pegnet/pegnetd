package pegnet

import (
	"context"
	"encoding/binary"
)

const createTableMetadata = `CREATE TABLE IF NOT EXISTS "pn_metadata" (
	"name" TEXT NOT NULL,
	"value" BLOB,
	
	UNIQUE("name")
);
`

func (p *Pegnet) CreateTableMetadata() error {
	_, err := p.DB.Exec(createTableMetadata)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pegnet) InsertSynced(ctx context.Context, height uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, height)
	_, err := p.DB.ExecContext(ctx, "REPLACE INTO pn_metadata (name, value) VALUES ($1, $2)", "synced", b)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pegnet) SelectSynced(ctx context.Context) (uint32, error) {
	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT value FROM pn_metadata WHERE name = $1", "synced").Scan(&data)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(data), nil
}
