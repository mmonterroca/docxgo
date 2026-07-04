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

// Package main implements the docxgo CLI binary with JSON-RPC over stdin/stdout.
package main

import "encoding/json"

// Request represents a JSON-RPC request.
type Request struct {
	ID     interface{}     `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response.
type Response struct {
	ID     interface{} `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
}

// ProtocolVersion is the JSON-RPC protocol version for feature negotiation.
const ProtocolVersion = "1.0"

// RPCError represents an error in a JSON-RPC response.
type RPCError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Operation string                 `json:"operation,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// errorResponse is a convenience helper that builds an error Response.
func errorResponse(id interface{}, code, message, operation string) Response {
	return Response{
		ID: id,
		Error: &RPCError{
			Code:      code,
			Message:   message,
			Operation: operation,
		},
	}
}

// errorResponseWithData builds an error Response with extra data.
func errorResponseWithData(id interface{}, code, message, operation string, data map[string]interface{}) Response {
	return Response{
		ID: id,
		Error: &RPCError{
			Code:      code,
			Message:   message,
			Operation: operation,
			Data:      data,
		},
	}
}
