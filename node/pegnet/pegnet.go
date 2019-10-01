package pegnet

import (
	"container/list"
	"database/sql"
	"fmt"
	"os"
	"os/user"

	"github.com/pegnet/pegnetd/config"

	"github.com/pegnet/pegnet/modules/grader"

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
	path := os.ExpandEnv(viper.GetString(config.SqliteDBPath))
	usr, err := user.Current()
	if err != nil {
		return err
	}
	path = fmt.Sprintf("%s/%s", usr.HomeDir, path)
	// TODO: Idc which sqlite to use. Change this if you want.T
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	p.DB = db
	err = p.CreateTableAddresses()
	if err != nil {
		// TODO: implement better schema validation
		if err.Error() != "table \"pn_addresses\" already exists" {
			return err
		}
	}
	err = p.CreateTableGrade()
	if err != nil {
		return err
	}
	err = p.CreateTableRate()
	if err != nil {
		return err
	}
	err = p.CreateTableMetadata()
	if err != nil {
		return err
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
