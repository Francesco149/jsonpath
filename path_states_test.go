package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var pathTests = []lexTest{
	{"simple", `$.akey`, []Item{i(pathRoot), i(pathPeriod), i(pathKey)}},
	{"simple 2", `$.akey.akey2`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey)}},
	{"simple 3", `$.akey.akey2.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathKey)}},
	{"simple 3", `$.akey.*.akey3`, []Item{i(pathRoot), i(pathPeriod), i(pathKey), i(pathPeriod), i(pathWildcard), i(pathPeriod), i(pathKey)}},
}

func TestValidPaths(t *testing.T) {
	as := assert.New(t)

	for _, test := range pathTests {
		lexer := NewBytesLexer([]byte(test.input), PATH)
		items := readerToArray(lexer)

		as.True(typeIsEqual(items, test.items, true), "Testing of %s: \nactual\n\t%+v\nexpected\n\t%v", test.name, itemsDescription(items, pathTokenNames), itemsDescription(test.items, pathTokenNames))
	}
}
