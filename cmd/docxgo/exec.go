/*
MIT License

Copyright (c) 2025 Misael Monterroca <misael@monterroca.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
