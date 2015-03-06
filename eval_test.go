package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	name     string
	json     string
	path     string
	expected []Result
}

var tests = []test{
	test{`key selection`, `{"aKey":32}`, `$.aKey+`, []Result{newResult(`32`, `aKey`)}},
	test{`nested key selection`, `{"aKey":{"bKey":32}}`, `$.aKey+`, []Result{newResult(`{"bKey":32}`, `aKey`)}},
	test{`empty array`, `{"aKey":[]}`, `$.aKey+`, []Result{newResult(`[]`, `aKey`)}},
	test{`multiple same-level keys, weird spacing`, `{    "aKey" 	: true ,    "bKey":  [	1 , 2	], "cKey" 	: true		} `, `$.bKey+`, []Result{newResult(`[1,2]`, `bKey`)}},

	test{`array index selection`, `{"aKey":[123,456]}`, `$.aKey[1]+`, []Result{newResult(`456`, `aKey`, 1)}},
	test{`array wild index selection`, `{"aKey":[123,456]}`, `$.aKey[*]+`, []Result{newResult(`123`, `aKey`, 0), newResult(`456`, `aKey`, 1)}},
	test{`array range index selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:3]+`, []Result{newResult(`22`, `aKey`, 1), newResult(`33`, `aKey`, 2)}},
	test{`array range (no index) selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:1]+`, []Result{}},
	test{`array range (no upper bound) selection`, `{"aKey":[11,22,33]}`, `$.aKey[1:]+`, []Result{newResult(`22`, `aKey`, 1), newResult(`33`, `aKey`, 2)}},

	test{`empty array - try selection`, `{"aKey":[]}`, `$.aKey[1]+`, []Result{}},
	test{`empty object`, `{"aKey":{}}`, `$.aKey+`, []Result{newResult(`{}`, `aKey`)}},
	test{`object w/ height=2`, `{"aKey":{"bKey":32}}`, `$.aKey.bKey+`, []Result{newResult(`32`, `aKey`, `bKey`)}},
	test{`array of multiple types`, `{"aKey":[1,{"s":true},"asdf"]}`, `$.aKey[1]+`, []Result{newResult(`{"s":true}`, `aKey`, 1)}},
	test{`nested array selection`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey+`, []Result{newResult(`[123,456]`, `aKey`, `bKey`)}},
	test{`nested array`, `[[[[[]], [true, false, []]]]]`, `$[0][0][1][2]+`, []Result{newResult(`[]`, 0, 0, 1, 2)}},
	test{`index of array selection`, `{"aKey":{"bKey":[123, 456, 789]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`456`, `aKey`, `bKey`, 1)}},
	test{`index of array selection (more than one)`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`456`, `aKey`, `bKey`, 1)}},
	test{`multi-level object/array`, `{"1Key":{"aKey": null, "bKey":{"trash":[1,2]}, "cKey":[123,456] }, "2Key":false}`, `$.1Key.bKey.trash[0]+`, []Result{newResult(`1`, `1Key`, `bKey`, `trash`, 0)}},
	test{`multi-level array`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*].michael[1]+`, []Result{newResult(`6`, `aKey`, 3, `michael`, 1)}},
	test{`multi-level array 2`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*][1]+`, []Result{newResult(`"3"`, `aKey`, 4, 1)}},
}

func TestPathQuery(t *testing.T) {
	as := assert.New(t)

	for _, t := range tests {
		eval, err := GetPathsInBytes([]byte(t.json), t.path)
		as.NoError(err, "Testing: %s", t.name)
		res := toInterfaceArray(eval.Results)
		// fmt.Println("--------")
		// for _, r := range res {
		// 	PrintResult(r, true)
		// }
		// fmt.Println("--------")
		as.NoError(eval.Error)
		as.Equal(t.expected, res, "Testing of %q", t.name)
	}
}

func newResult(value string, keys ...interface{}) Result {
	keysChanged := make([]interface{}, len(keys))
	for i, k := range keys {
		switch v := k.(type) {
		case string:
			keysChanged[i] = []byte(v)
		default:
			keysChanged[i] = v
		}
	}

	return Result{
		Value: []byte(value),
		Keys:  keysChanged,
	}
}

func toInterfaceArray(ch <-chan Result) []Result {
	vals := make([]Result, 0)
	for l := range ch {
		vals = append(vals, l)
	}
	return vals
}
