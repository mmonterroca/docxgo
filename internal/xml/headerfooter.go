// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package xml

import "encoding/xml"

// namespaceWordprocessingDrawing is the xmlns:wp namespace every <w:drawing>
// wrapper element (wp:inline, wp:anchor, wp:extent, wp:docPr, ...) carries.
// Document declares it; a header or footer holding an image needs it for the
// same reason, or the part is an undeclared-prefix error and Word offers to
// repair the package. Spelled out here rather than taken from pkg/constants
// because this package deliberately does not depend on it -- the two
// namespaces in NewHeader/NewFooter below are literals for the same reason.
//
// Note that xmlns:a and xmlns:pic are *not* needed here: a:graphic and pic:pic
// self-declare their prefixes on the drawing subtree itself (see drawing.go),
// so wp is the only one a containing part has to provide.
const namespaceWordprocessingDrawing = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"

// Header represents a Word header document (header1.xml, header2.xml, etc.)
//
// Content holds the header's top-level elements (*Paragraph, *Table) in
// document order, the same mixed-content pattern TableCell uses: there is no
// custom MarshalXML anywhere in this package, so ordering is slice order and
// each element's own XMLName decides its tag.
type Header struct {
	XMLName xml.Name `xml:"w:hdr"`
	Xmlns   string   `xml:"xmlns:w,attr"`
	XmlnsR  string   `xml:"xmlns:r,attr"`
	// XmlnsWP only matters to a header holding an image, but is declared
	// unconditionally: an unused namespace declaration is valid OOXML and
	// costs one attribute, whereas omitting it when a drawing *is* present
	// produces a package Word rejects.
	XmlnsWP string        `xml:"xmlns:wp,attr,omitempty"`
	Content []interface{} `xml:",any"`
}

// Footer represents a Word footer document (footer1.xml, footer2.xml, etc.)
//
// See Header's doc comment for the Content field's mixed-content contract and
// for why XmlnsWP is always declared.
type Footer struct {
	XMLName xml.Name      `xml:"w:ftr"`
	Xmlns   string        `xml:"xmlns:w,attr"`
	XmlnsR  string        `xml:"xmlns:r,attr"`
	XmlnsWP string        `xml:"xmlns:wp,attr,omitempty"`
	Content []interface{} `xml:",any"`
}

// NewHeader creates a new header document.
func NewHeader() *Header {
	return &Header{
		Xmlns:   "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:  "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		XmlnsWP: namespaceWordprocessingDrawing,
		Content: make([]interface{}, 0),
	}
}

// NewFooter creates a new footer document.
func NewFooter() *Footer {
	return &Footer{
		Xmlns:   "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR:  "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		XmlnsWP: namespaceWordprocessingDrawing,
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
