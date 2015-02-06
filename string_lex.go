package jsonpath

import (
	"fmt"
)

type slexer struct {
	input          string // the string being scanned.
	start          Pos    // start position of this Item.
	pos            Pos    // current position in the input
	width          Pos    // width of last rune read from input
	initialState   stateFn
	currentStateFn stateFn
	emittedItem    Item
	hasItem        bool
	state          interface{}
}

func NewStringLexer(input string, initial stateFn) *slexer {
	l := &slexer{
		input:          input,
		initialState:   initial,
		currentStateFn: initial,
	}
	return l
}

func (l *slexer) take() int {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r := int(l.input[l.pos])
	l.width = 1
	l.pos += l.width
	return r
}

func (l *slexer) peek() int {
	if int(l.pos) >= len(l.input) {
		return eof
	}
	return int(l.input[l.pos])
}

func (l *slexer) emit(t int) {
	i := Item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos

	l.emittedItem = i
	l.hasItem = true
}

func (l *slexer) ignore() {
	l.start = l.pos
}

func (l *slexer) next() (Item, bool) {
	for {
		if l.currentStateFn == nil {
			return Item{}, false
		}

		l.currentStateFn = l.currentStateFn(l, l.state)

		if l.hasItem {
			v := l.emittedItem
			l.hasItem = false
			return v, true
		}
	}
	return Item{}, false
}

func (l *slexer) setState(val interface{}) {
	l.state = val
}

func (l *slexer) errorf(format string, args ...interface{}) stateFn {
	i := Item{lexError, l.start, fmt.Sprintf(format, args...)}
	l.start = l.pos
	l.emittedItem = i
	l.hasItem = true
	return nil
}

func (l *slexer) reset() {
	l.start = 0
	l.pos = 0
	l.width = 0
	l.hasItem = false
	l.currentStateFn = l.initialState
	l.state = nil
}
