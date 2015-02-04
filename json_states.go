package jsonpath

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
func lexJsonRoot(l lexer, state interface{}) stateFn {
	l.setState(newStack())
	ignoreSpaceRun(l)
	cur := l.peek()
	var next stateFn
	switch cur {
	case '{':
		next = stateJsonObjectOpen
	case '[':
		next = stateJsonArrayOpen
	default:
		next = l.errorf("Expected '{' or '[' at root of JSON instead of '%c' %#U", cur, cur)
	}
	return next
}

func stateJsonObjectOpen(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	if cur != '{' {
		return l.errorf("Expected '{' as start of object instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonBraceLeft)
	state.(stack).push(jsonBraceLeft)

	return stateJsonObject
}

func stateJsonArrayOpen(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	if cur != '[' {
		return l.errorf("Expected '[' as start of array instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonBracketLeft)
	state.(stack).push(jsonBracketLeft)

	return stateJsonArray
}

func stateJsonObject(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	var next stateFn
	cur := l.peek()
	switch cur {
	case '}':
		if top := state.(stack).peek(); top != nil && top.(int) != jsonBraceLeft {
			next = l.errorf("Received %#U but has no matching '{'", cur)
			break
		}
		l.take()
		l.emit(jsonBraceRight)
		state.(stack).pop()
		next = stateJsonAfterValue
	default:
		next = stateJsonKey
	}
	return next
}

func stateJsonArray(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	var next stateFn
	cur := l.peek()
	switch cur {
	case ']':
		if top := state.(stack).peek(); top != nil && top.(int) != jsonBracketLeft {
			next = l.errorf("Received %#U but has no matching '['", cur)
			break
		}
		l.take()
		l.emit(jsonBracketRight)
		state.(stack).pop()
		next = stateJsonAfterValue
	default:
		next = stateJsonValue
	}
	return next
}

func stateJsonAfterValue(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	top := state.(stack).peek()
	empty := noValue
	val := empty
	if top != nil {
		val = top.(int)
	}

	switch cur {
	case ',':
		l.take()
		l.emit(jsonComma)
		switch int(val) {
		case jsonBraceLeft:
			return stateJsonKey
		case jsonBracketLeft:
			return stateJsonValue
		default:
			return l.errorf("Unexpected character in lexer stack: '%c' %#U", cur, cur)
		}
	case '}':
		l.take()
		l.emit(jsonBraceRight)
		state.(stack).pop()
		switch int(val) {
		case jsonBraceLeft:
			return stateJsonAfterValue
		case jsonBracketLeft:
			return l.errorf("Unexpected } in array")
		case empty:
			return nil
		}
	case ']':
		l.take()
		l.emit(jsonBracketRight)
		state.(stack).pop()
		switch int(val) {
		case jsonBraceLeft:
			return l.errorf("Unexpected ] in object")
		case jsonBracketLeft:
			return stateJsonAfterValue
		case empty:
			return nil
		}
	case eof:
		if state.(stack).len() == 0 {
			return nil
		} else {
			return l.errorf("Unexpected EOF instead of value")
		}
	default:
		return l.errorf("Unexpected character after json value token: '%c' %#U", cur, cur)
	}
	return nil
}

func stateJsonKey(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	if err := takeString(l); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonKey)

	return stateJsonColon
}

func stateJsonColon(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)

	cur := l.peek()
	if cur != ':' {
		return l.errorf("Expected ':' after key string instead of '%c' %#U", cur, cur)
	}
	l.take()
	l.emit(jsonColon)

	return stateJsonValue
}

func stateJsonValue(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()

	switch cur {
	case eof:
		return l.errorf("Unexpected EOF instead of value")
	case '"':
		return stateJsonString
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return stateJsonNumber
	case 't', 'f':
		return stateJsonBool
	case 'n':
		return stateJsonNull
	case '{':
		return stateJsonObjectOpen
	case '[':
		return stateJsonArrayOpen
	default:
		return l.errorf("Unexpected character as start of value: '%c' %#U", cur, cur)
	}
}

func stateJsonString(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	if err := takeString(l); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonString)
	return stateJsonAfterValue
}

func stateJsonNumber(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	if err := takeNumeric(l); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonNumber)
	return stateJsonAfterValue
}

func stateJsonBool(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	if !takeExactSequence(l, "true") {
		if !takeExactSequence(l, "false") {
			return l.errorf("Could not parse value as 'true' or 'false'")
		}
	}
	l.emit(jsonBool)
	return stateJsonAfterValue
}

func stateJsonNull(l lexer, state interface{}) stateFn {
	ignoreSpaceRun(l)
	if !takeExactSequence(l, "null") {
		return l.errorf("Could not parse value as 'null'")
	}
	l.emit(jsonNull)
	return stateJsonAfterValue
}
