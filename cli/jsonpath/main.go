package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NodePrime/jsonpath"
	flag "github.com/ogier/pflag"
)

func main() {
	var paths pathSlice
	filePtr := flag.StringP("file", "f", "", "Path to json file")
	jsonPtr := flag.StringP("json", "j", "", "JSON text")
	flag.VarP(&paths, "path", "p", "One or more paths to target in JSON")
	showKeysPtr := flag.BoolP("keys", "k", false, "Print keys & indexes that lead to value")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Pipe JSON to StdIn by not specifying --file or --json ")
	}
	flag.Parse()

	if len(paths) == 0 {
		fmt.Println("Must specify one or more paths with the   -paths   flag")
		os.Exit(1)
	}

	if filePtr != nil && *filePtr != "" {
		f, err := os.Open(*filePtr)
		if err != nil {
			fmt.Println(fmt.Errorf("Failed to open file: %q", err.Error()))
			os.Exit(1)
		}

		eval, err := jsonpath.GetPathsInReader(f, paths...)
		checkAndHandleError(err)
		for l := range eval.Results {
			jsonpath.PrintResult(l, *showKeysPtr)
		}
		checkAndHandleError(eval.Error)
		f.Close()

	} else if jsonPtr != nil && *jsonPtr != "" {
		eval, err := jsonpath.GetPathsInBytes([]byte(*jsonPtr), paths...)
		checkAndHandleError(err)
		for l := range eval.Results {
			jsonpath.PrintResult(l, *showKeysPtr)
		}
		checkAndHandleError(eval.Error)
	} else {
		reader := bufio.NewReader(os.Stdin)
		eval, err := jsonpath.GetPathsInReader(reader, paths...)
		checkAndHandleError(err)
		for l := range eval.Results {
			jsonpath.PrintResult(l, *showKeysPtr)
		}
		checkAndHandleError(eval.Error)
	}
}

func checkAndHandleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type pathSlice []string

func (i *pathSlice) Set(value string) error {
	for _, dt := range strings.Split(value, ",") {
		*i = append(*i, dt)
	}
	return nil
}

func (i *pathSlice) String() string {
	return fmt.Sprint(*i)
}
