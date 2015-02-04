package jsonpath

import (
	"bytes"
	"fmt"
	"io"
)

type rlexer struct {
	input          io.Reader
	nextByte       int
	lexeme         bytes.Buffer
	width          Pos
	currentStateFn stateFn
	emittedItem    Item
	hasItem        bool
	state          interface{}
}

func NewReaderLexer(rr io.Reader, initial stateFn) *rlexer {
	l := rlexer{
		input:          rr,
		nextByte:       noValue,
		currentStateFn: initial,
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

	var r [1]byte
	_, err := l.input.Read(r[:])

	if err == io.EOF {
		l.nextByte = eof
		return eof
	}

	l.nextByte = int(r[0])
	return l.nextByte
}

func (l *rlexer) emit(t int) {
	i := Item{t, l.width + 1, l.lexeme.String()}
	l.lexeme.Truncate(0)
	l.emittedItem = i
	l.hasItem = true
}

func (l *rlexer) ignore() {
	l.lexeme.Truncate(0)
}

func (l *rlexer) next() (Item, bool) {
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

func (l *rlexer) setState(val interface{}) {
	l.state = val
}

func (l *rlexer) errorf(format string, args ...interface{}) stateFn {
	i := Item{jsonError, l.width + 1, fmt.Sprintf(format, args...)}
	l.lexeme.Truncate(0)
	l.emittedItem = i
	l.hasItem = true
	return nil
}
