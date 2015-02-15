package jsonpath

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type rlexer struct {
	lex
	bufInput *bufio.Reader
	input    io.Reader
	pos      Pos
	nextByte int
	lexeme   bytes.Buffer
}

func NewReaderLexer(rr io.Reader, initial stateFn) *rlexer {
	l := rlexer{
		input:    rr,
		bufInput: bufio.NewReader(rr),
		nextByte: noValue,
		lex:      newLex(initial),
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
	c := make([]byte, l.lexeme.Len())
	copy(c, l.lexeme.Bytes())
	l.setItem(t, l.pos, c)

	l.pos += Pos(l.lexeme.Len())
	l.hasItem = true
}

func (l *rlexer) setItem(typ int, pos Pos, val []byte) {
	l.item.typ = typ
	l.item.pos = pos
	l.item.val = val
}

func (l *rlexer) ignore() {
	l.pos += Pos(l.lexeme.Len())
	l.lexeme.Truncate(0)
}

func (l *rlexer) next() (*Item, bool) {
	l.lexeme.Truncate(0)
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
	l.lex = newLex(l.initialState)
}
