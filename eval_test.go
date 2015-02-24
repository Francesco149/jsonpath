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
	test{`key selection`, `{"aKey":32}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`32`)}}},
	test{`nested key selection`, `{"aKey":{"bKey":32}}`, `$.aKey+`, []Result{Result{b("aKey"), b(`{"bKey":32}`)}}},
	test{`empty array`, `{"aKey":[]}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`[]`)}}},
	test{`multiple same-level keys, weird spacing`, `{    "aKey" 	: true ,    "bKey":  [	1 , 2	], "cKey" 	: true		} `, `$.bKey+`, []Result{Result{b(`bKey`), b(`[1,2]`)}}},

	test{`array index selection`, `{"aKey":[123,456]}`, `$.aKey[1]+`, []Result{Result{b("aKey"), 1, b(`456`)}}},
	test{`array wild index selection`, `{"aKey":[123,456]}`, `$.aKey[*]+`, []Result{Result{b("aKey"), 0, b(`123`)}, Result{b("aKey"), 1, b(`456`)}}},
	test{`array range index selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:3]+`, []Result{Result{b("aKey"), 1, b(`22`)}, Result{b("aKey"), 2, b(`33`)}}},
	test{`array range (no index) selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:1]+`, []Result{}},
	test{`array range (no upper bound) selection`, `{"aKey":[11,22,33]}`, `$.aKey[1:]+`, []Result{Result{b("aKey"), 1, b(`22`)}, Result{b("aKey"), 2, b(`33`)}}},

	test{`empty array - try selection`, `{"aKey":[]}`, `$.aKey[1]+`, []Result{}},
	test{`empty object`, `{"aKey":{}}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`{}`)}}},
	test{`object w/ height=2`, `{"aKey":{"bKey":32}}`, `$.aKey.bKey+`, []Result{Result{b("aKey"), b("bKey"), b(`32`)}}},
	test{`array of multiple types`, `{"aKey":[1,{"s":true},"asdf"]}`, `$.aKey[1]+`, []Result{Result{b("aKey"), 1, b(`{"s":true}`)}}},
	test{`nested array selection`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey+`, []Result{Result{b("aKey"), b("bKey"), b(`[123,456]`)}}},
	test{`nested array`, `[[[[[]], [true, false, []]]]]`, `$[0][0][1][2]+`, []Result{Result{0, 0, 1, 2, b(`[]`)}}},
	test{`index of array selection`, `{"aKey":{"bKey":[123, 456, 789]}}`, `$.aKey.bKey[1]+`, []Result{Result{b("aKey"), b("bKey"), 1, b(`456`)}}},
	test{`index of array selection (more than one)`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]+`, []Result{Result{b("aKey"), b("bKey"), 1, b(`456`)}}},
	test{`multi-level object/array`, `{"1Key":{"aKey": null, "bKey":{"trash":[1,2]}, "cKey":[123,456] }, "2Key":false}`, `$.1Key.bKey.trash[0]+`, []Result{Result{b("1Key"), b("bKey"), b("trash"), 0, b(`1`)}}},
	test{`multi-level array`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*].michael[1]+`, []Result{Result{b("aKey"), 3, b("michael"), 1, b(`6`)}}},
	test{`multi-level array 2`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*][1]+`, []Result{Result{b("aKey"), 4, 1, b(`"3"`)}}},
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

func b(v string) []byte {
	return []byte(v)
}

func toInterfaceArray(ch <-chan Result) []Result {
	vals := make([]Result, 0)
	for l := range ch {
		vals = append(vals, l)
	}
	return vals
}
