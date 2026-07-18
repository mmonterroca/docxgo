// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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
