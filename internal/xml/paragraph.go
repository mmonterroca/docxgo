// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package xml provides XML structures and marshaling/unmarshaling for Office Open XML (OOXML) format.
//
// This package defines the low-level XML structures that map directly to the OOXML specification.
// Each struct corresponds to an XML element in the Word document format with proper namespace
// attributes (w:, r:, wp:, etc.).
//
// Structures include:
// - Document: w:document root element
// - Paragraph: w:p (paragraph) element
// - Run: w:r (run) element
// - Table: w:tbl (table) element
// - Drawing: w:drawing (drawing) element
// - Style: w:style (style) element
// - Section: w:sectPr (section properties) element
// - Field: w:fldSimple, w:fldChar (field) elements
//
// These structures are used by the serializer package to convert domain objects
// into XML that conforms to the Office Open XML standard.
package xml

import "encoding/xml"

// Paragraph represents w:p element.
type Paragraph struct {
	XMLName    xml.Name             `xml:"w:p"`
	Properties *ParagraphProperties `xml:"w:pPr,omitempty"`
	Elements   []interface{}        `xml:",any"`
}

// BookmarkStart represents w:bookmarkStart element.
type BookmarkStart struct {
	XMLName xml.Name `xml:"w:bookmarkStart"`
	ID      string   `xml:"w:id,attr"`
	Name    string   `xml:"w:name,attr"`
}

// BookmarkEnd represents w:bookmarkEnd element.
type BookmarkEnd struct {
	XMLName xml.Name `xml:"w:bookmarkEnd"`
	ID      string   `xml:"w:id,attr"`
}

// ParagraphProperties represents w:pPr element.
// Element order follows ISO/IEC 29500 (OOXML) schema sequence requirements.
type ParagraphProperties struct {
	XMLName           xml.Name             `xml:"w:pPr"`
	Style             *ParagraphStyleRef   `xml:"w:pStyle,omitempty"`      // 1. pStyle
	Numbering         *NumberingProperties `xml:"w:numPr,omitempty"`       // 2. numPr (after pStyle, before pBdr)
	Borders           *ParagraphBorders    `xml:"w:pBdr,omitempty"`        // 3. pBdr
	Spacing           *Spacing             `xml:"w:spacing,omitempty"`     // 4. spacing
	Indentation       *Indentation         `xml:"w:ind,omitempty"`         // 5. ind
	Justification     *Justification       `xml:"w:jc,omitempty"`          // 6. jc (near the end)
	SectionProperties *SectionProperties   `xml:"w:sectPr,omitempty"`      // 7. sectPr (always last)
}

// ParagraphBorders represents w:pBdr element (paragraph borders).
type ParagraphBorders struct {
	XMLName xml.Name `xml:"w:pBdr"`
	Top     *Border  `xml:"w:top,omitempty"`
	Bottom  *Border  `xml:"w:bottom,omitempty"`
	Left    *Border  `xml:"w:left,omitempty"`
	Right   *Border  `xml:"w:right,omitempty"`
}

// ParagraphStyleRef represents w:pStyle element (reference to a style).
type ParagraphStyleRef struct {
	Val string `xml:"w:val,attr"`
}

// Justification represents w:jc element (alignment).
type Justification struct {
	Val string `xml:"w:val,attr"`
}

// Indentation represents w:ind element.
type Indentation struct {
	Left      *int `xml:"w:left,attr,omitempty"`
	Right     *int `xml:"w:right,attr,omitempty"`
	FirstLine *int `xml:"w:firstLine,attr,omitempty"`
	Hanging   *int `xml:"w:hanging,attr,omitempty"`
}

// Spacing represents w:spacing element.
type Spacing struct {
	Before   *int    `xml:"w:before,attr,omitempty"`
	After    *int    `xml:"w:after,attr,omitempty"`
	Line     *int    `xml:"w:line,attr,omitempty"`
	LineRule *string `xml:"w:lineRule,attr,omitempty"`
}

// NumberingProperties represents w:numPr element (numbering reference).
type NumberingProperties struct {
	XMLName xml.Name       `xml:"w:numPr"`
	Level   *DecimalNumber `xml:"w:ilvl,omitempty"`
	NumID   *DecimalNumber `xml:"w:numId,omitempty"`
}

// DecimalNumber represents simple numeric value elements (w:val attr).
type DecimalNumber struct {
	Val int `xml:"w:val,attr"`
}

// Hyperlink represents w:hyperlink element.
// Can use either r:id (external link via relationship) or w:anchor (internal bookmark link).
type Hyperlink struct {
	XMLName xml.Name `xml:"w:hyperlink"`
	ID      string   `xml:"r:id,attr,omitempty"` // External hyperlink via relationship
	Anchor  string   `xml:"w:anchor,attr,omitempty"` // Internal bookmark anchor (e.g., _Toc123456)
	History string   `xml:"w:history,attr,omitempty"` // "1" to add to history
	Runs    []*Run   `xml:"w:r"`
}
