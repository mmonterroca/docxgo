// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import "fmt"

// MergeData is the data to merge into a template document.
// Keys correspond to placeholder names without delimiters.
type MergeData map[string]string

// MergeOptions configures merge behavior.
type MergeOptions struct {
	// OpenDelimiter is the opening delimiter for placeholders (default: "{{").
	OpenDelimiter string
	// CloseDelimiter is the closing delimiter for placeholders (default: "}}").
	CloseDelimiter string
	// StrictMode when true causes MergeTemplate to return an error if any
	// placeholder in the document has no corresponding key in MergeData.
	StrictMode bool
}

// DefaultMergeOptions returns the default merge options.
func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		OpenDelimiter:  "{{",
		CloseDelimiter: "}}",
		StrictMode:     false,
	}
}

// ValidationSeverity indicates the severity of a validation finding.
type ValidationSeverity int

const (
	// SeverityError indicates a problem that will prevent correct merge.
	SeverityError ValidationSeverity = iota
	// SeverityWarning indicates a non-critical finding.
	SeverityWarning
)

// ValidationError describes a template validation finding.
type ValidationError struct {
	// Severity indicates whether this is an error or warning.
	Severity ValidationSeverity
	// Key is the placeholder or data key involved.
	Key string
	// Message describes the issue.
	Message string
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	severity := "ERROR"
	if v.Severity == SeverityWarning {
		severity = "WARNING"
	}
	return fmt.Sprintf("[%s] %s: %s", severity, v.Key, v.Message)
}
