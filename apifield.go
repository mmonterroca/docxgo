/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)
   Copyright (c) 2025 SlideLang Enhanced Fork

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

package docx

import (
	"encoding/xml"
)

// FldChar represents a w:fldChar element in Word XML
// Used for field codes like TOC, PAGEREF, etc.
type FldChar struct {
	XMLName     xml.Name `xml:"w:fldChar"`
	FldCharType string   `xml:"w:fldCharType,attr"` // "begin", "separate", "end"
}

// InstrText represents a w:instrText element in Word XML
// Contains the field instruction text
type InstrText struct {
	XMLName xml.Name `xml:"w:instrText"`
	Text    string   `xml:",chardata"`
}

// Field represents a complete Word field with all its components
// A field consists of: fldChar(begin) -> instrText -> fldChar(separate) -> result text -> fldChar(end)
type Field struct {
	Begin     *Run // Run with fldChar type="begin"
	InstrText *Run // Run with instrText containing the field instruction
	Separate  *Run // Run with fldChar type="separate"
	Result    *Run // Run with the field result text
	End       *Run // Run with fldChar type="end"
}

// NewFldChar creates a new field character
func NewFldChar(fldCharType string) *FldChar {
	return &FldChar{
		FldCharType: fldCharType,
	}
}

// NewInstrText creates a new instruction text element
func NewInstrText(text string) *InstrText {
	return &InstrText{
		Text: text,
	}
}

// AddFldChar adds a field character to a run
func (r *Run) AddFldChar(fldCharType string) *FldChar {
	fldChar := NewFldChar(fldCharType)
	r.Children = append(r.Children, fldChar)
	return fldChar
}

// AddInstrText adds instruction text to a run
func (r *Run) AddInstrText(text string) *InstrText {
	instrText := NewInstrText(text)
	r.Children = append(r.Children, instrText)
	return instrText
}

// AddField adds a complete field to a paragraph
// This creates the full field structure: begin -> instrText -> separate -> result -> end
func (p *Paragraph) AddField(instrText string, resultText string) *Field {
	// Begin field
	beginRun := &Run{file: p.file}
	beginRun.AddFldChar("begin")
	p.Children = append(p.Children, beginRun)

	// Instruction text
	instrRun := &Run{file: p.file}
	instrRun.AddInstrText(instrText)
	p.Children = append(p.Children, instrRun)

	// Separate field and result
	sepRun := &Run{file: p.file}
	sepRun.AddFldChar("separate")
	p.Children = append(p.Children, sepRun)

	// Result text (what's displayed before field is updated)
	resultRun := &Run{file: p.file}
	resultRun.AddText(resultText)
	p.Children = append(p.Children, resultRun)

	// End field
	endRun := &Run{file: p.file}
	endRun.AddFldChar("end")
	p.Children = append(p.Children, endRun)

	return &Field{
		Begin:     beginRun,
		InstrText: instrRun,
		Separate:  sepRun,
		Result:    resultRun,
		End:       endRun,
	}
}

// AddTOCField adds a Table of Contents field
func (p *Paragraph) AddTOCField(depth int, useHyperlinks bool, usePageNumbers bool) *Field {
	instrText := "TOC"

	// Add outline levels (e.g., \o "1-3" for levels 1-3)
	if depth > 0 {
		instrText += " \\o \"1-" + string(rune('0'+depth)) + "\""
	}

	// Add hyperlinks flag
	if useHyperlinks {
		instrText += " \\h"
	}

	// Add page numbers formatting
	if usePageNumbers {
		instrText += " \\z \\u"
	}

	return p.AddField(instrText, "Table of Contents")
}

// AddPageRefField adds a PAGEREF field for referencing page numbers of bookmarks
func (p *Paragraph) AddPageRefField(bookmarkName string, useHyperlink bool) *Field {
	instrText := "PAGEREF " + bookmarkName
	if useHyperlink {
		instrText += " \\h"
	}
	return p.AddField(instrText, "1")
}

// AddPageField adds a PAGE field for current page number
func (p *Paragraph) AddPageField() *Field {
	return p.AddField("PAGE", "1")
}

// AddNumPagesField adds a NUMPAGES field for total page count
func (p *Paragraph) AddNumPagesField() *Field {
	return p.AddField("NUMPAGES", "1")
}

// AddRefField adds a REF field for cross-references
func (p *Paragraph) AddRefField(bookmarkName string, useHyperlink bool) *Field {
	instrText := "REF " + bookmarkName
	if useHyperlink {
		instrText += " \\h"
	}
	return p.AddField(instrText, bookmarkName)
}

// AddSeqField adds a SEQ field for auto-numbering sequences (figures, tables, etc.)
func (p *Paragraph) AddSeqField(identifier string, format string) *Field {
	instrText := "SEQ " + identifier
	if format != "" {
		instrText += " \\* " + format
	}
	return p.AddField(instrText, "1")
}
