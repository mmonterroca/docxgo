// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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
	"encoding/json"
	"fmt"
	"os"
	"runtime"

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
		runRPC()

	case "version":
		jsonFlag := false
		for _, a := range os.Args[2:] {
			if a == "--json" {
				jsonFlag = true
			}
		}
		if jsonFlag {
			info := map[string]interface{}{
				"name":            "docxgo",
				"version":         docx.Version,
				"protocolVersion": ProtocolVersion,
				"goVersion":       runtime.Version(),
				"platform":        runtime.GOOS,
				"arch":            runtime.GOARCH,
				"features":        capabilitiesMap(),
			}
			out, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "docxgo: failed to marshal version info: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(out))
		} else {
			fmt.Println(docx.Version)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

// capabilitiesMap returns the feature capabilities of this binary.
func capabilitiesMap() map[string]bool {
	return map[string]bool{
		"rpc":           true,
		"template":      true,
		"mailMerge":     true,
		"inspect":       true,
		"validate":      true,
		"batch":         true,
		"applyPatch":    true,
		"setLanguage":   true,
		"replaceText":   true,
		"cellEdit":      true,
		"streaming":     false,
		"partialUpdate": false,
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: docxgo <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  exec [--request JSON]  Execute a single JSON-RPC request (reads from stdin if omitted)")
	fmt.Fprintln(os.Stderr, "  rpc                    Start a persistent JSON-RPC server (newline-delimited)")
	fmt.Fprintln(os.Stderr, "  version [--json]       Print the docxgo library version")
}
