package serializer
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

package serializer_test

import (
	"encoding/xml"
	"testing"

	"github.com/SlideLang/go-docx/v2/domain"
	"github.com/SlideLang/go-docx/v2/internal/core"
	"github.com/SlideLang/go-docx/v2/internal/serializer"
)

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
	data, err := xml.MarshalIndent(xmlRun, "", "  ")
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

	if len(xmlPara.Runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(xmlPara.Runs))
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
	if len(firstCell.Paragraphs) == 0 {
		t.Fatal("expected at least one paragraph in first cell")
	}
	if len(firstCell.Paragraphs[0].Runs) == 0 {
		t.Fatal("expected at least one run in first paragraph")
	}
	if firstCell.Paragraphs[0].Runs[0].Text.Content != "Cell 0,0" {
		t.Errorf("expected 'Cell 0,0', got %q", firstCell.Paragraphs[0].Runs[0].Text.Content)
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

	if len(xmlDoc.Body.Paragraphs) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(xmlDoc.Body.Paragraphs))
	}

	if len(xmlDoc.Body.Tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(xmlDoc.Body.Tables))
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
	doc.SetMetadata(meta)

	// Add content
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("Hello, World!")
	run.SetBold(true)

	ser := serializer.NewDocumentSerializer()
	
	// Serialize document
	xmlDoc := ser.SerializeDocument(doc)
	data, err := xml.MarshalIndent(xmlDoc, "", "  ")
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

	// Serialize core properties
	coreProps := ser.SerializeCoreProperties(meta)
	propsData, err := xml.MarshalIndent(coreProps, "", "  ")
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
