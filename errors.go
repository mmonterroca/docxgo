/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)
   Copyright (c) 2025 SlideLang

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package docx

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrNilDocument is returned when a nil document is used
	ErrNilDocument = errors.New("docx: nil document")

	// ErrNilParagraph is returned when a nil paragraph is used
	ErrNilParagraph = errors.New("docx: nil paragraph")

	// ErrNilRun is returned when a nil run is used
	ErrNilRun = errors.New("docx: nil run")

	// ErrInvalidIndent is returned when indent values are out of range
	ErrInvalidIndent = errors.New("docx: invalid indent value")

	// ErrInvalidJustification is returned when an invalid justification value is provided
	ErrInvalidJustification = errors.New("docx: invalid justification value")

	// ErrConflictingIndent is returned when both firstLine and hanging indents are specified
	ErrConflictingIndent = errors.New("docx: cannot specify both firstLine and hanging indent")

	// ErrInvalidUnderline is returned when an invalid underline style is provided
	ErrInvalidUnderline = errors.New("docx: invalid underline style")
)

// DocxError represents a structured error from the docx package
type DocxError struct {
	// Op is the operation that failed
	Op string

	// Err is the underlying error
	Err error

	// Context provides additional context about the error
	Context string
}

// Error implements the error interface
func (e *DocxError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("docx: %s: %s: %v", e.Op, e.Context, e.Err)
	}
	return fmt.Sprintf("docx: %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *DocxError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	// Field is the field that failed validation
	Field string

	// Value is the invalid value
	Value interface{}

	// Constraint describes what constraint was violated
	Constraint string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("docx: validation failed for %s: %v violates constraint: %s",
		e.Field, e.Value, e.Constraint)
}

// Helper functions to create common errors

// newDocxError creates a new DocxError
func newDocxError(op string, err error, context string) *DocxError {
	return &DocxError{
		Op:      op,
		Err:     err,
		Context: context,
	}
}

// newValidationError creates a new ValidationError
func newValidationError(field string, value interface{}, constraint string) *ValidationError {
	return &ValidationError{
		Field:      field,
		Value:      value,
		Constraint: constraint,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// IsDocxError checks if an error is a DocxError
func IsDocxError(err error) bool {
	var de *DocxError
	return errors.As(err, &de)
}
