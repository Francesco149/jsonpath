package jsonpath

import (
	"bytes"
	"fmt"
	"io"
)

type Pos uint64

const (
	eof     = -1
	noValue = -2

	ErrorErroneousUnicode = "Malformed UTF-8"
	ErrorEarlyTermination = "Lexer was stopped early"
)

type Item struct {
	typ int
	pos Pos // The starting position, in bytes, of this item in the input string.
	val string
}

func itemsDescription(items []*Item, nameMap map[int]string) []string {
	vals := make([]string, len(items))

	for i, item := range items {
		var found bool
		vals[i], found = nameMap[item.typ]
		if !found {
			vals[i] = item.val
		}
	}
	return vals
}

type lexer struct {
	input    io.Reader
	nextByte int
	lexeme   bytes.Buffer
	width    Pos // width of all items until now
	stack    *intstack
	lastItem *Item
	items    chan *Item
	kill     chan struct{}
	stopped  bool
	err      error
}

func NewLexer(rr io.Reader, bufferSize int) *lexer {
	l := lexer{
		items:    make(chan *Item, bufferSize),
		kill:     make(chan struct{}),
		input:    rr,
		stack:    newIntStack(),
		nextByte: noValue,
	}
	return &l
}

type stateFn func(*lexer) stateFn

func (l *lexer) Run(initial stateFn) {
	go func() {
		for state := initial; state != nil; {
			state = state(l)
		}
		if !l.stopped {
			close(l.items)
		}
	}()
}

func (l *lexer) Kill() {
	if !l.stopped { // not a cure-all
		close(l.kill)
	}
}

func (l *lexer) take() int {
	if l.nextByte == noValue {
		l.peek()
	}

	nr := l.nextByte
	l.nextByte = noValue
	l.lexeme.WriteByte(byte(nr))
	return nr
}

func (l *lexer) skip() {
	if l.nextByte != noValue {
		l.nextByte = noValue
		l.width++
		l.peek()
	}
}

func (l *lexer) emit(t int) {
	if l.stopped == true {
		return
	}

	if t == jsonBracketLeft || t == jsonBraceLeft {
		l.stack.Push(t)
	}

	if t == jsonBracketRight || t == jsonBraceRight {
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

func (l *lexer) peek() int {
	if l.nextByte != noValue {
		return l.nextByte
	}

	var r [1]byte
	_, err := l.input.Read(r[:])

	if err == io.EOF {
		l.nextByte = eof
		return eof
	}

	l.nextByte = int(r[0])
	return l.nextByte
}

func (l *lexer) acceptWhere(where func(int) bool) {
	for where(l.peek()) {
		l.take()
	}
}

func (l *lexer) acceptString(str string) bool {
	for _, r := range []byte(str) {
		if v := l.peek(); v == int(r) {
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

	i := &Item{jsonError, l.width + 1, fmt.Sprintf(format, args...)}
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
func isSpace(r int) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}
