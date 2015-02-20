package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

type Result []interface{}

type queryStateFn func(*query, *Item) queryStateFn

type query struct {
	path
	state  *eval
	next   queryStateFn
	last   int
	buffer bytes.Buffer
	valLoc stack
}

type key string
type index int

type eval struct {
	tr         tokenReader
	levelStack intStack
	location   stack
	queries    map[string]*query
	state      evalStateFn
	prevIndex  int
	nextKey    []byte
	copyValues bool

	Results chan Result
	Error   error
}

func newEvaluation(tr tokenReader, paths ...*path) *eval {
	e := &eval{
		tr:         tr,
		Results:    make(chan Result, 100),
		location:   *newStack(),
		levelStack: *newIntStack(),
		state:      evalRoot,
		queries:    make(map[string]*query, 0),
		prevIndex:  -1,
		nextKey:    nil,
		copyValues: true, // depends on which lexer is used
	}

	for _, p := range paths {
		e.queries[p.stringValue] = &query{
			path:   *p,
			state:  e,
			next:   pathMatchNextOp,
			last:   -1,
			buffer: *bytes.NewBuffer(make([]byte, 0, 50)),
			valLoc: *newStack(),
		}
	}

	// Determine whether to copy item values from lexer
	switch tr.(type) {
	case *rlexer:
		e.copyValues = true
	default:
		e.copyValues = false
	}

	return e
}

func (e *eval) run() {
	// f, _ := os.Create("/tmp/trace")
	// pprof.StartCPUProfile(f)
	for {
		t, ok := e.tr.next()
		if !ok || e.state == nil {
			break
		}

		// run evaluator function
		e.state = e.state(e, t)

		anyRunning := false
		// run path function for each path
		for str, query := range e.queries {
			if query.next != nil {
				anyRunning = true
				query.next = query.next(query, t)
				if query.next == nil {
					delete(e.queries, str)
				}
			}
		}

		if !anyRunning {
			break
		}

		if e.Error != nil {
			break
		}
	}

	// pprof.StopCPUProfile()

	close(e.Results)
}

type evalStateFn func(*eval, *Item) evalStateFn

func evalRoot(e *eval, i *Item) evalStateFn {
	switch i.typ {
	case jsonBraceLeft:
		e.levelStack.push(i.typ)
		return evalObjectAfterOpen
	case jsonBracketLeft:
		e.levelStack.push(i.typ)
		return evalArrayAfterOpen
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func evalObjectAfterOpen(e *eval, i *Item) evalStateFn {
	switch i.typ {
	case jsonKey:
		c := i.val[1 : len(i.val)-1]
		if e.copyValues {
			d := make([]byte, len(c))
			copy(d, c)
			c = d
		}
		e.nextKey = c
		return evalObjectColon
	case jsonBraceRight:
		return rightBraceOrBracket(e)
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func evalObjectColon(e *eval, i *Item) evalStateFn {
	switch i.typ {
	case jsonColon:
		return evalObjectValue
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}

	return nil
}

func evalObjectValue(e *eval, i *Item) evalStateFn {
	e.location.push(e.nextKey)

	switch i.typ {
	case jsonNull, jsonNumber, jsonString, jsonBool:
		// printLoc(e.location.toArray())
		// fmt.Println("Value: ", string(i.val))
		return evalObjectAfterValue
	case jsonBraceLeft:
		e.levelStack.push(i.typ)
		return evalObjectAfterOpen
	case jsonBracketLeft:
		e.levelStack.push(i.typ)
		return evalArrayAfterOpen
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func evalObjectAfterValue(e *eval, i *Item) evalStateFn {
	e.location.pop()
	switch i.typ {
	case jsonComma:
		return evalObjectAfterOpen
	case jsonBraceRight:
		return rightBraceOrBracket(e)
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func rightBraceOrBracket(e *eval) evalStateFn {
	e.levelStack.pop()

	lowerTyp, ok := e.levelStack.peek()
	if !ok {
		return evalRootEnd
	} else {
		switch lowerTyp {
		case jsonBraceLeft:
			return evalObjectAfterValue
		case jsonBracketLeft:
			return evalArrayAfterValue
		}
	}
	return nil
}

func evalArrayAfterOpen(e *eval, i *Item) evalStateFn {
	e.prevIndex = -1

	switch i.typ {
	case jsonNull, jsonNumber, jsonString, jsonBool, jsonBraceLeft, jsonBracketLeft:
		return evalArrayValue(e, i)
	case jsonBracketRight:
		setPrevIndex(e)
		return rightBraceOrBracket(e)
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func evalArrayValue(e *eval, i *Item) evalStateFn {
	e.prevIndex++
	e.location.push(e.prevIndex)

	switch i.typ {
	case jsonNull, jsonNumber, jsonString, jsonBool:
		return evalArrayAfterValue
	case jsonBraceLeft:
		e.levelStack.push(i.typ)
		return evalObjectAfterOpen
	case jsonBracketLeft:
		e.levelStack.push(i.typ)
		return evalArrayAfterOpen
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func evalArrayAfterValue(e *eval, i *Item) evalStateFn {
	switch i.typ {
	case jsonComma:
		if val, ok := e.location.pop(); ok {
			if valIndex, ok := val.(int); ok {
				e.prevIndex = valIndex
			}
		}
		return evalArrayValue
	case jsonBracketRight:
		e.location.pop()
		setPrevIndex(e)
		return rightBraceOrBracket(e)
	case jsonError:
		return evalError(e, i)
	default:
		e.Error = errors.New(UnexpectedToken)
	}
	return nil
}

func setPrevIndex(e *eval) {
	e.prevIndex = -1
	peeked, ok := e.location.peek()
	if ok {
		if peekedIndex, intOk := peeked.(int); intOk {
			e.prevIndex = peekedIndex
		}
	}
}

func evalRootEnd(e *eval, i *Item) evalStateFn {
	if i != nil {
		if i.typ == jsonError {
			evalError(e, i)
		} else {
			e.Error = errors.New(BadStructure)
		}
	}
	return nil
}

func evalError(e *eval, i *Item) evalStateFn {
	e.Error = fmt.Errorf("%s at byte index %d", string(i.val), i.pos)
	return nil
}

func print(q *query, i *Item) queryStateFn {
	printLoc(q.state.location.toArray(), i.val)
	return print
}

func pathMatchNextOp(q *query, i *Item) queryStateFn {
	if q.last > q.state.location.len()-1 {
		q.last -= 1
		return pathMatchNextOp
	}

	if q.last == q.state.location.len()-2 {
		nextOp := q.operators[q.last+1]
		current, ok := q.state.location.peek()
		if ok {
			if itemMatchOperator(current, i, nextOp) {
				// printLoc(q.state.location.toArray())
				q.last += 1
			}
		}
	}

	if q.last == len(q.operators)-1 {
		if q.captureEndValue {
			q.buffer.Write(i.val)
		}
		q.valLoc = *q.state.location.clone()
		return pathEndValue
	}

	return pathMatchNextOp
}

func pathEndValue(q *query, i *Item) queryStateFn {
	if q.state.location.len() >= q.valLoc.len() {
		if q.captureEndValue {
			q.buffer.Write(i.val)
		}
	} else {
		if q.buffer.Len() > 0 {
			val := make([]byte, q.buffer.Len())
			copy(val, q.buffer.Bytes())
			q.valLoc.push(val)
		}
		q.state.Results <- q.valLoc.toArray()

		q.valLoc = *newStack()
		q.buffer.Truncate(0)
		q.last -= 1
		return pathMatchNextOp
	}
	return pathEndValue
}

func itemMatchOperator(loc interface{}, i *Item, op *operator) bool {
	topBytes, isKey := loc.([]byte)
	topInt, isIndex := loc.(int)
	if isKey {
		switch op.typ {
		case opTypeNameWild:
			return true
		case opTypeName, opTypeNameList:
			_, found := op.keyStrings[string(topBytes)]
			return found
		}
	} else if isIndex {
		switch op.typ {
		case opTypeIndexWild:
			return true
		case opTypeIndex, opTypeIndexRange:
			return topInt >= op.indexStart && topInt <= op.indexEnd
		}
	}
	return false
}

func printLoc(s []interface{}, vals ...interface{}) {
	for _, s := range s {
		switch v := s.(type) {
		case []byte:
			fmt.Printf("%s ", string(v))
		default:
			fmt.Printf("%v ", v)
		}
	}
	for _, v := range vals {
		switch i := v.(type) {
		case []byte:
			fmt.Printf("%s ", string(i))
		default:
			fmt.Printf("%v ", i)
		}
	}
	fmt.Println("")
}

func printLevels(s []int) {
	for _, s := range s {
		fmt.Printf("%v ", jsonTokenNames[s])
	}
	fmt.Println("")
}
