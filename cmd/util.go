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

// FactoidToFactoshi takes a Factoid amount as a string and returns the value in
// factoids
func FactoidToFactoshi(amt string) int64 {
	valid := regexp.MustCompile(`^([0-9]+)?(\.[0-9]+)?$`)
	if !valid.MatchString(amt) {
		return -1
	}

	var total int64 = 0

	dot := regexp.MustCompile(`\.`)
	pieces := dot.Split(amt, 2)
	whole, _ := strconv.Atoi(pieces[0])
	total += int64(whole) * 1e8

	if len(pieces) > 1 {
		a := regexp.MustCompile(`(0*)([0-9]+)$`)

		as := a.FindStringSubmatch(pieces[1])
		part, _ := strconv.Atoi(as[0])
		power := len(as[1]) + len(as[2])
		total += int64(part * 1e8 / int(math.Pow10(power)))
	}

	return total
}
