package node

// Special structure for the mint amount
type MintSupply struct {
	Token  string  // token
	Amount float64 // amount of token to mint
}

var (
	MintTotalSupply = []MintSupply{
		{"PEG", 334509613},
		{"pUSD", 3184409},
		{"pKRW", 118},
		{"pXAU", 1},
		{"pXAG", 599},
		{"pXBT", 2},
		{"pETH", 5476},
		{"pLTC", 2004},
		{"pRVN", 13124813},
		{"pXBC", 243},
		{"pBNB", 3461},
		{"pXLM", 45892},
		{"pADA", 1414096},
		{"pXMR", 682},
		{"pDASH", 6001},
		{"pZEC", 2696},
		{"pEOS", 2059},
		{"pLINK", 9110},
		{"pATOM", 101},
		{"pNEO", 2},
		{"pCRO", 164},
		{"pETC", 5},
		{"pVET", 22400000},
		{"pHT", 5},
		{"pDCR", 1049},
		{"pAUD", 9},
		{"pNOK", 59},
		{"pXTZ", 11117},
		{"pDOGE", 9870},
		{"pALGO", 457602},
		{"pDGB", 51175},
	}
)
