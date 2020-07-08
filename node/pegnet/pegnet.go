package pegnet

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pegnet/pegnetd/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Pegnet struct {
	Config *viper.Viper

	// This is the sqlite db to store state
	DB *sql.DB
}

func New(conf *viper.Viper) *Pegnet {
	p := new(Pegnet)
	p.Config = conf
	return p
}

func (p *Pegnet) Init() error {
	// The path should contain a $HOME env variable.
	rawpath := p.Config.GetString(config.SqliteDBPath)
	if runtime.GOOS == "windows" {
		rawpath = strings.Replace(rawpath, "$HOME", "$USERPROFILE", -1)
	}
	path := os.ExpandEnv(rawpath)
	// TODO: Come up with actual migrations.
	// 		until then, we can just bump this version number
	//		and make the database reset when we need to.
	path += ".v4"

	// Ensure the path exists
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	openmode := path
	modes := ""
	if p.Config.GetBool(config.SQLDBWalMode) {
		modes += "_journal=WAL&"
	}
	modes += p.Config.GetString(config.CustomSQLDBMode)
	if modes != "" {
		openmode += "?" + modes
	}

	log.Infof("Opening database from '%s'", path)
	db, err := sql.Open("sqlite3", openmode)
	if err != nil {
		return err
	}
	p.DB = db
	err = p.createTables()
	if err != nil {
		return err
	}
	return nil
}

func (p *Pegnet) createTables() error {
	for _, sql := range []string{
		createTableAddresses,
		createTableGrade,
		createTableRate,
		createTableMetadata,
		createTableWinners,
		createTableTransactions,
		createTableTransactionBatchHolding,
		createTableTxHistoryBatch,
		createTableTxHistoryTx,
		createTableTxHistoryLookup,
		createTableSyncVersion,
		createTableBank,
	} {
		if _, err := p.DB.Exec(sql); err != nil {
			return fmt.Errorf("createTables: %v", err)
		}
	}

	err := p.migrations()
	if err != nil {
		return fmt.Errorf("migrations: %v", err)
	}
	return nil
}

// migrations runs after create tables. It is not versioned like standard
// migrations.
func (p *Pegnet) migrations() error {
	// migrate pn_history_lookup alter column
	txhistoryMigrateLookup1(p)

	v4Migrate, err := p.v4MigrationNeeded()
	if err != nil {
		return err
	}
	if v4Migrate {
		log.Infof("Running v4 migrations")
		for _, sql := range []string{
			addressTableV4Migration,
		} {
			if _, err := p.DB.Exec(sql); err != nil {
				return err
			}
		}
	}

	return nil
}

// QueryAble is so we can swap db and tx interactions
type QueryAble interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
