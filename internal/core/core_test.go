// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core_test

import (
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestNewDocument(t *testing.T) {
	doc := core.NewDocument()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}

	paras := doc.Paragraphs()
	if len(paras) != 0 {
		t.Errorf("expected 0 paragraphs, got %d", len(paras))
	}
}

func TestDocument_AddParagraph(t *testing.T) {
	doc := core.NewDocument()

	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph failed: %v", err)
	}
	if para == nil {
		t.Fatal("expected non-nil paragraph")
	}

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(paras))
	}
}

func TestDocument_AddTable(t *testing.T) {
	doc := core.NewDocument()

	table, err := doc.AddTable(3, 4)
	if err != nil {
		t.Fatalf("AddTable failed: %v", err)
	}
	if table == nil {
		t.Fatal("expected non-nil table")
	}

	if table.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", table.RowCount())
	}
	if table.ColumnCount() != 4 {
		t.Errorf("expected 4 columns, got %d", table.ColumnCount())
	}
}

func TestDocument_AddSectionWithBreak(t *testing.T) {
	doc := core.NewDocument()

	para1, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph failed: %v", err)
	}
	_, _ = para1.AddRun()

	secondSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeContinuous)
	if err != nil {
		t.Fatalf("AddSectionWithBreak failed: %v", err)
	}
	if secondSection == nil {
		t.Fatal("expected non-nil section")
	}

	_, err = doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph after section failed: %v", err)
	}

	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	blocks := doc.Blocks()
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks (paragraph, break, paragraph), got %d", len(blocks))
	}

	if blocks[0].Paragraph == nil {
		t.Error("expected first block to be a paragraph")
	}
	if blocks[1].SectionBreak == nil {
		t.Fatal("expected second block to be a section break")
	}
	if blocks[1].SectionBreak.Type != domain.SectionBreakTypeContinuous {
		t.Errorf("expected section break type Continuous, got %v", blocks[1].SectionBreak.Type)
	}
	if blocks[2].Paragraph == nil {
		t.Error("expected third block to be a paragraph")
	}

	// Mutating returned slice must not affect document internals.
	blocks[0] = domain.Block{}
	if len(doc.Blocks()) != 3 {
		t.Error("mutating returned blocks slice should not affect document")
	}
}

func TestDocument_AddTable_InvalidDimensions(t *testing.T) {
	doc := core.NewDocument()

	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"zero rows", 0, 3},
		{"zero cols", 3, 0},
		{"negative rows", -1, 3},
		{"negative cols", 3, -1},
		{"too many rows", 1001, 3},
		{"too many cols", 3, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := doc.AddTable(tt.rows, tt.cols)
			if err == nil {
				t.Errorf("expected error for rows=%d, cols=%d", tt.rows, tt.cols)
			}
		})
	}
}

func TestParagraph_AddRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun failed: %v", err)
	}
	if run == nil {
		t.Fatal("expected non-nil run")
	}

	runs := para.Runs()
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
}

func TestRun_TextFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	// Test text
	text := "Hello, World!"
	err := run.SetText(text)
	if err != nil {
		t.Fatalf("SetText failed: %v", err)
	}
	if run.Text() != text {
		t.Errorf("expected text %q, got %q", text, run.Text())
	}

	// Test bold
	err = run.SetBold(true)
	if err != nil {
		t.Fatalf("SetBold failed: %v", err)
	}
	if !run.Bold() {
		t.Error("expected bold to be true")
	}

	// Test italic
	err = run.SetItalic(true)
	if err != nil {
		t.Fatalf("SetItalic failed: %v", err)
	}
	if !run.Italic() {
		t.Error("expected italic to be true")
	}

	// Test color
	err = run.SetColor(domain.ColorRed)
	if err != nil {
		t.Fatalf("SetColor failed: %v", err)
	}
	if run.Color() != domain.ColorRed {
		t.Error("expected color to be red")
	}

	// Test font size
	err = run.SetSize(24) // 12pt
	if err != nil {
		t.Fatalf("SetSize failed: %v", err)
	}
	if run.Size() != 24 {
		t.Errorf("expected size 24, got %d", run.Size())
	}
}

func TestRun_Language(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	if run.Language() != nil {
		t.Fatal("expected no language override by default")
	}

	lang := &domain.Language{Val: "fr", EastAsia: "fr", Bidi: "fr"}
	if err := run.SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage failed: %v", err)
	}
	got := run.Language()
	if got == nil || *got != *lang {
		t.Errorf("expected language %+v, got %+v", lang, got)
	}

	// Mutating the caller's copy after SetLanguage must not affect the run,
	// and mutating the returned copy must not affect the run either — same
	// copy-semantics contract as Document.SetLanguage/Language.
	lang.Val = "mutated"
	if run.Language().Val != "fr" {
		t.Error("SetLanguage must copy its argument")
	}
	got.Val = "mutated"
	if run.Language().Val != "fr" {
		t.Error("Language must return a copy")
	}

	if err := run.SetLanguage(nil); err != nil {
		t.Fatalf("SetLanguage(nil) failed: %v", err)
	}
	if run.Language() != nil {
		t.Error("expected SetLanguage(nil) to clear the override")
	}
}

func TestRun_SetLanguage_RejectsEmptyTag(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	if err := run.SetLanguage(&domain.Language{}); err == nil {
		t.Error("expected an error for a Language with no Val/EastAsia/Bidi set")
	}
}

func TestRun_SetSize_Validation(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()

	tests := []struct {
		name string
		size int
		ok   bool
	}{
		{"minimum size", constants.MinFontSize, true},
		{"maximum size", constants.MaxFontSize, true},
		{"below minimum", constants.MinFontSize - 1, false},
		{"above maximum", constants.MaxFontSize + 1, false},
		{"normal size", 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run.SetSize(tt.size)
			if tt.ok && err != nil {
				t.Errorf("expected no error for size %d, got %v", tt.size, err)
			}
			if !tt.ok && err == nil {
				t.Errorf("expected error for size %d", tt.size)
			}
		})
	}
}

func TestParagraph_Alignment(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	tests := []domain.Alignment{
		domain.AlignmentLeft,
		domain.AlignmentCenter,
		domain.AlignmentRight,
		domain.AlignmentJustify,
		domain.AlignmentDistribute,
	}

	for _, align := range tests {
		err := para.SetAlignment(align)
		if err != nil {
			t.Fatalf("SetAlignment(%v) failed: %v", align, err)
		}
		if para.Alignment() != align {
			t.Errorf("expected alignment %v, got %v", align, para.Alignment())
		}
	}
}

func TestParagraph_Indentation(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	indent := domain.Indentation{
		Left:      720, // 0.5 inch
		Right:     720,
		FirstLine: 360, // 0.25 inch
	}

	err := para.SetIndent(indent)
	if err != nil {
		t.Fatalf("SetIndent failed: %v", err)
	}

	result := para.Indent()
	if result.Left != indent.Left {
		t.Errorf("expected left indent %d, got %d", indent.Left, result.Left)
	}
	if result.Right != indent.Right {
		t.Errorf("expected right indent %d, got %d", indent.Right, result.Right)
	}
	if result.FirstLine != indent.FirstLine {
		t.Errorf("expected first line indent %d, got %d", indent.FirstLine, result.FirstLine)
	}
}

func TestParagraph_Indentation_BothFirstLineAndHanging(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Cannot have both first line and hanging indent
	indent := domain.Indentation{
		FirstLine: 360,
		Hanging:   360,
	}

	err := para.SetIndent(indent)
	if err == nil {
		t.Error("expected error when setting both first line and hanging indent")
	}
}

func TestTable_RowOperations(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(2, 3)

	// Test initial rows
	if table.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", table.RowCount())
	}

	// Add row
	row, err := table.AddRow()
	if err != nil {
		t.Fatalf("AddRow failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected non-nil row")
	}
	if table.RowCount() != 3 {
		t.Errorf("expected 3 rows after AddRow, got %d", table.RowCount())
	}

	// Insert row
	_, err = table.InsertRow(1)
	if err != nil {
		t.Fatalf("InsertRow failed: %v", err)
	}
	if table.RowCount() != 4 {
		t.Errorf("expected 4 rows after InsertRow, got %d", table.RowCount())
	}

	// Delete row
	err = table.DeleteRow(0)
	if err != nil {
		t.Fatalf("DeleteRow failed: %v", err)
	}
	if table.RowCount() != 3 {
		t.Errorf("expected 3 rows after DeleteRow, got %d", table.RowCount())
	}
}

func TestTableCell_AddParagraph(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 1)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	para, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph failed: %v", err)
	}
	if para == nil {
		t.Fatal("expected non-nil paragraph")
	}

	paras := cell.Paragraphs()
	if len(paras) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(paras))
	}
}

func TestTableCell_RemoveParagraph(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 1)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	p1, _ := cell.AddParagraph()
	p1.AddRun()
	p1.Runs()[0].SetText("first")

	p2, _ := cell.AddParagraph()
	p2.AddRun()
	p2.Runs()[0].SetText("second")

	p3, _ := cell.AddParagraph()
	p3.AddRun()
	p3.Runs()[0].SetText("third")

	if err := cell.RemoveParagraph(1); err != nil {
		t.Fatalf("RemoveParagraph(1) failed: %v", err)
	}

	paras := cell.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(paras))
	}
	if paras[0].Text() != "first" {
		t.Errorf("paras[0].Text() = %q, want %q", paras[0].Text(), "first")
	}
	if paras[1].Text() != "third" {
		t.Errorf("paras[1].Text() = %q, want %q", paras[1].Text(), "third")
	}
}

func TestTableCell_RemoveParagraph_OutOfRange(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 1)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)
	cell.AddParagraph()

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"equal to length", 1},
		{"way out of range", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cell.RemoveParagraph(tt.index); err == nil {
				t.Errorf("expected error for index %d, got nil", tt.index)
			}
		})
	}
}

func TestParagraph_ClearRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Add 3 runs
	r1, _ := para.AddRun()
	r1.SetText("one")
	r2, _ := para.AddRun()
	r2.SetText("two")
	r3, _ := para.AddRun()
	r3.SetText("three")

	if len(para.Runs()) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(para.Runs()))
	}

	para.ClearRuns()

	if len(para.Runs()) != 0 {
		t.Errorf("expected 0 runs after ClearRuns, got %d", len(para.Runs()))
	}
	if para.Text() != "" {
		t.Errorf("expected empty text after ClearRuns, got %q", para.Text())
	}
}

func TestParagraph_RemoveRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("first")
	r2, _ := para.AddRun()
	r2.SetText("second")
	r3, _ := para.AddRun()
	r3.SetText("third")

	// Remove middle run
	err := para.RemoveRun(1)
	if err != nil {
		t.Fatalf("RemoveRun(1) failed: %v", err)
	}

	runs := para.Runs()
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	if runs[0].Text() != "first" {
		t.Errorf("expected first run text 'first', got %q", runs[0].Text())
	}
	if runs[1].Text() != "third" {
		t.Errorf("expected second run text 'third', got %q", runs[1].Text())
	}
}

func TestParagraph_RemoveRun_OutOfRange(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	para.AddRun()

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"equal to length", 1},
		{"way out of range", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := para.RemoveRun(tt.index)
			if err == nil {
				t.Errorf("expected error for index %d, got nil", tt.index)
			}
		})
	}
}

func TestParagraph_InsertRunAt(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("A")
	r2, _ := para.AddRun()
	r2.SetText("C")

	// Insert at beginning
	rBegin, err := para.InsertRunAt(0)
	if err != nil {
		t.Fatalf("InsertRunAt(0) failed: %v", err)
	}
	rBegin.SetText("Z")

	// Insert in middle (between Z and A, which are now at 0 and 1)
	rMid, err := para.InsertRunAt(2)
	if err != nil {
		t.Fatalf("InsertRunAt(2) failed: %v", err)
	}
	rMid.SetText("B")

	// Insert at end
	rEnd, err := para.InsertRunAt(len(para.Runs()))
	if err != nil {
		t.Fatalf("InsertRunAt(end) failed: %v", err)
	}
	rEnd.SetText("D")

	runs := para.Runs()
	if len(runs) != 5 {
		t.Fatalf("expected 5 runs, got %d", len(runs))
	}

	expected := []string{"Z", "A", "B", "C", "D"}
	for i, exp := range expected {
		if runs[i].Text() != exp {
			t.Errorf("run[%d]: expected %q, got %q", i, exp, runs[i].Text())
		}
	}
}

func TestParagraph_InsertRunAt_OutOfRange(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"beyond length", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := para.InsertRunAt(tt.index)
			if err == nil {
				t.Errorf("expected error for index %d, got nil", tt.index)
			}
		})
	}
}
