package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/NodePrime/jsonpath"
)

func main() {
	filePtr := flag.String("file", "", "Path to json file")
	jsonPtr := flag.String("json", "", "JSON text")
	pathPtr := flag.String("path", "", "Path to target in JSON")
	valuePtr := flag.Bool("value", true, "Print value at end of path")
	flag.Parse()

	if pathPtr == nil && *pathPtr != "" {
		fmt.Println("Must specify path")
		os.Exit(1)
	}

	if filePtr != nil && *filePtr != "" {
		f, err := os.Open(*filePtr)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		reader := bufio.NewReader(f)
		results := jsonpath.Get(reader, *pathPtr)
		for l := range results {
			printLine(l, *valuePtr)
		}
		f.Close()

	} else if jsonPtr != nil && *jsonPtr != "" {
		results := jsonpath.GetByString(*jsonPtr, *pathPtr)
		for l := range results {
			printLine(l, *valuePtr)
		}
	} else {
		fmt.Println("Must specify file or string")
		os.Exit(1)
	}
}

func printLine(l []interface{}, printValue bool) {
	for i, v := range l {
		if i == len(l)-1 {
			if !printValue {
				continue
			}
		}
		switch v := v.(type) {
		case string:
			fmt.Printf("%q", v)
		case int64:
			fmt.Printf("%d", v)
		}
		if i < len(l)-1 {
			fmt.Printf(" ")
		}
	}
	fmt.Println()
}
