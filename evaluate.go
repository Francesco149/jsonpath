package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

type evaluator struct {
	locStack *stack
	results  [][]interface{}
}

func newEvaluator() *evaluator {
	e := &evaluator{
		locStack: newStack(),
		results:  make([][]interface{}, 0),
	}
	return e
}

func (e *evaluator) perform(jsonStream <-chan *Item, keys []*key) (bool, error) {
	if len(keys) == 0 {
		val, err := traverseValueTree(jsonStream, true)
		if err != nil {
			return true, err
		}
		line := e.locStack.Clone()
		line.Push(val)
		e.results = append(e.results, line.ToArray())
		return true, nil
	}

	k := keys[0]
	var nextKeys []*key
	if len(keys) > 1 {
		nextKeys = keys[1:]
	}

	switch k.typ {
	case keyTypeIndex, keyTypeIndexRange, keyTypeIndexWild:
		start := <-jsonStream
		if start.typ != jsonBracketLeft {
			return true, fmt.Errorf("Expected [ instead of %s", jsonTokenNames[start.typ])
		}

		i := int64(0)
	arrayValues:
		for {
			if i < k.indexStart || i > k.indexEnd {
				traverseValueTree(jsonStream, false)
			} else if i >= k.indexStart && i <= k.indexEnd {
				e.locStack.Push(i)
				if ok, err := e.perform(jsonStream, nextKeys); !ok || err != nil {
					return ok, err
				}
				e.locStack.Pop()

			}
			comma := <-jsonStream
			switch comma.typ {
			case jsonComma:
				break
			case jsonBracketRight:
				break arrayValues
			default:
				return true, fmt.Errorf("Expected ',' or ']' instead of %s", jsonTokenNames[comma.typ])
			}
			i++
		}
	case keyTypeName, keyTypeNameList, keyTypeNameWild:
		start := <-jsonStream
		if start.typ != jsonBraceLeft {
			return true, fmt.Errorf("Expected { instead of %s", jsonTokenNames[start.typ])
		}

		keysCaptured := 0
	nameValuePairs:
		for {
			name := <-jsonStream
			if name.typ != jsonKey {
				return true, fmt.Errorf("Expected key instead of %s", jsonTokenNames[name.typ])
			}
			trimmedName := name.val[1 : len(name.val)-1]

			capture := false
			switch k.typ {
			case keyTypeName, keyTypeNameList:
				if _, found := k.keys[trimmedName]; found {
					capture = true
				}
			case keyTypeNameWild:
				capture = true
			}

			colon := <-jsonStream
			if colon.typ != jsonColon {
				return true, fmt.Errorf("Expected colon instead of %s", jsonTokenNames[colon.typ])
			}

			if capture {
				keysCaptured++

				e.locStack.Push(trimmedName)
				if ok, err := e.perform(jsonStream, nextKeys); !ok || err != nil {
					return ok, err
				}
				e.locStack.Pop()
			} else {
				traverseValueTree(jsonStream, false)
			}

			comma := <-jsonStream
			switch comma.typ {
			case jsonComma:
				//if k.typ != keyTypeNameWild && keysCaptured == len(k.keys) {
				//	break keyOp // early terminate operation
				//} else {
				break
				//}
			case jsonBraceRight:
				break nameValuePairs
			default:
				return true, fmt.Errorf("Expected ',' or '}' instead of %s", jsonTokenNames[comma.typ])
			}
		}
	}
	return true, nil
}

func (e *evaluator) run(l *lexer, keys []*key) ([][]interface{}, error) {
	_, err := e.perform(l.items, keys)

	l.Kill()
	for _ = range l.items {
		// deflate buffer
	}

	return e.results, err
}

func traverseValueTree(jsonStream <-chan *Item, capture bool) (string, error) {
	jsonStack := &stack{}
	buffer := bytes.NewBufferString("")

	for {
		item, ok := <-jsonStream
		if !ok {
			return "", errors.New("Premature end of stream")
		}
		switch item.typ {
		case jsonBraceLeft, jsonBracketLeft:
			jsonStack.Push(item.typ)
		case jsonBraceRight, jsonBracketRight:
			jsonStack.Pop()
		case jsonError:
			return "", errors.New(item.val)
		case jsonEOF:
			return "", errors.New("Premature EOF")
		}

		if capture {
			buffer.WriteString(item.val)
		}

		if jsonStack.Len() == 0 {
			break
		}
	}
	return buffer.String(), nil
}
