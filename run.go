package jsonpath

import (
	"fmt"
	"io"
)

func GetPathsInBytes(input []byte, pathStrings ...string) (*eval, error) {
	paths, err := genPaths(pathStrings)
	if err != nil {
		return nil, err
	}
	lexer := NewSliceLexer(input, JSON)
	eval := newEvaluation(lexer, paths...)
	go eval.run()
	return eval, nil
}

func GetPathsInReader(r io.Reader, pathStrings ...string) (*eval, error) {
	paths, err := genPaths(pathStrings)
	if err != nil {
		return nil, err
	}

	lexer := NewReaderLexer(r, JSON)
	eval := newEvaluation(lexer, paths...)
	go eval.run()
	return eval, nil
}

func genPaths(pathStrings []string) ([]*path, error) {
	paths := make([]*path, len(pathStrings))
	for x, p := range pathStrings {
		path, err := parsePath(p)
		if err != nil {
			return nil, err
		}
		paths[x] = path
	}
	return paths, nil
}

func Iterate(l lexer) {
	for {
		_, ok := l.next()
		if !ok {
			break
		}
	}
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
