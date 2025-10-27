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
