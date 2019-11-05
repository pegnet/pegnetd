package cmd

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// TODO: We should really have a simple module for these

// FactoshiToFactoid converts a uint64 factoshi ammount into a fixed point
// number represented as a string
func FactoshiToFactoid(i int64) string {
	d := i / 1e8
	r := i % 1e8
	ds := fmt.Sprintf("%d", d)
	rs := fmt.Sprintf("%08d", r)
	rs = strings.TrimRight(rs, "0")
	if len(rs) > 0 {
		ds = ds + "."
	}
	return fmt.Sprintf("%s%s", ds, rs)
}

// FactoidToFactoshi is taken from the factom lib, but errors when extra decimals provided
func FactoidToFactoshi(amt string) (uint64, error) {
	valid := regexp.MustCompile(`^([0-9]+)?(\.[0-9]+)?$`)
	if !valid.MatchString(amt) {
		return 0, nil
	}

	var total uint64 = 0

	dot := regexp.MustCompile(`\.`)
	pieces := dot.Split(amt, 2)
	whole, _ := strconv.Atoi(pieces[0])
	total += uint64(whole) * 1e8

	if len(pieces) > 1 {
		if len(pieces[1]) > 8 {
			return 0, fmt.Errorf("factoids are only subdivisible up to 1e-8, trim back on the number of decimal places")
		}

		a := regexp.MustCompile(`(0*)([0-9]+)$`)

		as := a.FindStringSubmatch(pieces[1])
		part, _ := strconv.Atoi(as[0])
		power := len(as[1]) + len(as[2])
		total += uint64(part * 1e8 / int(math.Pow10(power)))
	}

	return total, nil
}
