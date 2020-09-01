package cmd_test

import (
	"testing"

	. "github.com/pegnet/pegnetd/cmd"
)

func TestFactoidToFactoshi(t *testing.T) {
	_, err := FactoidToFactoshi(`3\9763.76826965`)
	if err == nil {
		t.Error("Should have an error")
	}
}
