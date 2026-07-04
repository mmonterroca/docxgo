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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// runRPC starts a persistent JSON-RPC server that reads newline-delimited JSON
// requests from stdin and writes newline-delimited JSON responses to stdout.
// It runs until stdin is closed (EOF) or a SIGTERM/SIGINT signal is received.
func runRPC() int {
	s := newServer()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	lineCh := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		// Increase the scanner buffer so large JSON-RPC requests don't hit the
		// default 64K token limit and get treated as an unexpected EOF.
		scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "docxgo rpc: scanner error: %v\n", err)
		}
		close(lineCh)
	}()

	for {
		select {
		case line, ok := <-lineCh:
			if !ok {
				// EOF: stdin closed, graceful shutdown
				return 0
			}
			if line == "" {
				continue
			}
			var req Request
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				resp := Response{
					Error: &RPCError{
						Code:    "PARSE_ERROR",
						Message: "failed to parse request: " + err.Error(),
					},
				}
				writeResponse(resp)
				continue
			}
			resp := s.dispatch(&req)
			writeResponse(resp)
		case <-sigCh:
			// Graceful shutdown on signal
			return 0
		}
	}
}
