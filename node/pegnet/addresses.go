package pegnet

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
)

// pn_addresses

func createTableAddressesWithTableName(tableName string) string {
	return fmt.Sprintf(createTableAddresses, tableName)
}

const createTableAddresses = `CREATE TABLE IF NOT EXISTS "%s" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "peg_balance"   INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("peg_balance" >= 0),
        "pusd_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pusd_balance" >= 0),
        "peur_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("peur_balance" >= 0),
        "pjpy_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pjpy_balance" >= 0),
        "pgbp_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pgbp_balance" >= 0),
        "pcad_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pcad_balance" >= 0),
        "pchf_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pchf_balance" >= 0),
        "pinr_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pinr_balance" >= 0),
        "psgd_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("psgd_balance" >= 0),
        "pcny_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pcny_balance" >= 0),
        "phkd_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("phkd_balance" >= 0),
        "pkrw_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pkrw_balance" >= 0),
        "pbrl_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pbrl_balance" >= 0),
        "pphp_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pphp_balance" >= 0),
        "pmxn_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pmxn_balance" >= 0),
        "pxau_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxau_balance" >= 0),
        "pxag_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxag_balance" >= 0),
        "pxbt_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxbt_balance" >= 0),
        "peth_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("peth_balance" >= 0),
        "pltc_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pltc_balance" >= 0),
        "prvn_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("prvn_balance" >= 0),
        "pxbc_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxbc_balance" >= 0),
        "pfct_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pfct_balance" >= 0),
        "pbnb_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pbnb_balance" >= 0),
        "pxlm_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxlm_balance" >= 0),
        "pada_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pada_balance" >= 0),
        "pxmr_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pxmr_balance" >= 0),
        "pdash_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pdash_balance" >= 0),
        "pzec_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pzec_balance" >= 0),
        "pdcr_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pdcr_balance" >= 0),
        -- v4 additions
        "paud_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("paud_balance" >= 0),
        "pnzd_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pnzd_balance" >= 0),
        "psek_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("psek_balance" >= 0),
        "pnok_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pnok_balance" >= 0),
        "prub_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("prub_balance" >= 0),
        "pzar_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pzar_balance" >= 0),
        "ptry_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("ptry_balance" >= 0),
        "peos_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("peos_balance" >= 0),
        "plink_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("plink_balance" >= 0),
        "patom_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("patom_balance" >= 0),
        "pbat_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pbat_balance" >= 0),
        "pxtz_balance"  INTEGER NOT NULL DEFAULT 0 
                        CONSTRAINT "insufficient balance" CHECK ("pxtz_balance" >= 0),
		-- v5 additions
        "phbar_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("phbar_balance" >= 0),
        "pneo_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pneo_balance" >= 0),
        "pcro_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pcro_balance" >= 0),
        "petc_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("petc_balance" >= 0),
        "pont_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pont_balance" >= 0),
        "pdoge_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pdoge_balance" >= 0),
        "pvet_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pvet_balance" >= 0),
        "pht_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pht_balance" >= 0),
        "palgo_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("palgo_balance" >= 0),
        "pdgb_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pdgb_balance" >= 0),
        "paed_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("paed_balance" >= 0),
        "pars_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pars_balance" >= 0),
        "ptwd_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("ptwd_balance" >= 0),
        "prwf_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("prwf_balance" >= 0),
        "pkes_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pkes_balance" >= 0),
        "pugx_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pugx_balance" >= 0),
        "ptzs_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("ptzs_balance" >= 0),
        "pbif_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pbif_balance" >= 0),
        "petb_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("petb_balance" >= 0),
        "pngn_balance"  INTEGER NOT NULL DEFAULT 0
                        CONSTRAINT "insufficient balance" CHECK ("pngn_balance" >= 0)		
);
CREATE INDEX IF NOT EXISTS "idx_address_balances_address_id" ON "pn_addresses"("address");
`

// Add additional columns if they do not exist
const v4migrationNeeded = `
SELECT CASE
   -- If the new currencies do not already exist, run the query to add the columns.
    WHEN (SELECT COUNT(*) FROM pragma_table_info('pn_addresses') WHERE name = 'pxtz_balance') > 0 THEN
        false -- No migration needed
    ELSE
        true
END AS migrate;
`

const addressTableV4Migration = `
ALTER TABLE pn_addresses
        ADD "paud_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("paud_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pnzd_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("pnzd_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "psek_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("psek_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pnok_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("pnok_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "prub_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("prub_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pzar_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("pzar_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "ptry_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("ptry_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "peos_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("peos_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "plink_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("plink_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "patom_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("patom_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pbat_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("pbat_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pxtz_balance"  INTEGER NOT NULL DEFAULT 0
			CONSTRAINT "insufficient balance" CHECK ("pxtz_balance" >= 0);
`

func (p *Pegnet) v4MigrationNeeded() (migrate bool, err error) {
	row := p.DB.QueryRow(v4migrationNeeded)
	err = row.Scan(&migrate)
	return
}

// Add additional columns if they do not exist
const v5migrationNeeded = `
SELECT CASE
   -- If the new currencies do not already exist, run the query to add the columns.
    WHEN (SELECT COUNT(*) FROM pragma_table_info('pn_addresses') WHERE name = 'pht_balance') > 0 THEN
        false -- No migration needed
    ELSE
        true
END AS migrate;
`

const addressTableV5Migration = `
ALTER TABLE pn_addresses
        ADD "phbar_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("phbar_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pneo_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pneo_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pcro_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pcro_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "petc_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("petc_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pont_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pont_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pdoge_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pdoge_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pvet_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pvet_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pht_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pht_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "palgo_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("palgo_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pdgb_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pdgb_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "paed_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("paed_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pars_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pars_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "ptwd_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("ptwd_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "prwf_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("prwf_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pkes_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pkes_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pugx_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pugx_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "ptzs_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("ptzs_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pbif_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pbif_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "petb_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("petb_balance" >= 0);
ALTER TABLE pn_addresses
        ADD "pngn_balance"  INTEGER NOT NULL DEFAULT 0
            CONSTRAINT "insufficient balance" CHECK ("pngn_balance" >= 0);
`

func (p *Pegnet) v5MigrationNeeded() (migrate bool, err error) {
	row := p.DB.QueryRow(v5migrationNeeded)
	err = row.Scan(&migrate)
	return
}

// Use addressSelectCols instead of '*' to ensure the order is always the same
var addressSelectCols = ``

// snapshotMinSelectCols is for querying against both snapshot tables
var snapshotMinSelectCols = ``

func init() {
	// +2 for 2 extra cols, -1 since the max is +1
	cols := make([]string, fat2.PTickerMax+2-1)
	cols[0] = "id"
	cols[1] = "address"
	for i := 1; i < int(fat2.PTickerMax); i++ {
		cols[i+1] = strings.ToLower(fat2.PTicker(i).String()) + "_balance"
	}
	addressSelectCols = strings.Join(cols, ",") + " "

	// Query from the min value of the 2 balances across the 2 snapshots
	snCols := make([]string, fat2.PTickerMax+2-1)
	snCols[0] = "sn_current.id as id"
	snCols[1] = "sn_current.address as address"
	for i := 1; i < int(fat2.PTickerMax); i++ {
		token := strings.ToLower(fat2.PTicker(i).String()) + "_balance"
		snCols[i+1] = fmt.Sprintf("MIN(sn_current.%s, sn_past.%s) AS %s", token, token, token)
	}

	snapshotMinSelectCols = strings.Join(snCols, ",") + " "
}

func (p *Pegnet) CreateTableAddresses() error {
	_, err := p.DB.Exec(createTableAddressesWithTableName("pn_addresses"))
	if err != nil {
		return err
	}

	_, err = p.DB.Exec(createTableAddressesWithTableName("snapshot_current"))
	if err != nil {
		return err
	}

	_, err = p.DB.Exec(createTableAddressesWithTableName("snapshot_past"))
	if err != nil {
		return err
	}
	return nil
}

// AddToBalance adds value to the typed balance of adr, creating a new row in
// "pn_addresses" if it does not exist. If successful, the row id is returned.
func (p *Pegnet) AddToBalance(tx *sql.Tx, adr *factom.FAAddress, ticker fat2.PTicker, value uint64) (int64, error) {
	stmtStringFmt := `INSERT INTO "pn_addresses"
                ("address", "%[1]s_balance") VALUES (?, ?)
                ON CONFLICT("address") DO
                UPDATE SET "%[1]s_balance" = "%[1]s_balance" + "excluded"."%[1]s_balance";`
	stmt, err := tx.Prepare(fmt.Sprintf(stmtStringFmt, strings.ToLower(ticker.String())))
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(adr[:], value)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

// SubFromBalance subtracts value from the typed balance of adr, creating a new row in
// "pn_addresses" if it does not exist and value is 0. If successful, the row id is returned,
// otherwise 0. If subtracting sub would result in a negative balance, txErr is not nil
// and starts with "insufficient balance".
func (p *Pegnet) SubFromBalance(tx *sql.Tx, adr *factom.FAAddress, ticker fat2.PTicker, value uint64) (id int64, txError, err error) {
	if value == 0 {
		// Allow tx's with zeros to result in an INSERT.
		id, err = p.AddToBalance(tx, adr, ticker, 0)
		return id, nil, err
	}
	balance, err := p.SelectPendingBalance(tx, adr, ticker)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, InsufficientBalanceErr, nil
		}
		return 0, nil, err
	}
	if balance < value {
		return 0, InsufficientBalanceErr, nil
	}

	stmtStringFmt := `UPDATE pn_addresses SET %[1]s_balance = %[1]s_balance - ? WHERE address = ?;`
	tickerLower := strings.ToLower(ticker.String())
	stmt, err := tx.Prepare(fmt.Sprintf(stmtStringFmt, tickerLower))
	if err != nil {
		return 0, nil, err
	}
	res, err := stmt.Exec(value, adr[:])
	if err != nil {
		return 0, nil, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, nil, err
	}
	return lastID, nil, nil
}

// SelectPendingBalance returns the balance of an individual token type for the given
// address, in the context of a given sql transaction. If the address is not in the
// database or will not be in the database after the tx is committed, 0 will be returned.
func (p *Pegnet) SelectPendingBalance(tx *sql.Tx, adr *factom.FAAddress, ticker fat2.PTicker) (uint64, error) {
	if ticker <= fat2.PTickerInvalid || fat2.PTickerMax <= ticker {
		return 0, fmt.Errorf("invalid token type")
	}
	var balance uint64
	stmtStringFmt := `SELECT %s_balance FROM pn_addresses WHERE address = ?;`
	stmt := fmt.Sprintf(stmtStringFmt, strings.ToLower(ticker.String()))
	err := tx.QueryRow(stmt, adr[:]).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

// SelectBalance returns the balance of an individual token type for the given
// address. If the address is not in the database, 0 will be returned.
func (p *Pegnet) SelectBalance(adr *factom.FAAddress, ticker fat2.PTicker) (uint64, error) {
	if ticker <= fat2.PTickerInvalid || fat2.PTickerMax <= ticker {
		return 0, fmt.Errorf("invalid token type")
	}
	var balance uint64
	stmtStringFmt := `SELECT %s_balance FROM pn_addresses WHERE address = ?;`
	stmt := fmt.Sprintf(stmtStringFmt, strings.ToLower(ticker.String()))
	err := p.DB.QueryRow(stmt, adr[:]).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

type BalancePair struct {
	Address *factom.FAAddress
	Balance uint64
}
type BalancesPair struct {
	Address  *factom.FAAddress
	Balances []uint64
}

func (p *Pegnet) IsIncludedTopPEGAddress(address []byte) bool {
	count := 100 // Top 100 addresses
	stmtStringFmt := `SELECT address, %[1]s_balance FROM pn_addresses WHERE %[1]s_balance > 0 ORDER BY %[1]s_balance DESC LIMIT ?;`
	stmt := fmt.Sprintf(stmtStringFmt, strings.ToLower("PEG"))
	rows, err := p.DB.Query(stmt, count)
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var Balance uint64
		var adr []byte
		if err := rows.Scan(&adr, &Balance); err != nil {
			return false
		}
		if bytes.Equal(address, adr) {
			return true
		}
	}
	return false
}

// SelectRichList returns the balance of all addresses for a given ticker
func (p *Pegnet) SelectRichList(ticker fat2.PTicker, count int) ([]BalancePair, error) {
	if ticker <= fat2.PTickerInvalid || fat2.PTickerMax <= ticker {
		return nil, fmt.Errorf("invalid token type")
	}
	if count < 1 {
		return nil, fmt.Errorf("invalid count")
	}
	var res []BalancePair
	stmtStringFmt := `SELECT address, %[1]s_balance FROM pn_addresses WHERE %[1]s_balance > 0 ORDER BY %[1]s_balance DESC LIMIT ?;`
	stmt := fmt.Sprintf(stmtStringFmt, strings.ToLower(ticker.String()))
	rows, err := p.DB.Query(stmt, count)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pair BalancePair
		var adr []byte

		if err := rows.Scan(&adr, &pair.Balance); err != nil {
			return nil, err
		}

		var fa factom.FAAddress
		copy(fa[:], adr)
		pair.Address = &fa

		res = append(res, pair)
	}

	return res, nil
}

// SelectPendingBalances returns a map of all valid PTickers and their associated
// balances for the given address. If the address is not in the database,
// the map will contain 0 for all valid PTickers. This works on the pending tx
func (p *Pegnet) SelectPendingBalances(tx *sql.Tx, adr *factom.FAAddress) (map[fat2.PTicker]uint64, error) {
	return p.selectBalances(tx, adr)
}

// SelectBalances returns a map of all valid PTickers and their associated
// balances for the given address. If the address is not in the database,
// the map will contain 0 for all valid PTickers.
func (p *Pegnet) SelectBalances(adr *factom.FAAddress) (map[fat2.PTicker]uint64, error) {
	return p.selectBalances(p.DB, adr)
}

// SelectPendingBalances returns a map of all valid PTickers and their associated
// balances for the given address. If the address is not in the database,
// the map will contain 0 for all valid PTickers. This works on the pending tx
func (Pegnet) selectBalances(q QueryAble, adr *factom.FAAddress) (map[fat2.PTicker]uint64, error) {
	balanceMap := make(map[fat2.PTicker]uint64, int(fat2.PTickerMax))
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		balanceMap[i] = 0
	}
	// Can't make pointers of map elements, so a temporary array must be used
	balances := make([]uint64, int(fat2.PTickerMax))
	var id int
	var address []byte
	query := fmt.Sprintf(`SELECT %s FROM pn_addresses WHERE address = ?;`, addressSelectCols)
	err := q.QueryRow(query, adr[:]).Scan(
		&id,
		&address,
		&balances[fat2.PTickerPEG],
		&balances[fat2.PTickerUSD],
		&balances[fat2.PTickerEUR],
		&balances[fat2.PTickerJPY],
		&balances[fat2.PTickerGBP],
		&balances[fat2.PTickerCAD],
		&balances[fat2.PTickerCHF],
		&balances[fat2.PTickerINR],
		&balances[fat2.PTickerSGD],
		&balances[fat2.PTickerCNY],
		&balances[fat2.PTickerHKD],
		&balances[fat2.PTickerKRW],
		&balances[fat2.PTickerBRL],
		&balances[fat2.PTickerPHP],
		&balances[fat2.PTickerMXN],
		&balances[fat2.PTickerXAU],
		&balances[fat2.PTickerXAG],
		&balances[fat2.PTickerXBT],
		&balances[fat2.PTickerETH],
		&balances[fat2.PTickerLTC],
		&balances[fat2.PTickerRVN],
		&balances[fat2.PTickerXBC],
		&balances[fat2.PTickerFCT],
		&balances[fat2.PTickerBNB],
		&balances[fat2.PTickerXLM],
		&balances[fat2.PTickerADA],
		&balances[fat2.PTickerXMR],
		&balances[fat2.PTickerDASH],
		&balances[fat2.PTickerZEC],
		&balances[fat2.PTickerDCR],
		// V4 Additions
		&balances[fat2.PTickerAUD],
		&balances[fat2.PTickerNZD],
		&balances[fat2.PTickerSEK],
		&balances[fat2.PTickerNOK],
		&balances[fat2.PTickerRUB],
		&balances[fat2.PTickerZAR],
		&balances[fat2.PTickerTRY],
		&balances[fat2.PTickerEOS],
		&balances[fat2.PTickerLINK],
		&balances[fat2.PTickerATOM],
		&balances[fat2.PTickerBAT],
		&balances[fat2.PTickerXTZ],
		// V5 Additions
		&balances[fat2.PTickerHBAR],
		&balances[fat2.PTickerNEO],
		&balances[fat2.PTickerCRO],
		&balances[fat2.PTickerETC],
		&balances[fat2.PTickerONT],
		&balances[fat2.PTickerDOGE],
		&balances[fat2.PTickerVET],
		&balances[fat2.PTickerHT],
		&balances[fat2.PTickerALGO],
		&balances[fat2.PTickerDGB],
		&balances[fat2.PTickerAED],
		&balances[fat2.PTickerARS],
		&balances[fat2.PTickerTWD],
		&balances[fat2.PTickerRWF],
		&balances[fat2.PTickerKES],
		&balances[fat2.PTickerUGX],
		&balances[fat2.PTickerTZS],
		&balances[fat2.PTickerBIF],
		&balances[fat2.PTickerETB],
		&balances[fat2.PTickerNGN],
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return balanceMap, nil
		}
		return nil, err
	}
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		balanceMap[i] = balances[i]
	}
	return balanceMap, nil
}

// SelectPendingBalances returns a map of all valid PTickers and their associated
// balances for the given address. If the address is not in the database,
// the map will contain 0 for all valid PTickers. This works on the pending tx
func (p *Pegnet) SelectAllBalances() ([]BalancesPair, error) {
	query := fmt.Sprintf(`SELECT %s FROM pn_addresses;`, addressSelectCols)
	rows, err := p.DB.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var res []BalancesPair
	for rows.Next() {
		var bp BalancesPair
		bp.Balances = make([]uint64, int(fat2.PTickerMax))

		var id int
		var address []byte
		err = rows.Scan(
			&id,
			&address,
			&bp.Balances[fat2.PTickerPEG],
			&bp.Balances[fat2.PTickerUSD],
			&bp.Balances[fat2.PTickerEUR],
			&bp.Balances[fat2.PTickerJPY],
			&bp.Balances[fat2.PTickerGBP],
			&bp.Balances[fat2.PTickerCAD],
			&bp.Balances[fat2.PTickerCHF],
			&bp.Balances[fat2.PTickerINR],
			&bp.Balances[fat2.PTickerSGD],
			&bp.Balances[fat2.PTickerCNY],
			&bp.Balances[fat2.PTickerHKD],
			&bp.Balances[fat2.PTickerKRW],
			&bp.Balances[fat2.PTickerBRL],
			&bp.Balances[fat2.PTickerPHP],
			&bp.Balances[fat2.PTickerMXN],
			&bp.Balances[fat2.PTickerXAU],
			&bp.Balances[fat2.PTickerXAG],
			&bp.Balances[fat2.PTickerXBT],
			&bp.Balances[fat2.PTickerETH],
			&bp.Balances[fat2.PTickerLTC],
			&bp.Balances[fat2.PTickerRVN],
			&bp.Balances[fat2.PTickerXBC],
			&bp.Balances[fat2.PTickerFCT],
			&bp.Balances[fat2.PTickerBNB],
			&bp.Balances[fat2.PTickerXLM],
			&bp.Balances[fat2.PTickerADA],
			&bp.Balances[fat2.PTickerXMR],
			&bp.Balances[fat2.PTickerDASH],
			&bp.Balances[fat2.PTickerZEC],
			&bp.Balances[fat2.PTickerDCR],
			// V4 Additions
			&bp.Balances[fat2.PTickerAUD],
			&bp.Balances[fat2.PTickerNZD],
			&bp.Balances[fat2.PTickerSEK],
			&bp.Balances[fat2.PTickerNOK],
			&bp.Balances[fat2.PTickerRUB],
			&bp.Balances[fat2.PTickerZAR],
			&bp.Balances[fat2.PTickerTRY],
			&bp.Balances[fat2.PTickerEOS],
			&bp.Balances[fat2.PTickerLINK],
			&bp.Balances[fat2.PTickerATOM],
			&bp.Balances[fat2.PTickerBAT],
			&bp.Balances[fat2.PTickerXTZ],
			// V5 Additions
			&bp.Balances[fat2.PTickerHBAR],
			&bp.Balances[fat2.PTickerNEO],
			&bp.Balances[fat2.PTickerCRO],
			&bp.Balances[fat2.PTickerETC],
			&bp.Balances[fat2.PTickerONT],
			&bp.Balances[fat2.PTickerDOGE],
			&bp.Balances[fat2.PTickerVET],
			&bp.Balances[fat2.PTickerHT],
			&bp.Balances[fat2.PTickerALGO],
			&bp.Balances[fat2.PTickerDGB],
			&bp.Balances[fat2.PTickerAED],
			&bp.Balances[fat2.PTickerARS],
			&bp.Balances[fat2.PTickerTWD],
			&bp.Balances[fat2.PTickerRWF],
			&bp.Balances[fat2.PTickerKES],
			&bp.Balances[fat2.PTickerUGX],
			&bp.Balances[fat2.PTickerTZS],
			&bp.Balances[fat2.PTickerBIF],
			&bp.Balances[fat2.PTickerETB],
			&bp.Balances[fat2.PTickerNGN],
		)
		if err != nil {
			return nil, err
		}

		var fa factom.FAAddress
		copy(fa[:], address)
		bp.Address = &fa

		res = append(res, bp)
	}
	return res, nil
}

func (p *Pegnet) SelectIssuances() (map[fat2.PTicker]uint64, error) {
	issuanceMap := make(map[fat2.PTicker]uint64, int(fat2.PTickerMax))
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		issuanceMap[i] = 0
	}
	// Can't make pointers of map elements, so a temporary array must be used
	issuances := make([]uint64, int(fat2.PTickerMax))
	queryFmt := `SELECT %v FROM pn_addresses`
	var sb strings.Builder
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax-1; i++ {
		tickerLower := strings.ToLower(i.String())
		sb.WriteString(fmt.Sprintf("IFNULL(SUM(%s_balance), 0), ", tickerLower))
	}
	tickerLower := strings.ToLower((fat2.PTickerMax - 1).String())
	sb.WriteString(fmt.Sprintf("IFNULL(SUM(%s_balance), 0) ", tickerLower))
	err := p.DB.QueryRow(fmt.Sprintf(queryFmt, sb.String())).Scan(
		&issuances[fat2.PTickerPEG],
		&issuances[fat2.PTickerUSD],
		&issuances[fat2.PTickerEUR],
		&issuances[fat2.PTickerJPY],
		&issuances[fat2.PTickerGBP],
		&issuances[fat2.PTickerCAD],
		&issuances[fat2.PTickerCHF],
		&issuances[fat2.PTickerINR],
		&issuances[fat2.PTickerSGD],
		&issuances[fat2.PTickerCNY],
		&issuances[fat2.PTickerHKD],
		&issuances[fat2.PTickerKRW],
		&issuances[fat2.PTickerBRL],
		&issuances[fat2.PTickerPHP],
		&issuances[fat2.PTickerMXN],
		&issuances[fat2.PTickerXAU],
		&issuances[fat2.PTickerXAG],
		&issuances[fat2.PTickerXBT],
		&issuances[fat2.PTickerETH],
		&issuances[fat2.PTickerLTC],
		&issuances[fat2.PTickerRVN],
		&issuances[fat2.PTickerXBC],
		&issuances[fat2.PTickerFCT],
		&issuances[fat2.PTickerBNB],
		&issuances[fat2.PTickerXLM],
		&issuances[fat2.PTickerADA],
		&issuances[fat2.PTickerXMR],
		&issuances[fat2.PTickerDASH],
		&issuances[fat2.PTickerZEC],
		&issuances[fat2.PTickerDCR],
		// V4 Additions
		&issuances[fat2.PTickerAUD],
		&issuances[fat2.PTickerNZD],
		&issuances[fat2.PTickerSEK],
		&issuances[fat2.PTickerNOK],
		&issuances[fat2.PTickerRUB],
		&issuances[fat2.PTickerZAR],
		&issuances[fat2.PTickerTRY],
		&issuances[fat2.PTickerEOS],
		&issuances[fat2.PTickerLINK],
		&issuances[fat2.PTickerATOM],
		&issuances[fat2.PTickerBAT],
		&issuances[fat2.PTickerXTZ],
		// V5 Additions
		&issuances[fat2.PTickerHBAR],
		&issuances[fat2.PTickerNEO],
		&issuances[fat2.PTickerCRO],
		&issuances[fat2.PTickerETC],
		&issuances[fat2.PTickerONT],
		&issuances[fat2.PTickerDOGE],
		&issuances[fat2.PTickerVET],
		&issuances[fat2.PTickerHT],
		&issuances[fat2.PTickerALGO],
		&issuances[fat2.PTickerDGB],
		&issuances[fat2.PTickerAED],
		&issuances[fat2.PTickerARS],
		&issuances[fat2.PTickerTWD],
		&issuances[fat2.PTickerRWF],
		&issuances[fat2.PTickerKES],
		&issuances[fat2.PTickerUGX],
		&issuances[fat2.PTickerTZS],
		&issuances[fat2.PTickerBIF],
		&issuances[fat2.PTickerETB],
		&issuances[fat2.PTickerNGN],
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return issuanceMap, nil
		}
		return nil, err
	}
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		issuanceMap[i] = issuances[i]
	}
	return issuanceMap, nil
}
