/*
MIT License

Copyright (c) 2025 Misael Monterroca <misael@monterroca.com>

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

package template

import (
	"os"
	"testing"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
)

// TestIntegration_CreateTemplateAndMerge tests the full round trip:
// create doc → add placeholders → save → open → merge → save → verify
func TestIntegration_CreateTemplateAndMerge(t *testing.T) {
	// Create template
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Dear {{name}}, your order #{{order_id}} is confirmed.")
	r.SetBold(true)

	templatePath := t.TempDir() + "/template.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	// Open and merge
	doc2, err := docx.OpenDocument(templatePath)
	if err != nil {
		t.Fatalf("failed to open template: %v", err)
	}

	data := MergeData{
		"name":     "Alice",
		"order_id": "12345",
	}
	if err := MergeTemplate(doc2, data); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Save merged
	outputPath := t.TempDir() + "/merged.docx"
	if err := doc2.SaveAs(outputPath); err != nil {
		t.Fatalf("failed to save merged: %v", err)
	}

	// Verify merged doc
	doc3, err := docx.OpenDocument(outputPath)
	if err != nil {
		t.Fatalf("failed to reopen merged: %v", err)
	}

	paras := doc3.Paragraphs()
	if len(paras) == 0 {
		t.Fatal("expected at least 1 paragraph")
	}

	var fullText string
	for _, run := range paras[0].Runs() {
		fullText += run.Text()
	}
	if fullText != "Dear Alice, your order #12345 is confirmed." {
		t.Errorf("expected merged text, got %q", fullText)
	}
}

// TestIntegration_ComplexTemplate tests multiple sections, tables, and mixed content.
func TestIntegration_ComplexTemplate(t *testing.T) {
	doc := core.NewDocument()

	// Body paragraph
	p1, _ := doc.AddParagraph()
	r1, _ := p1.AddRun()
	r1.SetText("Report for {{company}}")
	r1.SetBold(true)

	// Table with placeholders
	table, _ := doc.AddTable(2, 2)
	rows := table.Rows()

	// Header row
	c00, _ := rows[0].Cell(0)
	ph00, _ := c00.AddParagraph()
	rh0, _ := ph00.AddRun()
	rh0.SetText("Metric")

	c01, _ := rows[0].Cell(1)
	ph01, _ := c01.AddParagraph()
	rh1, _ := ph01.AddRun()
	rh1.SetText("Value")

	// Data row with placeholders
	c10, _ := rows[1].Cell(0)
	pd0, _ := c10.AddParagraph()
	rd0, _ := pd0.AddRun()
	rd0.SetText("{{metric_name}}")

	c11, _ := rows[1].Cell(1)
	pd1, _ := c11.AddParagraph()
	rd1, _ := pd1.AddRun()
	rd1.SetText("{{metric_value}}")

	// Another paragraph after table
	p2, _ := doc.AddParagraph()
	r2, _ := p2.AddRun()
	r2.SetText("Generated on {{date}}")

	// Save, open, merge
	templatePath := t.TempDir() + "/complex_template.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	doc2, err := docx.OpenDocument(templatePath)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	data := MergeData{
		"company":      "TechCorp",
		"metric_name":  "Revenue",
		"metric_value": "$1.2M",
		"date":         "2025-01-15",
	}

	if err := MergeTemplate(doc2, data); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	outputPath := t.TempDir() + "/complex_merged.docx"
	if err := doc2.SaveAs(outputPath); err != nil {
		t.Fatalf("failed to save merged: %v", err)
	}

	// Verify
	doc3, err := docx.OpenDocument(outputPath)
	if err != nil {
		t.Fatalf("failed to reopen: %v", err)
	}

	paras := doc3.Paragraphs()
	if len(paras) == 0 {
		t.Fatal("no paragraphs found")
	}

	firstText := paras[0].Runs()[0].Text()
	if firstText != "Report for TechCorp" {
		t.Errorf("expected 'Report for TechCorp', got %q", firstText)
	}

	// Check table cells were merged
	tables := doc3.Tables()
	if len(tables) == 0 {
		t.Fatal("no tables found")
	}

	dataRow := tables[0].Rows()[1]
	cell0, _ := dataRow.Cell(0)
	cell0Paras := cell0.Paragraphs()
	if len(cell0Paras) > 0 && len(cell0Paras[0].Runs()) > 0 {
		cellText := cell0Paras[0].Runs()[0].Text()
		if cellText != "Revenue" {
			t.Errorf("expected cell text 'Revenue', got %q", cellText)
		}
	}
}

// TestIntegration_BatchMerge tests merging the same template with multiple data sets.
func TestIntegration_BatchMerge(t *testing.T) {
	// Create simple template
	doc := core.NewDocument()
	p, _ := doc.AddParagraph()
	r, _ := p.AddRun()
	r.SetText("Hello {{name}}, welcome to {{event}}!")

	templatePath := t.TempDir() + "/batch_template.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Batch merge
	records := []MergeData{
		{"name": "Alice", "event": "GoConf 2025"},
		{"name": "Bob", "event": "GoConf 2025"},
		{"name": "Charlie", "event": "GoConf 2025"},
		{"name": "Diana", "event": "GoConf 2025"},
		{"name": "Eve", "event": "GoConf 2025"},
	}

	outputDir := t.TempDir()
	for i, data := range records {
		d, err := docx.OpenDocument(templatePath)
		if err != nil {
			t.Fatalf("record %d: failed to open: %v", i, err)
		}

		if err := MergeTemplate(d, data); err != nil {
			t.Fatalf("record %d: merge failed: %v", i, err)
		}

		outPath := outputDir + "/" + data["name"] + ".docx"
		if err := d.SaveAs(outPath); err != nil {
			t.Fatalf("record %d: failed to save: %v", i, err)
		}

		// Verify each output
		d2, err := docx.OpenDocument(outPath)
		if err != nil {
			t.Fatalf("record %d: failed to reopen: %v", i, err)
		}

		paras := d2.Paragraphs()
		if len(paras) == 0 {
			t.Fatalf("record %d: no paragraphs", i)
		}

		var text string
		for _, run := range paras[0].Runs() {
			text += run.Text()
		}
		expected := "Hello " + data["name"] + ", welcome to GoConf 2025!"
		if text != expected {
			t.Errorf("record %d: expected %q, got %q", i, expected, text)
		}
	}
}

// TestIntegration_PreservesDocumentStructure verifies formatting and structure survive merge.
func TestIntegration_PreservesDocumentStructure(t *testing.T) {
	doc := core.NewDocument()

	// Bold paragraph
	p1, _ := doc.AddParagraph()
	p1.SetAlignment(domain.AlignmentCenter)
	r1, _ := p1.AddRun()
	r1.SetText("{{title}}")
	r1.SetBold(true)
	r1.SetSize(28)

	// Italic paragraph
	p2, _ := doc.AddParagraph()
	r2, _ := p2.AddRun()
	r2.SetText("By {{author}}")
	r2.SetItalic(true)

	// Table
	table, _ := doc.AddTable(1, 2)
	c0, _ := table.Rows()[0].Cell(0)
	cp0, _ := c0.AddParagraph()
	cr0, _ := cp0.AddRun()
	cr0.SetText("{{label}}")
	cr0.SetBold(true)

	c1, _ := table.Rows()[0].Cell(1)
	cp1, _ := c1.AddParagraph()
	cr1, _ := cp1.AddRun()
	cr1.SetText("{{value}}")

	templatePath := t.TempDir() + "/structure_template.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	doc2, err := docx.OpenDocument(templatePath)
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	data := MergeData{
		"title":  "Annual Report",
		"author": "Jane Doe",
		"label":  "Status",
		"value":  "Complete",
	}

	if err := MergeTemplate(doc2, data); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	outputPath := t.TempDir() + "/structure_merged.docx"
	if err := doc2.SaveAs(outputPath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify structure preservation
	doc3, err := docx.OpenDocument(outputPath)
	if err != nil {
		t.Fatalf("failed to reopen: %v", err)
	}

	paras := doc3.Paragraphs()
	if len(paras) < 2 {
		t.Fatalf("expected at least 2 paragraphs, got %d", len(paras))
	}

	// Check title formatting
	titleRuns := paras[0].Runs()
	if len(titleRuns) > 0 {
		if titleRuns[0].Text() != "Annual Report" {
			t.Errorf("title text: expected 'Annual Report', got %q", titleRuns[0].Text())
		}
		if !titleRuns[0].Bold() {
			t.Error("expected title to be bold")
		}
		if titleRuns[0].Size() != 28 {
			t.Errorf("expected title size 28, got %d", titleRuns[0].Size())
		}
	}

	// Check subtitle formatting
	subtitleRuns := paras[1].Runs()
	if len(subtitleRuns) > 0 {
		if subtitleRuns[0].Text() != "By Jane Doe" {
			t.Errorf("subtitle text: expected 'By Jane Doe', got %q", subtitleRuns[0].Text())
		}
		if !subtitleRuns[0].Italic() {
			t.Error("expected subtitle to be italic")
		}
	}

	// Check table survived
	tables := doc3.Tables()
	if len(tables) == 0 {
		t.Error("expected at least 1 table")
	}

	// Verify output file exists and has reasonable size
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file stat: %v", err)
	}
	if info.Size() < 1000 {
		t.Errorf("output file seems too small: %d bytes", info.Size())
	}
}
