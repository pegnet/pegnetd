package node

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/pegnet/pegnetd/config"

	"github.com/pegnet/pegnetd/fat/fat2"

	"github.com/pegnet/pegnetd/exit"
	"github.com/spf13/viper"
)

func TestAveragePeriod(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	exit.GlobalExitHandler.AddCancel(cancel)
	_ = ctx

	fctDat, _ := os.Create("FCT.tsv")
	defer fctDat.Close()

	// Get the config
	conf := viper.GetViper()
	conf.Set(config.Network, "MainNet")
	conf.SetConfigName("pegnetd-conf")
	conf.Set(config.SqliteDBPath, "$HOME/.pegnetd/mainnet/sql.db")
	configpath := os.ExpandEnv("$HOME/.pegnetd/pegnetd-conf.toml")
	conf.SetConfigFile(os.ExpandEnv(configpath))

	n, err := NewPegnetd(ctx, conf)
	if err != nil {
		t.Fatal(err)
	}
	// Min 206422 293471
	for i := uint32(208500); i < 210500; i++ {

		averages := n.GetPegNetRateAverages(ctx, i).(map[fat2.PTicker]uint64)
		for pAsset, v := range averages {
			fmt.Sprintf("%6s %15d", pAsset, v)
		}
		// Get the rate for FCT at the current height
		price := n.LastAveragesData[fat2.PTickerFCT][len(n.LastAveragesData[fat2.PTickerFCT])-1]
		avgPrice := averages[fat2.PTickerFCT]
		fctDat.WriteString(fmt.Sprintf("%f\t%f\n", float64(price)/100000000, float64(avgPrice)/100000000))
		_ = averages

		if i%10000 == 0 {
			fmt.Printf("%6d ", i)
			println()
			if i == 206709 {
				println()
			}
		}
	}
}
