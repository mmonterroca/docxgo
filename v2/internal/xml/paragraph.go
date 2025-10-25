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

// Paragraph represents w:p element.
type Paragraph struct {
	XMLName    xml.Name              `xml:"w:p"`
	Properties *ParagraphProperties  `xml:"w:pPr,omitempty"`
	Runs       []*Run                `xml:"w:r,omitempty"`
	Hyperlinks []*Hyperlink          `xml:"w:hyperlink,omitempty"`
}

// ParagraphProperties represents w:pPr element.
type ParagraphProperties struct {
	XMLName       xml.Name       `xml:"w:pPr"`
	Style         *Style         `xml:"w:pStyle,omitempty"`
	Justification *Justification `xml:"w:jc,omitempty"`
	Indentation   *Indentation   `xml:"w:ind,omitempty"`
	Spacing       *Spacing       `xml:"w:spacing,omitempty"`
}

// Style represents w:pStyle element.
type Style struct {
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
	Before      *int    `xml:"w:before,attr,omitempty"`
	After       *int    `xml:"w:after,attr,omitempty"`
	Line        *int    `xml:"w:line,attr,omitempty"`
	LineRule    *string `xml:"w:lineRule,attr,omitempty"`
}

// Hyperlink represents w:hyperlink element.
type Hyperlink struct {
	XMLName xml.Name `xml:"w:hyperlink"`
	ID      string   `xml:"r:id,attr"`
	Runs    []*Run   `xml:"w:r"`
}
