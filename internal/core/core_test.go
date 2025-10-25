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

package core_test

import (
	"testing"

	"github.com/SlideLang/go-docx/domain"
	"github.com/SlideLang/go-docx/internal/core"
	"github.com/SlideLang/go-docx/pkg/constants"
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
		Left:      720,  // 0.5 inch
		Right:     720,
		FirstLine: 360,  // 0.25 inch
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
