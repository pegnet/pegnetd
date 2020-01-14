package pegnet

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type MinerDominanceResult struct {
	Start       int `json:"startheight"`
	Stop        int `json:"stopheight"`
	Miners      map[string]MinerDominance
	TotalWins   int32 `json:"totalwins"`
	TotalGraded int32 `json:"totalgraded"`
}

type MinerDominance struct {
	Identities       []string `json:"identities"`
	TotalWins        int32    `json:"totalwins"`
	TotalGraded      int32    `json:"totalgraded"`
	WinPercentage    float64  `json:"winpercent"`
	GradedPercentage float64  `json:"gradedpercent"`
}

// SelectMinerDominance returns information around which miners are winning PEG
// and being graded in a block range
// Params are the start and stop block height. The stop height is inclusive.
func (p *Pegnet) SelectMinerDominance(ctx context.Context, start, stop int) (MinerDominanceResult, error) {
	result := MinerDominanceResult{Start: start, Stop: stop, Miners: make(map[string]MinerDominance)}
	if stop < start {
		return result, fmt.Errorf("invalid stop, must be >= start")
	}

	// First check the start/stop bounds and tighten them if we need to
	stmtString := `SELECT COALESCE(MIN(height), 0) AS min, COALESCE(MAX(height), 0) AS max FROM pn_winners`
	row := p.DB.QueryRow(stmtString)
	var min, max int
	err := row.Scan(&min, &max)
	if err != nil {
		return result, err
	}
	if result.Start < min {
		result.Start = min
	}
	if result.Stop > max {
		result.Stop = max
	}

	// Group by unique addresses and count the number of >0 payouts (wins)
	// and the number of count (graded).
	// Also select their identities
	stmtString = `
	SELECT address, COUNT(NULLIF(0, payout)) AS wins, COUNT(*) AS graded, group_concat(DISTINCT minerid) 
	FROM pn_winners
	WHERE pn_winners.height >= ? AND pn_winners.height <= ? GROUP BY pn_winners.address;
	`

	rows, err := p.DB.QueryContext(ctx, stmtString, result.Start, result.Stop)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var address, ids string
		var wins, graded int32
		err := rows.Scan(&address, &wins, &graded, &ids)
		if err != nil {
			return result, err
		}

		// Add to running total
		result.TotalGraded += graded
		result.TotalWins += wins

		// Add address results
		result.Miners[address] = MinerDominance{
			Identities:  strings.Split(ids, ","),
			TotalWins:   wins,
			TotalGraded: graded,
		}
	}

	// Prevent a divide by 0
	if !(result.TotalWins == 0 || result.TotalGraded == 0) {
		for add, m := range result.Miners {
			m.WinPercentage = float64(m.TotalWins) / float64(result.TotalWins)
			m.GradedPercentage = float64(m.TotalGraded) / float64(result.TotalGraded)
			result.Miners[add] = m
		}
	}

	return result, nil
}
