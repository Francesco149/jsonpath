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
	test{`key-value object`, `{"aKey":32}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`32`)}}},
	// test{`object selection`, `{"aKey":{"bKey":32}}`, `$.aKey+`, []Result{Result{b("aKey"), b(`{"bKey":32}`)}}},
	// test{`empty array`, `{"aKey":[]}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`[]`)}}},
	// test{`multiple same-level keys`, `{ "aKey" : true, "bKey" : [ 1 , 2 ], "cKey" : true } `, `$.bKey+`, []Result{Result{b(`bKey`), b(`[1,2]`)}}},
	// test{`array selection`, `{"aKey":[123,456]}`, `$.aKey[1]+`, []Result{Result{b("aKey"), 1, b(`456`)}}},
	// test{`empty array - try selection`, `{"aKey":[]}`, `$.aKey[1]+`, []Result{}},
	// test{`empty object`, `{"aKey":{}}`, `$.aKey+`, []Result{Result{b(`aKey`), b(`{}`)}}},
	// test{`object w/ height=2`, `{"aKey":{"bKey":32}}`, `$.aKey.bKey+`, []Result{Result{b("aKey"), b("bKey"), b(`32`)}}},
	// test{`array of multiple types`, `{"aKey":[1,{"s":true},"asdf"]}`, `$.aKey[1]+`, []Result{Result{b("aKey"), 1, b(`{"s":true}`)}}},
	// test{`nested array selection`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey+`, []Result{Result{b("aKey"), b("bKey"), b(`[123,456]`)}}},
	// test{`nested array`, `[[[[[]], [true, false, []]]]]`, `$[0][0][0][0]+`, []Result{Result{0, 0, 0, 0, b(`[]`)}}},
	// test{`index of array selection`, `{"aKey":{"bKey":[123]}}`, `$.aKey.bKey[0]+`, []Result{Result{b("aKey"), b("bKey"), 0, b(`123`)}}},
	// test{`index of array selection (more than one)`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]+`, []Result{Result{b("aKey"), b("bKey"), 1, b(`456`)}}},
	// test{`multi-level object/array`, `{"1Key":{"aKey": null, "bKey":{"trash":[1,2]}, "cKey":[123,456] }, "2Key":false}`, `$.1Key.bKey.trash[0]+`, []Result{Result{b("1Key"), b("bKey"), b("trash"), 0, b(`1`)}}},
	// test{`multi-level array`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*].michael[1]+`, []Result{Result{b("aKey"), 3, b("michael"), 1, b(`6`)}}},
}

func b(v string) []byte {
	return []byte(v)
}

func TestPathQuery(t *testing.T) {
	as := assert.New(t)

	for _, t := range tests {
		eval, err := GetPathsInBytes([]byte(t.json), t.path)
		as.NoError(err)
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

func toInterfaceArray(ch <-chan Result) []Result {
	vals := make([]Result, 0)
	for l := range ch {
		vals = append(vals, l)
	}
	return vals
}
