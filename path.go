package jsonpath

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

func GetByString(input, path string) <-chan []interface{} {
	reader := strings.NewReader(input)
	return Get(reader, path)
}

func Get(rs io.Reader, path string) <-chan []interface{} {
	query, err := parsePath(path)
	if err != nil {
		return nil
	}

	lexer := NewLexer(rs, 10)
	lexer.Run(JSON)

	eval := newEvaluator()

	return eval.run(lexer, query)
}

const (
	keyTypeIndex = iota
	keyTypeIndexRange
	keyTypeIndexWild
	keyTypeName
	keyTypeNameList
	keyTypeNameWild
)

type key struct {
	typ int

	indexStart int64
	indexEnd   int64

	keys map[string]struct{}
}

func genIndexKey(path <-chan *Item) (*key, error) {
	k := &key{}
	first := <-path
	switch first.typ {
	case pathWildcard:
		k.typ = keyTypeIndexWild
		k.indexStart = 0
		k.indexEnd = math.MaxInt64
		end := <-path
		if end.typ != pathBracketRight {
			return nil, fmt.Errorf("Expected ] after * instead of %q", first.val)
		}

	case pathIndex:
		v, err := strconv.ParseInt(first.val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Could not parse %q into int64", first.val)
		}
		k.indexStart = v
		k.indexEnd = v

		second := <-path
		switch second.typ {
		case pathBracketRight:
			k.typ = keyTypeIndex
		// case path range
		default:
			return nil, fmt.Errorf("Unexpected value within brackets: %q", first.val)
		}

	default:
		return nil, fmt.Errorf("Unexpected value within brackets: %q", first.val)
	}

	return k, nil
}

func parsePath(path string) ([]*key, error) {
	reader := strings.NewReader(path)
	lexer := NewLexer(reader, 10)
	lexer.Run(PATH)

	return toQuery(lexer.items)
}

func toQuery(path <-chan *Item) ([]*key, error) {
	query := make([]*key, 0)
	for p := range path {
		switch p.typ {
		case pathRoot:
			if len(query) != 0 {
				return nil, errors.New("Unexpected root after start")
			}
			continue
		case pathPeriod:
			continue
		case pathBracketLeft:
			k, err := genIndexKey(path)
			if err != nil {
				return nil, err
			}
			query = append(query, k)
		case pathKey:
			query = append(query, &key{typ: keyTypeName, keys: map[string]struct{}{p.val: struct{}{}}})
		case pathWildcard:
			query = append(query, &key{typ: keyTypeNameWild})
		}
	}

	return query, nil
}
