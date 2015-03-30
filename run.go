package jsonpath

import "io"

func EvalPathsInBytes(input []byte, paths []*path) (*Eval, error) {
	lexer := NewSliceLexer(input, JSON)
	eval := newEvaluation(lexer, paths...)
	return eval, nil
}

func EvalPathsInReader(r io.Reader, paths []*path) (*Eval, error) {
	lexer := NewReaderLexer(r, JSON)
	eval := newEvaluation(lexer, paths...)
	return eval, nil
}

func ParsePaths(pathStrings ...string) ([]*path, error) {
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
