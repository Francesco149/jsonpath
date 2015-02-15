package jsonpath

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

const (
	opTypeIndex = iota
	opTypeIndexRange
	opTypeIndexWild
	opTypeName
	opTypeNameList
	opTypeNameWild
)

type query struct {
	captureEndValue bool
	operators       []*operator
}

type operator struct {
	typ int

	indexStart int
	indexEnd   int

	keyStrings map[string]struct{}
}

func genIndexKey(tr tokenReader) (*operator, error) {
	k := &operator{}
	var t *Item
	var ok bool
	if t, ok = tr.next(); !ok {
		return nil, errors.New("Expected number or *, but got none")
	}

	switch t.typ {
	case pathWildcard:
		k.typ = opTypeIndexWild
		k.indexStart = 0
		k.indexEnd = math.MaxInt64
		if t, ok = tr.next(); !ok {
			return nil, errors.New("Expected number or *, but got none")
		}
		if t.typ != pathBracketRight {
			return nil, fmt.Errorf("Expected ] after * instead of %q", t.val)
		}

	case pathIndex:
		v, err := strconv.Atoi(string(t.val))
		if err != nil {
			return nil, fmt.Errorf("Could not parse %q into int64", t.val)
		}
		k.indexStart = v
		k.indexEnd = v

		if t, ok = tr.next(); !ok {
			return nil, errors.New("Expected number or *, but got none")
		}
		switch t.typ {
		case pathBracketRight:
			k.typ = opTypeIndex
		// case path range
		default:
			return nil, fmt.Errorf("Unexpected value within brackets: %q", t.val)
		}

	default:
		return nil, fmt.Errorf("Unexpected value within brackets: %q", t.val)
	}

	return k, nil
}

func parsePath(path string) (*query, error) {
	lexer := NewBytesLexer([]byte(path), PATH)
	return toQuery(lexer)
}

func toQuery(tr tokenReader) (*query, error) {
	q := &query{
		captureEndValue: false,
		operators:       make([]*operator, 0),
	}
	for {
		p, ok := tr.next()
		if !ok {
			break
		}
		switch p.typ {
		case pathRoot:
			if len(q.operators) != 0 {
				return nil, errors.New("Unexpected root after start")
			}
			continue
		case pathPeriod:
			continue
		case pathBracketLeft:
			k, err := genIndexKey(tr)
			if err != nil {
				return nil, err
			}
			q.operators = append(q.operators, k)
		case pathKey:
			q.operators = append(q.operators, &operator{typ: opTypeName, keyStrings: map[string]struct{}{string(p.val): struct{}{}}})
		case pathWildcard:
			q.operators = append(q.operators, &operator{typ: opTypeNameWild})
		case pathValue:
			q.captureEndValue = true
		case pathError:
			return q, errors.New(string(p.val))
		}
	}
	return q, nil
}
