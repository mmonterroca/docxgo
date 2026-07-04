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
