package jsonpath

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type rlexer struct {
	bufInput       *bufio.Reader
	input          io.Reader
	pos            Pos
	nextByte       int
	lexeme         bytes.Buffer
	initialState   stateFn
	currentStateFn stateFn
	emittedItem    *Item
	hasItem        bool
	state          interface{}
}

func NewReaderLexer(rr io.Reader, initial stateFn) *rlexer {
	l := rlexer{
		input:          rr,
		bufInput:       bufio.NewReader(rr),
		nextByte:       noValue,
		initialState:   initial,
		currentStateFn: initial,
		emittedItem:    &Item{},
	}
	return &l
}

func (l *rlexer) take() int {
	if l.nextByte == noValue {
		l.peek()
	}

	nr := l.nextByte
	l.nextByte = noValue
	l.lexeme.WriteByte(byte(nr))
	return nr
}

func (l *rlexer) peek() int {
	if l.nextByte != noValue {
		return l.nextByte
	}

	r, err := l.bufInput.ReadByte()
	if err == io.EOF {
		l.nextByte = eof
		return eof
	}

	l.nextByte = int(r)
	return l.nextByte
}

func (l *rlexer) emit(t int) {
	l.setItem(t, l.pos, l.lexeme.Bytes())

	l.pos += Pos(l.lexeme.Len())
	l.hasItem = true
}

func (l *rlexer) setItem(typ int, pos Pos, val []byte) {
	l.emittedItem.typ = typ
	l.emittedItem.pos = pos
	l.emittedItem.val = val
}

func (l *rlexer) ignore() {
	l.pos += Pos(l.lexeme.Len())
	l.lexeme.Truncate(0)
}

func (l *rlexer) next() (*Item, bool) {
	l.lexeme.Truncate(0)
	for {
		if l.currentStateFn == nil {
			return l.emittedItem, false
		}

		l.currentStateFn = l.currentStateFn(l, l.state)

		if l.hasItem {
			l.hasItem = false
			return l.emittedItem, true
		}
	}
	return l.emittedItem, false
}

func (l *rlexer) setState(val interface{}) {
	l.state = val
}

func (l *rlexer) errorf(format string, args ...interface{}) stateFn {
	l.setItem(lexError, l.pos, []byte(fmt.Sprintf(format, args...)))
	l.lexeme.Truncate(0)
	l.hasItem = true
	return nil
}

func (l *rlexer) reset() {
	l.bufInput.Reset(l.input)
	l.lexeme.Truncate(0)
	l.nextByte = noValue
	l.pos = 0
	l.hasItem = false
	l.currentStateFn = l.initialState
	l.state = nil
}
