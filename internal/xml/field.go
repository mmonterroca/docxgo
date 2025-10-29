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
