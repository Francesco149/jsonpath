package jsonpath

import (
	"fmt"
	"math"
	"unicode"
)

func funnelToArray(ch chan *Item) []*Item {
	vals := make([]*Item, 0)
	for l := range ch {
		vals = append(vals, l)
	}
	return vals
}

func isNumericStart(r rune) bool {
	return r == '-' || unicode.IsDigit(r)
}

// Testing
type lexTest struct {
	name  string
	input string
	items []*Item
}

func i(tokenType int) *Item {
	return &Item{tokenType, 0, ""}
}

func typeIsEqual(i1, i2 []*Item, printError bool) bool {
	for k := 0; k < int(math.Max(float64(len(i1)), float64(len(i2)))); k++ {
		if k < len(i1) {
			if i1[k].typ == jsonError && printError {
				fmt.Println(i1[k].val)
			}
		} else if k < len(i2) {
			if i2[k].typ == jsonError && printError {
				fmt.Println(i2[k].val)
			}
		}

		if k >= len(i1) || k > len(i2) {
			return false
		}

		if i1[k].typ != i2[k].typ {
			return false
		}
	}

	return true
}
