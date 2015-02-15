package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

type Result []interface{}

type path struct {
	query
	state  *eval
	next   pathStateFn
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
	paths      []*path
	state      evalStateFn
	prevIndex  int
	nextKey    []byte

	Results chan Result
	Error   error
}

func newEvals(tr tokenReader, q *query) *eval {
	if len(q.operators) == 0 {
		return nil
	}

	e := &eval{
		tr:         tr,
		Results:    make(chan Result, 100),
		location:   *newStack(),
		levelStack: *newIntStack(),
		state:      evalRoot,
		paths:      []*path{},
		prevIndex:  -1,
		nextKey:    nil,
	}

	p := &path{*q, e, pathMatchNextOp, -1, bytes.Buffer{}, stack{}}

	e.paths = append(e.paths, p)
	return e
}

func (e *eval) run() {
	for {
		t, ok := e.tr.next()
		if !ok || e.state == nil {
			break
		}

		// printLevels(e.levelStack.toArray())
		// printLoc(e.location.toArray())
		// run evaluator function
		e.state = e.state(e, t)
		if e.Error != nil {
			break
		}
		// printLoc(e.location.toArray(), interface{}(t.val))
		// fmt.Println("")

		anyRunning := false
		// run path function for each path
		for _, p := range e.paths {
			if p.next != nil {
				anyRunning = true
				p.next = p.next(p, t)
			}
		}
		// TODO: Improve early termination to stop iterating over nil paths
		if !anyRunning {
			break
		}
	}

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
		e.nextKey = i.val
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
	trimmed := e.nextKey[1 : len(e.nextKey)-1]
	e.location.push(trimmed)

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

type pathStateFn func(*path, *Item) pathStateFn

func print(p *path, i *Item) pathStateFn {
	printLoc(p.state.location.toArray(), i.val)
	return print
}

func pathMatchNextOp(p *path, i *Item) pathStateFn {
	if p.last > p.state.location.len()-1 {
		p.last -= 1
		return pathMatchNextOp
	}

	if p.last == p.state.location.len()-2 {
		nextOp := p.operators[p.last+1]
		current, ok := p.state.location.peek()
		if ok {
			if itemMatchOperator(current, i, nextOp) {
				// printLoc(p.state.location.toArray())
				p.last += 1
			}
		}
	}

	if p.last == len(p.operators)-1 {
		if p.captureEndValue {
			p.buffer.Write(i.val)
		}
		p.valLoc = *p.state.location.clone()
		return pathEndValue
	}

	return pathMatchNextOp
}

func pathEndValue(p *path, i *Item) pathStateFn {
	if p.state.location.len() >= p.valLoc.len() {
		if p.captureEndValue {
			p.buffer.Write(i.val)
		}
	} else {
		if p.buffer.Len() > 0 {
			val := make([]byte, p.buffer.Len())
			copy(val, p.buffer.Bytes())
			p.valLoc.push(val)
		}
		p.state.Results <- p.valLoc.toArray()

		p.valLoc = *newStack()
		p.buffer.Truncate(0)
		p.last -= 1
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
			// trimmedName := topBytes[1 : len(topBytes)-1]
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
