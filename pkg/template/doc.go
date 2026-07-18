// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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
