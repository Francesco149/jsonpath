package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

type Pos uint64

const (
	TOKEN_EOF = iota
	TOKEN_ERROR
	TOKEN_BRACE_LEFT
	TOKEN_BRACE_RIGHT
	TOKEN_BRACKET_LEFT
	TOKEN_BRACKET_RIGHT
	TOKEN_COLON
	TOKEN_COMMA
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_NULL
	TOKEN_KEY
	TOKEN_BOOL
)

const (
	eof = -1

	ErrorErroneousUnicode = "Malformed UTF-8"
	ErrorEarlyTermination = "Lexer was stopped early"
)

var tokenNames = map[int]string{
	TOKEN_EOF:           "EOF",
	TOKEN_ERROR:         "ERROR",
	TOKEN_BRACE_LEFT:    "{",
	TOKEN_BRACE_RIGHT:   "}",
	TOKEN_BRACKET_LEFT:  "[",
	TOKEN_BRACKET_RIGHT: "]",
	TOKEN_COLON:         ":",
	TOKEN_COMMA:         ",",
	TOKEN_NUMBER:        "NUMBER",
	TOKEN_STRING:        "STRING",
	TOKEN_NULL:          "NULL",
	TOKEN_KEY:           "KEY",
	TOKEN_BOOL:          "BOOL",
}

type Item struct {
	typ int
	pos Pos // The starting position, in bytes, of this item in the input string.
	val string
}

func (i Item) String() string {
	name, exists := tokenNames[i.typ]
	if exists {
		return name
		//return fmt.Sprintf("%s(%s)", i.val, name)
	}
	return fmt.Sprintf("%q", i.val)
}

type lexer struct {
	input    io.RuneReader
	nextRune *rune
	lexeme   bytes.Buffer
	width    Pos // width of all items until now
	stack    *stack
	lastItem *Item
	items    chan *Item
	kill     chan struct{}
	stopped  bool
	err      error
}

func NewLexer(rr io.RuneScanner, bufferSize int) *lexer {
	l := lexer{
		items: make(chan *Item, bufferSize),
		kill:  make(chan struct{}),
		input: rr,
		stack: &stack{},
	}
	return &l
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexRoot; state != nil; {
		state = state(l)
	}
	if !l.stopped {
		close(l.items)
	}
}

func (l *lexer) take() rune {
	if l.nextRune != nil {
		nr := *l.nextRune
		l.nextRune = nil
		return nr
	}

	r, size, err := l.input.ReadRune()

	if r == 0xEF && size == 1 { // Replacement Character
		l.err = errors.New(ErrorErroneousUnicode)
		return r
	}

	if err == io.EOF {
		return eof
	}

	l.lexeme.WriteRune(r)
	return r
}

func (l *lexer) skip() {
	if l.nextRune != nil {
		nr := *l.nextRune
		l.nextRune = nil
		size := utf8.RuneLen(nr)
		l.width += Pos(size)
		l.peek()
	}
}

func (l *lexer) emit(t int) {
	if l.stopped == true {
		return
	}

	if t == TOKEN_BRACKET_LEFT || t == TOKEN_BRACE_LEFT {
		l.stack.Push(t)
	}

	if t == TOKEN_BRACKET_RIGHT || t == TOKEN_BRACE_RIGHT {
		l.stack.Pop()
	}

	i := &Item{t, l.width + 1, l.lexeme.String()}
	select {
	case l.items <- i:
		l.lastItem = i
	case <-l.kill:
		close(l.items)
		l.stopped = true
	}
	l.width += Pos(l.lexeme.Len())
	l.lexeme.Truncate(0)
}

func (l *lexer) peek() rune {
	if l.nextRune != nil {
		return *l.nextRune
	}

	r, size, err := l.input.ReadRune()
	l.nextRune = &r

	if r == 0xEF && size == 1 { // Replacement Character
		l.err = errors.New(ErrorErroneousUnicode)
		return r
	}

	if err == io.EOF {
		e := rune(-1)
		l.nextRune = &e
		return e
	}

	return r
}

func (l *lexer) acceptWhere(where func(rune) bool) {
	for where(l.peek()) {
		l.take()
	}
}

func (l *lexer) acceptString(str string) bool {
	for _, r := range str {
		if v := l.peek(); v == r {
			l.take()
		} else {
			return false
		}
	}
	return true
}

func (l *lexer) ignoreSpaceRun() {
	for isSpace(l.peek()) {
		l.skip()
	}
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	if l.stopped == true {
		return nil
	}

	i := &Item{TOKEN_ERROR, l.width + 1, fmt.Sprintf(format, args...)}
	select {
	case l.items <- i:
		close(l.items)
		l.stopped = true
		l.lastItem = i
	case <-l.kill:
		close(l.items)
		l.stopped = true
	}
	l.width += Pos(l.lexeme.Len())
	l.lexeme.Truncate(0)
	return nil
}

// isSpace reports whether r is a space character or newline.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isNumericStart(r rune) bool {
	return r == '-' || unicode.IsDigit(r)
}
