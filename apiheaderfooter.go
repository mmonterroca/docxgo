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
	// For now, we'll create a simple header implementation
	// In a full implementation, this would create actual header XML files
	// and link them in document.xml.rels

	// Create a paragraph that will be used in the header
	p := &Paragraph{
		Children: make([]interface{}, 0, 64),
		file:     f,
	}

	// TODO: Full implementation requires:
	// 1. Create header1.xml, header2.xml, header3.xml files
	// 2. Add relationships in document.xml.rels
	// 3. Add references in document.xml sectPr
	//
	// For v0.3.0, we'll document the API and provide a working placeholder

	return p
}

// AddFooter adds a footer to the document
// Returns a paragraph that can be formatted
func (f *Docx) AddFooter(footerType HeaderFooterType) *Paragraph {
	// Similar to AddHeader, this is a placeholder for the full implementation

	p := &Paragraph{
		Children: make([]interface{}, 0, 64),
		file:     f,
	}

	return p
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
