package jsonpath

import (
	"bytes"
	"errors"
	"fmt"
)

type evaluator struct {
	tokenReader
	locStack stack
	results  chan []interface{}
}

const (
	AbruptTokenStreamEnd = "Token reader is not sending anymore tokens"
)

func newEvaluator(tr tokenReader) *evaluator {
	e := &evaluator{
		locStack:    stack(newStack()),
		tokenReader: tr,
		results:     make(chan []interface{}, 100),
	}
	return e
}

func (e *evaluator) perform(keys []*key) error {
	if len(keys) == 0 {
		val, err := traverseValueTree(&e.tokenReader, true)
		if err != nil {
			return err
		}
		line := e.locStack.clone()
		line.push(val)
		e.results <- line.toArray()
		return nil
	}

	k := keys[0]
	var nextKeys []*key
	if len(keys) > 1 {
		nextKeys = keys[1:]
	}

	switch k.typ {
	case keyTypeIndex, keyTypeIndexRange, keyTypeIndexWild:
		start, ok := e.tokenReader.next()
		if !ok {
			return fmt.Errorf(AbruptTokenStreamEnd)
		}
		if start.typ != jsonBracketLeft {
			return nil
		}
		foundEnd := false

		i := int64(0)
	arrayValues:
		for {
			if i < k.indexStart || i > k.indexEnd {
				traverseValueTree(&e.tokenReader, false)
			} else if i >= k.indexStart && i <= k.indexEnd {
				e.locStack.push(i)
				if err := e.perform(nextKeys); err != nil {
					fmt.Println(err)
					return err
				}
				e.locStack.pop()
			}
			comma, ok := e.tokenReader.next()
			if !ok {
				return fmt.Errorf(AbruptTokenStreamEnd)
			}
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
			if err := traverseUntilEnd(&e.tokenReader, jsonBraceLeft, jsonBraceRight); err != nil {
				return err
			}
		}
	case keyTypeName, keyTypeNameList, keyTypeNameWild:
		start, ok := e.tokenReader.next()
		if !ok {
			return fmt.Errorf(AbruptTokenStreamEnd)
		}
		if start.typ != jsonBraceLeft {
			return nil
		}
		foundEnd := false

		keysCaptured := 0
	nameValuePairs:
		for {
			name, ok := e.tokenReader.next()
			if !ok {
				return fmt.Errorf(AbruptTokenStreamEnd)
			}
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

			colon, ok := e.tokenReader.next()
			if !ok {
				return fmt.Errorf(AbruptTokenStreamEnd)
			}
			if colon.typ != jsonColon {
				return fmt.Errorf("Expected colon instead of %s", jsonTokenNames[colon.typ])
			}

			if capture {
				keysCaptured++

				e.locStack.push(trimmedName)
				if err := e.perform(nextKeys); err != nil {
					return err
				}
				e.locStack.pop()
			} else {
				traverseValueTree(&e.tokenReader, false)
			}

			comma, ok := e.tokenReader.next()
			if !ok {
				return fmt.Errorf(AbruptTokenStreamEnd)
			}
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
			if err := traverseUntilEnd(&e.tokenReader, jsonBraceLeft, jsonBraceRight); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *evaluator) run(keys []*key) chan []interface{} {
	err := e.perform(keys)
	if err != nil {
		fmt.Println(err)
	}

	close(e.results)
	return e.results
}

func traverseUntilEnd(tr *tokenReader, open, end int) error {
	jsonStack := newStack()
	jsonStack.push(open)

looper:
	for {
		item, ok := (*tr).next()
		if !ok {
			return fmt.Errorf(AbruptTokenStreamEnd)
		}
		switch item.typ {
		case jsonBraceLeft, jsonBracketLeft:
			jsonStack.push(item.typ)
		case jsonBraceRight, jsonBracketRight:
			jsonStack.pop()
		case jsonError:
			return errors.New(item.val)
		case jsonEOF:
			return errors.New("Premature EOF")
		}

		if jsonStack.len() == 0 {
			break looper
		}
	}
	return nil
}

func traverseValueTree(tr *tokenReader, capture bool) (string, error) {
	jsonStack := newStack()
	buffer := bytes.NewBufferString("")

	for {
		item, ok := (*tr).next()
		if !ok {
			return "", fmt.Errorf(AbruptTokenStreamEnd)
		}
		switch item.typ {
		case jsonBraceLeft, jsonBracketLeft:
			jsonStack.push(item.typ)
		case jsonBraceRight, jsonBracketRight:
			jsonStack.pop()
		case jsonError:
			return "", errors.New(item.val)
		case jsonEOF:
			return "", errors.New("Premature EOF")
		}

		if capture {
			buffer.WriteString(item.val)
		}

		if jsonStack.len() == 0 {
			break
		}
	}
	return buffer.String(), nil
}
