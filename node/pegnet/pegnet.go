package pegnet

import (
	"container/list"

	"github.com/pegnet/pegnet/modules/grader"

	"github.com/spf13/viper"
)

type Pegnet struct {
	Config *viper.Viper

	// TODO: Make this a database
	PegnetChain *list.List
}

func New(conf *viper.Viper) *Pegnet {
	p := new(Pegnet)
	p.Config = conf
	p.PegnetChain = list.New()

	return p
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
