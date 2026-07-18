// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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

	// AddSectionWithBreak adds a new section specifying how the transition
	// to the new section should behave (next page, continuous, even page, odd page).
	AddSectionWithBreak(breakType SectionBreakType) (Section, error)

	// AddPageBreak adds a page break to the document.
	// Creates a new paragraph with a page break run.
	AddPageBreak() error

	// StyleManager returns the style manager for this document.
	// Use this to query, add, or modify document styles.
	StyleManager() StyleManager

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

	// Blocks returns all top-level content elements (paragraphs, tables,
	// and section breaks) in the order they were added to the document.
	// The returned slice is a copy and modifications won't affect the document.
	Blocks() []Block

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

	// SetBackgroundColor sets the page background color for the entire document.
	// Pass an RGB color to render a solid page background. Use the zero-value
	// color together with BackgroundColor() to determine whether a custom color
	// is applied.
	SetBackgroundColor(color Color) error

	// BackgroundColor returns the configured page background color.
	// The boolean result indicates whether a background color is explicitly set.
	BackgroundColor() (Color, bool)

	// SetLanguage sets the document's default proofing language, used by Word
	// for spell-checking, grammar-checking, and hyphenation. Pass nil to clear
	// it. Returns an error if lang is non-nil but has no tag set (Val,
	// EastAsia, and Bidi all empty), and on a document opened via
	// OpenDocument/OpenDocumentFromBytes/OpenDocumentFromReader whose
	// styles.xml or settings.xml were preserved verbatim for round-trip
	// fidelity — on such a document the language could never actually reach
	// the saved file, so SetLanguage refuses rather than silently no-op.
	SetLanguage(lang *Language) error

	// Language returns a copy of the document's default proofing language, or
	// nil if unset. Mutating the returned value has no effect on the document.
	Language() *Language
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

// Language represents a document's default proofing language, expressed as
// BCP 47 language tags (e.g. "es-MX", "en-US").
type Language struct {
	Val      string // primary language, applied to Latin-script text
	EastAsia string // optional, for East Asian scripts (Chinese, Japanese, Korean)
	Bidi     string // optional, for right-to-left scripts (Arabic, Hebrew)
}
