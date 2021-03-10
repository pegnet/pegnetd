package node

import "github.com/pegnet/pegnetd/fat/fat2"

// Special structure for the mint amount
type MintSupply struct {
	Ticker fat2.PTicker // token
	Amount uint64       // amount of token to mint
}

var (
	MintTotalSupplyMap = []MintSupply{
		{fat2.PTickerPEG, 334509613},
		{fat2.PTickerUSD, 3184409},
		{fat2.PTickerKRW, 118},
		{fat2.PTickerXAU, 1},
		{fat2.PTickerXAG, 599},
		{fat2.PTickerXBT, 2},
		{fat2.PTickerETH, 5476},
		{fat2.PTickerLTC, 2004},
		{fat2.PTickerRVN, 13124813},
		{fat2.PTickerXBC, 243},
		{fat2.PTickerBNB, 3461},
		{fat2.PTickerXLM, 45892},
		{fat2.PTickerADA, 1414096},
		{fat2.PTickerXMR, 682},
		{fat2.PTickerDASH, 6001},
		{fat2.PTickerZEC, 2696},
		{fat2.PTickerEOS, 2059},
		{fat2.PTickerLINK, 9110},
		{fat2.PTickerATOM, 101},
		{fat2.PTickerNEO, 2},
		{fat2.PTickerCRO, 164},
		{fat2.PTickerETC, 5},
		{fat2.PTickerVET, 22400000},
		{fat2.PTickerHT, 5},
		{fat2.PTickerDCR, 1049},
		{fat2.PTickerAUD, 9},
		{fat2.PTickerNOK, 59},
		{fat2.PTickerXTZ, 11117},
		{fat2.PTickerDOGE, 9870},
		{fat2.PTickerALGO, 457602},
		{fat2.PTickerDGB, 51175},
	}
)
