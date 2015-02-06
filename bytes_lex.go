package jsonpath

import (
	"fmt"
)

type blexer struct {
	lex
	input []byte // the []byte being scanned.
	start Pos    // start position of this Item.
	pos   Pos    // current position in the input
	width Pos    // width of last rune read from input
}

func NewBytesLexer(input []byte, initial stateFn) *blexer {
	l := &blexer{
		lex:   newLex(initial),
		input: input,
	}
	return l
}

func (l *blexer) take() int {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r := int(l.input[l.pos])
	l.width = 1
	l.pos += l.width
	return r
}

func (l *blexer) peek() int {
	if int(l.pos) >= len(l.input) {
		return eof
	}
	return int(l.input[l.pos])
}

func (l *blexer) emit(t int) {
	l.setItem(t, l.start, l.input[l.start:l.pos])
	l.start = l.pos
	l.hasItem = true
}

func (l *blexer) setItem(typ int, pos Pos, val []byte) {
	l.item.typ = typ
	l.item.pos = pos
	l.item.val = val
}

func (l *blexer) ignore() {
	l.start = l.pos
}

func (l *blexer) next() (*Item, bool) {
	for {
		if l.currentStateFn == nil {
			return &l.item, false
		}

		l.currentStateFn = l.currentStateFn(l, &l.stack)

		if l.hasItem {
			l.hasItem = false
			return &l.item, true
		}
	}
	return &l.item, false
}

func (l *blexer) errorf(format string, args ...interface{}) stateFn {
	l.setItem(lexError, l.start, []byte(fmt.Sprintf(format, args...)))
	l.start = l.pos
	l.hasItem = true
	return nil
}

func (l *blexer) reset() {
	l.start = 0
	l.pos = 0
	l.width = 0
	l.lex = newLex(l.initialState)
}
