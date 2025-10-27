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
	Style   *TableStyle    `xml:"w:tblStyle,omitempty"`
	Width   *TableWidth    `xml:"w:tblW,omitempty"`
	Jc      *Justification `xml:"w:jc,omitempty"`
}

// TableStyle represents w:tblStyle element.
type TableStyle struct {
	Val string `xml:"w:val,attr"`
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
	Tables     []*Table             `xml:"w:tbl,omitempty"` // Nested tables
}

// TableCellProperties represents w:tcPr element.
type TableCellProperties struct {
	XMLName  xml.Name       `xml:"w:tcPr"`
	Width    *TableWidth    `xml:"w:tcW,omitempty"`
	GridSpan *GridSpan      `xml:"w:gridSpan,omitempty"`
	VMerge   *VMerge        `xml:"w:vMerge,omitempty"`
	VAlign   *VerticalAlign `xml:"w:vAlign,omitempty"`
	Borders  *TableBorders  `xml:"w:tcBorders,omitempty"`
	Shading  *Shading       `xml:"w:shd,omitempty"`
}

// GridSpan represents w:gridSpan element for horizontal cell merging.
type GridSpan struct {
	Val int `xml:"w:val,attr"`
}

// VMerge represents w:vMerge element for vertical cell merging.
type VMerge struct {
	Val string `xml:"w:val,attr,omitempty"` // "restart" or omitted for "continue"
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
