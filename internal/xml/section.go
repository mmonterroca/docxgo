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

package xml

import "encoding/xml"

// SectionProperties represents w:sectPr element (section properties).
type SectionProperties struct {
	XMLName     xml.Name     `xml:"w:sectPr"`
	PageSize    *PageSize    `xml:"w:pgSz,omitempty"`
	PageMargins *PageMargins `xml:"w:pgMar,omitempty"`
	Columns     *Columns     `xml:"w:cols,omitempty"`
	HeaderRef   []HeaderRef  `xml:"w:headerReference,omitempty"`
	FooterRef   []FooterRef  `xml:"w:footerReference,omitempty"`
	Type        *SectionType `xml:"w:type,omitempty"`
}

// PageSize represents w:pgSz element (page size).
type PageSize struct {
	XMLName xml.Name `xml:"w:pgSz"`
	Width   int      `xml:"w:w,attr"`              // Width in twips
	Height  int      `xml:"w:h,attr"`              // Height in twips
	Orient  string   `xml:"w:orient,attr,omitempty"` // portrait or landscape
}

// PageMargins represents w:pgMar element (page margins).
type PageMargins struct {
	XMLName xml.Name `xml:"w:pgMar"`
	Top     int      `xml:"w:top,attr"`
	Right   int      `xml:"w:right,attr"`
	Bottom  int      `xml:"w:bottom,attr"`
	Left    int      `xml:"w:left,attr"`
	Header  int      `xml:"w:header,attr"`
	Footer  int      `xml:"w:footer,attr"`
	Gutter  int      `xml:"w:gutter,attr,omitempty"`
}

// Columns represents w:cols element (column definition).
type Columns struct {
	XMLName xml.Name `xml:"w:cols"`
	Num     int      `xml:"w:num,attr,omitempty"`     // Number of columns
	Space   int      `xml:"w:space,attr,omitempty"`   // Space between columns
	Sep     *bool    `xml:"w:sep,attr,omitempty"`     // Draw separator line
}

// HeaderRef represents w:headerReference element.
type HeaderRef struct {
	XMLName xml.Name `xml:"w:headerReference"`
	Type    string   `xml:"w:type,attr"` // default, first, even
	ID      string   `xml:"r:id,attr"`   // Relationship ID
}

// FooterRef represents w:footerReference element.
type FooterRef struct {
	XMLName xml.Name `xml:"w:footerReference"`
	Type    string   `xml:"w:type,attr"` // default, first, even
	ID      string   `xml:"r:id,attr"`   // Relationship ID
}

// SectionType represents w:type element (section type).
type SectionType struct {
	XMLName xml.Name `xml:"w:type"`
	Val     string   `xml:"w:val,attr"` // nextPage, continuous, evenPage, oddPage
}

// NewSectionProperties creates a new section properties element.
func NewSectionProperties() *SectionProperties {
	return &SectionProperties{
		HeaderRef: make([]HeaderRef, 0),
		FooterRef: make([]FooterRef, 0),
	}
}

// SetPageSize sets the page size for the section.
func (sp *SectionProperties) SetPageSize(width, height int, landscape bool) {
	orient := "portrait"
	if landscape {
		orient = "landscape"
	}
	sp.PageSize = &PageSize{
		Width:  width,
		Height: height,
		Orient: orient,
	}
}

// SetPageMargins sets the page margins for the section.
func (sp *SectionProperties) SetPageMargins(top, right, bottom, left, header, footer int) {
	sp.PageMargins = &PageMargins{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
		Header: header,
		Footer: footer,
	}
}

// SetColumns sets the number of columns for the section.
func (sp *SectionProperties) SetColumns(num int) {
	sp.Columns = &Columns{
		Num: num,
	}
}

// AddHeaderRef adds a header reference.
func (sp *SectionProperties) AddHeaderRef(headerType, rID string) {
	sp.HeaderRef = append(sp.HeaderRef, HeaderRef{
		Type: headerType,
		ID:   rID,
	})
}

// AddFooterRef adds a footer reference.
func (sp *SectionProperties) AddFooterRef(footerType, rID string) {
	sp.FooterRef = append(sp.FooterRef, FooterRef{
		Type: footerType,
		ID:   rID,
	})
}
