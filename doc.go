/*
Package docx provides a comprehensive, clean-architecture Go library for creating
Microsoft Word (.docx) documents programmatically.

# Overview

go-docx v2 is a complete rewrite focused on clean architecture, comprehensive OOXML
support, and developer experience. It provides a type-safe, well-documented API for
creating professional Word documents.

# Features

  - Clean Architecture with domain-driven design
  - Complete OOXML (Office Open XML) support
  - Section management (page size, orientation, margins, columns)
  - Headers and footers (Default, First, Even)
  - Fields (TOC, page numbers, hyperlinks, dates, etc.)
  - Style system with 40+ built-in styles
  - Tables with formatting control
  - Thread-safe operations
  - Comprehensive error handling

# Quick Start

Creating a simple document:

	package main

	import (
	    docx "github.com/SlideLang/go-docx"
	    "github.com/SlideLang/go-docx/domain"
	)

	func main() {
	    doc := docx.NewDocument()

	    para, _ := doc.AddParagraph()
	    para.SetStyle(domain.StyleIDHeading1)
	    run, _ := para.AddRun()
	    run.AddText("Hello, go-docx!")

	    doc.SaveToFile("hello.docx")
	}

# Page Layout

Configure page size, orientation, and margins:

	section, _ := doc.DefaultSection()
	section.SetPageSize(domain.PageSizeA4)
	section.SetOrientation(domain.OrientationPortrait)
	section.SetMargins(domain.Margins{
	    Top:    1440, // 1 inch in twips
	    Right:  1440,
	    Bottom: 1440,
	    Left:   1440,
	    Header: 720,  // 0.5 inch
	    Footer: 720,
	})

# Headers and Footers

Add headers and footers with dynamic content:

	// Header
	header, _ := section.Header(domain.HeaderDefault)
	headerPara, _ := header.AddParagraph()
	headerPara.SetAlignment(domain.AlignmentRight)
	headerRun, _ := headerPara.AddRun()
	headerRun.AddText("Document Title")

	// Footer with page numbers
	footer, _ := section.Footer(domain.FooterDefault)
	footerPara, _ := footer.AddParagraph()

	r1, _ := footerPara.AddRun()
	r1.AddText("Page ")

	r2, _ := footerPara.AddRun()
	r2.AddField(docx.NewPageNumberField())

# Fields

Create dynamic fields for TOC, page numbers, hyperlinks, etc.:

	// Table of Contents
	tocRun, _ := para.AddRun()
	tocOptions := map[string]string{
	    "levels":     "1-3",
	    "hyperlinks": "true",
	}
	tocRun.AddField(docx.NewTOCField(tocOptions))

	// Hyperlink
	linkRun, _ := para.AddRun()
	linkField := docx.NewHyperlinkField(
	    "https://github.com/SlideLang/go-docx",
	    "Visit go-docx",
	)
	linkRun.SetColor(0x0000FF)
	linkRun.SetUnderline(domain.UnderlineSingle)
	linkRun.AddField(linkField)

# Styles

Use built-in styles with type-safe constants:

	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)

	heading, _ := doc.AddParagraph()
	heading.SetStyle(domain.StyleIDHeading1)

	quote, _ := doc.AddParagraph()
	quote.SetStyle(domain.StyleIDIntenseQuote)

	// Query styles
	styleMgr := doc.StyleManager()
	if styleMgr.HasStyle(domain.StyleIDHeading2) {
	    para.SetStyle(domain.StyleIDHeading2)
	}

# Tables

Create and format tables:

	table, _ := doc.AddTable(3, 4) // 3 rows, 4 columns

	// Access cells
	rows := table.Rows()
	cells := rows[0].Cells()

	cell := cells[0]
	para, _ := cell.AddParagraph()
	run, _ := para.AddRun()
	run.AddText("Cell content")
	run.SetBold(true)

# Character Formatting

Apply rich formatting to text runs:

	run, _ := para.AddRun()
	run.AddText("Formatted Text")
	run.SetBold(true)
	run.SetItalic(true)
	run.SetColor(0xFF0000) // Red
	run.SetFontSize(14)
	run.SetFontFamily("Arial")
	run.SetUnderline(domain.UnderlineSingle)

# Saving Documents

Save to file or writer:

	// Save to file
	err := doc.SaveToFile("document.docx")

	// Write to io.Writer
	var buf bytes.Buffer
	n, err := doc.WriteTo(&buf)

# Architecture

go-docx v2 follows clean architecture principles:

  - domain/: Core interfaces and business logic
  - internal/: Implementation details (core, manager, serializer, xml)
  - pkg/: Shared utilities (color, constants)

This structure provides testability, maintainability, and extensibility.

# Units and Measurements

Word documents use "twips" (twentieth of a point) for most measurements:

  - 1 inch = 1440 twips
  - 1 cm ≈ 567 twips
  - 1 point = 20 twips

Common conversions:

	oneinch := 1440
	halfInch := 720
	onePoint := 20

# Compatibility

Compatible with:
  - Microsoft Word 2007+ (Windows, macOS)
  - LibreOffice Writer 7.0+
  - Google Docs (upload)
  - Pages (macOS)
  - Any OOXML-compatible word processor

# Thread Safety

All document operations are thread-safe and protected with RWMutex.

# Error Handling

Most operations return (result, error). Always check errors:

	para, err := doc.AddParagraph()
	if err != nil {
	    log.Fatal(err)
	}

# Examples

See the examples/ directory for comprehensive examples:
  - examples/basic/: Simple document creation
  - examples/04_fields/: Fields demonstration
  - examples/05_styles/: Style management
  - examples/06_sections/: Page layout and sections
  - examples/07_advanced/: Complete integration

# Documentation

  - API Documentation: docs/API_DOCUMENTATION.md
  - Migration Guide: MIGRATION.md (v1 to v2)
  - Design Document: docs/V2_DESIGN.md
  - Contributing: CONTRIBUTING.md

# License

GNU Affero General Public License v3.0 (AGPL-3.0)
See LICENSE file for details.

# Support

  - Issues: https://github.com/SlideLang/go-docx/issues
  - Discussions: https://github.com/SlideLang/go-docx/discussions
*/
package docx
