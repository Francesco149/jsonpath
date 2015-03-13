package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	JsonObject = iota
	JsonArray
	JsonString
	JsonNumber
	JsonNull
	JsonBool
)

type Result struct {
	Keys  []interface{}
	Value []byte
	Type  int
}

type queryStateFn func(*query, *Item) queryStateFn

type query struct {
	path
	state  *Eval
	next   queryStateFn
	last   int
	buffer bytes.Buffer
	result *Result
	valLoc stack
}

type key string
type index int

type Eval struct {
	tr         tokenReader
	levelStack intStack
	location   stack
	queries    map[string]*query
	state      evalStateFn
	prevIndex  int
	nextKey    []byte
	copyValues bool

	resultQueue *Results
	Error       error
}

func newEvaluation(tr tokenReader, paths ...*path) *Eval {
	e := &Eval{
		tr:          tr,
		location:    *newStack(),
		levelStack:  *newIntStack(),
		state:       evalRoot,
		queries:     make(map[string]*query, 0),
		prevIndex:   -1,
		nextKey:     nil,
		copyValues:  true, // depends on which lexer is used
		resultQueue: newResults(),
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
	case *readerLexer:
		e.copyValues = true
	default:
		e.copyValues = false
	}

	return e
}

func (e *Eval) Iterate() (*Results, bool) {
	e.resultQueue.clear()

	t, ok := e.tr.next()
	if !ok || e.state == nil {
		return nil, false
	}

	// run evaluator function
	e.state = e.state(e, t)

	anyRunning := false
	// run path function for each path
	for str, query := range e.queries {
		// safety check
		if query.next != nil {
			anyRunning = true
			query.next = query.next(query, t)
			if query.next == nil {
				delete(e.queries, str)
			}

			if query.result != nil {
				e.resultQueue.push(query.result)
				query.result = nil
			}
		} else {
			delete(e.queries, str)
		}
	}

	if !anyRunning {
		return nil, false
	}

	if e.Error != nil {
		return nil, false
	}

	return e.resultQueue, true
}

func (e *Eval) Next() (*Result, bool) {
	if e.resultQueue.len() > 0 {
		return e.resultQueue.Pop(), true
	}

	for {
		if _, ok := e.Iterate(); ok {
			if e.resultQueue.len() > 0 {
				return e.resultQueue.Pop(), true
			}
		} else {
			break
		}

	}
	return nil, false
}

type evalStateFn func(*Eval, *Item) evalStateFn

func evalRoot(e *Eval, i *Item) evalStateFn {
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

func evalObjectAfterOpen(e *Eval, i *Item) evalStateFn {
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

func evalObjectColon(e *Eval, i *Item) evalStateFn {
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

func evalObjectValue(e *Eval, i *Item) evalStateFn {
	e.location.push(e.nextKey)

	switch i.typ {
	case jsonNull, jsonNumber, jsonString, jsonBool:
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

func evalObjectAfterValue(e *Eval, i *Item) evalStateFn {
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

func rightBraceOrBracket(e *Eval) evalStateFn {
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

func evalArrayAfterOpen(e *Eval, i *Item) evalStateFn {
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

func evalArrayValue(e *Eval, i *Item) evalStateFn {
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

func evalArrayAfterValue(e *Eval, i *Item) evalStateFn {
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

func setPrevIndex(e *Eval) {
	e.prevIndex = -1
	peeked, ok := e.location.peek()
	if ok {
		if peekedIndex, intOk := peeked.(int); intOk {
			e.prevIndex = peekedIndex
		}
	}
}

func evalRootEnd(e *Eval, i *Item) evalStateFn {
	if i.typ != jsonEOF {
		if i.typ == jsonError {
			evalError(e, i)
		} else {
			e.Error = errors.New(BadStructure)
		}
	}
	return nil
}

func evalError(e *Eval, i *Item) evalStateFn {
	e.Error = fmt.Errorf("%s at byte index %d", string(i.val), i.pos)
	return nil
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
		r := &Result{Keys: q.valLoc.toArray()}
		if q.buffer.Len() > 0 {
			val := make([]byte, q.buffer.Len())
			copy(val, q.buffer.Bytes())
			r.Value = val
			r.Type = determineType(val)
		}
		q.result = r

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

func determineType(val []byte) int {
	switch val[0] {
	case '{':
		return JsonObject
	case '"':
		return JsonString
	case '[':
		return JsonArray
	case 'n':
		return JsonNull
	case 't', 'b':
		return JsonBool
	default:
		return JsonNumber
	}
}

func (r *Result) Pretty(showPath bool) string {
	b := bytes.NewBufferString("")
	printed := false
	if showPath {
		for _, k := range r.Keys {
			switch v := k.(type) {
			case int:
				b.WriteString(fmt.Sprintf("%d", v))
			default:
				b.WriteString(fmt.Sprintf("%q", v))
			}
			b.WriteRune('\t')
			printed = true
		}
	} else if r.Value == nil {
		if len(r.Keys) > 0 {
			printed = true
			switch v := r.Keys[len(r.Keys)-1].(type) {
			case int:
				b.WriteString(fmt.Sprintf("%d", v))
			default:
				b.WriteString(fmt.Sprintf("%q", v))
			}
		}
	}

	if r.Value != nil {
		printed = true
		b.WriteString(fmt.Sprintf("%s", r.Value))
	}
	if printed {
		b.WriteRune('\n')
	}
	return b.String()
}
