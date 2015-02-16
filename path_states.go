package jsonpath

const (
	pathError = iota
	pathEOF

	pathRoot
	pathKey
	pathBracketLeft
	pathBracketRight
	pathIndex
	pathOr
	pathLength
	pathWildcard
	pathPeriod
	pathValue
)

var pathTokenNames = map[int]string{
	pathError: "ERROR",
	pathEOF:   "EOF",

	pathRoot:         "$",
	pathKey:          "KEY",
	pathBracketLeft:  "[",
	pathBracketRight: "]",
	pathIndex:        "INDEX",
	pathOr:           "|",
	pathLength:       "LENGTH",
	pathWildcard:     "*",
	pathPeriod:       ".",
	pathValue:        "+",
}

var PATH = lexPathRoot

func lexPathRoot(l lexer, state *intStack) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	if cur != '$' {
		return l.errorf("Expected $ at start of path instead of  %#U", cur)
	}

	l.take()
	l.emit(pathRoot)

	return lexPathAfterKey
}

func lexPathAfterKey(l lexer, state *intStack) stateFn {
	cur := l.peek()
	switch cur {
	case '.':
		l.take()
		l.emit(pathPeriod)
		return lexKey
	case '[':
		l.take()
		l.emit(pathBracketLeft)
		return lexPathArray
	case '+':
		l.take()
		l.emit(pathValue)
		return lexPathAfterValue
	case eof:
		return nil
	default:
		return l.errorf("Unrecognized rune after path element %#U", cur)
	}
	return nil
}

func lexKey(l lexer, state *intStack) stateFn {
	// TODO: Support globbing of keys
	cur := l.peek()
	switch cur {
	case '*':
		l.take()
		l.emit(pathWildcard)
		return lexPathAfterKey
	case '"':
		takeString(l, false)
		l.emit(pathKey)

		cur = l.take()
		if cur != '"' {
			return l.errorf("Expected \" after quoted key instead of %#U", cur)
		}
		l.ignore() // skip the end quote
		return lexPathAfterKey
	case eof:
		return nil
	default:
		for {
			v := l.peek()
			if v == '.' || v == '[' || v == '+' || v == eof {
				break
			}
			l.take()
		}
		l.emit(pathKey)
		return lexPathAfterKey
	}
}

func lexPathArray(l lexer, state *intStack) stateFn {
	// TODO: Expand supported operations
	// Currently only supports single index or wildcard (1 or all)
	cur := l.peek()
	switch {
	case isNumericStart(cur):
		takeDigits(l)
		l.emit(pathIndex)
	case cur == '*':
		l.take()
		l.emit(pathWildcard)
	default:
		return l.errorf("Expected digit instead of  %#U", cur)
	}

	return lexPathArrayClose
}

func lexPathArrayClose(l lexer, state *intStack) stateFn {
	cur := l.peek()
	if cur != ']' {
		return l.errorf("Expected ] instead of  %#U", cur)
	}
	l.take()
	l.emit(pathBracketRight)
	return lexPathAfterKey
}

func lexPathAfterValue(l lexer, state *intStack) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	if cur != eof {
		return l.errorf("Expected EOF instead of %#U", cur)
	}
	return nil
}
