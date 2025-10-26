/*
MIT License

Copyright (c) 2025 Misael Montero
Copyright (c) 2020-2023 fumiama (original go-docx)

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



// Package errors provides structured error types for go-docx v2.
package errors

import (
	"fmt"
	"strings"
)

// Error codes for categorizing errors
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInvalidState = "INVALID_STATE"
	ErrCodeIO           = "IO_ERROR"
	ErrCodeXML          = "XML_ERROR"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeUnsupported  = "UNSUPPORTED"
)

// DocxError represents a structured error in go-docx v2.
type DocxError struct {
	Code    string                 // Error code (e.g., "VALIDATION_ERROR")
	Op      string                 // Operation that failed (e.g., "Document.AddParagraph")
	Err     error                  // Underlying error
	Message string                 // Human-readable message
	Context map[string]interface{} // Additional context
}

// Error implements the error interface.
func (e *DocxError) Error() string {
	var parts []string

	if e.Op != "" {
		parts = append(parts, fmt.Sprintf("operation=%s", e.Op))
	}

	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("code=%s", e.Code))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Err != nil {
		parts = append(parts, fmt.Sprintf("cause=%v", e.Err))
	}

	if len(e.Context) > 0 {
		var ctx []string
		for k, v := range e.Context {
			ctx = append(ctx, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("context={%s}", strings.Join(ctx, ", ")))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying error.
func (e *DocxError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error.
func (e *DocxError) Is(target error) bool {
	t, ok := target.(*DocxError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field      string      // Field name that failed validation
	Value      interface{} // Invalid value
	Constraint string      // Constraint that was violated
	Message    string      // Human-readable message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("validation error: field=%s, value=%v, constraint=%s, message=%s",
			e.Field, e.Value, e.Constraint, e.Message)
	}
	return fmt.Sprintf("validation error: field=%s, value=%v, constraint=%s",
		e.Field, e.Value, e.Constraint)
}

// BuilderError wraps an error and allows method chaining to continue
// while capturing the first error that occurred.
type BuilderError struct {
	err error
}

// Error implements the error interface.
func (b *BuilderError) Error() string {
	if b.err == nil {
		return "no error"
	}
	return b.err.Error()
}

// Unwrap returns the underlying error.
func (b *BuilderError) Unwrap() error {
	return b.err
}

// HasError returns true if an error has been captured.
func (b *BuilderError) HasError() bool {
	return b.err != nil
}

// Get returns the captured error (may be nil).
func (b *BuilderError) Get() error {
	return b.err
}

// Set sets the error if one hasn't been set already.
func (b *BuilderError) Set(err error) {
	if b.err == nil && err != nil {
		b.err = err
	}
}

// Helper functions for creating common errors

// Errorf creates a new DocxError with formatted message.
func Errorf(code, op, format string, args ...interface{}) error {
	return &DocxError{
		Code:    code,
		Op:      op,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an error with operation context.
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &DocxError{
		Code: ErrCodeInternal,
		Op:   op,
		Err:  err,
	}
}

// WrapWithCode wraps an error with operation and error code.
func WrapWithCode(err error, code, op string) error {
	if err == nil {
		return nil
	}
	return &DocxError{
		Code: code,
		Op:   op,
		Err:  err,
	}
}

// WrapWithContext wraps an error with operation and additional context.
func WrapWithContext(err error, op string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &DocxError{
		Code:    ErrCodeInternal,
		Op:      op,
		Err:     err,
		Context: context,
	}
}

// NotFound creates a "not found" error.
func NotFound(op, item string) error {
	return &DocxError{
		Code:    ErrCodeNotFound,
		Op:      op,
		Message: fmt.Sprintf("%s not found", item),
	}
}

// InvalidState creates an "invalid state" error.
func InvalidState(op, message string) error {
	return &DocxError{
		Code:    ErrCodeInvalidState,
		Op:      op,
		Message: message,
	}
}

// Validation creates a validation error.
func Validation(field string, value interface{}, constraint, message string) error {
	return &ValidationError{
		Field:      field,
		Value:      value,
		Constraint: constraint,
		Message:    message,
	}
}

// NewValidationError creates a validation error with operation context.
// This is a convenience function for backward compatibility.
func NewValidationError(op, field string, value interface{}, message string) error {
	return &DocxError{
		Code: ErrCodeValidation,
		Op:   op,
		Err: &ValidationError{
			Field:   field,
			Value:   value,
			Message: message,
		},
	}
}

// NewNotFoundError creates a "not found" error.
// This is a convenience function for backward compatibility.
func NewNotFoundError(op, field string, value interface{}, message string) error {
	return &DocxError{
		Code:    ErrCodeNotFound,
		Op:      op,
		Message: fmt.Sprintf("%s: %v - %s", field, value, message),
	}
}

// InvalidArgument creates a validation error for invalid arguments.
func InvalidArgument(op, field string, value interface{}, message string) error {
	return &DocxError{
		Code: ErrCodeValidation,
		Op:   op,
		Err: &ValidationError{
			Field:   field,
			Value:   value,
			Message: message,
		},
	}
}

// Unsupported creates an "unsupported" error.
func Unsupported(op, feature string) error {
	return &DocxError{
		Code:    ErrCodeUnsupported,
		Op:      op,
		Message: fmt.Sprintf("%s is not supported", feature),
	}
}
