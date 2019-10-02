package pegnet

import (
	"context"
	"encoding/json"

	"github.com/pegnet/pegnet/modules/opr"
)

const createTableGrade = `CREATE TABLE IF NOT EXISTS "pn_grade" (
	"height" INTEGER NOT NULL,
	"winners" BLOB,
	
	UNIQUE("height")
);
`

const createTableRate = `CREATE TABLE IF NOT EXISTS "pn_rate" (
	"height" INTEGER NOT NULL,
	"token" TEXT,
	"value" INTEGER,
	
	UNIQUE("height", "token")
);
`

func (p *Pegnet) InsertRate(ctx context.Context, height uint32, rates []opr.AssetUint) error {
	for _, r := range rates {
		_, err := p.DB.ExecContext(ctx, "INSERT INTO pn_rate (height, token, value) VALUES ($1, $2, $3)", height, r.Name, r.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pegnet) InsertGrade(ctx context.Context, height uint32, winners []string) error {
	data, err := json.Marshal(winners)
	if err != nil {
		return err
	}
	_, err = p.DB.ExecContext(ctx, "INSERT INTO pn_grade (height, winners) VALUES ($1, $2)", height, data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pegnet) SelectGrade(ctx context.Context, height uint32) ([]string, error) {
	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT winners FROM pn_grade WHERE height = $1", height).Scan(&data)
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

func (p *Pegnet) SelectPrevious(ctx context.Context, height uint32) ([]string, error) {
	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT winners FROM pn_grade WHERE height < $1 ORDER BY height DESC LIMIT 1", height).Scan(&data)
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
