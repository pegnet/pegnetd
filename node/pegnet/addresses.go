package pegnet

import (
	"fmt"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
	"strings"
)

const createTableAddresses = `CREATE TABLE "pn_addresses" (
        "id"            INTEGER PRIMARY KEY,
        "address"       BLOB NOT NULL UNIQUE,
        "peg_balance"   INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peg_balance" >= 0),
        "pusd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pusd_balance" >= 0),
        "peur_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peur_balance" >= 0),
        "pjpy_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pjpy_balance" >= 0),
        "pgbp_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pgbp_balance" >= 0),
        "pcad_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pcad_balance" >= 0),
        "pchf_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pchf_balance" >= 0),
        "pinr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pinr_balance" >= 0),
        "psgd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("psgd_balance" >= 0),
        "pcny_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pcny_balance" >= 0),
        "phkd_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("phkd_balance" >= 0),
        "pkrw_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pkrw_balance" >= 0),
        "pbrl_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pbrl_balance" >= 0),
        "pphp_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pphp_balance" >= 0),
        "pmxn_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pmxn_balance" >= 0),
        "pxau_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxau_balance" >= 0),
        "pxag_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxag_balance" >= 0),
        "pxbt_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxbt_balance" >= 0),
        "peth_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("peth_balance" >= 0),
        "pltc_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pltc_balance" >= 0),
        "prvn_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("prvn_balance" >= 0),
        "pxbc_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxbc_balance" >= 0),
        "pfct_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pfct_balance" >= 0),
        "pbnb_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pbnb_balance" >= 0),
        "pxlm_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxlm_balance" >= 0),
        "pada_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pada_balance" >= 0),
        "pxmr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pxmr_balance" >= 0),
        "pdas_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pdas_balance" >= 0),
        "pzec_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pzec_balance" >= 0),
        "pdcr_balance"  INTEGER NOT NULL
                        CONSTRAINT "insufficient balance" CHECK ("pdcr_balance" >= 0)
);
`

func (p *Pegnet) CreateTableAddresses() error {
	_, err := p.DB.Exec(createTableAddresses)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pegnet) AddToBalance(adr *factom.FAAddress, ticker fat2.PTicker, add uint64) (int64, error) {
	// TODO: implement AddToBalance
	return 0, nil
}

func (p *Pegnet) SubFromBalance(adr *factom.FAAddress, ticker fat2.PTicker, add uint64) (int64, error) {
	// TODO: implement SubFromBalance
	return 0, nil
}

// SelectBalance returns the balance of an individual token type for the given
// address. If the address is not in the database, 0 will be returned.
func (p *Pegnet) SelectBalance(adr *factom.FAAddress, ticker fat2.PTicker) (uint64, error) {
	if ticker <= fat2.PTickerInvalid || fat2.PTickerMax <= ticker {
		return 0, fmt.Errorf("invalid token type")
	}
	stmtStringFmt := `SELECT %s_balance FROM pn_addresses WHERE address = ?;`
	tickerLower := strings.ToLower(ticker.String())
	rows, err := p.DB.Query(fmt.Sprintf(stmtStringFmt, tickerLower), adr[:])
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() == false {
		return 0, nil
	}
	var balance uint64
	err = rows.Scan(&balance)
	if err != nil {
		return 0, err
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}
	return balance, nil
}

// SelectBalances returns a map of all valid PTickers and their associated
// balances for the given address. If the address is not in the database,
// the map will contain 0 for all valid PTickers.
func (p *Pegnet) SelectBalances(adr *factom.FAAddress) (map[fat2.PTicker]uint64, error) {
	rows, err := p.DB.Query(`SELECT * FROM pn_addresses WHERE address = ?;`, adr[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balanceMap := make(map[fat2.PTicker]uint64, int(fat2.PTickerMax))
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		balanceMap[i] = 0
	}
	if rows.Next() == false {
		return balanceMap, nil
	}

	// Can't make pointers of map elements, so a temporary array must be used
	balances := make([]uint64, int(fat2.PTickerMax))
	var id int
	var address []byte
	err = rows.Scan(
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
		&balances[fat2.PTickerDAS],
		&balances[fat2.PTickerZEC],
		&balances[fat2.PTickerDCR],
	)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		balanceMap[i] = balances[i]
	}
	return balanceMap, nil
}
