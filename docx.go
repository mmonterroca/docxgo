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
	"github.com/SlideLang/go-docx/domain"
	"github.com/SlideLang/go-docx/internal/core"
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
//	    "https://github.com/SlideLang/go-docx",
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
