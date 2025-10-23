/*
   Copyright (c) 2025 SlideLang Enhanced Fork

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

package docx

import (
	"encoding/xml"
	"strconv"
	"sync/atomic"
)

// Header represents a header XML structure
type Header struct {
	XMLName xml.Name `xml:"w:hdr" json:"-"`
	XMLW    string   `xml:"xmlns:w,attr"`
	XMLR    string   `xml:"xmlns:r,attr"`
	Items   []interface{}
}

// Footer represents a footer XML structure
type Footer struct {
	XMLName xml.Name `xml:"w:ftr" json:"-"`
	XMLW    string   `xml:"xmlns:w,attr"`
	XMLR    string   `xml:"xmlns:r,attr"`
	Items   []interface{}
}

// HeaderFooterType defines the type of header or footer
type HeaderFooterType string

const (
	// HeaderFooterDefault is for all pages except first and even (if different)
	HeaderFooterDefault HeaderFooterType = "default"
	// HeaderFooterFirst is for the first page only
	HeaderFooterFirst HeaderFooterType = "first"
	// HeaderFooterEven is for even pages when using different odd/even headers
	HeaderFooterEven HeaderFooterType = "even"
)

// AddHeader adds a header to the document
// Returns a paragraph that can be formatted
func (f *Docx) AddHeader(headerType HeaderFooterType) *Paragraph {
	// Ensure headers map exists
	if f.headers == nil {
		f.headers = make(map[HeaderFooterType]*Header)
	}
	
	// Create header structure
	header := &Header{
		XMLW:  XMLNS_W,
		XMLR:  XMLNS_R,
		Items: make([]interface{}, 0, 8),
	}
	
	// Store header
	f.headers[headerType] = header
	
	// Create and return a paragraph that will be added to this header
	p := &Paragraph{
		Children: make([]interface{}, 0, 64),
		file:     f,
	}
	
	header.Items = append(header.Items, p)
	
	// Add relationship for this header
	rID := f.addHeaderRelationship(headerType)
	
	// Add sectPr reference to document body if not exists
	f.addHeaderReference(headerType, rID)
	
	return p
}

// AddFooter adds a footer to the document
// Returns a paragraph that can be formatted
func (f *Docx) AddFooter(footerType HeaderFooterType) *Paragraph {
	// Ensure footers map exists
	if f.footers == nil {
		f.footers = make(map[HeaderFooterType]*Footer)
	}
	
	// Create footer structure
	footer := &Footer{
		XMLW:  XMLNS_W,
		XMLR:  XMLNS_R,
		Items: make([]interface{}, 0, 8),
	}
	
	// Store footer
	f.footers[footerType] = footer
	
	// Create and return a paragraph
	p := &Paragraph{
		Children: make([]interface{}, 0, 64),
		file:     f,
	}
	
	footer.Items = append(footer.Items, p)
	
	// Add relationship for this footer
	rID := f.addFooterRelationship(footerType)
	
	// Add sectPr reference to document body if not exists
	f.addFooterReference(footerType, rID)
	
	return p
}

// addHeaderRelationship adds a header relationship and returns the rID
func (f *Docx) addHeaderRelationship(headerType HeaderFooterType) string {
	rID := "rId" + strconv.Itoa(int(atomic.AddUintptr(&f.rID, 1)))
	
	filename := "header1.xml"
	switch headerType {
	case HeaderFooterFirst:
		filename = "header2.xml"
	case HeaderFooterEven:
		filename = "header3.xml"
	}
	
	f.docRelation.Relationship = append(f.docRelation.Relationship, Relationship{
		ID:     rID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header",
		Target: filename,
	})
	
	return rID
}

// addFooterRelationship adds a footer relationship and returns the rID
func (f *Docx) addFooterRelationship(footerType HeaderFooterType) string {
	rID := "rId" + strconv.Itoa(int(atomic.AddUintptr(&f.rID, 1)))
	
	filename := "footer1.xml"
	switch footerType {
	case HeaderFooterFirst:
		filename = "footer2.xml"
	case HeaderFooterEven:
		filename = "footer3.xml"
	}
	
	f.docRelation.Relationship = append(f.docRelation.Relationship, Relationship{
		ID:     rID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer",
		Target: filename,
	})
	
	return rID
}

// addHeaderReference adds header reference to sectPr
func (f *Docx) addHeaderReference(headerType HeaderFooterType, rID string) {
	// Get or create sectPr
	sectPr := f.getSectPr()
	
	// Determine type string
	typeStr := "default"
	switch headerType {
	case HeaderFooterFirst:
		typeStr = "first"
	case HeaderFooterEven:
		typeStr = "even"
	}
	
	// Add header reference
	sectPr.HeaderReference = append(sectPr.HeaderReference, HeaderReference{
		Type: typeStr,
		RID:  rID,
	})
}

// addFooterReference adds footer reference to sectPr
func (f *Docx) addFooterReference(footerType HeaderFooterType, rID string) {
	// Get or create sectPr
	sectPr := f.getSectPr()
	
	// Determine type string
	typeStr := "default"
	switch footerType {
	case HeaderFooterFirst:
		typeStr = "first"
	case HeaderFooterEven:
		typeStr = "even"
	}
	
	// Add footer reference
	sectPr.FooterReference = append(sectPr.FooterReference, FooterReference{
		Type: typeStr,
		RID:  rID,
	})
}

// getSectPr gets or creates the sectPr element in the document body
func (f *Docx) getSectPr() *SectPr {
	// Check if last item is already a sectPr
	if len(f.Document.Body.Items) > 0 {
		if sectPr, ok := f.Document.Body.Items[len(f.Document.Body.Items)-1].(*SectPr); ok {
			return sectPr
		}
	}
	
	// Create new sectPr
	sectPr := &SectPr{}
	f.Document.Body.Items = append(f.Document.Body.Items, sectPr)
	return sectPr
}

// AddPageNumberFooter is a convenience method to add a simple page number footer
func (f *Docx) AddPageNumberFooter() *Paragraph {
	footer := f.AddFooter(HeaderFooterDefault)
	footer.AddText("Page ")
	footer.AddPageField()
	footer.AddText(" of ")
	footer.AddNumPagesField()
	footer.Justification("center")
	return footer
}

// AddDocumentTitleHeader is a convenience method to add a document title header
func (f *Docx) AddDocumentTitleHeader(title string) *Paragraph {
	header := f.AddHeader(HeaderFooterDefault)
	header.AddText(title).Size("20").Color("666666")
	header.Justification("center")
	return header
}