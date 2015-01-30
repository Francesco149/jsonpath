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
	pathPtr := flag.String("path", "", "Path to target in JOSN")
	flag.Parse()

	if pathPtr == nil && *pathPtr != "" {
		fmt.Println("Must specify path")
		os.Exit(1)
	}

	if filePtr != nil && *filePtr != "" {
		f, err := os.Open("large.json")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		reader := bufio.NewReader(f)
		results := jsonpath.Get(reader, *pathPtr)
		for l := range results {
			printLine(l)
		}

	} else if jsonPtr != nil && *jsonPtr != "" {
		results := jsonpath.GetByString(*jsonPtr, *pathPtr)
		for l := range results {
			printLine(l)
		}
	} else {
		fmt.Println("Must specify file or string")
		os.Exit(1)
	}
}

func printLine(l []interface{}) {
	for i, v := range l {
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
