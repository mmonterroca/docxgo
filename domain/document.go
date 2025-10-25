/*
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

// Package domain defines the core interfaces for go-docx v2.
// These interfaces provide a clean, testable API for working with Word documents.
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
