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

// Field represents a w:fldSimple or w:fldChar complex field.
// Complex fields use: w:fldChar (begin) -> w:instrText -> w:fldChar (separate) -> result -> w:fldChar (end)
type FieldSimple struct {
	XMLName xml.Name `xml:"w:fldSimple"`
	Instr   string   `xml:"w:instr,attr"`
	Text    *Text    `xml:"w:t,omitempty"`
}

// FieldChar represents w:fldChar element (field character).
type FieldChar struct {
	XMLName  xml.Name `xml:"w:fldChar"`
	FldType  string   `xml:"w:fldCharType,attr"` // begin, separate, end
	Dirty    *bool    `xml:"w:dirty,attr,omitempty"`
	FldLock  *bool    `xml:"w:fldLock,attr,omitempty"`
}

// InstrText represents w:instrText element (field instruction text).
type InstrText struct {
	XMLName xml.Name `xml:"w:instrText"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Content string   `xml:",chardata"`
}

// FieldBegin creates a field begin character.
func NewFieldBegin() *FieldChar {
	return &FieldChar{
		FldType: "begin",
	}
}

// FieldSeparate creates a field separate character.
func NewFieldSeparate() *FieldChar {
	return &FieldChar{
		FldType: "separate",
	}
}

// FieldEnd creates a field end character.
func NewFieldEnd() *FieldChar {
	return &FieldChar{
		FldType: "end",
	}
}

// NewInstrText creates a field instruction text element.
func NewInstrText(instruction string) *InstrText {
	return &InstrText{
		Space:   "preserve",
		Content: instruction,
	}
}
