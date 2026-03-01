// Command docxgo provides a CLI binary for creating and manipulating Word
// documents via JSON-RPC over stdin/stdout.
//
// Usage:
//
//	docxgo exec [--request JSON]   Execute a single JSON-RPC request
//	docxgo rpc                     Start a persistent JSON-RPC server
//	docxgo version                 Print the version string
package main

import (
	"fmt"
	"os"

	docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "exec":
		var requestFlag string
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--request" && i+1 < len(os.Args) {
				requestFlag = os.Args[i+1]
				i++
			}
		}
		os.Exit(runExec(requestFlag))

	case "rpc":
		os.Exit(runRPC())

	case "version":
		fmt.Println(docx.Version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: docxgo <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  exec [--request JSON]  Execute a single JSON-RPC request (reads from stdin if omitted)")
	fmt.Fprintln(os.Stderr, "  rpc                    Start a persistent JSON-RPC server (newline-delimited)")
	fmt.Fprintln(os.Stderr, "  version                Print the docxgo library version")
}
