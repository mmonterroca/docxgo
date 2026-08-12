// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package serializer_test

import (
	stdxml "encoding/xml"
	"io"
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

func TestRunSerializer_Language(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("bonjour")
	if err := run.SetLanguage(&domain.Language{Val: "fr", EastAsia: "fr", Bidi: "fr"}); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}

	ser := serializer.NewRunSerializer()
	xmlRun := ser.Serialize(run)

	if xmlRun.Properties == nil || xmlRun.Properties.Lang == nil {
		t.Fatal("expected run language to be set")
	}
	if xmlRun.Properties.Lang.Val != "fr" || xmlRun.Properties.Lang.EastAsia != "fr" || xmlRun.Properties.Lang.Bidi != "fr" {
		t.Errorf("unexpected Lang: %+v", xmlRun.Properties.Lang)
	}
}

func TestRunSerializer_Language_UnsetOmitsLangElement(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("plain")

	ser := serializer.NewRunSerializer()
	xmlRun := ser.Serialize(run)

	if xmlRun.Properties != nil && xmlRun.Properties.Lang != nil {
		t.Error("expected no run-level language when unset — should inherit the document default")
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

func TestTableSerializer_CellBorders(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 1)

	// Get first cell and set borders
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	// Set borders with specific properties
	borders := domain.TableBorders{
		Top: domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 4,
			Color: domain.ColorRed,
		},
		Bottom: domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 4,
			Color: domain.ColorRed,
		},
		Left: domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 4,
			Color: domain.ColorRed,
		},
		Right: domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 4,
			Color: domain.ColorRed,
		},
	}
	cell.SetBorders(borders)

	// Serialize the table
	ser := serializer.NewTableSerializer()
	xmlTable := ser.Serialize(table)

	// Get first cell from serialized table
	if len(xmlTable.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(xmlTable.Rows))
	}
	if len(xmlTable.Rows[0].Cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(xmlTable.Rows[0].Cells))
	}

	firstCell := xmlTable.Rows[0].Cells[0]

	// Check that borders are set
	if firstCell.Properties == nil {
		t.Fatal("expected cell properties to be set")
	}
	if firstCell.Properties.Borders == nil {
		t.Fatal("expected borders to be set")
	}

	// Check top border properties
	if firstCell.Properties.Borders.Top == nil {
		t.Fatal("expected border top to be set")
	}
	if firstCell.Properties.Borders.Top.Val != "single" {
		t.Errorf("expected border top style to be 'single', got %q", firstCell.Properties.Borders.Top.Val)
	}
	if firstCell.Properties.Borders.Top.Sz != 4 {
		t.Errorf("expected border top width (Sz) to be 4, got %d", firstCell.Properties.Borders.Top.Sz)
	}
	if firstCell.Properties.Borders.Top.Color != "FF0000" {
		t.Errorf("expected border top color to be 'FF0000' (red), got %q", firstCell.Properties.Borders.Top.Color)
	}

	// Check all borders have the same properties
	borders_to_check := []*xmlstructs.Border{
		firstCell.Properties.Borders.Bottom,
		firstCell.Properties.Borders.Left,
		firstCell.Properties.Borders.Right,
	}
	borderNames := []string{"bottom", "left", "right"}

	for i, border := range borders_to_check {
		if border == nil {
			t.Errorf("expected border %s to be set", borderNames[i])
			continue
		}
		if border.Val != "single" {
			t.Errorf("expected border %s style to be 'single', got %q", borderNames[i], border.Val)
		}
		if border.Sz != 4 {
			t.Errorf("expected border %s width (Sz) to be 4, got %d", borderNames[i], border.Sz)
		}
		if border.Color != "FF0000" {
			t.Errorf("expected border %s color to be 'FF0000' (red), got %q", borderNames[i], border.Color)
		}
	}
}

func TestTableSerializer_GridColumnWidths(t *testing.T) {
	t.Run("derives widths from first row cells", func(t *testing.T) {
		doc := core.NewDocument()
		table, _ := doc.AddTable(2, 3)

		row, _ := table.Row(0)
		c0, _ := row.Cell(0)
		c0.SetWidth(2000)
		c1, _ := row.Cell(1)
		c1.SetWidth(3000)
		c2, _ := row.Cell(2)
		c2.SetWidth(4000)

		ser := serializer.NewTableSerializer()
		xmlTable := ser.Serialize(table)

		if xmlTable.Grid == nil {
			t.Fatal("expected grid to be set")
		}
		if len(xmlTable.Grid.Cols) != 3 {
			t.Fatalf("expected 3 grid columns, got %d", len(xmlTable.Grid.Cols))
		}

		expected := []int{2000, 3000, 4000}
		for i, col := range xmlTable.Grid.Cols {
			if col.W == nil {
				t.Errorf("column %d: expected width to be set", i)
				continue
			}
			if *col.W != expected[i] {
				t.Errorf("column %d: expected width %d, got %d", i, expected[i], *col.W)
			}
		}
	})

	t.Run("omits widths when cells have no explicit width", func(t *testing.T) {
		doc := core.NewDocument()
		table, _ := doc.AddTable(1, 2)

		ser := serializer.NewTableSerializer()
		xmlTable := ser.Serialize(table)

		if xmlTable.Grid == nil {
			t.Fatal("expected grid to be set")
		}
		for i, col := range xmlTable.Grid.Cols {
			if col.W != nil {
				t.Errorf("column %d: expected width to be omitted (nil), got %d", i, *col.W)
			}
		}
	})

	t.Run("distributes spanned cell width across grid columns", func(t *testing.T) {
		doc := core.NewDocument()
		table, _ := doc.AddTable(1, 4)

		row, _ := table.Row(0)
		// Merge first two cells horizontally (span=2)
		c0, _ := row.Cell(0)
		c0.SetWidth(6000)
		c0.Merge(2, 1) // merge 2 cols, 1 row

		c2, _ := row.Cell(2)
		c2.SetWidth(2000)
		c3, _ := row.Cell(3)
		c3.SetWidth(1000)

		ser := serializer.NewTableSerializer()
		xmlTable := ser.Serialize(table)

		if xmlTable.Grid == nil {
			t.Fatal("expected grid to be set")
		}
		if len(xmlTable.Grid.Cols) != 4 {
			t.Fatalf("expected 4 grid columns, got %d", len(xmlTable.Grid.Cols))
		}

		// 6000 / 2 = 3000 per spanned column
		expected := []int{3000, 3000, 2000, 1000}
		for i, col := range xmlTable.Grid.Cols {
			if col.W == nil {
				t.Errorf("column %d: expected width to be set", i)
				continue
			}
			if *col.W != expected[i] {
				t.Errorf("column %d: expected width %d, got %d", i, expected[i], *col.W)
			}
		}
	})
}

func TestDocumentSerializer_TableStyleBorders(t *testing.T) {
	doc := core.NewDocument()

	// The built-in style manager should include TableGrid with borders
	sm := doc.StyleManager()

	ser := serializer.NewDocumentSerializer()
	xmlStyles := ser.SerializeStyles(sm, nil, nil, nil)

	// Find the TableGrid style
	var tableGridStyle *xmlstructs.Style
	for _, s := range xmlStyles.Styles {
		if s.StyleID == "TableGrid" {
			tableGridStyle = s
			break
		}
	}

	if tableGridStyle == nil {
		t.Fatal("expected TableGrid style to be present in serialized styles")
	}

	if tableGridStyle.Type != "table" {
		t.Errorf("expected TableGrid type to be 'table', got %q", tableGridStyle.Type)
	}

	if tableGridStyle.TblPr == nil {
		t.Fatal("expected TableGrid to have w:tblPr element")
	}

	if tableGridStyle.TblPr.Borders == nil {
		t.Fatal("expected TableGrid w:tblPr to have w:tblBorders element")
	}

	borders := tableGridStyle.TblPr.Borders

	// TableGrid should have all 6 border sides (top, left, bottom, right, insideH, insideV)
	sides := map[string]*xmlstructs.Border{
		"top":     borders.Top,
		"left":    borders.Left,
		"bottom":  borders.Bottom,
		"right":   borders.Right,
		"insideH": borders.InsideH,
		"insideV": borders.InsideV,
	}

	for name, border := range sides {
		if border == nil {
			t.Errorf("expected %s border to be set", name)
			continue
		}
		if border.Val != "single" {
			t.Errorf("%s border: expected style 'single', got %q", name, border.Val)
		}
		if border.Sz != 4 {
			t.Errorf("%s border: expected width 4, got %d", name, border.Sz)
		}
	}

	// Verify XML output contains expected elements
	data, err := stdxml.MarshalIndent(tableGridStyle, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal TableGrid style: %v", err)
	}
	xmlStr := string(data)

	if !contains(xmlStr, "<w:tblPr>") {
		t.Error("expected <w:tblPr> in serialized TableGrid style")
	}
	if !contains(xmlStr, "<w:tblBorders>") {
		t.Error("expected <w:tblBorders> in serialized TableGrid style")
	}
	if !contains(xmlStr, "<w:insideH") {
		t.Error("expected <w:insideH> border in serialized TableGrid style")
	}
	if !contains(xmlStr, "<w:insideV") {
		t.Error("expected <w:insideV> border in serialized TableGrid style")
	}
}

// ─── Zero paragraph spacing regression tests ───────────────────────────────
//
// A paragraph at the domain's own defaults (SpacingBefore/After = 0, single
// line spacing) never emits a direct <w:spacing> element — that's existing,
// correct behavior (see serializeProperties). The bug was that nothing backed
// that omission up: word/styles.xml's w:docDefaults carried no w:pPrDefault
// spacing, so Word fell back to its own defaults (8pt after, 1.15 line)
// instead of the 0/0/240 the domain models. These tests pin both halves: the
// paragraph still omits direct formatting, and the document defaults now
// actually state 0/0/240 so that omission means what it's supposed to mean.

func TestDocumentSerializer_DocDefaultsStateZeroSpacing(t *testing.T) {
	doc := core.NewDocument()
	sm := doc.StyleManager()

	ser := serializer.NewDocumentSerializer()
	xmlStyles := ser.SerializeStyles(sm, nil, nil, nil)

	if xmlStyles.DocDefaults == nil || xmlStyles.DocDefaults.ParaDefaults == nil {
		t.Fatal("expected w:docDefaults/w:pPrDefault to be set")
	}
	spacing := xmlStyles.DocDefaults.ParaDefaults.Properties.Spacing
	if spacing == nil {
		t.Fatal("expected w:pPrDefault to carry a w:spacing element")
	}
	if spacing.Before == nil || *spacing.Before != 0 {
		t.Errorf("docDefaults before = %v, want 0", spacing.Before)
	}
	if spacing.After == nil || *spacing.After != 0 {
		t.Errorf("docDefaults after = %v, want 0", spacing.After)
	}
	if spacing.Line == nil || *spacing.Line != 240 {
		t.Errorf("docDefaults line = %v, want 240", spacing.Line)
	}
	if spacing.LineRule != "auto" {
		t.Errorf("docDefaults lineRule = %q, want %q", spacing.LineRule, "auto")
	}
}

func TestParagraphSerializer_DefaultSpacingOmitsDirectFormatting(t *testing.T) {
	// A paragraph that never touched its spacing must not emit a direct
	// <w:spacing> element — direct formatting wins over docDefaults/style in
	// OOXML's cascade, so emitting one unconditionally (the #63 regression)
	// would make every paragraph opt out of inheriting from its style.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties != nil && xmlPara.Properties.Spacing != nil {
		t.Errorf("expected no direct w:spacing for a paragraph at defaults, got %+v", xmlPara.Properties.Spacing)
	}
}

func TestParagraphSerializer_StyledParagraphSpacingNotClobbered(t *testing.T) {
	// A Heading1 paragraph that never sets its own spacing must not emit
	// direct <w:spacing>, so it inherits the style's 240/120 rather than
	// docDefaults' 0/0. Forcing direct 0/0 formatting on every paragraph
	// (the #63 regression) would silently strip spacing from every heading.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	if err := para.SetStyle(domain.StyleIDHeading1); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties != nil && xmlPara.Properties.Spacing != nil {
		t.Errorf("expected no direct w:spacing on a styled paragraph that didn't set its own, got %+v", xmlPara.Properties.Spacing)
	}

	// Confirm the style itself still carries its own non-zero spacing —
	// i.e. there's something real for the paragraph to inherit.
	sm := doc.StyleManager()
	docSer := serializer.NewDocumentSerializer()
	xmlStyles := docSer.SerializeStyles(sm, nil, nil, nil)

	var heading1 *xmlstructs.Style
	for _, s := range xmlStyles.Styles {
		if s.StyleID == domain.StyleIDHeading1 {
			heading1 = s
			break
		}
	}
	if heading1 == nil {
		t.Fatal("expected Heading1 style to be present in serialized styles")
	}
	if heading1.ParaProps == nil || heading1.ParaProps.Spacing == nil {
		t.Fatal("expected Heading1 style to carry its own w:spacing")
	}
	if heading1.ParaProps.Spacing.Before == nil || *heading1.ParaProps.Spacing.Before != 240 {
		t.Errorf("Heading1 style before = %v, want 240", heading1.ParaProps.Spacing.Before)
	}
	if heading1.ParaProps.Spacing.After == nil || *heading1.ParaProps.Spacing.After != 120 {
		t.Errorf("Heading1 style after = %v, want 120", heading1.ParaProps.Spacing.After)
	}
}

func TestParagraphSerializer_ExactLineSpacingAtDefaultValueIsEmitted(t *testing.T) {
	// An Exact/AtLeast line-spacing rule is a real departure from the document
	// defaults even when its value happens to equal DefaultLineSpacing (240).
	// If the emit gate looked only at the value, such a paragraph would emit
	// nothing and silently inherit the lineRule="auto" that w:pPrDefault now
	// installs — turning a caller's exact 12pt line height into auto spacing.
	for _, tt := range []struct {
		name string
		rule domain.LineSpacingRule
		want string
	}{
		{"Exact", domain.LineSpacingExact, "exact"},
		{"AtLeast", domain.LineSpacingAtLeast, "atLeast"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := core.NewDocument()
			para, _ := doc.AddParagraph()
			if err := para.SetLineSpacing(domain.LineSpacing{
				Rule:  tt.rule,
				Value: 240, // deliberately the default value
			}); err != nil {
				t.Fatalf("SetLineSpacing: %v", err)
			}

			ser := serializer.NewParagraphSerializer()
			xmlPara := ser.Serialize(para)

			if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
				t.Fatalf("expected direct w:spacing to be emitted for rule %s at the default value", tt.name)
			}
			got := xmlPara.Properties.Spacing.LineRule
			if got == nil {
				t.Fatalf("lineRule not set, want %q", tt.want)
			}
			if *got != tt.want {
				t.Errorf("lineRule = %q, want %q", *got, tt.want)
			}
		})
	}

	// An explicit SetLineSpacing(Auto, 240) is a real call the caller made,
	// even though it names the default values — since it's distinguishable
	// from "never called" (see LineSpacingSet, #69), it must still emit, so
	// it can override a style's own non-auto line spacing rather than being
	// mistaken for "unset" and silently losing to the style.
	t.Run("ExplicitAutoAtDefaultIsEmitted", func(t *testing.T) {
		doc := core.NewDocument()
		para, _ := doc.AddParagraph()
		if err := para.SetLineSpacing(domain.LineSpacing{
			Rule:  domain.LineSpacingAuto,
			Value: 240,
		}); err != nil {
			t.Fatalf("SetLineSpacing: %v", err)
		}

		ser := serializer.NewParagraphSerializer()
		xmlPara := ser.Serialize(para)
		if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
			t.Fatal("expected direct w:spacing for an explicit SetLineSpacing call, even at default values")
		}
	})

	// A paragraph that never calls SetLineSpacing at all is the genuine
	// "no departure" case and must stay out of the way of the style.
	t.Run("NeverSetStaysSilent", func(t *testing.T) {
		doc := core.NewDocument()
		para, _ := doc.AddParagraph()

		ser := serializer.NewParagraphSerializer()
		xmlPara := ser.Serialize(para)
		if xmlPara.Properties != nil && xmlPara.Properties.Spacing != nil {
			t.Errorf("expected no direct w:spacing when SetLineSpacing was never called, got %+v", xmlPara.Properties.Spacing)
		}
	})
}

func TestParagraphSerializer_PartialIndentOmitsOtherSides(t *testing.T) {
	// A paragraph that only sets Left indent must not emit w:right/
	// w:firstLine/w:hanging as explicit zero — that would override a style's
	// indentation on those sides instead of leaving them to inherit.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	if err := para.SetIndent(domain.Indentation{Left: 720}); err != nil {
		t.Fatalf("SetIndent: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Indentation == nil {
		t.Fatal("expected w:ind to be set")
	}
	ind := xmlPara.Properties.Indentation
	if ind.Left == nil || *ind.Left != 720 {
		t.Errorf("ind.Left = %v, want 720", ind.Left)
	}
	if ind.Right != nil {
		t.Errorf("ind.Right = %v, want omitted (nil)", ind.Right)
	}
	if ind.FirstLine != nil {
		t.Errorf("ind.FirstLine = %v, want omitted (nil)", ind.FirstLine)
	}
	if ind.Hanging != nil {
		t.Errorf("ind.Hanging = %v, want omitted (nil)", ind.Hanging)
	}
}

// TestParagraphSerializer_SetIndentClearsStalePerSideFlags is a regression
// test for a paragraph whose left indent was set via SetIndentLeft (the
// reader's per-attribute path — see applyParagraphIndentation) and then
// overwritten via a single SetIndent call. SetIndent used to leave
// indentLeftSet at true from the earlier call, so the serializer emitted an
// explicit w:left="0" and clobbered a style's own left indentation on a side
// this SetIndent call never mentioned. SetIndent must clear all four
// indent*Set flags along with the struct it replaces.
func TestParagraphSerializer_SetIndentClearsStalePerSideFlags(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	if err := para.SetIndentLeft(720); err != nil {
		t.Fatalf("SetIndentLeft: %v", err)
	}
	if err := para.SetIndent(domain.Indentation{Right: 360}); err != nil {
		t.Fatalf("SetIndent: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Indentation == nil {
		t.Fatal("expected w:ind to be set")
	}
	ind := xmlPara.Properties.Indentation
	if ind.Left != nil {
		t.Errorf("ind.Left = %v, want omitted (nil) — SetIndent must clear the stale indentLeftSet flag", ind.Left)
	}
	if ind.Right == nil || *ind.Right != 360 {
		t.Errorf("ind.Right = %v, want 360", ind.Right)
	}
	if ind.FirstLine != nil {
		t.Errorf("ind.FirstLine = %v, want omitted (nil)", ind.FirstLine)
	}
	if ind.Hanging != nil {
		t.Errorf("ind.Hanging = %v, want omitted (nil)", ind.Hanging)
	}
}

// TestParagraphSerializer_NeverEmitsBothFirstLineAndHanging is a regression
// test: SetIndent rejects setting both FirstLine and Hanging together, but
// the per-side setters SetIndentFirstLine/SetIndentHanging don't share that
// check — reachable from the reader hydrating a source <w:ind> that already
// carries both attributes (applyParagraphIndentation calls each setter
// independently per attribute present). CT_Ind's schema allows both, but
// Word treats them as mutually exclusive, so the serializer must emit only
// one — hanging wins, matching what Word itself does.
func TestParagraphSerializer_NeverEmitsBothFirstLineAndHanging(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	if err := para.SetIndentFirstLine(200); err != nil {
		t.Fatalf("SetIndentFirstLine: %v", err)
	}
	if err := para.SetIndentHanging(300); err != nil {
		t.Fatalf("SetIndentHanging: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Indentation == nil {
		t.Fatal("expected w:ind to be set")
	}
	ind := xmlPara.Properties.Indentation
	if ind.FirstLine != nil {
		t.Errorf("ind.FirstLine = %v, want omitted (nil) — hanging must win when both are set", ind.FirstLine)
	}
	if ind.Hanging == nil || *ind.Hanging != 300 {
		t.Errorf("ind.Hanging = %v, want 300", ind.Hanging)
	}
}

// TestParagraphSerializer_ExplicitZeroAfterOnStyledParagraphIsEmitted guards
// against the #69 bug: SpacingAfter() int can't distinguish "never called"
// from "explicitly set to 0", so an explicit SetSpacingAfter(0) on a
// paragraph whose style supplies non-zero spacing used to be silently
// dropped, and the paragraph inherited the style's value instead of the
// caller's explicit 0.
func TestParagraphSerializer_ExplicitZeroAfterOnStyledParagraphIsEmitted(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	if err := para.SetStyle(domain.StyleIDHeading1); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	if err := para.SetSpacingAfter(0); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
		t.Fatal("expected direct w:spacing for an explicit SetSpacingAfter(0) on a styled paragraph")
	}
	after := xmlPara.Properties.Spacing.After
	if after == nil || *after != 0 {
		t.Errorf("spacing.After = %v, want a pointer to 0", after)
	}
	// Before was never set, so it must still be omitted — otherwise the
	// caller's explicit After:0 would carry an unintended Before:0 with it,
	// clobbering the style's own before-spacing too.
	if xmlPara.Properties.Spacing.Before != nil {
		t.Errorf("spacing.Before = %v, want omitted (nil), since it was never set", xmlPara.Properties.Spacing.Before)
	}
	// Line/LineRule were never set either, so they must also be omitted —
	// otherwise a paragraph that only set spacingAfter would gain an
	// unintended direct w:line="240" w:lineRule="auto", overriding the
	// style's own line spacing (e.g. 1.5 lines) with single-spacing. This
	// was a real regression: the old code always filled Line/LineRule
	// whenever the w:spacing element was emitted for any reason.
	if xmlPara.Properties.Spacing.Line != nil {
		t.Errorf("spacing.Line = %v, want omitted (nil), since line spacing was never set", xmlPara.Properties.Spacing.Line)
	}
	if xmlPara.Properties.Spacing.LineRule != nil {
		t.Errorf("spacing.LineRule = %v, want omitted (nil), since line spacing was never set", xmlPara.Properties.Spacing.LineRule)
	}
}

// TestParagraphSerializer_ExplicitSpacingDoesNotClobberStyleLineSpacing pins
// the bug found in code review of #77: the emit gate correctly opened the
// <w:spacing> element on beforeSet/afterSet alone, but Line and LineRule were
// filled unconditionally inside it — lineSpacingRuleToString never returns
// nil and a never-touched LineSpacing still carries the constructor's
// {Auto, 240} defaults, so any paragraph that only set spacingBefore/After
// silently gained direct w:line="240" w:lineRule="auto", clobbering a
// style's real line spacing (here, Heading1's own line spacing, simulated by
// a style with an explicit non-default line spacing).
func TestParagraphSerializer_ExplicitSpacingDoesNotClobberStyleLineSpacing(t *testing.T) {
	doc := core.NewDocument()
	sm := doc.StyleManager()
	style, err := sm.GetStyle(domain.StyleIDHeading1)
	if err != nil {
		t.Fatalf("GetStyle: %v", err)
	}
	paraStyle, ok := style.(domain.ParagraphStyle)
	if !ok {
		t.Fatal("Heading1 style does not implement domain.ParagraphStyle")
	}
	// Give the style a non-default line spacing (1.5 lines) so an
	// unintended direct override on the paragraph is observable.
	if err := paraStyle.SetLineSpacing(360); err != nil {
		t.Fatalf("style SetLineSpacing: %v", err)
	}

	para, _ := doc.AddParagraph()
	if err := para.SetStyle(domain.StyleIDHeading1); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	if err := para.SetSpacingAfter(0); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}
	// The paragraph's own LineSpacing was never touched.

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
		t.Fatal("expected direct w:spacing for the explicit SetSpacingAfter(0)")
	}
	if xmlPara.Properties.Spacing.Line != nil || xmlPara.Properties.Spacing.LineRule != nil {
		t.Errorf("got direct w:line=%v w:lineRule=%v, want both omitted so the paragraph inherits the style's 1.5-line spacing",
			xmlPara.Properties.Spacing.Line, xmlPara.Properties.Spacing.LineRule)
	}
}

// TestParagraphSerializer_ExplicitBeforeAndZeroAfterBothEmitted confirms both
// an explicit non-zero Before and an explicit zero After are emitted
// together, rather than the zero-value gate hiding one of them.
func TestParagraphSerializer_ExplicitBeforeAndZeroAfterBothEmitted(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	if err := para.SetStyle(domain.StyleIDHeading1); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	if err := para.SetSpacingBefore(240); err != nil {
		t.Fatalf("SetSpacingBefore: %v", err)
	}
	if err := para.SetSpacingAfter(0); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(para)

	if xmlPara.Properties == nil || xmlPara.Properties.Spacing == nil {
		t.Fatal("expected direct w:spacing")
	}
	spacing := xmlPara.Properties.Spacing
	if spacing.Before == nil || *spacing.Before != 240 {
		t.Errorf("spacing.Before = %v, want a pointer to 240", spacing.Before)
	}
	if spacing.After == nil || *spacing.After != 0 {
		t.Errorf("spacing.After = %v, want a pointer to 0", spacing.After)
	}
}

// wrappedParagraph is the plain embedding decorator a third-party consumer
// would write to add behavior around a domain.Paragraph.
type wrappedParagraph struct{ domain.Paragraph }

// TestParagraphSerializer_WrappedParagraphDegradesGracefully confirms that a
// domain.Paragraph implementation which doesn't expose the concrete type's
// *Set() methods (any third-party or wrapped implementation) neither panics
// nor gets an explicit zero it never asked for — it falls back to the
// zero-value gate, exactly like before #69.
func TestParagraphSerializer_WrappedParagraphDegradesGracefully(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	if err := para.SetStyle(domain.StyleIDHeading1); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	if err := para.SetSpacingAfter(0); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}

	ser := serializer.NewParagraphSerializer()
	xmlPara := ser.Serialize(wrappedParagraph{para})

	if xmlPara.Properties != nil && xmlPara.Properties.Spacing != nil {
		t.Errorf("expected no direct w:spacing through a wrapped domain.Paragraph (degraded behavior), got %+v", xmlPara.Properties.Spacing)
	}
}

// TestSerializeSectionParts_HeaderTable pins SerializeSectionParts' output
// for a header and a footer that each contain a paragraph/table/paragraph
// sequence. It's the only place that exercises the headerFooterPart
// interface assertion for both header AND footer at once — a missing method
// on either docxHeader or docxFooter would make that assertion fail
// silently (SerializeSectionParts just `continue`s past it), dropping the
// part from the output map with no build error. Asserting the map here
// turns that into a loud test failure instead.
func TestSerializeSectionParts_HeaderTable(t *testing.T) {
	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection() error = %v", err)
	}

	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header() error = %v", err)
	}
	if _, err := header.AddParagraph(); err != nil {
		t.Fatalf("header.AddParagraph() error = %v", err)
	}
	if _, err := header.AddTable(1, 1); err != nil {
		t.Fatalf("header.AddTable() error = %v", err)
	}
	if _, err := header.AddParagraph(); err != nil {
		t.Fatalf("header.AddParagraph() [2] error = %v", err)
	}

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer() error = %v", err)
	}
	if _, err := footer.AddParagraph(); err != nil {
		t.Fatalf("footer.AddParagraph() error = %v", err)
	}
	if _, err := footer.AddTable(1, 1); err != nil {
		t.Fatalf("footer.AddTable() error = %v", err)
	}
	if _, err := footer.AddParagraph(); err != nil {
		t.Fatalf("footer.AddParagraph() [2] error = %v", err)
	}

	// Relationship IDs/target paths for headers and footers are only
	// assigned as a side effect of WriteTo (prepareHeaderFooterRelationships)
	// -- SerializeSectionParts skips any part without one, by design (it
	// mirrors what a document that never attached the part to a section
	// looks like). Write to a discarded buffer to trigger that assignment,
	// then inspect the same in-memory doc.
	if _, err := doc.WriteTo(io.Discard); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	ser := serializer.NewDocumentSerializer()
	headers, footers := ser.SerializeSectionParts(doc)

	xmlHeader, ok := headers["header1.xml"]
	if !ok {
		t.Fatalf("headers map missing \"header1.xml\"; got keys %v", mapKeys(headers))
	}
	if len(xmlHeader.Content) != 3 {
		t.Fatalf("len(header Content) = %d, want 3", len(xmlHeader.Content))
	}
	if _, ok := xmlHeader.Content[0].(*xmlstructs.Paragraph); !ok {
		t.Errorf("header Content[0] = %T, want *xmlstructs.Paragraph", xmlHeader.Content[0])
	}
	if _, ok := xmlHeader.Content[1].(*xmlstructs.Table); !ok {
		t.Errorf("header Content[1] = %T, want *xmlstructs.Table", xmlHeader.Content[1])
	}
	if _, ok := xmlHeader.Content[2].(*xmlstructs.Paragraph); !ok {
		t.Errorf("header Content[2] = %T, want *xmlstructs.Paragraph", xmlHeader.Content[2])
	}

	xmlFooter, ok := footers["footer1.xml"]
	if !ok {
		t.Fatalf("footers map missing \"footer1.xml\"; got keys %v", mapKeys(footers))
	}
	if len(xmlFooter.Content) != 3 {
		t.Fatalf("len(footer Content) = %d, want 3", len(xmlFooter.Content))
	}
	if _, ok := xmlFooter.Content[0].(*xmlstructs.Paragraph); !ok {
		t.Errorf("footer Content[0] = %T, want *xmlstructs.Paragraph", xmlFooter.Content[0])
	}
	if _, ok := xmlFooter.Content[1].(*xmlstructs.Table); !ok {
		t.Errorf("footer Content[1] = %T, want *xmlstructs.Table", xmlFooter.Content[1])
	}
	if _, ok := xmlFooter.Content[2].(*xmlstructs.Paragraph); !ok {
		t.Errorf("footer Content[2] = %T, want *xmlstructs.Paragraph", xmlFooter.Content[2])
	}
}

func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
