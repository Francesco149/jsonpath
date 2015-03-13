package jsonpath

import (
	"bytes"
)

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
