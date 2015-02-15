package main

import (
	// "bufio"

	"flag"
	"fmt"
	"os"

	"github.com/NodePrime/jsonpath"
)

func main() {
	filePtr := flag.String("file", "", "Path to json file")
	jsonPtr := flag.String("json", "", "JSON text")
	pathPtr := flag.String("path", "", "Path to target in JSON")
	showPathPtr := flag.Bool("showPath", false, "Print keys & indexes that arrive to value")
	flag.Parse()

	if pathPtr == nil && *pathPtr != "" {
		fmt.Println("Must specify path")
		os.Exit(1)
	}

	if filePtr != nil && *filePtr != "" {
		f, err := os.Open(*filePtr)
		if err != nil {
			fmt.Println(fmt.Errorf("Failed to open file: %q", err.Error()))
			os.Exit(1)
		}

		eval, err := jsonpath.FindPathInReader(f, *pathPtr)
		checkAndHandleError(err)
		for l := range eval.Results {
			jsonpath.PrintResult(l, *showPathPtr)
		}
		checkAndHandleError(eval.Error)
		f.Close()

	} else if jsonPtr != nil && *jsonPtr != "" {
		eval, err := jsonpath.FindPathInBytes([]byte(*jsonPtr), *pathPtr)
		checkAndHandleError(err)
		for l := range eval.Results {
			jsonpath.PrintResult(l, *showPathPtr)
		}
		checkAndHandleError(eval.Error)
	} else {
		fmt.Println("Must specify file or string")
		os.Exit(1)
	}
}

func checkAndHandleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
