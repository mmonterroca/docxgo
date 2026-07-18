// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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
