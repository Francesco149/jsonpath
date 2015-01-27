package jsonpath

import "unicode"

func lexRoot(l *lexer) stateFn {
	if err := lexObject; err != nil {
		return err
	}
	if l.stack.Peek() != nil {
		return l.errorf("Missing end bracket or brace")
	}
	return nil
}

func lexObject(l *lexer) stateFn {
	if l.stopped {
		return l.errorf(ErrorEarlyTermination)
	}
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '{' {
		return l.errorf("Expected '{' as start of object instead of '%#U'", cur)
	}
	l.take()
	l.emit(TOKEN_BRACE_LEFT)

	l.ignoreSpaceRun()
	cur = l.peek()
	var next stateFn
	switch cur {
	case '"':
		next = takeKeyValuePairs
	case '}':
		if top := l.stack.Peek(); top != nil && top.(int) != TOKEN_BRACE_LEFT {
			next = l.errorf("Received '%#U' but has no matching '{", cur)
			break
		}
		l.take()
		l.emit(TOKEN_BRACE_RIGHT)
		next = nil
	default:
		next = l.errorf("Expected \" or } inside an object instead of '%#U'", cur)
	}
	return next
}

func hasMatchingBracket(l *lexer, t int) bool {
	top := l.stack.Peek()
	if top == nil {
		return false
	}
	openToken := top.(int)
	if t == TOKEN_BRACKET_RIGHT {
		return openToken == TOKEN_BRACKET_LEFT
	} else if t == TOKEN_BRACE_RIGHT {
		return openToken == TOKEN_BRACE_LEFT
	}
	return false
}

func takeKeyValuePairs(l *lexer) stateFn {
	for {
		if l.stopped {
			return l.errorf(ErrorEarlyTermination)
		}

		if err := takeKeyAndColon(l); err != nil {
			return err
		}
		if err := takeValue(l); err != nil {
			return err
		}
		l.ignoreSpaceRun()
		cur := l.peek()
		switch cur {
		case ',':
			l.take()
			l.emit(TOKEN_COMMA)
			continue
		case '}':
			if top := l.stack.Peek(); top != nil && top.(int) != TOKEN_BRACE_LEFT {
				return l.errorf("Received '%#U' but has no matching '{", cur)
			}
			l.take()
			l.emit(TOKEN_BRACE_RIGHT)
			return nil
		default:
			return l.errorf("Unexpected character after value: %#U", cur)
		}
	}
}

func lexArray(l *lexer) stateFn {
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '[' {
		return l.errorf("Expected '[' as start of array instead of '%#U'", cur)
	}
	l.take()
	l.emit(TOKEN_BRACKET_LEFT)
	l.ignoreSpaceRun()
	cur = l.peek()
	// var next stateFn
	switch cur {
	case ']':
		l.take()
		l.emit(TOKEN_BRACKET_RIGHT)
	default:
	valueLoop:
		for {
			if l.stopped {
				return l.errorf(ErrorEarlyTermination)
			}

			if err := takeValue(l); err != nil {
				return err
			}
			cur = l.peek()
			switch cur {
			case ',':
				l.take()
				l.emit(TOKEN_COMMA)
			case ']':
				l.take()
				l.emit(TOKEN_BRACKET_RIGHT)
				break valueLoop
			default:
				return l.errorf("Expected character after array value: '%#U'", cur)
			}
		}
	}
	return nil
}

func takeKeyAndColon(l *lexer) stateFn {
	if l.stopped {
		return l.errorf(ErrorEarlyTermination)
	}

	l.ignoreSpaceRun()
	if err := takeString(l); err != nil {
		return err
	}
	l.emit(TOKEN_KEY)
	l.ignoreSpaceRun()

	cur := l.peek()
	if cur != ':' {
		return l.errorf("Expected ':' after key string instead of '%#U'", cur)
	}
	l.take()
	l.emit(TOKEN_COLON)
	return nil
}

func takeValue(l *lexer) stateFn {
	if l.stopped {
		return l.errorf(ErrorEarlyTermination)
	}

	l.ignoreSpaceRun()
	cur := l.peek()
	var err stateFn

	switch {
	case cur == eof:
		return l.errorf("Unexpected EOF instead of value")
	case cur == '"':
		if err = takeString(l); err != nil {
			return err
		}
		l.emit(TOKEN_STRING)
	case isNumericStart(cur):
		if err := takeNumeric(l); err != nil {
			return err
		}
		l.emit(TOKEN_NUMBER)
	case cur == 't':
		if success := l.acceptString("true"); !success {
			return l.errorf("Could not parse value as 'true'")
		}
		l.emit(TOKEN_BOOL)
	case cur == 'f':
		if success := l.acceptString("false"); !success {
			return l.errorf("Could not parse value as 'false'")
		}
		l.emit(TOKEN_BOOL)
	case cur == 'n':
		if success := l.acceptString("null"); !success {
			return l.errorf("Could not parse value as 'null'")
		}
		l.emit(TOKEN_NULL)
	case cur == '{':
		for state := lexObject; state != nil; {
			state = state(l)
		}
	case cur == '[':
		for state := lexArray; state != nil; {
			state = state(l)
		}
	default:
		return l.errorf("Unexpected character as start of value: %#U", cur)
	}

	return nil
}

func takeNumeric(l *lexer) stateFn {
	// TODO: Handle digit 0 separately
	cur := l.peek()
	switch {
	case cur == '-':
		l.take()
		cur = l.peek()
		if !unicode.IsDigit(cur) {
			return l.errorf("Expected digit after dash instead of '%#U'", cur)
		}
		l.acceptWhere(unicode.IsDigit)
	case unicode.IsDigit(cur):
		l.take()
		l.acceptWhere(unicode.IsDigit)
	default:
		return l.errorf("Expected digit or dash instead of '%#U'", cur)
	}

	takeExponent := func(l *lexer) stateFn {
		r := l.peek()
		if r != 'e' && r != 'E' {
			return nil
		}
		l.take()
		r = l.peek()
		switch {
		case r == '+', r == '-':
			l.take()
			if r = l.peek(); !unicode.IsDigit(r) {
				return l.errorf("Expected digit after numeric sign instead of '%#U'", cur)
			}
			l.acceptWhere(unicode.IsDigit)
		case unicode.IsDigit(r):
			l.acceptWhere(unicode.IsDigit)
		default:
			return l.errorf("Expected digit after 'e' instead of '%#U'", cur)
		}
		return nil
	}

	// fraction or exponent
	cur = l.peek()
	switch cur {
	case '.':
		l.take()
		cur = l.peek()
		if !unicode.IsDigit(cur) {
			return l.errorf("Expected digit after '.' instead of '%#U'", cur)
		}
		l.acceptWhere(unicode.IsDigit)
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

func takeString(l *lexer) stateFn {
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '"' {
		return l.errorf("Expected \" as start of string instead of '%#U'", cur)
	}
	l.take()

	var previous *rune
	for {
		cur := l.peek()
		if cur == eof {
			return l.errorf("Unexpected EOF in string")
		} else if cur == '"' && (previous == nil || *previous != '\\') {
			l.take()
			break
		} else {
			l.take()
		}

		previous = &cur
	}
	return nil
}
