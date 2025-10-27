/*
MIT License

Copyright (c) 2025 Misael Monterroca

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

package core

import (
	"testing"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
)

func TestTableCellMerge_Horizontal(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl1", 3, 4, idGen, relManager)

	// Get first row, first cell
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0) error = %v", err)
	}

	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) error = %v", err)
	}

	// Merge 3 columns horizontally
	if err := cell.Merge(3, 1); err != nil {
		t.Fatalf("Merge(3, 1) error = %v", err)
	}

	// Check GridSpan
	if cell.GridSpan() != 3 {
		t.Errorf("GridSpan() = %v, want 3", cell.GridSpan())
	}

	// VMerge should be None for horizontal-only merge
	if cell.VMerge() != domain.VMergeNone {
		t.Errorf("VMerge() = %v, want VMergeNone", cell.VMerge())
	}
}

func TestTableCellMerge_Vertical(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl2", 4, 3, idGen, relManager)

	// Get first row, first cell
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0) error = %v", err)
	}

	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) error = %v", err)
	}

	// Merge 3 rows vertically
	if err := cell.Merge(1, 3); err != nil {
		t.Fatalf("Merge(1, 3) error = %v", err)
	}

	// Check VMerge
	if cell.VMerge() != domain.VMergeRestart {
		t.Errorf("VMerge() = %v, want VMergeRestart", cell.VMerge())
	}

	// GridSpan should be 1 for vertical-only merge
	if cell.GridSpan() != 1 {
		t.Errorf("GridSpan() = %v, want 1", cell.GridSpan())
	}
}

func TestTableCellMerge_Both(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl3", 4, 4, idGen, relManager)

	// Get cell at (0,0)
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0) error = %v", err)
	}

	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) error = %v", err)
	}

	// Merge 2x2 region
	if err := cell.Merge(2, 2); err != nil {
		t.Fatalf("Merge(2, 2) error = %v", err)
	}

	// Check both directions
	if cell.GridSpan() != 2 {
		t.Errorf("GridSpan() = %v, want 2", cell.GridSpan())
	}

	if cell.VMerge() != domain.VMergeRestart {
		t.Errorf("VMerge() = %v, want VMergeRestart", cell.VMerge())
	}
}

func TestTableCellSetGridSpan(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl4", 2, 4, idGen, relManager)

	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	// Set GridSpan directly
	if err := cell.SetGridSpan(3); err != nil {
		t.Fatalf("SetGridSpan(3) error = %v", err)
	}

	if cell.GridSpan() != 3 {
		t.Errorf("GridSpan() = %v, want 3", cell.GridSpan())
	}

	// Invalid GridSpan
	if err := cell.SetGridSpan(0); err == nil {
		t.Error("SetGridSpan(0) should return error")
	}
}

func TestTableCellSetVMerge(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl5", 3, 2, idGen, relManager)

	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	// Set VMerge to restart
	if err := cell.SetVMerge(domain.VMergeRestart); err != nil {
		t.Fatalf("SetVMerge(VMergeRestart) error = %v", err)
	}

	if cell.VMerge() != domain.VMergeRestart {
		t.Errorf("VMerge() = %v, want VMergeRestart", cell.VMerge())
	}

	// Set to continue
	row2, _ := table.Row(1)
	cell2, _ := row2.Cell(0)

	if err := cell2.SetVMerge(domain.VMergeContinue); err != nil {
		t.Fatalf("SetVMerge(VMergeContinue) error = %v", err)
	}

	if cell2.VMerge() != domain.VMergeContinue {
		t.Errorf("VMerge() = %v, want VMergeContinue", cell2.VMerge())
	}
}

func TestTableCellNestedTable(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl6", 2, 2, idGen, relManager)

	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	// Add nested table
	nested, err := cell.AddTable(2, 3)
	if err != nil {
		t.Fatalf("AddTable(2, 3) error = %v", err)
	}

	if nested == nil {
		t.Fatal("AddTable returned nil table")
	}

	// Verify nested table dimensions
	if len(nested.Rows()) != 2 {
		t.Errorf("Nested table rows = %v, want 2", len(nested.Rows()))
	}

	nestedRow, _ := nested.Row(0)
	if len(nestedRow.Cells()) != 3 {
		t.Errorf("Nested table cols = %v, want 3", len(nestedRow.Cells()))
	}

	// Verify Tables() accessor
	tables := cell.Tables()
	if len(tables) != 1 {
		t.Errorf("Tables() length = %v, want 1", len(tables))
	}
}

func TestTableCellMultipleNestedTables(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl7", 1, 1, idGen, relManager)

	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	// Add multiple nested tables
	nested1, _ := cell.AddTable(2, 2)
	nested2, _ := cell.AddTable(3, 3)

	tables := cell.Tables()
	if len(tables) != 2 {
		t.Errorf("Tables() length = %v, want 2", len(tables))
	}

	// Verify both tables exist
	if len(nested1.Rows()) != 2 {
		t.Errorf("Nested1 rows = %v, want 2", len(nested1.Rows()))
	}

	if len(nested2.Rows()) != 3 {
		t.Errorf("Nested2 rows = %v, want 3", len(nested2.Rows()))
	}
}

func TestTableStyle(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl8", 3, 3, idGen, relManager)

	// Default style
	style := table.Style()
	if style.Name != "" {
		t.Errorf("Default style name = %v, want empty", style.Name)
	}

	// Set predefined style
	if err := table.SetStyle(domain.TableStyleGrid); err != nil {
		t.Fatalf("SetStyle(TableStyleGrid) error = %v", err)
	}

	style = table.Style()
	if style.Name != "TableGrid" {
		t.Errorf("Style name = %v, want TableGrid", style.Name)
	}

	// Set custom style
	custom := domain.TableStyle{Name: "MyCustomStyle"}
	if err := table.SetStyle(custom); err != nil {
		t.Fatalf("SetStyle(custom) error = %v", err)
	}

	style = table.Style()
	if style.Name != "MyCustomStyle" {
		t.Errorf("Style name = %v, want MyCustomStyle", style.Name)
	}
}

func TestTableCellMerge_InvalidArguments(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl9", 2, 2, idGen, relManager)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	tests := []struct {
		name    string
		cols    int
		rows    int
		wantErr bool
	}{
		{"zero cols", 0, 1, true},
		{"zero rows", 1, 0, true},
		{"negative cols", -1, 1, true},
		{"negative rows", 1, -1, true},
		{"valid 1x1", 1, 1, false},
		{"valid 2x2", 2, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cell.Merge(tt.cols, tt.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge(%d, %d) error = %v, wantErr %v", tt.cols, tt.rows, err, tt.wantErr)
			}
		})
	}
}

func TestTableCellAddTable_InvalidArguments(t *testing.T) {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)

	table := NewTable("tbl10", 1, 1, idGen, relManager)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"zero rows", 0, 2, true},
		{"zero cols", 2, 0, true},
		{"negative rows", -1, 2, true},
		{"negative cols", 2, -1, true},
		{"valid 1x1", 1, 1, false},
		{"valid 5x5", 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cell.AddTable(tt.rows, tt.cols)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTable(%d, %d) error = %v, wantErr %v", tt.rows, tt.cols, err, tt.wantErr)
			}
		})
	}
}
