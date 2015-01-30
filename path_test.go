package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathToQuery(t *testing.T) {
	as := assert.New(t)
	items := make(chan *Item, 100)
	items <- &Item{typ: pathRoot, val: `$`}
	items <- &Item{typ: pathPeriod, val: `.`}
	items <- &Item{typ: pathKey, val: `aKey`}
	items <- &Item{typ: pathPeriod, val: `.`}
	items <- &Item{typ: pathKey, val: `bKey`}
	items <- &Item{typ: pathBracketLeft, val: `[`}
	items <- &Item{typ: pathIndex, val: `234`}
	items <- &Item{typ: pathBracketRight, val: `]`}

	close(items)

	q, err := toQuery(items)
	as.NoError(err)
	as.Equal(len(q), 3, "Key count")
	as.Equal(q[0].typ, keyTypeName)
	as.Equal(q[0].keys, map[string]struct{}{"aKey": struct{}{}})
	as.Equal(q[1].typ, keyTypeName)
	as.Equal(q[1].keys, map[string]struct{}{"bKey": struct{}{}})
	as.Equal(q[2].typ, keyTypeIndex)
	as.Equal(q[2].indexStart, 234)
	as.Equal(q[2].indexEnd, 234)
}

func TestTraverseValueTree(t *testing.T) {
	as := assert.New(t)

	items := make(chan *Item, 100)
	items <- &Item{typ: jsonBraceLeft, val: `{`}
	items <- &Item{typ: jsonKey, val: `"aKey"`}
	items <- &Item{typ: jsonColon, val: `:`}
	items <- &Item{typ: jsonBraceLeft, val: `{`}
	items <- &Item{typ: jsonKey, val: `"bKey"`}
	items <- &Item{typ: jsonColon, val: `:`}
	items <- &Item{typ: jsonBracketLeft, val: `[`}
	items <- &Item{typ: jsonNumber, val: `123`}
	items <- &Item{typ: jsonComma, val: `,`}
	items <- &Item{typ: jsonNumber, val: `456`}
	items <- &Item{typ: jsonBracketRight, val: `]`}
	items <- &Item{typ: jsonBraceRight, val: `}`}
	items <- &Item{typ: jsonBraceRight, val: `}`}

	// Should not capture these
	items <- &Item{typ: jsonNumber, val: `332`}
	items <- &Item{typ: jsonComma, val: `,`}
	items <- &Item{typ: jsonBraceLeft, val: `{`}
	close(items)

	res, err := traverseValueTree(items, true)
	as.NoError(err)
	as.Equal(`{"aKey":{"bKey":[123,456]}}`, res)
}

type test struct {
	json     string
	path     string
	expected [][]interface{}
}

var tests = []test{
	test{`{"aKey":32}}`, `$.aKey`, [][]interface{}{[]interface{}{"aKey", `32`}}},
	test{`{"aKey":{"bKey":32}}`, `$.aKey`, [][]interface{}{[]interface{}{"aKey", `{"bKey":32}`}}},
	test{`{"aKey":{"bKey":32}}`, `$.aKey.bKey`, [][]interface{}{[]interface{}{"aKey", "bKey", "32"}}},
	test{`{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey`, [][]interface{}{[]interface{}{"aKey", "bKey", `[123,456]`}}},
	test{`{"aKey":{"bKey":[123]}}`, `$.aKey.bKey[0]`, [][]interface{}{[]interface{}{"aKey", "bKey", 0, `123`}}},
	test{`{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]`, [][]interface{}{[]interface{}{"aKey", "bKey", 1, `456`}}},
	test{`{"aKey":{"bKey":{"trash":[0]}, "cKey":[123,456]}}`, `$.aKey.cKey[0]`, [][]interface{}{[]interface{}{"aKey", "cKey", 0, `123`}}},
}

func TestPathQuery(t *testing.T) {
	as := assert.New(t)

	for _, t := range tests {
		results, err := GetByString(t.json, t.path)
		as.NoError(err)
		as.Equal(t.expected, results)
	}
}
