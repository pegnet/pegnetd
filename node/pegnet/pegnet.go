package pegnet

import (
	"container/list"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Pegnet struct {
	Config *viper.Viper

	// TODO: Make this a database
	PegnetChain *list.List

	// This is the sqlite db to store state
	DB *sql.DB
}

func New(conf *viper.Viper) *Pegnet {
	p := new(Pegnet)
	p.Config = conf
	p.PegnetChain = list.New()

	return p
}

func (p *Pegnet) Init() error {
	// The path should contain a $HOME env variable.
	// TODO: Check that works on windows....
	path := os.ExpandEnv(viper.GetString(config.SqliteDBPath))

	// Ensure the path exists
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	log.Infof("Opening database from '%s'", path)
	// TODO: Idc which sqlite to use. Change this if you want.T
	db, err := sql.Open("sqlite3", path)
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
	} {
		if _, err := p.DB.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pegnet) InsertGradedBlock(block grader.GradedBlock) {
	p.PegnetChain.PushBack(block)
}

func (p *Pegnet) FetchPreviousBlock() grader.GradedBlock {
	mark := p.PegnetChain.Back()
	if mark == nil {
		return nil
	}

	return mark.Value.(grader.GradedBlock)
}
