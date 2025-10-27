/*
MIT License

Copyright (c) 2025 Misael Monterroca
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

// Package domain defines the core domain interfaces for go-docx v2.
//
// This package provides a clean, testable API for working with Word documents.
// All functionality is exposed through interfaces to promote loose coupling
// and enable easy testing and mocking.
//
// # Core Interfaces
//
// The main interfaces are:
//   - Document: The top-level document structure
//   - Paragraph: A paragraph containing runs of formatted text
//   - Run: A run of text with consistent formatting
//   - Table: A table with rows and cells
//   - Section: A section with page layout settings
//   - Image: An embedded image with positioning
//   - Field: Dynamic fields (TOC, page numbers, etc.)
//   - Style: Paragraph and character styles
//
// # Example Usage
//
// Create a simple document:
//
//	doc := docx.NewDocument()
//	para, _ := doc.AddParagraph()
//	run, _ := para.AddRun()
//	run.SetText("Hello, World!")
//	run.SetBold(true)
//	doc.SaveAs("hello.docx")
//
// Create a table:
//
//	table, _ := doc.AddTable(3, 2) // 3 rows, 2 columns
//	cell, _ := table.Row(0).Cell(0)
//	para, _ := cell.AddParagraph()
//	run, _ := para.AddRun()
//	run.SetText("Cell content")
//
// Add an image:
//
//	para, _ := doc.AddParagraph()
//	img, _ := para.AddImage("photo.png")
//	size := domain.NewImageSizeInches(3.0, 2.0)
//	img.SetSize(size)
//
// # Architecture
//
// The domain package defines "what" can be done with a document.
// Implementation details ("how") are handled by the internal/core package.
// This separation enables:
//   - Clean architecture with dependency inversion
//   - Easy unit testing with mocks
//   - Future alternative implementations
//
// # Thread Safety
//
// Document instances are NOT thread-safe. If you need concurrent access,
// use external synchronization (e.g., sync.Mutex).
package domain

import "io"

// Document represents a Word document (.docx file).
// It provides methods to add content, manage structure, and persist to disk.
type Document interface {
	// AddParagraph adds a new paragraph to the document.
	// Returns an error if the operation fails.
	AddParagraph() (Paragraph, error)

	// AddTable adds a new table with the specified dimensions.
	// Returns an error if rows or cols are invalid.
	AddTable(rows, cols int) (Table, error)

	// AddSection adds a new section to the document.
	// Sections allow different page layouts within the same document.
	AddSection() (Section, error)

	// DefaultSection returns the default (first) section of the document.
	// Every document has at least one section.
	DefaultSection() (Section, error)

	// Paragraphs returns all paragraphs in the document.
	// The returned slice is a copy and modifications won't affect the document.
	Paragraphs() []Paragraph

	// Tables returns all tables in the document.
	// The returned slice is a copy and modifications won't affect the document.
	Tables() []Table

	// Sections returns all sections in the document.
	// The returned slice is a copy and modifications won't affect the document.
	Sections() []Section

	// WriteTo writes the document to the provided writer in .docx format.
	// Returns the number of bytes written and any error encountered.
	WriteTo(w io.Writer) (int64, error)

	// SaveAs saves the document to the specified file path.
	// Creates the file if it doesn't exist, overwrites if it does.
	SaveAs(path string) error

	// Validate checks if the document structure is valid.
	// Returns an error describing what's invalid, or nil if valid.
	Validate() error

	// Metadata returns the document's metadata (title, author, etc.)
	Metadata() *Metadata

	// SetMetadata updates the document's metadata.
	SetMetadata(meta *Metadata) error
}

// Metadata contains document properties like title, author, etc.
type Metadata struct {
	Title       string
	Subject     string
	Creator     string
	Keywords    []string
	Description string
	Created     string // ISO 8601 format
	Modified    string // ISO 8601 format
}
