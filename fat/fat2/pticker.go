package fat2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PPTicker is an internal representation of a PegNet asset type
type PTicker int

const (
	PTickerInvalid PTicker = iota
	PTickerPEG
	PTickerUSD
	PTickerEUR
	PTickerJPY
	PTickerGBP
	PTickerCAD
	PTickerCHF
	PTickerINR
	PTickerSGD
	PTickerCNY
	PTickerHKD
	PTickerKRW
	PTickerBRL
	PTickerPHP
	PTickerMXN
	PTickerXAU
	PTickerXAG
	PTickerXBT
	PTickerETH
	PTickerLTC
	PTickerRVN
	PTickerXBC
	PTickerFCT
	PTickerBNB
	PTickerXLM
	PTickerADA
	PTickerXMR
	PTickerDASH
	PTickerZEC
	PTickerDCR
	// V4 Additions
	PTickerAUD
	PTickerNZD
	PTickerSEK
	PTickerNOK
	PTickerRUB
	PTickerZAR
	PTickerTRY
	PTickerEOS
	PTickerLINK
	PTickerATOM
	PTickerBAT
	PTickerXTZ
	// V5 Additions
	PTickerHBAR
	PTickerNEO
	PTickerCRO
	PTickerETC
	PTickerONT
	PTickerDOGE
	PTickerVET
	PTickerHT
	PTickerALGO
	PTickerDGB
	PTickerAED
	PTickerARS
	PTickerTWD
	PTickerRWF
	PTickerKES
	PTickerUGX
	PTickerTZS
	PTickerBIF
	PTickerETB
	PTickerNGN
	PTickerMax
)

var validPTickerStrings = []string{
	"PEG",
	"pUSD",
	"pEUR",
	"pJPY",
	"pGBP",
	"pCAD",
	"pCHF",
	"pINR",
	"pSGD",
	"pCNY",
	"pHKD",
	"pKRW",
	"pBRL",
	"pPHP",
	"pMXN",
	"pXAU",
	"pXAG",
	"pXBT",
	"pETH",
	"pLTC",
	"pRVN",
	"pXBC",
	"pFCT",
	"pBNB",
	"pXLM",
	"pADA",
	"pXMR",
	"pDASH",
	"pZEC",
	"pDCR",
	// V4 Additions
	"pAUD",
	"pNZD",
	"pSEK",
	"pNOK",
	"pRUB",
	"pZAR",
	"pTRY",
	"pEOS",
	"pLINK",
	"pATOM",
	"pBAT",
	"pXTZ",
	// V5 Additions
	"pHBAR",
	"pNEO",
	"pCRO",
	"pETC",
	"pONT",
	"pDOGE",
	"pVET",
	"pHT",
	"pALGO",
	"pDGB",
	"pAED",
	"pARS",
	"pTWD",
	"pRWF",
	"pKES",
	"pUGX",
	"pTZS",
	"pBIF",
	"pETB",
	"pNGN",
}

var validPTickers = func() map[string]PTicker {
	pTickers := make(map[string]PTicker, len(validPTickerStrings))
	for i, str := range validPTickerStrings {
		pTickers[str] = PTicker(i + 1)
	}
	return pTickers
}()

func StringToTicker(str string) PTicker {
	return validPTickers[str]
}

// UnmarshalJSON unmarshals the bytes into a PTicker and returns an error
// if the ticker is invalid
func (t *PTicker) UnmarshalJSON(data []byte) error {
	ticker := string(data)
	if ticker[0] == '"' {
		ticker = strings.Trim(ticker, `"`)
	}
	// When unmarshalling, the bytes passed in are []byte("\"PEG\"") rather
	// than just[]byte("PEG") so we must ensure that we take the quotes into
	// account here
	if len(ticker) < 3 {
		return fmt.Errorf("invalid token type")
	}

	pTicker, ok := validPTickers[ticker]
	if !ok {
		*t = PTickerInvalid
		return fmt.Errorf("invalid token type")
	}
	*t = pTicker
	return nil
}

// MarshalJSON marshals the PTicker into the bytes that represent it in JSON
func (t PTicker) MarshalJSON() ([]byte, error) {
	if t <= PTickerInvalid || PTickerMax <= t {
		return nil, fmt.Errorf("invalid token type")
	}
	return json.Marshal(t.String())
}

// String returns the string representation of this PTicker
func (t PTicker) String() string {
	if t <= PTickerInvalid || PTickerMax <= t {
		return fmt.Errorf("invalid token type").Error()
	}
	return validPTickerStrings[int(t)-1]
}
