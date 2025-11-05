/*
   Copyright (c) 2025 Misael Monterroca

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

package serializer_test

import (
	stdxml "encoding/xml"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
	"github.com/mmonterroca/docxgo/v2/internal/serializer"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
)

func collectRuns(p *xmlstructs.Paragraph) []*xmlstructs.Run {
	runs := make([]*xmlstructs.Run, 0)
	for _, el := range p.Elements {
		switch v := el.(type) {
		case *xmlstructs.Run:
			runs = append(runs, v)
		case *xmlstructs.Hyperlink:
			runs = append(runs, v.Runs...)
		default:
			// ignore other elements (bookmarks, field chars handled via runs)
		}
	}
	return runs
}

func TestRunSerializer(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	run.SetText("Hello, World!")
	run.SetBold(true)
	run.SetItalic(true)
	run.SetSize(24)
	run.SetColor(domain.ColorRed)

	ser := serializer.NewRunSerializer()
	xmlRun := ser.Serialize(run)

	if xmlRun.Text == nil {
		t.Fatal("expected text to be set")
	}
	if xmlRun.Text.Content != "Hello, World!" {
		t.Errorf("expected text 'Hello, World!', got %q", xmlRun.Text.Content)
	}

	if xmlRun.Properties == nil {
		t.Fatal("expected properties to be set")
	}
	if xmlRun.Properties.Bold == nil {
		t.Error("expected bold to be set")
	}
	if xmlRun.Properties.Italic == nil {
		t.Error("expected italic to be set")
	}
	if xmlRun.Properties.Size == nil || xmlRun.Properties.Size.Val != 24 {
		t.Error("expected size to be 24")
	}
	if xmlRun.Properties.Color == nil || xmlRun.Properties.Color.Val != "FF0000" {
		t.Error("expected color to be FF0000 (red)")
	}
}

func TestRunSerializer_XMLOutput(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	run.SetText("Test")
	run.SetBold(true)

	ser := serializer.NewRunSerializer()
	xmlRun := ser.Serialize(run)

	// Marshal to XML
	data, err := stdxml.MarshalIndent(xmlRun, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	xmlStr := string(data)
	if xmlStr == "" {
		t.Error("expected non-empty XML")
	}

	// Check for expected elements
	if !contains(xmlStr, "<w:r>") {
		t.Error("expected <w:r> element")
	}
	if !contains(xmlStr, "<w:rPr>") {
		t.Error("expected <w:rPr> element")
	}
	if !contains(xmlStr, "<w:b") {
		t.Error("expected <w:b> element")
	}
	if !contains(xmlStr, "<w:t>Test</w:t>") {
		t.Error("expected <w:t>Test</w:t> element")
	}
}

func TestParagraphSerializer(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	run1, _ := para.AddRun()
	run1.SetText("First ")

	run2, _ := para.AddRun()
	run2.SetText("Second")
	run2.SetBold(true)

	para.SetAlignment(domain.AlignmentCenter)

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if runs := collectRuns(xmlPara); len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}

	if xmlPara.Properties == nil {
		t.Fatal("expected properties to be set")
	}
	if xmlPara.Properties.Justification == nil {
		t.Error("expected justification to be set")
	}
	if xmlPara.Properties.Justification.Val != "center" {
		t.Errorf("expected justification 'center', got %q", xmlPara.Properties.Justification.Val)
	}
}

func TestParagraphSerializer_Indentation(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	indent := domain.Indentation{
		Left:      720,
		Right:     360,
		FirstLine: 360,
	}
	para.SetIndent(indent)

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Indentation == nil {
		t.Fatal("expected indentation to be set")
	}

	ind := xmlPara.Properties.Indentation
	if ind.Left == nil || *ind.Left != 720 {
		t.Error("expected left indent 720")
	}
	if ind.Right == nil || *ind.Right != 360 {
		t.Error("expected right indent 360")
	}
	if ind.FirstLine == nil || *ind.FirstLine != 360 {
		t.Error("expected first line indent 360")
	}
}

func TestTableSerializer(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(2, 3)

	// Fill first cell
	row0, _ := table.Row(0)
	cell, _ := row0.Cell(0)
	cellPara, _ := cell.AddParagraph()
	cellRun, _ := cellPara.AddRun()
	cellRun.SetText("Cell 0,0")

	ser := serializer.NewTableSerializer()
	xmlTable := ser.Serialize(table)

	if len(xmlTable.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(xmlTable.Rows))
	}

	if xmlTable.Grid == nil {
		t.Fatal("expected grid to be set")
	}
	if len(xmlTable.Grid.Cols) != 3 {
		t.Errorf("expected 3 columns, got %d", len(xmlTable.Grid.Cols))
	}

	// Check first cell
	if len(xmlTable.Rows[0].Cells) != 3 {
		t.Errorf("expected 3 cells in first row, got %d", len(xmlTable.Rows[0].Cells))
	}

	firstCell := xmlTable.Rows[0].Cells[0]
	if len(firstCell.Content) == 0 {
		t.Fatal("expected at least one element in first cell content")
	}
	firstPara, ok := firstCell.Content[0].(*xmlstructs.Paragraph)
	if !ok {
		t.Fatalf("expected first content element to be paragraph, got %T", firstCell.Content[0])
	}
	cellRuns := collectRuns(firstPara)
	if len(cellRuns) == 0 {
		t.Fatal("expected at least one run in first paragraph")
	}
	if cellRuns[0].Text == nil || cellRuns[0].Text.Content != "Cell 0,0" {
		t.Errorf("expected 'Cell 0,0', got %v", cellRuns[0].Text)
	}
}

func TestDocumentSerializer(t *testing.T) {
	doc := core.NewDocument()

	// Add paragraph
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("Test Document")

	// Add table
	_, _ = doc.AddTable(2, 2)

	ser := serializer.NewDocumentSerializer()
	xmlDoc := ser.SerializeDocument(doc)

	if xmlDoc.Body == nil {
		t.Fatal("expected body to be set")
	}

	if len(xmlDoc.Body.Content) != 2 {
		t.Fatalf("expected 2 body elements, got %d", len(xmlDoc.Body.Content))
	}

	if _, ok := xmlDoc.Body.Content[0].(*xmlstructs.Paragraph); !ok {
		t.Errorf("expected first body element to be paragraph, got %T", xmlDoc.Body.Content[0])
	}

	if _, ok := xmlDoc.Body.Content[1].(*xmlstructs.Table); !ok {
		t.Errorf("expected second body element to be table, got %T", xmlDoc.Body.Content[1])
	}

	if xmlDoc.XMLnsW == "" {
		t.Error("expected XMLnsW to be set")
	}
	if xmlDoc.XMLnsR == "" {
		t.Error("expected XMLnsR to be set")
	}
}

func TestDocumentSerializer_CompleteXML(t *testing.T) {
	doc := core.NewDocument()

	// Set metadata
	meta := &domain.Metadata{
		Title:   "Test Document",
		Creator: "Test Suite",
		Subject: "Testing",
	}
	if err := doc.SetMetadata(meta); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Add content
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	if err := run.SetText("Hello, World!"); err != nil {
		t.Fatalf("SetText failed: %v", err)
	}
	run.SetBold(true)

	ser := serializer.NewDocumentSerializer()

	// Serialize document
	xmlDoc := ser.SerializeDocument(doc)
	data, err := stdxml.MarshalIndent(xmlDoc, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal document: %v", err)
	}

	xmlStr := string(data)
	if !contains(xmlStr, "<w:document") {
		t.Error("expected <w:document> element")
	}
	if !contains(xmlStr, "<w:body>") {
		t.Error("expected <w:body> element")
	}

	body := xmlDoc.Body
	if body == nil {
		t.Fatal("expected body to be set")
	}

	if len(body.Content) == 0 {
		t.Error("expected document body content")
	}

	if body.SectPr == nil {
		t.Error("expected final section properties")
	}

	// Serialize core properties
	coreProps := ser.SerializeCoreProperties(meta)
	propsData, err := stdxml.MarshalIndent(coreProps, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal core properties: %v", err)
	}

	propsStr := string(propsData)
	if !contains(propsStr, "Test Document") {
		t.Error("expected title in core properties")
	}
	if !contains(propsStr, "Test Suite") {
		t.Error("expected creator in core properties")
	}
}

func TestDocumentSerializer_SectionBreaks(t *testing.T) {
	doc := core.NewDocument()

	defaultSection, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("failed to obtain default section: %v", err)
	}
	_ = defaultSection.SetOrientation(domain.OrientationLandscape)
	_ = defaultSection.SetColumns(2)

	para1, _ := doc.AddParagraph()
	run1, _ := para1.AddRun()
	run1.SetText("Section one")

	newSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeEvenPage)
	if err != nil {
		t.Fatalf("failed to add section: %v", err)
	}
	_ = newSection.SetColumns(3)

	para2, _ := doc.AddParagraph()
	run2, _ := para2.AddRun()
	run2.SetText("Section two")

	ser := serializer.NewDocumentSerializer()
	xmlDoc := ser.SerializeDocument(doc)
	body := xmlDoc.Body
	if body == nil {
		t.Fatal("expected body to be set")
	}

	if len(body.Content) != 3 {
		t.Fatalf("expected 3 body elements (para, break, para), got %d", len(body.Content))
	}

	breakPara, ok := body.Content[1].(*xmlstructs.Paragraph)
	if !ok {
		t.Fatalf("expected second element to be paragraph break, got %T", body.Content[1])
	}

	if breakPara.Properties == nil || breakPara.Properties.SectionProperties == nil {
		t.Fatal("expected section properties on break paragraph")
	}
	breakSect := breakPara.Properties.SectionProperties
	if breakSect.Type == nil || breakSect.Type.Val != "evenPage" {
		t.Errorf("expected section break type evenPage, got %v", breakSect.Type)
	}
	if breakSect.PageSize == nil || breakSect.PageSize.Orient != "landscape" {
		t.Errorf("expected landscape orientation on break, got %+v", breakSect.PageSize)
	}
	if breakSect.PageSize.Width <= breakSect.PageSize.Height {
		t.Errorf("expected width greater than height for landscape, got %+v", breakSect.PageSize)
	}
	if breakSect.Columns == nil || breakSect.Columns.Num != 2 {
		t.Errorf("expected 2 columns on first section, got %+v", breakSect.Columns)
	}

	if body.SectPr == nil {
		t.Fatal("expected final section properties on body")
	}
	if body.SectPr.Columns == nil || body.SectPr.Columns.Num != 3 {
		t.Errorf("expected 3 columns on final section, got %+v", body.SectPr.Columns)
	}
	if body.SectPr.Type != nil {
		t.Errorf("did not expect section type on final section, got %+v", body.SectPr.Type)
	}
	if body.SectPr.PageSize == nil {
		t.Fatal("expected final section page size")
	}
	if body.SectPr.PageSize.Width >= body.SectPr.PageSize.Height {
		t.Errorf("expected portrait dimensions on final section, got %+v", body.SectPr.PageSize)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRunSerializer_Underline(t *testing.T) {
	tests := []struct {
		name      string
		style     domain.UnderlineStyle
		wantEmpty bool
	}{
		{"Single", domain.UnderlineSingle, false},
		{"Double", domain.UnderlineDouble, false},
		{"Thick", domain.UnderlineThick, false},
		{"Dotted", domain.UnderlineDotted, false},
		{"Dashed", domain.UnderlineDashed, false},
		{"Wave", domain.UnderlineWave, false},
		{"None", domain.UnderlineNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			para, _ := doc.AddParagraph()
			run, _ := para.AddRun()
			run.SetText("Underlined")
			run.SetUnderline(tt.style)

			ser := serializer.NewRunSerializer()
			xmlRun := ser.Serialize(run)

			if !tt.wantEmpty {
				if xmlRun.Properties == nil || xmlRun.Properties.Underline == nil {
					t.Error("expected underline to be set")
				}
			}
		})
	}
}

func TestRunSerializer_Highlight(t *testing.T) {
	tests := []struct {
		name  string
		color domain.HighlightColor
	}{
		{"Yellow", domain.HighlightYellow},
		{"Green", domain.HighlightGreen},
		{"Cyan", domain.HighlightCyan},
		{"Magenta", domain.HighlightMagenta},
		{"Blue", domain.HighlightBlue},
		{"Red", domain.HighlightRed},
		{"DarkBlue", domain.HighlightDarkBlue},
		{"DarkGreen", domain.HighlightDarkGreen},
		{"LightGray", domain.HighlightLightGray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			para, _ := doc.AddParagraph()
			run, _ := para.AddRun()
			run.SetText("Highlighted")
			run.SetHighlight(tt.color)

			ser := serializer.NewRunSerializer()
			xmlRun := ser.Serialize(run)

			if xmlRun.Properties == nil || xmlRun.Properties.Highlight == nil {
				t.Error("expected highlight to be set")
			}
		})
	}
}

func TestParagraphSerializer_LineSpacing(t *testing.T) {
	tests := []struct {
		name    string
		spacing domain.LineSpacing
	}{
		{"Exact", domain.LineSpacing{Rule: domain.LineSpacingExact, Value: 360}},
		{"AtLeast", domain.LineSpacing{Rule: domain.LineSpacingAtLeast, Value: 480}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			para, _ := doc.AddParagraph()
			para.SetLineSpacing(tt.spacing)

			ser := serializer.NewParagraphSerializer()
			xmlPara := ser.Serialize(para)

			if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
				t.Error("expected spacing to be set")
			}
		})
	}
}

func TestParagraphSerializer_Alignment(t *testing.T) {
	tests := []struct {
		name      string
		alignment domain.Alignment
		expected  string
	}{
		{"Center", domain.AlignmentCenter, "center"},
		{"Right", domain.AlignmentRight, "right"},
		{"Justify", domain.AlignmentJustify, "both"},
		{"Distribute", domain.AlignmentDistribute, "distribute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			para, _ := doc.AddParagraph()
			para.SetAlignment(tt.alignment)

			ser := serializer.NewParagraphSerializer()
			xmlPara := ser.Serialize(para)

			if xmlPara.Properties == nil || xmlPara.Properties.Justification == nil {
				t.Errorf("expected justification to be set for alignment %v", tt.alignment)
			} else if xmlPara.Properties.Justification.Val != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, xmlPara.Properties.Justification.Val)
			}
		})
	}
}

func TestTableSerializer_VerticalAlignment(t *testing.T) {
	tests := []struct {
		name  string
		align domain.VerticalAlignment
	}{
		{"Top", domain.VerticalAlignTop},
		{"Center", domain.VerticalAlignCenter},
		{"Bottom", domain.VerticalAlignBottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			table, _ := doc.AddTable(1, 1)
			row, _ := table.Row(0)
			cell, _ := row.Cell(0)
			cell.SetVerticalAlignment(tt.align)

			ser := serializer.NewTableSerializer()
			xmlTable := ser.Serialize(table)

			if len(xmlTable.Rows) == 0 {
				t.Fatal("expected at least one row")
			}
			if len(xmlTable.Rows[0].Cells) == 0 {
				t.Fatal("expected at least one cell")
			}
		})
	}
}

func TestTableSerializer_CellWidth(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 2)
	row, _ := table.Row(0)

	cell1, _ := row.Cell(0)
	cell1.SetWidth(2000) // width in twips

	cell2, _ := row.Cell(1)
	cell2.SetWidth(3000)

	ser := serializer.NewTableSerializer()
	xmlTable := ser.Serialize(table)

	if len(xmlTable.Rows) == 0 || len(xmlTable.Rows[0].Cells) < 2 {
		t.Fatal("expected cells to be serialized")
	}
}

func TestRunSerializer_WithTextBreaks(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("Line 1\nLine 2\nLine 3")

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	runs := collectRuns(xmlPara)
	// Should have multiple runs due to newline expansion
	if len(runs) < 3 {
		t.Errorf("expected at least 3 elements (text+break+text), got %d", len(runs))
	}
}
