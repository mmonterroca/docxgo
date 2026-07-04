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
	"regexp"
	"testing"

	"github.com/mmonterroca/docxgo/v2/internal/core"
)

func TestFindPlaceholders_Simple(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello {{Name}}, welcome!")

	results := FindPlaceholders(doc)

	if len(results) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(results))
	}
	if results[0].Name != "Name" {
		t.Errorf("expected name 'Name', got %q", results[0].Name)
	}
	if results[0].FullMatch != "{{Name}}" {
		t.Errorf("expected full match '{{Name}}', got %q", results[0].FullMatch)
	}
	if results[0].Location.Type != LocationParagraph {
		t.Errorf("expected LocationParagraph, got %d", results[0].Location.Type)
	}
}

func TestFindPlaceholders_Multiple(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{FirstName}} {{LastName}}")

	results := FindPlaceholders(doc)

	if len(results) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(results))
	}
	if results[0].Name != "FirstName" {
		t.Errorf("expected 'FirstName', got %q", results[0].Name)
	}
	if results[1].Name != "LastName" {
		t.Errorf("expected 'LastName', got %q", results[1].Name)
	}
}

func TestFindPlaceholders_MultipleParagraphs(t *testing.T) {
	doc := core.NewDocument()

	p1, _ := doc.AddParagraph()
	r1, _ := p1.AddRun()
	r1.SetText("Dear {{Name}},")

	p2, _ := doc.AddParagraph()
	r2, _ := p2.AddRun()
	r2.SetText("Your order {{OrderID}} is ready.")

	results := FindPlaceholders(doc)

	if len(results) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(results))
	}
	if results[0].Name != "Name" {
		t.Errorf("expected 'Name', got %q", results[0].Name)
	}
	if results[1].Name != "OrderID" {
		t.Errorf("expected 'OrderID', got %q", results[1].Name)
	}
}

func TestFindPlaceholders_SplitAcrossRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Simulate Word splitting "{{Name}}" across 3 runs
	r1, _ := para.AddRun()
	r1.SetText("Hello {{")
	r2, _ := para.AddRun()
	r2.SetText("Name")
	r3, _ := para.AddRun()
	r3.SetText("}}!")

	results := FindPlaceholders(doc)

	// ConsolidateRuns should heal the split, then find the placeholder
	if len(results) != 1 {
		t.Fatalf("expected 1 placeholder after consolidation, got %d", len(results))
	}
	if results[0].Name != "Name" {
		t.Errorf("expected 'Name', got %q", results[0].Name)
	}
}

func TestFindPlaceholders_NoPlaceholders(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Just plain text without any placeholders.")

	results := FindPlaceholders(doc)

	if len(results) != 0 {
		t.Errorf("expected 0 placeholders, got %d", len(results))
	}
}

func TestFindPlaceholders_InTable(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(2, 2)

	rows := table.Rows()
	cells := rows[0].Cells()
	para, _ := cells[0].AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{CellValue}}")

	results := FindPlaceholders(doc)

	if len(results) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(results))
	}
	if results[0].Name != "CellValue" {
		t.Errorf("expected 'CellValue', got %q", results[0].Name)
	}
	if results[0].Location.Type != LocationTableCell {
		t.Errorf("expected LocationTableCell, got %d", results[0].Location.Type)
	}
}

func TestFindPlaceholders_DottedName(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{.Name}}")

	results := FindPlaceholders(doc)

	if len(results) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(results))
	}
	if results[0].Name != "Name" {
		t.Errorf("expected 'Name', got %q", results[0].Name)
	}
}

func TestFindPlaceholders_CustomPattern(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello ${Name}, welcome to ${Company}!")

	pattern := regexp.MustCompile(`\$\{(\w+)\}`)
	results := FindPlaceholdersCustom(doc, pattern)

	if len(results) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(results))
	}
	if results[0].Name != "Name" {
		t.Errorf("expected 'Name', got %q", results[0].Name)
	}
	if results[1].Name != "Company" {
		t.Errorf("expected 'Company', got %q", results[1].Name)
	}
}

func TestPlaceholderNames_Deduplicated(t *testing.T) {
	doc := core.NewDocument()

	p1, _ := doc.AddParagraph()
	r1, _ := p1.AddRun()
	r1.SetText("{{Name}} is {{Name}}")

	p2, _ := doc.AddParagraph()
	r2, _ := p2.AddRun()
	r2.SetText("{{Email}}")

	names := PlaceholderNames(doc)

	if len(names) != 2 {
		t.Fatalf("expected 2 unique names, got %d", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["Name"] {
		t.Error("expected 'Name' in names")
	}
	if !nameSet["Email"] {
		t.Error("expected 'Email' in names")
	}
}

func TestFindPlaceholders_EmptyDocument(t *testing.T) {
	doc := core.NewDocument()

	results := FindPlaceholders(doc)

	if len(results) != 0 {
		t.Errorf("expected 0 placeholders, got %d", len(results))
	}
}
