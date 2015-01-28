package jsonpath

import "unicode"

const (
	jsonEOF = iota
	jsonError

	jsonBraceLeft
	jsonBraceRight
	jsonBracketLeft
	jsonBracketRight
	jsonColon
	jsonComma
	jsonNumber
	jsonString
	jsonNull
	jsonKey
	jsonBool
)

var jsonTokenNames = map[int]string{
	jsonEOF:   "EOF",
	jsonError: "ERROR",

	jsonBraceLeft:    "{",
	jsonBraceRight:   "}",
	jsonBracketLeft:  "[",
	jsonBracketRight: "]",
	jsonColon:        ":",
	jsonComma:        ",",
	jsonNumber:       "NUMBER",
	jsonString:       "STRING",
	jsonNull:         "NULL",
	jsonKey:          "KEY",
	jsonBool:         "BOOL",
}

var JSON = lexJsonRoot

// TODO: Handle array at root as well as object
func lexJsonRoot(l *lexer) stateFn {
	if err := lexJsonObject; err != nil {
		return err
	}
	if l.stack.Peek() != nil {
		return l.errorf("Missing end bracket or brace")
	}
	return nil
}

func lexJsonObject(l *lexer) stateFn {
	if l.stopped {
		return l.errorf(ErrorEarlyTermination)
	}
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '{' {
		return l.errorf("Expected '{' as start of object instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonBraceLeft)

	l.ignoreSpaceRun()
	cur = l.peek()
	var next stateFn
	switch cur {
	case '"':
		next = takeKeyValuePairs
	case '}':
		if top := l.stack.Peek(); top != nil && top.(int) != jsonBraceLeft {
			next = l.errorf("Received '%#U' but has no matching '{", cur)
			break
		}
		l.take()
		l.emit(jsonBraceRight)
		next = nil
	default:
		next = l.errorf("Expected \" or } inside an object instead of '%c' %#U", cur, cur)
	}
	return next
}

func lexJsonMatchingBracket(l *lexer, t int) bool {
	top := l.stack.Peek()
	if top == nil {
		return false
	}
	openToken := top.(int)
	if t == jsonBracketRight {
		return openToken == jsonBracketLeft
	} else if t == jsonBraceRight {
		return openToken == jsonBraceLeft
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
			l.emit(jsonComma)
			continue
		case '}':
			if top := l.stack.Peek(); top != nil && top.(int) != jsonBraceLeft {
				return l.errorf("Received '%#U' but has no matching '{", cur)
			}
			l.take()
			l.emit(jsonBraceRight)
			return nil
		default:
			return l.errorf("Unexpected character after value: '%c' %#U", cur, cur)
		}
	}
}

func lexJsonArray(l *lexer) stateFn {
	l.ignoreSpaceRun()
	cur := l.peek()
	if cur != '[' {
		return l.errorf("Expected '[' as start of array instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonBracketLeft)
	l.ignoreSpaceRun()
	cur = l.peek()
	// var next stateFn
	switch cur {
	case ']':
		l.take()
		l.emit(jsonBracketRight)
	default:
	valueLoop:
		for {
			if l.stopped {
				return l.errorf(ErrorEarlyTermination)
			}

			if err := takeValue(l); err != nil {
				return err
			}

			l.ignoreSpaceRun()
			cur = l.peek()
			switch cur {
			case ',':
				l.take()
				l.emit(jsonComma)
			case ']':
				l.take()
				l.emit(jsonBracketRight)
				break valueLoop
			default:
				return l.errorf("Unexpected character after array value: '%c' %#U", cur, cur)
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
	l.emit(jsonKey)
	l.ignoreSpaceRun()

	cur := l.peek()
	if cur != ':' {
		return l.errorf("Expected ':' after key string instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonColon)
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
		l.emit(jsonString)
	case isNumericStart(cur):
		if err := takeNumeric(l); err != nil {
			return err
		}
		l.emit(jsonNumber)
	case cur == 't':
		if success := l.acceptString("true"); !success {
			return l.errorf("Could not parse value as 'true'")
		}
		l.emit(jsonBool)
	case cur == 'f':
		if success := l.acceptString("false"); !success {
			return l.errorf("Could not parse value as 'false'")
		}
		l.emit(jsonBool)
	case cur == 'n':
		if success := l.acceptString("null"); !success {
			return l.errorf("Could not parse value as 'null'")
		}
		l.emit(jsonNull)
	case cur == '{':
		for state := lexJsonObject; state != nil; {
			state = state(l)
		}
	case cur == '[':
		for state := lexJsonArray; state != nil; {
			state = state(l)
		}
	default:
		return l.errorf("Unexpected character as start of value: '%c' %#U", cur, cur)
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
			return l.errorf("Expected digit after dash instead of '%c' %#U", cur, cur)
		}
		l.acceptWhere(unicode.IsDigit)
	case unicode.IsDigit(cur):
		l.take()
		l.acceptWhere(unicode.IsDigit)
	default:
		return l.errorf("Expected digit or dash instead of '%c' %#U", cur, cur)
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
				return l.errorf("Expected digit after numeric sign instead of '%c' %#U", cur, cur)
			}
			l.acceptWhere(unicode.IsDigit)
		case unicode.IsDigit(r):
			l.acceptWhere(unicode.IsDigit)
		default:
			return l.errorf("Expected digit after 'e' instead of '%c' %#U", cur, cur)
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
			return l.errorf("Expected digit after '.' instead of '%c' %#U", cur, cur)
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
		return l.errorf("Expected \" as start of string instead of '%c' %#U", cur, cur)
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
