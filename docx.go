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



// Package docx provides functionality for creating and manipulating
// Microsoft Word (.docx) documents.
//
// This is v2 of the go-docx library, a complete rewrite with improved
// architecture, better error handling, and comprehensive OOXML support.
//
// Example usage:
//
//	doc := docx.NewDocument()
//	para, _ := doc.AddParagraph()
//	run, _ := para.AddRun()
//	run.SetText("Hello, World!")
//	run.SetBold(true)
//	doc.SaveAs("hello.docx")
package docx

import (
	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/core"
	"github.com/mmonterroca/docxgo/pkg/color"
)

// NewDocument creates a new empty Word document.
// The document is created with default settings and an empty body.
//
// Example:
//
//	doc := docx.NewDocument()
//	para, _ := doc.AddParagraph()
//	run, _ := para.AddRun()
//	run.AddText("Hello, World!")
//	doc.SaveToFile("hello.docx")
func NewDocument() domain.Document {
	return core.NewDocument()
}

// Version is the library version.
const Version = "2.0.0-beta"

// Common color constants exported for convenience.
var (
	Black   = color.Black
	White   = color.White
	Red     = color.Red
	Green   = color.Green
	Blue    = color.Blue
	Yellow  = color.Yellow
	Cyan    = color.Cyan
	Magenta = color.Magenta
	Orange  = color.Orange
	Purple  = color.Purple
	Gray    = color.Gray
	Silver  = color.Silver
)

// Common alignment constants exported for convenience.
const (
	AlignmentLeft       = domain.AlignmentLeft
	AlignmentCenter     = domain.AlignmentCenter
	AlignmentRight      = domain.AlignmentRight
	AlignmentJustify    = domain.AlignmentJustify
	AlignmentDistribute = domain.AlignmentDistribute
)

// Common underline constants exported for convenience.
const (
	UnderlineNone   = domain.UnderlineNone
	UnderlineSingle = domain.UnderlineSingle
	UnderlineDouble = domain.UnderlineDouble
	UnderlineThick  = domain.UnderlineThick
	UnderlineDotted = domain.UnderlineDotted
	UnderlineDashed = domain.UnderlineDashed
	UnderlineWave   = domain.UnderlineWave
)

// Common break type constants exported for convenience.
const (
	BreakTypePage   = domain.BreakTypePage
	BreakTypeColumn = domain.BreakTypeColumn
	BreakTypeLine   = domain.BreakTypeLine
)

// Field creation functions

// NewField creates a new field of the specified type.
// Use the specific factory functions (NewPageNumberField, NewTOCField, etc.)
// for most use cases.
func NewField(fieldType domain.FieldType) domain.Field {
	return core.NewField(fieldType)
}

// NewPageNumberField creates a field that displays the current page number.
//
// Example:
//
//	footer, _ := section.Footer(domain.FooterDefault)
//	para, _ := footer.AddParagraph()
//	run, _ := para.AddRun()
//	run.AddField(docx.NewPageNumberField())
func NewPageNumberField() domain.Field {
	return core.NewPageNumberField()
}

// NewPageCountField creates a field that displays the total number of pages.
//
// Example:
//
//	run, _ := para.AddRun()
//	run.AddText("Page ")
//	run2, _ := para.AddRun()
//	run2.AddField(docx.NewPageNumberField())
//	run3, _ := para.AddRun()
//	run3.AddText(" of ")
//	run4, _ := para.AddRun()
//	run4.AddField(docx.NewPageCountField())
func NewPageCountField() domain.Field {
	return core.NewPageCountField()
}

// NewTOCField creates a Table of Contents field.
// The switches map accepts standard Word TOC switches:
//   - "levels": Heading levels to include (e.g., "1-3")
//   - "hyperlinks": Whether to make TOC entries clickable ("true"/"false")
//   - "hidePageNumbers": Whether to hide page numbers ("true"/"false")
//
// Example:
//
//	tocOptions := map[string]string{
//	    "levels":     "1-3",
//	    "hyperlinks": "true",
//	}
//	run.AddField(docx.NewTOCField(tocOptions))
func NewTOCField(switches map[string]string) domain.Field {
	return core.NewTOCField(switches)
}

// NewHyperlinkField creates a clickable hyperlink field.
// The url parameter should be a complete URL (http://, https://, mailto:, etc.)
// The displayText is what the user sees in the document.
//
// Example:
//
//	run, _ := para.AddRun()
//	linkField := docx.NewHyperlinkField(
//	    "https://github.com/mmonterroca/docxgo",
//	    "go-docx Repository",
//	)
//	run.SetColor(0x0000FF) // Blue
//	run.SetUnderline(domain.UnderlineSingle)
//	run.AddField(linkField)
func NewHyperlinkField(url, displayText string) domain.Field {
	return core.NewHyperlinkField(url, displayText)
}

// NewStyleRefField creates a field that references the last paragraph
// of the specified style (useful for running headers showing chapter titles).
//
// Example:
//
//	// In header, show the current chapter (last Heading1)
//	header, _ := section.Header(domain.HeaderDefault)
//	para, _ := header.AddParagraph()
//	run, _ := para.AddRun()
//	run.AddField(docx.NewStyleRefField("Heading 1"))
func NewStyleRefField(styleName string) domain.Field {
	return core.NewStyleRefField(styleName)
}
