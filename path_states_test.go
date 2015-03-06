package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var pathTests = []lexTest{
	{"simple", `$.akey`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathEOF)}},
	{"simple w/ value", `$.akey+`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathValue), i(pathEOF)}},
	{"simple 2", `$.akey.akey2`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathEOF)}},
	{"simple 3", `$.akey.akey2.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathEOF)}},
	{"quoted keys", `$.akey["akey2"].akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathBracketLeft), i(pathKey), i(pathBracketRight), i(pathPeriod), i(pathKey), i(pathEOF)}},
	{"wildcard key", `$.akey.*.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathWildcard), i(pathPeriod), i(pathKey), i(pathEOF)}},
	{"wildcard index", `$.akey[*]`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathBracketLeft), i(pathWildcard), i(pathBracketRight), i(pathEOF)}},
	{"bracket notation", `$["aKey"][*][32][23:42]`, []Item{i(pathRoot), i(pathBracketLeft), i(pathKey), i(pathBracketRight), i(pathBracketLeft), i(pathWildcard), i(pathBracketRight), i(pathBracketLeft), i(pathIndex), i(pathBracketRight), i(pathBracketLeft), i(pathIndex), i(pathIndexRange), i(pathIndex), i(pathBracketRight), i(pathEOF)}},
}

func TestValidPaths(t *testing.T) {
	as := assert.New(t)
	for _, test := range pathTests {
		lexer := NewSliceLexer([]byte(test.input), PATH)
		items := readerToArray(lexer)

		as.True(typeIsEqual(items, test.items, true), "Testing of %s: \nactual\n\t%+v\nexpected\n\t%v", test.name, itemsDescription(items, pathTokenNames), itemsDescription(test.items, pathTokenNames))
	}
}
