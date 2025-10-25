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

// Table represents w:tbl element.
type Table struct {
	XMLName    xml.Name         `xml:"w:tbl"`
	Properties *TableProperties `xml:"w:tblPr,omitempty"`
	Grid       *TableGrid       `xml:"w:tblGrid,omitempty"`
	Rows       []*TableRow      `xml:"w:tr"`
}

// TableProperties represents w:tblPr element.
type TableProperties struct {
	XMLName xml.Name       `xml:"w:tblPr"`
	Width   *TableWidth    `xml:"w:tblW,omitempty"`
	Jc      *Justification `xml:"w:jc,omitempty"`
	Style   *Style         `xml:"w:tblStyle,omitempty"`
}

// TableWidth represents w:tblW element.
type TableWidth struct {
	Type string `xml:"w:type,attr"`
	W    int    `xml:"w:w,attr"`
}

// TableGrid represents w:tblGrid element.
type TableGrid struct {
	XMLName xml.Name   `xml:"w:tblGrid"`
	Cols    []*GridCol `xml:"w:gridCol"`
}

// GridCol represents w:gridCol element.
type GridCol struct {
	W *int `xml:"w:w,attr,omitempty"`
}

// TableRow represents w:tr element.
type TableRow struct {
	XMLName    xml.Name            `xml:"w:tr"`
	Properties *TableRowProperties `xml:"w:trPr,omitempty"`
	Cells      []*TableCell        `xml:"w:tc"`
}

// TableRowProperties represents w:trPr element.
type TableRowProperties struct {
	XMLName xml.Name        `xml:"w:trPr"`
	Height  *TableRowHeight `xml:"w:trHeight,omitempty"`
}

// TableRowHeight represents w:trHeight element.
type TableRowHeight struct {
	Val  int    `xml:"w:val,attr"`
	Rule string `xml:"w:hRule,attr,omitempty"`
}

// TableCell represents w:tc element.
type TableCell struct {
	XMLName    xml.Name             `xml:"w:tc"`
	Properties *TableCellProperties `xml:"w:tcPr,omitempty"`
	Paragraphs []*Paragraph         `xml:"w:p"`
}

// TableCellProperties represents w:tcPr element.
type TableCellProperties struct {
	XMLName xml.Name       `xml:"w:tcPr"`
	Width   *TableWidth    `xml:"w:tcW,omitempty"`
	VAlign  *VerticalAlign `xml:"w:vAlign,omitempty"`
	Borders *TableBorders  `xml:"w:tcBorders,omitempty"`
	Shading *Shading       `xml:"w:shd,omitempty"`
}

// VerticalAlign represents w:vAlign element.
type VerticalAlign struct {
	Val string `xml:"w:val,attr"`
}

// TableBorders represents w:tcBorders element.
type TableBorders struct {
	XMLName xml.Name `xml:"w:tcBorders"`
	Top     *Border  `xml:"w:top,omitempty"`
	Left    *Border  `xml:"w:left,omitempty"`
	Bottom  *Border  `xml:"w:bottom,omitempty"`
	Right   *Border  `xml:"w:right,omitempty"`
}

// Border represents a border element.
type Border struct {
	Val   string `xml:"w:val,attr"`
	Sz    int    `xml:"w:sz,attr,omitempty"`
	Color string `xml:"w:color,attr,omitempty"`
}

// Shading represents w:shd element.
type Shading struct {
	Val   string `xml:"w:val,attr,omitempty"`
	Color string `xml:"w:color,attr,omitempty"`
	Fill  string `xml:"w:fill,attr,omitempty"`
}
