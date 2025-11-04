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

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func TestFieldChar_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		field   *FieldChar
		wantXML string
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
	if strings.Contains(xmlStr, `w:orient="portrait"`) {
		t.Error("Portrait orientation should omit explicit attribute")
	}

	// Check margins
	if !strings.Contains(xmlStr, `w:top="1440"`) {
		t.Error("Should contain top margin")
	}

	// Check columns
	if !strings.Contains(xmlStr, `w:num="2"`) {
		t.Error("Should contain column count")
	}
	if !strings.Contains(xmlStr, `w:space="720"`) {
		t.Error("Should contain default column spacing")
	}

	// Landscape orientation should emit attribute
	sectPrLandscape := NewSectionProperties()
	sectPrLandscape.SetPageSize(15840, 12240, true)
	dataLandscape, err := xml.Marshal(sectPrLandscape)
	if err != nil {
		t.Fatalf("Marshal() error for landscape = %v", err)
	}
	if !strings.Contains(string(dataLandscape), `w:orient="landscape"`) {
		t.Error("Landscape orientation should include orient attribute")
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
		Elements: []interface{}{
			&Run{
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
		Elements: []interface{}{
			&Run{
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
	heading1.ParaProps = &StyleParagraphProperties{
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

// Mock Image for testing
type mockImage struct {
	id              string
	format          domain.ImageFormat
	size            domain.ImageSize
	data            []byte
	relationshipID  string
	target          string
	description     string
	position        domain.ImagePosition
}

func (m *mockImage) ID() string                                  { return m.id }
func (m *mockImage) Format() domain.ImageFormat                  { return m.format }
func (m *mockImage) Size() domain.ImageSize                      { return m.size }
func (m *mockImage) SetSize(size domain.ImageSize) error         { m.size = size; return nil }
func (m *mockImage) Data() []byte                                { return m.data }
func (m *mockImage) RelationshipID() string                      { return m.relationshipID }
func (m *mockImage) Target() string                              { return m.target }
func (m *mockImage) Description() string                         { return m.description }
func (m *mockImage) SetDescription(desc string) error            { m.description = desc; return nil }
func (m *mockImage) Position() domain.ImagePosition              { return m.position }

func TestNewInlineDrawing(t *testing.T) {
	img := &mockImage{
		id:             "img1",
		format:         domain.ImageFormatPNG,
		size:           domain.NewImageSize(100, 100),
		relationshipID: "rId1",
		description:    "Test Image",
	}

	drawing := NewInlineDrawing(img, 1)

	if drawing == nil {
		t.Fatal("NewInlineDrawing() returned nil")
	}
	if drawing.Inline == nil {
		t.Fatal("NewInlineDrawing() Inline is nil")
	}
	if drawing.Inline.Extent == nil {
		t.Fatal("Inline.Extent is nil")
	}
	if drawing.Inline.Extent.Cx != img.size.WidthEMU {
		t.Errorf("Extent.Cx = %d; want %d", drawing.Inline.Extent.Cx, img.size.WidthEMU)
	}
	if drawing.Inline.Extent.Cy != img.size.HeightEMU {
		t.Errorf("Extent.Cy = %d; want %d", drawing.Inline.Extent.Cy, img.size.HeightEMU)
	}
	if drawing.Inline.DocPr.ID != 1 {
		t.Errorf("DocPr.ID = %d; want 1", drawing.Inline.DocPr.ID)
	}
	if !strings.Contains(drawing.Inline.DocPr.Name, "img1") {
		t.Errorf("DocPr.Name should contain img ID")
	}
	if drawing.Inline.DocPr.Descr != "Test Image" {
		t.Errorf("DocPr.Descr = %s; want 'Test Image'", drawing.Inline.DocPr.Descr)
	}
	if drawing.Inline.Graphic == nil {
		t.Fatal("Inline.Graphic is nil")
	}
}

func TestNewFloatingDrawing(t *testing.T) {
	img := &mockImage{
		id:             "img2",
		format:         domain.ImageFormatJPEG,
		size:           domain.NewImageSize(200, 150),
		relationshipID: "rId2",
		description:    "Floating Image",
		position: domain.ImagePosition{
			Type:       domain.ImagePositionFloating,
			HAlign:     domain.HAlignCenter,
			VAlign:     domain.VAlignTop,
			OffsetX:    0,
			OffsetY:    0,
			WrapText:   domain.WrapSquare,
			ZOrder:     5,
			BehindText: false,
		},
	}

	drawing := NewFloatingDrawing(img, 2)

	if drawing == nil {
		t.Fatal("NewFloatingDrawing() returned nil")
	}
	if drawing.Anchor == nil {
		t.Fatal("NewFloatingDrawing() Anchor is nil")
	}
	if drawing.Anchor.Extent == nil {
		t.Fatal("Anchor.Extent is nil")
	}
	if drawing.Anchor.Extent.Cx != img.size.WidthEMU {
		t.Errorf("Extent.Cx = %d; want %d", drawing.Anchor.Extent.Cx, img.size.WidthEMU)
	}
	if drawing.Anchor.RelativeHeight != 5 {
		t.Errorf("RelativeHeight = %d; want 5", drawing.Anchor.RelativeHeight)
	}
	if drawing.Anchor.BehindDoc {
		t.Error("BehindDoc should be false")
	}
	if drawing.Anchor.PositionH == nil {
		t.Fatal("PositionH is nil")
	}
	if drawing.Anchor.PositionV == nil {
		t.Fatal("PositionV is nil")
	}
	if drawing.Anchor.WrapType == nil {
		t.Fatal("WrapType should be set for WrapSquare")
	}
}

func TestConvertHAlign(t *testing.T) {
	tests := []struct {
		name     string
		align    domain.HorizontalAlign
		expected string
	}{
		{"Left", domain.HAlignLeft, "column"},
		{"Center", domain.HAlignCenter, "column"},
		{"Right", domain.HAlignRight, "column"},
		{"Inside", domain.HAlignInside, "margin"},
		{"Outside", domain.HAlignOutside, "margin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertHAlign(tt.align)
			if result != tt.expected {
				t.Errorf("convertHAlign(%v) = %s; want %s", tt.align, result, tt.expected)
			}
		})
	}
}

func TestConvertVAlign(t *testing.T) {
	tests := []struct {
		name     string
		align    domain.VerticalAlign
		expected string
	}{
		{"Top", domain.VAlignTop, "paragraph"},
		{"Center", domain.VAlignCenter, "paragraph"},
		{"Bottom", domain.VAlignBottom, "paragraph"},
		{"Inside", domain.VAlignInside, "margin"},
		{"Outside", domain.VAlignOutside, "margin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertVAlign(tt.align)
			if result != tt.expected {
				t.Errorf("convertVAlign(%v) = %s; want %s", tt.align, result, tt.expected)
			}
		})
	}
}

func TestNewGraphic(t *testing.T) {
	img := &mockImage{
		id:             "img3",
		format:         domain.ImageFormatPNG,
		size:           domain.NewImageSize(300, 200),
		relationshipID: "rId3",
		description:    "Graphic Test",
	}

	graphic := newGraphic(img, img.size)

	if graphic == nil {
		t.Fatal("newGraphic() returned nil")
	}
	if graphic.GraphicData == nil {
		t.Fatal("GraphicData is nil")
	}
	if graphic.GraphicData.Pic == nil {
		t.Fatal("Pic is nil")
	}
	if graphic.GraphicData.Pic.BlipFill == nil {
		t.Fatal("BlipFill is nil")
	}
	if graphic.GraphicData.Pic.BlipFill.Blip == nil {
		t.Fatal("Blip is nil")
	}
	if graphic.GraphicData.Pic.BlipFill.Blip.Embed != "rId3" {
		t.Errorf("Blip.Embed = %s; want 'rId3'", graphic.GraphicData.Pic.BlipFill.Blip.Embed)
	}
	if graphic.GraphicData.Pic.SpPr == nil {
		t.Fatal("SpPr is nil")
	}
	if graphic.GraphicData.Pic.SpPr.Xfrm == nil {
		t.Fatal("Xfrm is nil")
	}
	if graphic.GraphicData.Pic.SpPr.Xfrm.Ext == nil {
		t.Fatal("Ext is nil")
	}
	if graphic.GraphicData.Pic.SpPr.Xfrm.Ext.Cx != img.size.WidthEMU {
		t.Errorf("Ext.Cx = %d; want %d", graphic.GraphicData.Pic.SpPr.Xfrm.Ext.Cx, img.size.WidthEMU)
	}
}

