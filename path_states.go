package jsonpath

import "unicode"

const (
	pathEOF = iota
	pathError

	pathRoot
	pathKey
	pathBracketLeft
	pathBracketRight
	pathIndex
	pathOr
	pathLength
	pathWildcard
	pathPeriod
)

var pathTokenNames = map[int]string{
	pathEOF:   "EOF",
	pathError: "ERROR",

	pathRoot:         "$",
	pathKey:          "KEY",
	pathBracketLeft:  "[",
	pathBracketRight: "]",
	pathIndex:        "INDEX",
	pathOr:           "|",
	pathLength:       "LENGTH",
	pathWildcard:     "*",
	pathPeriod:       ".",
}

var PATH = lexPathRoot

func lexPathRoot(l *lexer) stateFn {
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '$' {
		return l.errorf("Expected $ at start of path instead of  '%#U'", cur)
	}

	l.take()
	l.emit(pathRoot)

	return lexAfterElement
}

func lexAfterElement(l *lexer) stateFn {
	cur := l.peek()
	switch {
	case cur == '.':
		l.take()
		l.emit(pathPeriod)
		return lexKey
	case cur == '[':
		return lexPathArray
	case cur == eof:
		return nil
	default:
		return l.errorf("Unrecognized rune after path element '%#U'", cur)
	}
	return nil
}

func lexKey(l *lexer) stateFn {
	// TODO: Support globbing of keys
	inQuotes := false
	cur := l.peek()
	if cur == '"' {
		l.skip()
		inQuotes = true
	}

looper:
	for {
		cur = l.peek()
		switch {
		case cur == eof:
			break looper
		case !inQuotes && (cur == '.' || cur == '['):
			break looper
		case inQuotes && cur == '"':
			break looper
		default:
			l.take()
		}
	}
	l.emit(pathKey)

	if inQuotes {
		cur = l.peek()
		if cur != '"' {
			return l.errorf("Expected \" instead of  '%#U'", cur)
		} else {
			l.skip()
		}
	}
	return lexAfterElement
}

func lexPathArray(l *lexer) stateFn {
	cur := l.peek()
	if cur != '[' {
		return l.errorf("Expected [ at start of array instead of  '%#U'", cur)
	}
	l.take()
	l.emit(pathBracketLeft)

	// TODO: Expand supported operations
	// Currently only supports single index or wildcard (1 or all)
	cur = l.peek()
	switch {
	case isNumericStart(cur):
		l.acceptWhere(unicode.IsDigit)
		l.emit(pathIndex)
	case cur == '*':
		l.take()
		l.emit(pathWildcard)
	default:
		return l.errorf("Expected digit instead of  '%#U'", cur)
	}

	cur = l.peek()
	if cur != ']' {
		return l.errorf("Expected ] instead of  '%#U'", cur)
	}
	l.take()
	l.emit(pathBracketRight)

	return lexAfterElement
}
