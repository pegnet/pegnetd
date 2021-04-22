package node

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/pegnet/pegnetd/node/conversions"

	"github.com/pegnet/pegnetd/config"

	"github.com/pegnet/pegnetd/fat/fat2"

	"github.com/pegnet/pegnetd/exit"
	"github.com/spf13/viper"
)

func TestAveragePeriod(t *testing.T) {

	//t.Skip("Should not be run as part of automated tests")

	ctx, cancel := context.WithCancel(context.Background())
	exit.GlobalExitHandler.AddCancel(cancel)
	_ = ctx

	// Open a file to write values that can be pulled into a spreadsheet and plotted.
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
	for i := uint32(208500); i < 209500; i++ {

		averages := n.GetPegNetRateAverages(ctx, i).(map[fat2.PTicker]uint64)
		_ = averages

		for pAsset, v := range averages {
			fmt.Sprintf("%6s %15d", pAsset, v)
		}

		// Write out a tab delineated file to plot to double check the averages against the values
		//
		// Get the rate for FCT at the current height
		ufp := func(x uint64) float64 { return float64(x) / 100000000 }
		fp := func(x int64) float64 { return float64(x) / 100000000 }
		FCTprice := n.LastAveragesData[fat2.PTickerFCT][len(n.LastAveragesData[fat2.PTickerFCT])-1]
		FCTavgPrice := averages[fat2.PTickerFCT]
		BTCprice := n.LastAveragesData[fat2.PTickerXBT][len(n.LastAveragesData[fat2.PTickerXBT])-1]
		BTCavgPrice := averages[fat2.PTickerXBT]

		AveragePeriod = 144 * 4
		AverageRequired = AveragePeriod / 2

		config.PIP10AverageActivation = i + 1
		convert1, err := conversions.Convert(i, 100000000000, FCTprice, FCTavgPrice, BTCprice, BTCavgPrice)
		if err != nil {
			t.Fatal("should not fail")
		}
		config.PIP10AverageActivation = i
		convert2, err := conversions.Convert(i, 100000000000, FCTprice, FCTavgPrice, BTCprice, BTCavgPrice)
		if err != nil {
			t.Fatal("should not fail")
		}
		fctDat.WriteString(fmt.Sprintf("%f\t%f\t%f\t%f\t%f\t%f\n",
			ufp(FCTprice), ufp(FCTavgPrice), ufp(BTCprice), ufp(BTCavgPrice), fp(convert1), fp(convert2)))

		if i%10000 == 0 {
			fmt.Printf("%6d ", i)
			println()
			if i == 206709 {
				println()
			}
		}
	}
}
