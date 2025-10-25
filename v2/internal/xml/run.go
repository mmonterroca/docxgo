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

// Package xml provides OOXML struct definitions for XML marshaling/unmarshaling.
package xml

import "encoding/xml"

// Run represents a w:r element (run of text with properties).
type Run struct {
	XMLName    xml.Name        `xml:"w:r"`
	Properties *RunProperties  `xml:"w:rPr,omitempty"`
	Text       *Text           `xml:"w:t,omitempty"`
	Tab        *struct{}       `xml:"w:tab,omitempty"`
	Break      *Break          `xml:"w:br,omitempty"`
	Drawing    *Drawing        `xml:"w:drawing,omitempty"`
}

// RunProperties represents w:rPr element (run properties).
type RunProperties struct {
	XMLName   xml.Name   `xml:"w:rPr"`
	Bold      *BoolValue `xml:"w:b,omitempty"`
	Italic    *BoolValue `xml:"w:i,omitempty"`
	Strike    *BoolValue `xml:"w:strike,omitempty"`
	Underline *Underline `xml:"w:u,omitempty"`
	Color     *Color     `xml:"w:color,omitempty"`
	Size      *HalfPt    `xml:"w:sz,omitempty"`
	SizeCS    *HalfPt    `xml:"w:szCs,omitempty"` // Complex script size
	Font      *Font      `xml:"w:rFonts,omitempty"`
	Highlight *Highlight `xml:"w:highlight,omitempty"`
}

// Text represents w:t element (text content).
type Text struct {
	XMLName   xml.Name `xml:"w:t"`
	Space     string   `xml:"xml:space,attr,omitempty"`
	Content   string   `xml:",chardata"`
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
	EastAsia string `xml:"w:eastAsia,attr,omitempty"`
	CS       string `xml:"w:cs,attr,omitempty"`
}

// Highlight represents w:highlight element.
type Highlight struct {
	Val string `xml:"w:val,attr"`
}

// Break represents w:br element (line break).
type Break struct {
	Type string `xml:"w:type,attr,omitempty"`
}

// Drawing represents w:drawing element (for images).
type Drawing struct {
	// TODO: Implement drawing structure
}
