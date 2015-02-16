package jsonpath

import (
	"fmt"
)

func takeExponent(l lexer) error {
	r := l.peek()
	if r != 'e' && r != 'E' {
		return nil
	}
	l.take()
	r = l.peek()
	switch r {
	case '+', '-':
		l.take()
		if p := l.peek(); !isDigit(p) {
			return fmt.Errorf("Expected digit after numeric sign instead of %#U", p, p)
		}
		takeDigits(l)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		takeDigits(l)
	default:
		return fmt.Errorf("Expected digit after 'e' instead of %#U", r, r)
	}
	return nil
}

func takeNumeric(l lexer) error {
	// TODO: Handle digit 0 separately
	cur := l.peek()
	switch cur {
	case '-':
		l.take()
		cur = l.peek()
		if !isDigit(cur) {
			return fmt.Errorf("Expected digit after dash instead of %#U", cur, cur)
		}
		takeDigits(l)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		l.take()
		takeDigits(l)
	default:
		return fmt.Errorf("Expected digit or dash instead of %#U", cur, cur)
	}

	// fraction or exponent
	cur = l.peek()
	switch cur {
	case '.':
		l.take()
		cur = l.peek()
		if !isDigit(cur) {
			return fmt.Errorf("Expected digit after '.' instead of %#U", cur, cur)
		}
		takeDigits(l)
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

func takeString(l lexer, includeQuotes bool) error {
	ignoreSpaceRun(l)
	cur := l.take()
	if cur != '"' {
		return fmt.Errorf("Expected \" as start of string instead of %#U", cur, cur)
	}

	if !includeQuotes {
		l.ignore()
	}

	var previous int
	for {
		cur = l.peek()
		if cur == eof {
			return fmt.Errorf("Unexpected EOF in string")
		} else if cur == '"' && (previous == noValue || previous != '\\') {
			if includeQuotes {
				l.take()
			} else {
				// handling function must catch "
			}
			break
		} else {
			l.take()
		}

		previous = cur
	}
	return nil
}

func takeDigits(l lexer) {
	for isDigit(l.peek()) {
		l.take()
	}
}

func ignoreSpaceRun(l lexer) {
	for isSpace(l.peek()) {
		l.take()
	}
	l.ignore()
}

func takeExactSequence(l lexer, str []byte) bool {
	for _, r := range str {
		if v := l.peek(); v == int(r) {
			l.take()
		} else {
			return false
		}
	}
	return true
}

func isSpace(r int) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isDigit(d int) bool {
	return d >= '0' && d <= '9'
}

func readerToArray(tr tokenReader) []Item {
	vals := make([]Item, 0)
	for {
		i, ok := tr.next()
		if !ok {
			break
		}
		v := *i
		s := make([]byte, len(v.val))
		copy(s, v.val)
		v.val = s
		vals = append(vals, v)
	}
	return vals
}

func isNumericStart(r int) bool {
	return r == '-' || isDigit(r)
}
