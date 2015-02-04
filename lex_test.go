package jsonpath

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testLexerMethods(l lexer, as *assert.Assertions) {
	s := l.peek()
	as.Equal('{', s, "First rune should match")
	r := l.take()
	as.Equal('{', r, "First rune should match")
	r = l.take()
	as.Equal('"', r, "Second rune should match")
	r = l.take()
	as.Equal('k', r, "Third rune should match")
	// Try peeking
	r = l.peek()
	as.Equal('e', r, "Peek fifth rune should match")
	// Second peek should yield same result
	r = l.peek()
	as.Equal('e', r, "Peek fifth rune should match")
	r = l.take()
	// Taking should yield peeked result
	as.Equal('e', r, "Rune should match")
	// Taking should yield next result
	r = l.take()
	as.Equal('y', r, "Rune should match")
	r = l.take()
	as.Equal('"', r, "Rune should match")
	r = l.peek()
	as.Equal(' ', r, "Rune should match")

	l.take()
	l.ignore()

	r = l.peek()
	as.Equal(':', r, "Rune should match")
}

func TestLexerMethods(t *testing.T) {
	as := assert.New(t)
	input := `{"key" :"value"}`

	sl := NewStringLexer(input, JSON)
	testLexerMethods(sl, as)

	r := strings.NewReader(input)
	rl := NewReaderLexer(r, JSON)
	testLexerMethods(rl, as)
}

func BenchmarkStringLexerJSON(b *testing.B) {
	for n := 0; n < b.N; n++ {
		lexer := NewStringLexer(jsonExamples[2], JSON)
		for {
			_, ok := lexer.next()
			if !ok {
				break
			}
		}
	}
}

func BenchmarkReaderLexerJSON(b *testing.B) {
	for n := 0; n < b.N; n++ {
		reader := strings.NewReader(jsonExamples[2])
		lexer := NewReaderLexer(reader, JSON)
		for {
			_, ok := lexer.next()
			if !ok {
				break
			}
		}
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var data interface{}
		json.Unmarshal([]byte(jsonExamples[2]), &data)
	}
}
