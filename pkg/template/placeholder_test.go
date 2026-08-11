// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"regexp"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
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

// TestFindPlaceholders_InHeaderTableCell pins walkHeaderFooterTables: before
// PR 2b, a header table's cells weren't reached by walkParagraphs at all
// (the header/footer arm only ever called Paragraphs()), so a placeholder
// there was invisible to FindPlaceholders. Location.Type stays LocationHeader
// (not a distinct table-cell type) so the existing skipHeaderFooter check in
// ReplaceText/MergeTemplate keeps working, but TableIndex/RowIndex/CellIndex
// are populated alongside it.
func TestFindPlaceholders_InHeaderTableCell(t *testing.T) {
	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	table, err := header.AddTable(1, 1)
	if err != nil {
		t.Fatalf("header.AddTable: %v", err)
	}
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)
	para, _ := cell.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{HeaderCellValue}}")

	results := FindPlaceholders(doc)

	if len(results) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(results))
	}
	if results[0].Name != "HeaderCellValue" {
		t.Errorf("expected \"HeaderCellValue\", got %q", results[0].Name)
	}
	if results[0].Location.Type != LocationHeader {
		t.Errorf("expected LocationHeader, got %d", results[0].Location.Type)
	}
	if results[0].Location.TableIndex != 0 || results[0].Location.RowIndex != 0 || results[0].Location.CellIndex != 0 {
		t.Errorf("expected TableIndex/RowIndex/CellIndex = 0/0/0, got %d/%d/%d",
			results[0].Location.TableIndex, results[0].Location.RowIndex, results[0].Location.CellIndex)
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
