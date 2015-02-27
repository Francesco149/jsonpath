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

func PrintResult(r Result, showPath bool) {
	printed := false
	if showPath {
		for _, k := range r.Keys {
			switch v := k.(type) {
			case int:
				fmt.Printf("%d", v)
			default:
				fmt.Printf("%q", v)
			}
			fmt.Print("\t")
			printed = true
		}
	} else if r.Value == nil {
		if len(r.Keys) > 0 {
			printed = true
			switch v := r.Keys[len(r.Keys)-1].(type) {
			case int:
				fmt.Printf("%d", v)
			default:
				fmt.Printf("%q", v)
			}
		}
	}

	if r.Value != nil {
		printed = true
		fmt.Printf("%s", r.Value)
	}
	if printed {
		fmt.Println()
	}
}
