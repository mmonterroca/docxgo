package xml

/*
MIT License

Copyright (c) 2025 Misael Monterroca

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

import "encoding/xml"

// Run represents a w:r element (run of text with properties).
type Run struct {
	XMLName    xml.Name       `xml:"w:r"`
	Properties *RunProperties `xml:"w:rPr,omitempty"`

	// Content can be text, fields, tabs, breaks, or drawings
	// Using interface{} with custom marshaling for flexibility
	Text    *Text     `xml:"w:t,omitempty"`
	Tab     *struct{} `xml:"w:tab,omitempty"`
	Break   *Break    `xml:"w:br,omitempty"`
	Drawing *Drawing  `xml:"w:drawing,omitempty"`

	// Field support - complex fields use multiple runs
	FieldChar *FieldChar `xml:"w:fldChar,omitempty"`
	InstrText *InstrText `xml:"w:instrText,omitempty"`
}

// RunProperties represents w:rPr element (run properties).
type RunProperties struct {
	XMLName   xml.Name   `xml:"w:rPr"`
	Style     *RunStyle  `xml:"w:rStyle,omitempty"`
	Font      *Font      `xml:"w:rFonts,omitempty"`
	Bold      *BoolValue `xml:"w:b,omitempty"`
	Italic    *BoolValue `xml:"w:i,omitempty"`
	Strike    *BoolValue `xml:"w:strike,omitempty"`
	Color     *Color     `xml:"w:color,omitempty"`
	Size      *HalfPt    `xml:"w:sz,omitempty"`
	SizeCS    *HalfPt    `xml:"w:szCs,omitempty"` // Complex script size
	Underline *Underline `xml:"w:u,omitempty"`
	Highlight *Highlight `xml:"w:highlight,omitempty"`
	Lang      *Language  `xml:"w:lang,omitempty"`
}

// Text represents w:t element (text content).
type Text struct {
	XMLName xml.Name `xml:"w:t"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Content string   `xml:",chardata"`
}

// BoolValue represents a boolean property.
type BoolValue struct {
	Val *bool `xml:"w:val,attr,omitempty"`
}

// Underline represents w:u element (underline).
type Underline struct {
	Val string `xml:"w:val,attr"`
}

// Color represents w:color element.
type Color struct {
	Val string `xml:"w:val,attr"`
}

// HalfPt represents font size in half-points.
type HalfPt struct {
	Val int `xml:"w:val,attr"`
}

// Font represents w:rFonts element.
type Font struct {
	ASCII    string `xml:"w:ascii,attr,omitempty"`
	HAnsi    string `xml:"w:hAnsi,attr,omitempty"`
	EastAsia string `xml:"w:eastAsia,attr,omitempty"`
	CS       string `xml:"w:cs,attr,omitempty"`
}

// Language represents w:lang element.
type Language struct {
	Val      string `xml:"w:val,attr,omitempty"`
	EastAsia string `xml:"w:eastAsia,attr,omitempty"`
	Bidi     string `xml:"w:bidi,attr,omitempty"`
}

// Highlight represents w:highlight element.
type Highlight struct {
	Val string `xml:"w:val,attr"`
}

// RunStyle represents w:rStyle element (run style reference).
type RunStyle struct {
	Val string `xml:"w:val,attr"`
}

// Break represents w:br element (line break).
type Break struct {
	Type string `xml:"w:type,attr,omitempty"`
}
