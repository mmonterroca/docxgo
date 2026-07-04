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

// Package template provides mail merge / template functionality for docx documents.
//
// The package allows you to define placeholder tokens in a Word document
// (using {{key}} syntax by default) and replace them with actual data values
// while preserving all document formatting.
//
// # Quick Start
//
// Open a template document, merge data, and save:
//
//	doc, _ := docx.OpenDocument("template.docx")
//	err := template.MergeTemplate(doc, template.MergeData{
//	    "customer_name": "Acme Corp",
//	    "invoice_date":  "2026-02-27",
//	    "total":         "$1,234.56",
//	})
//	doc.SaveAs("output.docx")
//
// # Features
//
//   - Automatic run consolidation to heal split placeholders
//   - Scans body paragraphs, table cells, headers, and footers
//   - Preserves all formatting (bold, italic, font, size, color, etc.)
//   - Custom delimiters (e.g., ${key} instead of {{key}})
//   - Strict mode for missing key detection
//   - Template validation with missing/unused key reporting
//   - Batch merge support (reopen template for each record)
//
// # Placeholder Syntax
//
// By default, placeholders use double-brace syntax: {{key}}
//
// An optional leading dot is supported for Go template compatibility: {{.key}}
//
// Nested keys use dots: {{customer.name}}
//
// Custom delimiters can be configured via MergeOptions.
package template
