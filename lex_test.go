package jsonpath

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			if i1[k].typ == TOKEN_ERROR && printError {
				fmt.Println(i1[k].val)
			}
		} else if k < len(i2) {
			if i2[k].typ == TOKEN_ERROR && printError {
				fmt.Println(i2[k].val)
			}
		}

		if i1[k].typ != i2[k].typ {
			return false
		}
	}

	return true
}

func TestLexerMethods(t *testing.T) {
	as := assert.New(t)

	reader := strings.NewReader(`{"key" :"value"}`)
	lexer := NewLexer(reader, 0)

	r := lexer.take()
	as.Equal(r, '{', "First rune should match")
	r = lexer.take()
	as.Equal(r, '"', "Second rune should match")
	r = lexer.take()
	as.Equal(r, 'k', "Third rune should match")
	// Try peeking
	r = lexer.peek()
	as.Equal(r, 'e', "Peek fifth rune should match")
	// Second peek should yield same result
	r = lexer.peek()
	as.Equal(r, 'e', "Peek fifth rune should match")
	r = lexer.take()
	// Taking should yield peeked result
	as.Equal(r, 'e', "Rune should match")
	// Taking should yield next result
	r = lexer.take()
	as.Equal(r, 'y', "Rune should match")
	r = lexer.take()
	as.Equal(r, '"', "Rune should match")
	r = lexer.peek()
	as.Equal(r, ' ', "Rune should match")
	lexer.skip()

	r = lexer.peek()
	as.Equal(r, ':', "Rune should match")
}

var tests = []lexTest{
	{"empty object", `{}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_BRACE_RIGHT)}},
	{"key string", `{"key" :"value"}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_STRING), i(TOKEN_BRACE_RIGHT)}},
	{"multiple pairs", `{"key" :"value","key2" :"value"}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_STRING), i(TOKEN_COMMA), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_STRING), i(TOKEN_BRACE_RIGHT)}},
	{"key number", `{"key" : 12.34e+56}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_NUMBER), i(TOKEN_BRACE_RIGHT)}},
	{"key true", `{"key" :true}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BOOL), i(TOKEN_BRACE_RIGHT)}},
	{"key false", `{"key" :false}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BOOL), i(TOKEN_BRACE_RIGHT)}},
	{"key null", `{"key" :null}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_NULL), i(TOKEN_BRACE_RIGHT)}},
	{"key arrayOf number", `{"key" :[23]}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_NUMBER), i(TOKEN_BRACKET_RIGHT), i(TOKEN_BRACE_RIGHT)}},
	{"key array", `{"key" :[23,"45",67]}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_NUMBER), i(TOKEN_COMMA), i(TOKEN_STRING), i(TOKEN_COMMA), i(TOKEN_NUMBER), i(TOKEN_BRACKET_RIGHT), i(TOKEN_BRACE_RIGHT)}},
	{"key array", `{"key" :["45",{}]}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_STRING), i(TOKEN_COMMA), i(TOKEN_BRACE_LEFT), i(TOKEN_BRACE_RIGHT), i(TOKEN_BRACKET_RIGHT), i(TOKEN_BRACE_RIGHT)}},
	{"key nestedObject", `{"key" :{"innerkey":"value"}}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_STRING), i(TOKEN_BRACE_RIGHT), i(TOKEN_BRACE_RIGHT)}},
}

func TestValidJson(t *testing.T) {
	as := assert.New(t)

	for _, test := range tests {
		reader := strings.NewReader(test.input)
		lexer := NewLexer(reader, 0)
		go lexer.run()
		items := funnelToArray(lexer.items)

		as.True(typeIsEqual(items, test.items, true), "Testing of %s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
	}
}

var errorTests = []lexTest{
	{"Missing end brace", `{`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_ERROR)}},
	{"Missing start brace", `}`, []*Item{i(TOKEN_ERROR)}},
	{"Missing key start quote", `{key":true}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_ERROR)}},
	{"Missing key end quote", `{"key:true}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_ERROR)}},
	{"Missing colon", `{"key"true}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_ERROR)}},
	{"Missing value", `{"key":}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_ERROR)}},
	{"Missing string start quote", `{"key":test"}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_ERROR)}},
	{"Missing embedded array bracket", `{"key":[}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_ERROR)}},
	{"Missing values in array", `{"key":[,]`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_ERROR)}},
	{"Missing value after comma", `{"key":[343,]}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_NUMBER), i(TOKEN_COMMA), i(TOKEN_ERROR)}},
	{"Missing comma in array", `{"key":[234 424]}`, []*Item{i(TOKEN_BRACE_LEFT), i(TOKEN_KEY), i(TOKEN_COLON), i(TOKEN_BRACKET_LEFT), i(TOKEN_NUMBER), i(TOKEN_ERROR)}},
}

func TestMalformedJson(t *testing.T) {
	as := assert.New(t)

	for _, test := range errorTests {
		reader := strings.NewReader(test.input)
		lexer := NewLexer(reader, 0)
		go lexer.run()
		items := funnelToArray(lexer.items)

		as.True(typeIsEqual(items, test.items, false), "Testing of %s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
	}
}

func TestEarlyTermination(t *testing.T) {
	as := assert.New(t)
	wg := sync.WaitGroup{}
	bufferSize := 5

	reader := strings.NewReader(`{"key":"value", "key2":{"ikey":3}, "key3":[1,2,3,4]}`)
	lexer := NewLexer(reader, bufferSize)
	wg.Add(1)
	go func() {
		lexer.run()
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

func funnelToArray(ch chan *Item) []*Item {
	vals := make([]*Item, 0)
	for l := range ch {
		vals = append(vals, l)
	}
	return vals
}
