package jsonpath

import (
	"fmt"
	"io"
)

func FindPathInBytes(input []byte, path string) (*eval, error) {
	query, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	lexer := NewBytesLexer(input, JSON)
	eval := newEvals(lexer, query)
	go eval.run()
	return eval, nil
}

func FindPathInReader(r io.Reader, path string) (*eval, error) {
	query, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	lexer := NewReaderLexer(r, JSON)
	eval := newEvals(lexer, query)
	go eval.run()
	return eval, nil
}

func PrintResult(l Result, showPath bool) {
	for i, v := range l {
		if !showPath && i < len(l)-1 {
			continue
		}

		switch v := v.(type) {
		case []byte:
			fmt.Printf("%s", v)
		case int:
			fmt.Printf("%d", v)
		default:
			fmt.Printf("%s", v)
		}
		if i < len(l)-1 {
			fmt.Print("\t")
		}
	}
	fmt.Println()
}
