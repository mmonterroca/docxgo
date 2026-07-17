// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package xml

import "encoding/xml"

// FieldSimple represents a w:fldSimple element for simple fields.
// Complex fields use: w:fldChar (begin) -> w:instrText -> w:fldChar (separate) -> result -> w:fldChar (end)
type FieldSimple struct {
	XMLName xml.Name `xml:"w:fldSimple"`
	Instr   string   `xml:"w:instr,attr"`
	Text    *Text    `xml:"w:t,omitempty"`
}

// FieldChar represents w:fldChar element (field character).
type FieldChar struct {
	XMLName xml.Name `xml:"w:fldChar"`
	FldType string   `xml:"w:fldCharType,attr"` // begin, separate, end
	Dirty   *bool    `xml:"w:dirty,attr,omitempty"`
	FldLock *bool    `xml:"w:fldLock,attr,omitempty"`
}

// InstrText represents w:instrText element (field instruction text).
type InstrText struct {
	XMLName xml.Name `xml:"w:instrText"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Content string   `xml:",chardata"`
}

// NewFieldBegin creates a field begin character.
func NewFieldBegin() *FieldChar {
	return &FieldChar{
		FldType: "begin",
	}
}

// NewFieldSeparate creates a field separate character.
func NewFieldSeparate() *FieldChar {
	return &FieldChar{
		FldType: "separate",
	}
}

// NewFieldEnd creates a field end character.
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
