// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package xml

import "encoding/xml"

// Header represents a Word header document (header1.xml, header2.xml, etc.)
//
// Content holds the header's top-level elements (*Paragraph, *Table) in
// document order, the same mixed-content pattern TableCell uses: there is no
// custom MarshalXML anywhere in this package, so ordering is slice order and
// each element's own XMLName decides its tag.
type Header struct {
	XMLName xml.Name      `xml:"w:hdr"`
	Xmlns   string        `xml:"xmlns:w,attr"`
	XmlnsR  string        `xml:"xmlns:r,attr"`
	Content []interface{} `xml:",any"`
}

// Footer represents a Word footer document (footer1.xml, footer2.xml, etc.)
//
// See Header's doc comment for the Content field's mixed-content contract.
type Footer struct {
	XMLName xml.Name      `xml:"w:ftr"`
	Xmlns   string        `xml:"xmlns:w,attr"`
	XmlnsR  string        `xml:"xmlns:r,attr"`
	Content []interface{} `xml:",any"`
}

// NewHeader creates a new header document.
func NewHeader() *Header {
	return &Header{
		Xmlns:   "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:  "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		Content: make([]interface{}, 0),
	}
}

// NewFooter creates a new footer document.
func NewFooter() *Footer {
	return &Footer{
		Xmlns:   "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:  "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		Content: make([]interface{}, 0),
	}
}

// AddParagraph adds a paragraph to the header.
func (h *Header) AddParagraph(p *Paragraph) {
	h.Content = append(h.Content, p)
}

// AddTable adds a table to the header.
func (h *Header) AddTable(t *Table) {
	h.Content = append(h.Content, t)
}

// AddParagraph adds a paragraph to the footer.
func (f *Footer) AddParagraph(p *Paragraph) {
	f.Content = append(f.Content, p)
}

// AddTable adds a table to the footer.
func (f *Footer) AddTable(t *Table) {
	f.Content = append(f.Content, t)
}
