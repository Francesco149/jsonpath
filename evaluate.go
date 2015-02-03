package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

type evaluator struct {
	locStack *stack
	results  chan []interface{}
}

func newEvaluator() *evaluator {
	e := &evaluator{
		locStack: newStack(),
		results:  make(chan []interface{}, 100),
	}
	return e
}

func (e *evaluator) perform(jsonStream <-chan Item, keys []*key) error {
	if len(keys) == 0 {
		val, err := traverseValueTree(jsonStream, true)
		if err != nil {
			return err
		}
		line := e.locStack.Clone()
		line.Push(val)
		e.results <- line.ToArray()
		return nil
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
			return nil
		}
		foundEnd := false

		i := int64(0)
	arrayValues:
		for {
			if i < k.indexStart || i > k.indexEnd {
				traverseValueTree(jsonStream, false)
			} else if i >= k.indexStart && i <= k.indexEnd {
				e.locStack.Push(i)
				if err := e.perform(jsonStream, nextKeys); err != nil {
					fmt.Println(err)
					return err
				}
				e.locStack.Pop()
			}
			comma := <-jsonStream
			switch comma.typ {
			case jsonComma:
				break
			case jsonBracketRight:
				foundEnd = true
				break arrayValues
			default:
				return fmt.Errorf("Expected ',' or ']' instead of %s", jsonTokenNames[comma.typ])
			}
			i++
		}
		if !foundEnd {
			if err := traverseUntilEnd(jsonStream, jsonBraceLeft, jsonBraceRight); err != nil {
				return err
			}
		}
	case keyTypeName, keyTypeNameList, keyTypeNameWild:
		start := <-jsonStream
		if start.typ != jsonBraceLeft {
			return nil
		}
		foundEnd := false

		keysCaptured := 0
	nameValuePairs:
		for {
			name := <-jsonStream
			if name.typ != jsonKey {
				return fmt.Errorf("Expected key instead of %s", jsonTokenNames[name.typ])
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
				return fmt.Errorf("Expected colon instead of %s", jsonTokenNames[colon.typ])
			}

			if capture {
				keysCaptured++

				e.locStack.Push(trimmedName)
				if err := e.perform(jsonStream, nextKeys); err != nil {
					return err
				}
				e.locStack.Pop()
			} else {
				traverseValueTree(jsonStream, false)
			}

			comma := <-jsonStream
			switch comma.typ {
			case jsonComma:
				if k.typ != keyTypeNameWild && keysCaptured == len(k.keys) {
					break nameValuePairs // early terminate operation
				} else {
					break
				}
			case jsonBraceRight:
				foundEnd = true
				break nameValuePairs
			default:
				return fmt.Errorf("Expected ',' or '}' instead of %s", jsonTokenNames[comma.typ])
			}
		}
		if !foundEnd {
			if err := traverseUntilEnd(jsonStream, jsonBraceLeft, jsonBraceRight); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *evaluator) run(l *lexer, keys []*key) chan []interface{} {
	go func() {
		err := e.perform(l.items, keys)
		if err != nil {
			fmt.Println(err)
		}

		l.Kill()
		for _ = range l.items {
			// deflate buffer
		}

		close(e.results)
	}()

	return e.results
}

func traverseUntilEnd(jsonStream <-chan Item, open, end int) error {
	jsonStack := &stack{}
	jsonStack.Push(open)

looper:
	for item := range jsonStream {
		switch item.typ {
		case jsonBraceLeft, jsonBracketLeft:
			jsonStack.Push(item.typ)
		case jsonBraceRight, jsonBracketRight:
			jsonStack.Pop()
		case jsonError:
			return errors.New(item.val)
		case jsonEOF:
			return errors.New("Premature EOF")
		}

		if jsonStack.Len() == 0 {
			break looper
		}
	}
	return nil
}

func traverseValueTree(jsonStream <-chan Item, capture bool) (string, error) {
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
