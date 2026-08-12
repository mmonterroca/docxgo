// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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
//
// Field order is the CT_TblPrBase sequence (tblStyle, tblW, jc, tblBorders,
// tblLook) and is load-bearing: this package has no custom MarshalXML, so
// struct order is document order, and CT_TblPrBase is an xsd:sequence -- a
// misordered child is a schema error, not a cosmetic difference.
type TableProperties struct {
	XMLName xml.Name           `xml:"w:tblPr"`
	Style   *TableStyle        `xml:"w:tblStyle,omitempty"`
	Width   *TableWidth        `xml:"w:tblW,omitempty"`
	Jc      *Justification     `xml:"w:jc,omitempty"`
	Borders *TableLevelBorders `xml:"w:tblBorders,omitempty"`
	Look    *TableLook         `xml:"w:tblLook,omitempty"`
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

// TableLook represents w:tblLook element that carries default visual hints for table features.
type TableLook struct {
	XMLName     xml.Name `xml:"w:tblLook"`
	Val         string   `xml:"w:val,attr,omitempty"`
	FirstRow    string   `xml:"w:firstRow,attr,omitempty"`
	LastRow     string   `xml:"w:lastRow,attr,omitempty"`
	FirstColumn string   `xml:"w:firstColumn,attr,omitempty"`
	LastColumn  string   `xml:"w:lastColumn,attr,omitempty"`
	NoHBand     string   `xml:"w:noHBand,attr,omitempty"`
	NoVBand     string   `xml:"w:noVBand,attr,omitempty"`
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
	Content    []interface{}        `xml:",any"`
}

// TableCellProperties represents w:tcPr element.
//
// Field order is the CT_TcPrBase sequence (tcW, gridSpan, vMerge, tcBorders,
// shd, vAlign). See TableProperties for why the order matters.
type TableCellProperties struct {
	XMLName  xml.Name       `xml:"w:tcPr"`
	Width    *TableWidth    `xml:"w:tcW,omitempty"`
	GridSpan *GridSpan      `xml:"w:gridSpan,omitempty"`
	VMerge   *VMerge        `xml:"w:vMerge,omitempty"`
	Borders  *TableBorders  `xml:"w:tcBorders,omitempty"`
	Shading  *Shading       `xml:"w:shd,omitempty"`
	VAlign   *VerticalAlign `xml:"w:vAlign,omitempty"`
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

// TableLevelBorders represents w:tblBorders element (used in table properties and table styles).
type TableLevelBorders struct {
	XMLName xml.Name `xml:"w:tblBorders"`
	Top     *Border  `xml:"w:top,omitempty"`
	Left    *Border  `xml:"w:left,omitempty"`
	Bottom  *Border  `xml:"w:bottom,omitempty"`
	Right   *Border  `xml:"w:right,omitempty"`
	InsideH *Border  `xml:"w:insideH,omitempty"`
	InsideV *Border  `xml:"w:insideV,omitempty"`
}

// Border represents a border element.
type Border struct {
	Val   string `xml:"w:val,attr"`
	Sz    int    `xml:"w:sz,attr,omitempty"`
	Space int    `xml:"w:space,attr,omitempty"`
	Color string `xml:"w:color,attr,omitempty"`
}

// Shading represents w:shd element.
//
// Attribute order follows CT_Shd: val, color, fill, then the theme*
// attributes -- themeFill and its tint/shade are the only theme attributes
// this package ever writes (see internal/reader's applyCellShading and
// internal/core.tableCell.ThemeFill for why a w:themeColor source still ends
// up here rather than on its own field).
type Shading struct {
	Val            string `xml:"w:val,attr,omitempty"`
	Color          string `xml:"w:color,attr,omitempty"`
	Fill           string `xml:"w:fill,attr,omitempty"`
	ThemeFill      string `xml:"w:themeFill,attr,omitempty"`
	ThemeFillTint  string `xml:"w:themeFillTint,attr,omitempty"`
	ThemeFillShade string `xml:"w:themeFillShade,attr,omitempty"`
}
