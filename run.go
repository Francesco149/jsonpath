package jsonpath

import "io"

func EvalPathsInBytes(input []byte, pathStrings ...string) (*Eval, error) {
	paths, err := genPaths(pathStrings)
	if err != nil {
		return nil, err
	}
	lexer := NewSliceLexer(input, JSON)
	eval := newEvaluation(lexer, paths...)
	return eval, nil
}

func EvalPathsInReader(r io.Reader, pathStrings ...string) (*Eval, error) {
	paths, err := genPaths(pathStrings)
	if err != nil {
		return nil, err
	}

	lexer := NewReaderLexer(r, JSON)
	eval := newEvaluation(lexer, paths...)
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
