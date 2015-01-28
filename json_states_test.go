package jsonpath

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var jsonTests = []lexTest{
	{"empty object", `{}`, []*Item{i(jsonBraceLeft), i(jsonBraceRight)}},
	{"key string", `{"key" :"value"}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonString), i(jsonBraceRight)}},
	{"multiple pairs", `{"key" :"value","key2" :"value"}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonString), i(jsonComma), i(jsonKey), i(jsonColon), i(jsonString), i(jsonBraceRight)}},
	{"key number", `{"key" : 12.34e+56}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonNumber), i(jsonBraceRight)}},
	{"key true", `{"key" :true}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBool), i(jsonBraceRight)}},
	{"key false", `{"key" :false}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBool), i(jsonBraceRight)}},
	{"key null", `{"key" :null}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonNull), i(jsonBraceRight)}},
	{"key arrayOf number", `{"key" :[23]}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonNumber), i(jsonBracketRight), i(jsonBraceRight)}},
	{"key array", `{"key" :[23,"45",67]}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonNumber), i(jsonComma), i(jsonString), i(jsonComma), i(jsonNumber), i(jsonBracketRight), i(jsonBraceRight)}},
	{"key array", `{"key" :["45",{}]}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonString), i(jsonComma), i(jsonBraceLeft), i(jsonBraceRight), i(jsonBracketRight), i(jsonBraceRight)}},
	{"key nestedObject", `{"key" :{"innerkey":"value"}}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonString), i(jsonBraceRight), i(jsonBraceRight)}},
}

func TestValidJson(t *testing.T) {
	as := assert.New(t)

	for _, test := range jsonTests {
		reader := strings.NewReader(test.input)
		lexer := NewLexer(reader, 0)
		go lexer.Run(JSON)
		items := funnelToArray(lexer.items)

		as.True(typeIsEqual(items, test.items, true), "Testing of %s: got\n\t%+v\nexpected\n\t%v", test.name, itemsDescription(items, jsonTokenNames), itemsDescription(test.items, jsonTokenNames))
	}
}

var errorJsonTests = []lexTest{
	{"Missing end brace", `{`, []*Item{i(jsonBraceLeft), i(jsonError)}},
	{"Missing start brace", `}`, []*Item{i(jsonError)}},
	{"Missing key start quote", `{key":true}`, []*Item{i(jsonBraceLeft), i(jsonError)}},
	{"Missing key end quote", `{"key:true}`, []*Item{i(jsonBraceLeft), i(jsonError)}},
	{"Missing colon", `{"key"true}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonError)}},
	{"Missing value", `{"key":}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonError)}},
	{"Missing string start quote", `{"key":test"}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonError)}},
	{"Missing embedded array bracket", `{"key":[}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonError)}},
	{"Missing values in array", `{"key":[,]`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonError)}},
	{"Missing value after comma", `{"key":[343,]}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonNumber), i(jsonComma), i(jsonError)}},
	{"Missing comma in array", `{"key":[234 424]}`, []*Item{i(jsonBraceLeft), i(jsonKey), i(jsonColon), i(jsonBracketLeft), i(jsonNumber), i(jsonError)}},
}

func TestMalformedJson(t *testing.T) {
	as := assert.New(t)

	for _, test := range errorJsonTests {
		reader := strings.NewReader(test.input)
		lexer := NewLexer(reader, 0)
		go lexer.Run(JSON)
		items := funnelToArray(lexer.items)

		as.True(typeIsEqual(items, test.items, false), "Testing of %s: got\n\t%+v\nexpected\n\t%v", test.name, itemsDescription(items, jsonTokenNames), itemsDescription(test.items, jsonTokenNames))
	}
}

func TestEarlyTerminationForJSON(t *testing.T) {
	as := assert.New(t)
	wg := sync.WaitGroup{}
	bufferSize := 3

	reader := strings.NewReader(`{"key":"value", "key2":{"ikey":3}, "key3":[1,2,3,4]}`)
	lexer := NewLexer(reader, bufferSize)
	wg.Add(1)
	go func() {
		lexer.Run(JSON)
		wg.Done()
	}()

	// Pop a few items
	<-lexer.items
	<-lexer.items
	// Kill command
	close(lexer.kill)

	wg.Wait()
	remainingItems := funnelToArray(lexer.items)
	as.True(len(remainingItems) <= bufferSize, "Count of remaining items should be less than buffer size: %d", len(remainingItems))
}

var jsonExamples = []string{
	`{"items":[
	  {
	    "name": "example document for wicked fast parsing of huge json docs",
	    "integer": 123,
	    "totally sweet scientific notation": -123.123e-2,
	    "unicode? you betcha!": "ú™£¢∞§\u2665",
	    "zero character": "0",
	    "null is boring": null
	  },
	  {
	    "name": "another object",
	    "cooler than first object?": true,
	    "nested object": {
	      "nested object?": true,
	      "is nested array the same combination i have on my luggage?": true,
	      "nested array": [1,2,3,4,5]
	    },
	    "false": false
	  }
]}`,
	`{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "Product set",
    "type": "array",
    "items": {
        "title": "Product",
        "type": "object",
        "properties": {
            "id": {
                "description": "The unique identifier for a product",
                "type": "number"
            },
            "name": {
                "type": "string"
            },
            "price": {
                "type": "number",
                "minimum": 0,
                "exclusiveMinimum": true
            },
            "tags": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "minItems": 1,
                "uniqueItems": true
            },
            "dimensions": {
                "type": "object",
                "properties": {
                    "length": {"type": "number"},
                    "width": {"type": "number"},
                    "height": {"type": "number"}
                },
                "required": ["length", "width", "height"]
            },
            "warehouseLocation": {
                "description": "Coordinates of the warehouse with the product",
                "$ref": "http://json-schema.org/geo"
            }
        },
        "required": ["id", "name", "price"]
    }
   }`,
}

func TestLargeJSON(t *testing.T) {
	as := assert.New(t)
	for _, json := range jsonExamples {
		reader := strings.NewReader(json)
		lexer := NewLexer(reader, 0)
		go lexer.Run(JSON)

		for i := range lexer.items {
			as.False(i.typ == jsonError, "Found error while parsing: %q", i.val)
		}
	}
}
