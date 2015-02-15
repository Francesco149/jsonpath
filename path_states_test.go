package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var pathTests = []lexTest{
	{"simple", `$.akey`, []Item{i(pathRoot), i(pathPeriod), i(pathKey)}},
	{"simple w/ value", `$.akey+`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathValue)}},
	{"simple 2", `$.akey.akey2`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey)}},
	{"simple 3", `$.akey.akey2.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey)}},
	{"quoted keys", `$.akey."akey2".akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey)}},
	{"wildcard key", `$.akey.*.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathWildcard), i(pathPeriod), i(pathKey)}},
	{"wildcard index", `$.akey[*]`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathBracketLeft), i(pathWildcard), i(pathBracketRight)}},
}

func TestValidPaths(t *testing.T) {
	as := assert.New(t)
	for _, test := range pathTests {
		lexer := NewBytesLexer([]byte(test.input), PATH)
		items := readerToArray(lexer)

		as.True(typeIsEqual(items, test.items, true), "Testing of %s: \nactual\n\t%+v\nexpected\n\t%v", test.name, itemsDescription(items, pathTokenNames), itemsDescription(test.items, pathTokenNames))
	}
}
