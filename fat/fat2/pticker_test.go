package fat2_test

import (
	"testing"

	. "github.com/pegnet/pegnetd/fat/fat2"
)

func TestStringToTicker(t *testing.T) {
	if PTickerInvalid != StringToTicker("random") {
		t.Error("wrong ticker")
	}
	if PTickerFCT != StringToTicker("pFCT") {
		t.Error("wrong ticker")
	}

	if PTickerDASH != StringToTicker("pDASH") {
		t.Error("wrong ticker")
	}

	if PTickerDASH.String() != "pDASH" {
		t.Error("wrong ticker")
	}

	d, _ := StringToTicker("pFCT").MarshalJSON()
	if string(d) != `"pFCT"` {
		t.Error("Marshal function was wrong")
	}
}
