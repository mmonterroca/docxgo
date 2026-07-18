// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package xml

import "encoding/xml"

// Header represents a Word header document (header1.xml, header2.xml, etc.)
type Header struct {
	XMLName    xml.Name     `xml:"w:hdr"`
	Xmlns      string       `xml:"xmlns:w,attr"`
	XmlnsR     string       `xml:"xmlns:r,attr"`
	Paragraphs []*Paragraph `xml:"w:p"`
}

// Footer represents a Word footer document (footer1.xml, footer2.xml, etc.)
type Footer struct {
	XMLName    xml.Name     `xml:"w:ftr"`
	Xmlns      string       `xml:"xmlns:w,attr"`
	XmlnsR     string       `xml:"xmlns:r,attr"`
	Paragraphs []*Paragraph `xml:"w:p"`
}

// NewHeader creates a new header document.
func NewHeader() *Header {
	return &Header{
		Xmlns:      "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:     "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		Paragraphs: make([]*Paragraph, 0),
	}
}

// NewFooter creates a new footer document.
func NewFooter() *Footer {
	return &Footer{
		Xmlns:      "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:     "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		Paragraphs: make([]*Paragraph, 0),
	}
}

// AddParagraph adds a paragraph to the header.
func (h *Header) AddParagraph(p *Paragraph) {
	h.Paragraphs = append(h.Paragraphs, p)
}

// AddParagraph adds a paragraph to the footer.
func (f *Footer) AddParagraph(p *Paragraph) {
	f.Paragraphs = append(f.Paragraphs, p)
}
