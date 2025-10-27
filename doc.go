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

/*
Package docx provides a comprehensive library for creating and manipulating
Microsoft Word (.docx) documents in Go.

# Overview

This is version 2 of go-docx, featuring a complete rewrite with:
  - Clean architecture with domain-driven design
  - Comprehensive OOXML (Office Open XML) support
  - Fluent builder API for easy document creation
  - Strong type safety and error handling
  - Full support for tables, images, styles, and fields

# Quick Start

Create a simple document:

	package main

	import (
	    "log"
	    "github.com/mmonterroca/docxgo"
	)

	func main() {
	    doc := docx.NewDocument()

	    // Add a paragraph with formatted text
	    para, _ := doc.AddParagraph()
	    run, _ := para.AddRun()
	    run.SetText("Hello, World!")
	    run.SetBold(true)
	    run.SetSize(28) // 14pt (size is in half-points)

	    // Save the document
	    if err := doc.SaveAs("hello.docx"); err != nil {
	        log.Fatal(err)
	    }
	}

# Using the Fluent Builder API

The builder API provides a more convenient way to create documents:

	builder := docx.NewDocumentBuilder()

	builder.AddParagraph().
	    Text("Welcome to go-docx v2!").
	    Bold().
	    FontSize(16).
	    Color(docx.Blue).
	    End()

	builder.AddParagraph().
	    Text("This is a simple paragraph with ").
	    Text("multiple runs. ").
	    Text("This one is italic.").
	    Italic().
	    End()

	doc, err := builder.Build()
	if err != nil {
	    log.Fatal(err)
	}

	doc.SaveAs("output.docx")

# Working with Tables

Create and populate tables:

	doc := docx.NewDocument()

	// Create a 3x2 table
	table, _ := doc.AddTable(3, 2)

	// Set table style
	table.SetStyle(docx.TableStyleGrid)

	// Populate cells
	row0, _ := table.Row(0)
	cell, _ := row0.Cell(0)
	para, _ := cell.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("Header 1")
	run.SetBold(true)

	// Access other cells
	cell2, _ := row0.Cell(1)
	para2, _ := cell2.AddParagraph()
	run2, _ := para2.AddRun()
	run2.SetText("Header 2")
	run2.SetBold(true)

	// Merge cells
	cell3, _ := table.Row(1).Cell(0)
	cell3.SetGridSpan(2) // Span across 2 columns

# Adding Images

Insert images into your documents:

	doc := docx.NewDocument()
	para, _ := doc.AddParagraph()

	// Add image with custom size
	img, _ := para.AddImage("logo.png")
	size := domain.NewImageSizeInches(2.0, 1.5) // 2" wide, 1.5" tall
	img.SetSize(size)
	img.SetDescription("Company Logo")

	// Add floating image with text wrapping
	para2, _ := doc.AddParagraph()
	pos := domain.ImagePosition{
	    Type:     domain.ImagePositionFloating,
	    HAlign:   domain.HAlignRight,
	    VAlign:   domain.VAlignTop,
	    WrapText: domain.WrapSquare,
	}
	img2, _ := para2.AddImageWithPosition("photo.jpg", size, pos)

# Using Styles

Apply built-in or custom styles:

	doc := docx.NewDocument()

	// Use built-in heading styles
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	run1, _ := h1.AddRun()
	run1.SetText("Chapter 1: Introduction")

	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	run2, _ := h2.AddRun()
	run2.SetText("Background")

	// Normal paragraph
	para, _ := doc.AddParagraph()
	run3, _ := para.AddRun()
	run3.SetText("This is normal body text.")

# Adding Fields

Insert dynamic fields like page numbers and table of contents:

	doc := docx.NewDocument()

	// Add table of contents
	tocPara, _ := doc.AddParagraph()
	toc, _ := tocPara.AddField(domain.FieldTypeTOC)

	// Add page number in footer
	section, _ := doc.DefaultSection()
	footer, _ := section.Footer(domain.FooterDefault)
	footerPara, _ := footer.AddParagraph()
	footerPara.SetAlignment(domain.AlignmentCenter)
	pageNum, _ := footerPara.AddField(domain.FieldTypePageNumber)

# Page Layout and Sections

Configure page size, margins, and orientation:

	doc := docx.NewDocument()

	// Configure default section
	section, _ := doc.DefaultSection()
	section.SetPageSize(domain.PageSizeLetter)
	section.SetOrientation(domain.OrientationLandscape)

	margins := domain.Margins{
	    Top:    1440, // 1 inch (1440 twips)
	    Right:  1440,
	    Bottom: 1440,
	    Left:   2880, // 2 inches
	}
	section.SetMargins(margins)

	// Add content to first section
	para1, _ := doc.AddParagraph()
	// ...

	// Create a new section with different layout
	section2, _ := doc.AddSection()
	section2.SetOrientation(domain.OrientationPortrait)

	// Content after this point uses section2 settings

# Error Handling

All methods that can fail return an error:

	doc := docx.NewDocument()

	para, err := doc.AddParagraph()
	if err != nil {
	    log.Fatalf("Failed to add paragraph: %v", err)
	}

	run, err := para.AddRun()
	if err != nil {
	    log.Fatalf("Failed to add run: %v", err)
	}

	if err := run.SetText("Hello"); err != nil {
	    log.Fatalf("Failed to set text: %v", err)
	}

	// Validate document before saving
	if err := doc.Validate(); err != nil {
	    log.Fatalf("Document validation failed: %v", err)
	}

# Architecture

The library is organized into several packages:

  - docx: The main public API and document builder
  - domain: Interface definitions for the document model
  - internal/core: Concrete implementations of domain interfaces
  - internal/manager: Resource managers (IDs, relationships, media, styles)
  - internal/serializer: Converts domain objects to XML structures
  - internal/writer: ZIP archive creation for .docx files
  - internal/xml: OOXML XML structure definitions
  - pkg/constants: OOXML constants and measurements
  - pkg/color: Color utilities
  - pkg/errors: Error types

This clean architecture enables:
  - Dependency inversion (depend on interfaces, not implementations)
  - Easy unit testing with mocks
  - Clear separation of concerns
  - Future extensibility

# Measurements

Word documents use several measurement units:

  - Twips: 1/1440 of an inch (used for margins, indents, spacing)
  - Half-points: Font sizes (e.g., 24 = 12pt)
  - EMUs: English Metric Units, 914400 per inch (used for images)
  - DXA: Twentieths of a point (used for borders)

Helper functions are provided for common conversions:

	// 1 inch margins
	margins := domain.Margins{
	    Top:    1440, // 1440 twips = 1 inch
	    Right:  1440,
	    Bottom: 1440,
	    Left:   1440,
	}

	// 14pt font (28 half-points)
	run.SetSize(28)

	// 3-inch wide image
	size := domain.NewImageSizeInches(3.0, 2.0)

# Thread Safety

Document instances are NOT thread-safe. Do not access the same document
from multiple goroutines without external synchronization (e.g., sync.Mutex).

Each Document should be created and used by a single goroutine, or protected
by appropriate locking.

# Compatibility

This library generates Office Open XML (OOXML) documents compatible with:
  - Microsoft Word 2007 and later
  - LibreOffice Writer
  - Google Docs
  - Apple Pages (with some limitations)
  - Other OOXML-compatible applications

# Version

Current version: 2.0.0-beta

This is a major rewrite of the original go-docx library with breaking changes.
See the migration guide in docs/V2_DESIGN.md for details.

# Examples

See the examples/ directory for complete working examples:
  - examples/basic/: Simple document creation
  - examples/04_fields/: Working with fields (TOC, page numbers)
  - More examples available in the repository

# Links

  - GitHub: https://github.com/mmonterroca/docxgo
  - Documentation: https://pkg.go.dev/github.com/mmonterroca/docxgo
  - Examples: https://github.com/mmonterroca/docxgo/tree/main/examples

# License

MIT License - see LICENSE file for details.

Copyright (c) 2025 Misael Monterroca
Original go-docx: Copyright (c) 2020-2023 fumiama
*/
package docx
