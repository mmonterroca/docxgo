package docx

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func TestOpenDocumentVariants(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Opened document"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sample.docx")
	if err := doc.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	opened, err := OpenDocument(path)
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}
	assertParagraphText(t, opened, "Opened document")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	openedBytes, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	assertParagraphText(t, openedBytes, "Opened document")

	openedReader, err := OpenDocumentFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenDocumentFromReader: %v", err)
	}
	assertParagraphText(t, openedReader, "Opened document")
}

func assertParagraphText(t *testing.T, doc domain.Document, expected string) {
	t.Helper()
	paras := doc.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs in opened document")
	}
	if got := paras[0].Text(); got != expected {
		t.Fatalf("unexpected paragraph text: %q", got)
	}
}

// TestGridSpanPreservedAfterRoundTrip reproduces issue #25:
// horizontal cell merges (w:gridSpan) are lost after save+reopen.
func TestGridSpanPreservedAfterRoundTrip(t *testing.T) {
	doc := NewDocument()
	table, err := doc.AddTable(2, 3)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	table.SetStyle(domain.TableStyleGrid)

	row0, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if err := cell.Merge(3, 1); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	p, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	r, err := p.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := r.AddText("Spans 3 columns"); err != nil {
		t.Fatalf("AddText: %v", err)
	}

	if got := cell.GridSpan(); got != 3 {
		t.Fatalf("before save: GridSpan = %d, want 3", got)
	}

	// Round-trip through bytes
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	tables := doc2.Tables()
	if len(tables) == 0 {
		t.Fatal("expected at least one table after reopen")
	}
	row0, err = tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0) after reopen: %v", err)
	}
	cell, err = row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) after reopen: %v", err)
	}

	if got := cell.GridSpan(); got != 3 {
		t.Errorf("after reopen: GridSpan = %d, want 3", got)
	}

	// Verify continuation cells are marked so the serializer skips them.
	cell1, err := row0.Cell(1)
	if err != nil {
		t.Fatalf("Cell(1) after reopen: %v", err)
	}
	if !cell1.IsHorizontallyMergedContinuation() {
		t.Errorf("Cell(1) should be a horizontal merge continuation")
	}
	cell2, err := row0.Cell(2)
	if err != nil {
		t.Fatalf("Cell(2) after reopen: %v", err)
	}
	if !cell2.IsHorizontallyMergedContinuation() {
		t.Errorf("Cell(2) should be a horizontal merge continuation")
	}
}

// TestVMergePreservedAfterRoundTrip verifies vertical cell merges survive save+reopen.
func TestVMergePreservedAfterRoundTrip(t *testing.T) {
	doc := NewDocument()
	table, err := doc.AddTable(3, 2)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	table.SetStyle(domain.TableStyleGrid)

	// Merge rows 0-2 in column 0 vertically (3 rows, 1 col).
	row0, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if err := cell.Merge(1, 3); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	p, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	r, err := p.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := r.AddText("Vertical span"); err != nil {
		t.Fatalf("AddText: %v", err)
	}

	if got := cell.VMerge(); got != domain.VMergeRestart {
		t.Fatalf("before save: VMerge = %v, want VMergeRestart", got)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	tables := doc2.Tables()
	if len(tables) == 0 {
		t.Fatal("expected at least one table after reopen")
	}

	// Row 0, Cell 0: should be VMergeRestart
	r0, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0) after reopen: %v", err)
	}
	c0, err := r0.Cell(0)
	if err != nil {
		t.Fatalf("Row(0).Cell(0) after reopen: %v", err)
	}
	if got := c0.VMerge(); got != domain.VMergeRestart {
		t.Errorf("Row(0).Cell(0) VMerge = %v, want VMergeRestart", got)
	}

	// Row 1, Cell 0: should be VMergeContinue
	r1, err := tables[0].Row(1)
	if err != nil {
		t.Fatalf("Row(1) after reopen: %v", err)
	}
	c1, err := r1.Cell(0)
	if err != nil {
		t.Fatalf("Row(1).Cell(0) after reopen: %v", err)
	}
	if got := c1.VMerge(); got != domain.VMergeContinue {
		t.Errorf("Row(1).Cell(0) VMerge = %v, want VMergeContinue", got)
	}
}
