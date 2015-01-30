package jsonpath

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexerMethods(t *testing.T) {
	as := assert.New(t)

	reader := strings.NewReader(`{"key" :"value"}`)
	lexer := NewLexer(reader, 0)

	r := lexer.take()
	as.Equal('{', r, "First rune should match")
	r = lexer.take()
	as.Equal('"', r, "Second rune should match")
	r = lexer.take()
	as.Equal('k', r, "Third rune should match")
	// Try peeking
	r = lexer.peek()
	as.Equal('e', r, "Peek fifth rune should match")
	// Second peek should yield same result
	r = lexer.peek()
	as.Equal('e', r, "Peek fifth rune should match")
	r = lexer.take()
	// Taking should yield peeked result
	as.Equal('e', r, "Rune should match")
	// Taking should yield next result
	r = lexer.take()
	as.Equal('y', r, "Rune should match")
	r = lexer.take()
	as.Equal('"', r, "Rune should match")
	r = lexer.peek()
	as.Equal(' ', r, "Rune should match")
	lexer.skip()

	r = lexer.peek()
	as.Equal(':', r, "Rune should match")
}
