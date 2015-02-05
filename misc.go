package jsonpath

import (
	"fmt"
	"math"
)

func takeNumeric(l lexer) error {
	// TODO: Handle digit 0 separately
	cur := l.peek()
	switch {
	case cur == '-':
		l.take()
		cur = l.peek()
		if !isDigit(cur) {
			return fmt.Errorf("Expected digit after dash instead of '%c' %#U", cur, cur)
		}
		takeWhere(l, isDigit)
	case isDigit(cur):
		l.take()
		takeWhere(l, isDigit)
	default:
		return fmt.Errorf("Expected digit or dash instead of '%c' %#U", cur, cur)
	}

	takeExponent := func(l lexer) error {
		r := l.peek()
		if r != 'e' && r != 'E' {
			return nil
		}
		l.take()
		r = l.peek()
		switch {
		case r == '+', r == '-':
			l.take()
			if r = l.peek(); !isDigit(r) {
				return fmt.Errorf("Expected digit after numeric sign instead of '%c' %#U", cur, cur)
			}
			takeWhere(l, isDigit)
		case isDigit(r):
			takeWhere(l, isDigit)
		default:
			return fmt.Errorf("Expected digit after 'e' instead of '%c' %#U", cur, cur)
		}
		return nil
	}

	// fraction or exponent
	cur = l.peek()
	switch cur {
	case '.':
		l.take()
		cur = l.peek()
		if !isDigit(cur) {
			return fmt.Errorf("Expected digit after '.' instead of '%c' %#U", cur, cur)
		}
		takeWhere(l, isDigit)
		if err := takeExponent(l); err != nil {
			return err
		}
	case 'e', 'E':
		if err := takeExponent(l); err != nil {
			return err
		}
	}

	return nil
}

func takeString(l lexer) error {
	ignoreSpaceRun(l)
	cur := l.peek()
	if cur != '"' {
		return fmt.Errorf("Expected \" as start of string instead of '%c' %#U", cur, cur)
	}
	l.take()

	var previous int
	for {
		cur := l.peek()
		if cur == eof {
			return fmt.Errorf("Unexpected EOF in string")
		} else if cur == '"' && (previous == noValue || previous != '\\') {
			l.take()
			break
		} else {
			l.take()
		}

		previous = cur
	}
	return nil
}

func ignoreSpaceRun(l lexer) {
	for isSpace(l.peek()) {
		l.take()
	}
	l.ignore()
}

func takeExactSequence(l lexer, str string) bool {
	for _, r := range []byte(str) {
		if v := l.peek(); v == int(r) {
			l.take()
		} else {
			return false
		}
	}
	return true
}

func takeWhere(l lexer, where func(int) bool) {
	for where(l.peek()) {
		l.take()
	}
}

func isSpace(r int) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func readerToArray(tr tokenReader) []Item {
	vals := make([]Item, 0)
	for {
		i, ok := tr.next()
		if !ok {
			break
		}
		vals = append(vals, i)
	}
	return vals
}

func isDigit(cur int) bool {
	return (cur >= '0' && cur <= '9')
}

func isNumericStart(r int) bool {
	return r == '-' || isDigit(r)
}

// Testing
type lexTest struct {
	name  string
	input string
	items []Item
}

func i(tokenType int) Item {
	return Item{tokenType, 0, ""}
}

func typeIsEqual(i1, i2 []Item, printError bool) bool {
	for k := 0; k < int(math.Max(float64(len(i1)), float64(len(i2)))); k++ {
		if k < len(i1) {
			if i1[k].typ == jsonError && printError {
				fmt.Println(i1[k].val)
			}
		} else if k < len(i2) {
			if i2[k].typ == jsonError && printError {
				fmt.Println(i2[k].val)
			}
		}

		if k >= len(i1) || k >= len(i2) {
			return false
		}

		if i1[k].typ != i2[k].typ {
			return false
		}
	}

	return true
}
