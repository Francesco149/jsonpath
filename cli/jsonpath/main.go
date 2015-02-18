package main

import (
	// "bufio"

	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NodePrime/jsonpath"
	//"github.com/davecheney/profile"
)

func main() {
	// cfg := profile.Config {
	//           //   MemProfile: true,
	// 	CPUProfile: true,
	//             // NoShutdownHook: true, // do not hook SIGINT
	//      }
	//      // p.Stop() must be called before the program exits to
	//      // ensure profiling information is written to disk.
	//      p := profile.Start(&cfg)
	//      defer p.Stop()

	var paths pathSlice
	filePtr := flag.String("file", "", "Path to json file")
	jsonPtr := flag.String("json", "", "JSON text")
	flag.Var(&paths, "paths", "One or more paths to target in JSON")
	showPathPtr := flag.Bool("showPath", false, "Print keys & indexes that arrive to value")
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
			jsonpath.PrintResult(l, *showPathPtr)
		}
		checkAndHandleError(eval.Error)
		f.Close()

	} else if jsonPtr != nil && *jsonPtr != "" {
		eval, err := jsonpath.GetPathsInBytes([]byte(*jsonPtr), paths...)
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
