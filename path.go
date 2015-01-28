package jsonpath

import (
	"io"
	"strings"
)

func jsonPath(rs io.RuneScanner, path string) ([]string, error) {
	lexer := NewLexer(rs, 10)
	_ = lexer
	return nil, nil
}

func jsonPathString(input, path string) ([]string, error) {
	reader := strings.NewReader(input)
	return jsonPath(reader, path)
}

func parsePath(path string) ([]*Item, error) {
	reader := strings.NewReader(path)
	lexer := NewLexer(reader, 10)
	go lexer.Run(PATH)
	pathItems := funnelToArray(lexer.items)
	return pathItems, nil
}
