// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// runExec executes a single JSON-RPC request and writes the response to stdout.
// The request is read from requestFlag if non-empty, otherwise from stdin.
func runExec(requestFlag string) int {
	var raw []byte
	var err error

	if requestFlag != "" {
		raw = []byte(requestFlag)
	} else {
		raw, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "docxgo: error reading stdin: %v\n", err)
			return 1
		}
	}

	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		resp := Response{
			Error: &RPCError{
				Code:    "PARSE_ERROR",
				Message: "failed to parse request: " + err.Error(),
			},
		}
		writeResponse(resp)
		return 1
	}

	s := newServer()
	resp := s.dispatch(&req)
	writeResponse(resp)

	if resp.Error != nil {
		return 1
	}
	return 0
}

// writeResponse encodes a Response as JSON and writes it to stdout followed by a newline.
func writeResponse(resp Response) {
	out, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docxgo: failed to marshal response: %v\n", err)
		return
	}
	fmt.Println(string(out))
}
