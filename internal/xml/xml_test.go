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

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestFieldChar_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		field    *FieldChar
		wantXML  string
	}{
		{
			name:    "Begin",
			field:   NewFieldBegin(),
			wantXML: `<w:fldChar w:fldCharType="begin"></w:fldChar>`,
		},
		{
			name:    "Separate",
			field:   NewFieldSeparate(),
			wantXML: `<w:fldChar w:fldCharType="separate"></w:fldChar>`,
		},
		{
			name:    "End",
			field:   NewFieldEnd(),
			wantXML: `<w:fldChar w:fldCharType="end"></w:fldChar>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := xml.Marshal(tt.field)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(data) != tt.wantXML {
				t.Errorf("Marshal() = %s, want %s", string(data), tt.wantXML)
			}
		})
	}
}

func TestInstrText_Marshal(t *testing.T) {
	instr := NewInstrText("PAGE")
	
	data, err := xml.Marshal(instr)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "PAGE") {
		t.Errorf("Marshal() should contain 'PAGE', got: %s", xmlStr)
	}
	if !strings.Contains(xmlStr, `xml:space="preserve"`) {
		t.Errorf("Marshal() should contain xml:space=\"preserve\", got: %s", xmlStr)
	}
}

func TestSectionProperties_Marshal(t *testing.T) {
	sectPr := NewSectionProperties()
	sectPr.SetPageSize(12240, 15840, false) // Letter size
	sectPr.SetPageMargins(1440, 1440, 1440, 1440, 720, 720)
	sectPr.SetColumns(2)

	data, err := xml.Marshal(sectPr)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	// Check page size
	if !strings.Contains(xmlStr, `w:w="12240"`) {
		t.Error("Should contain page width")
	}
	if !strings.Contains(xmlStr, `w:h="15840"`) {
		t.Error("Should contain page height")
	}
	if !strings.Contains(xmlStr, `w:orient="portrait"`) {
		t.Error("Should contain orientation")
	}

	// Check margins
	if !strings.Contains(xmlStr, `w:top="1440"`) {
		t.Error("Should contain top margin")
	}

	// Check columns
	if !strings.Contains(xmlStr, `w:num="2"`) {
		t.Error("Should contain column count")
	}
}

func TestSectionProperties_HeaderFooterRefs(t *testing.T) {
	sectPr := NewSectionProperties()
	sectPr.AddHeaderRef("default", "rId1")
	sectPr.AddFooterRef("default", "rId2")

	data, err := xml.Marshal(sectPr)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `w:headerReference`) {
		t.Error("Should contain header reference")
	}
	if !strings.Contains(xmlStr, `w:footerReference`) {
		t.Error("Should contain footer reference")
	}
	if !strings.Contains(xmlStr, `r:id="rId1"`) {
		t.Error("Should contain header relationship ID")
	}
	if !strings.Contains(xmlStr, `r:id="rId2"`) {
		t.Error("Should contain footer relationship ID")
	}
}

func TestHeader_Marshal(t *testing.T) {
	header := NewHeader()
	
	// Add a simple paragraph
	para := &Paragraph{
		Runs: []*Run{
			{
				Text: &Text{
					Space:   "preserve",
					Content: "Header Text",
				},
			},
		},
	}
	header.AddParagraph(para)

	data, err := xml.Marshal(header)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `<w:hdr`) {
		t.Error("Should contain hdr element")
	}
	if !strings.Contains(xmlStr, "Header Text") {
		t.Error("Should contain header text")
	}
	if !strings.Contains(xmlStr, `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`) {
		t.Error("Should contain namespace declaration")
	}
}

func TestFooter_Marshal(t *testing.T) {
	footer := NewFooter()
	
	// Add a simple paragraph
	para := &Paragraph{
		Runs: []*Run{
			{
				Text: &Text{
					Space:   "preserve",
					Content: "Footer Text",
				},
			},
		},
	}
	footer.AddParagraph(para)

	data, err := xml.Marshal(footer)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `<w:ftr`) {
		t.Error("Should contain ftr element")
	}
	if !strings.Contains(xmlStr, "Footer Text") {
		t.Error("Should contain footer text")
	}
}

func TestStyles_Marshal(t *testing.T) {
	styles := NewStyles()
	
	// Add Normal style
	normalStyle := NewParagraphStyle("Normal", "Normal", true)
	styles.AddStyle(normalStyle)
	
	// Add Heading1 style
	heading1 := NewParagraphStyle("Heading1", "Heading 1", false)
	heading1.BasedOn = &BasedOn{Val: "Normal"}
	heading1.ParaProps = &ParagraphProperties{
		OutlineLevel: &OutlineLevel{Val: 1},
	}
	styles.AddStyle(heading1)

	data, err := xml.Marshal(styles)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `<w:styles`) {
		t.Error("Should contain styles element")
	}
	if !strings.Contains(xmlStr, `w:type="paragraph"`) {
		t.Error("Should contain paragraph type")
	}
	if !strings.Contains(xmlStr, `w:styleId="Normal"`) {
		t.Error("Should contain Normal style ID")
	}
	if !strings.Contains(xmlStr, `w:default="true"`) {
		t.Error("Should mark Normal as default")
	}
	if !strings.Contains(xmlStr, `w:val="Normal"`) {
		t.Error("Should contain style name")
	}
}

func TestCharacterStyle_Marshal(t *testing.T) {
	styles := NewStyles()
	
	emphasis := NewCharacterStyle("Emphasis", "Emphasis", false)
	emphasis.RunProps = &RunProperties{
		Italic: &BoolValue{},
	}
	styles.AddStyle(emphasis)

	data, err := xml.Marshal(styles)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `w:type="character"`) {
		t.Error("Should contain character type")
	}
	if !strings.Contains(xmlStr, `w:styleId="Emphasis"`) {
		t.Error("Should contain Emphasis style ID")
	}
}

func TestRun_WithFieldChar(t *testing.T) {
	run := &Run{
		FieldChar: NewFieldBegin(),
	}

	data, err := xml.Marshal(run)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `<w:r>`) {
		t.Error("Should contain run element")
	}
	if !strings.Contains(xmlStr, `<w:fldChar`) {
		t.Error("Should contain field character")
	}
	if !strings.Contains(xmlStr, `w:fldCharType="begin"`) {
		t.Error("Should contain field type")
	}
}

func TestRun_WithInstrText(t *testing.T) {
	run := &Run{
		InstrText: NewInstrText("PAGE \\* Arabic"),
	}

	data, err := xml.Marshal(run)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	xmlStr := string(data)
	
	if !strings.Contains(xmlStr, `<w:r>`) {
		t.Error("Should contain run element")
	}
	if !strings.Contains(xmlStr, `<w:instrText`) {
		t.Error("Should contain instruction text")
	}
	if !strings.Contains(xmlStr, "PAGE") {
		t.Error("Should contain PAGE instruction")
	}
}

func TestComplexField_Sequence(t *testing.T) {
	// A complex field consists of:
	// 1. Run with fldChar begin
	// 2. Run with instrText
	// 3. Run with fldChar separate
	// 4. Run with result text
	// 5. Run with fldChar end

	runs := []*Run{
		{FieldChar: NewFieldBegin()},
		{InstrText: NewInstrText("PAGE")},
		{FieldChar: NewFieldSeparate()},
		{Text: &Text{Content: "1"}}, // Result
		{FieldChar: NewFieldEnd()},
	}

	for i, run := range runs {
		data, err := xml.Marshal(run)
		if err != nil {
			t.Fatalf("Run %d Marshal() error = %v", i, err)
		}

		xmlStr := string(data)
		if !strings.Contains(xmlStr, `<w:r>`) {
			t.Errorf("Run %d should contain run element", i)
		}
	}
}
