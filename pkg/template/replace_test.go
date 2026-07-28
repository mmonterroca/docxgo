// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
)

func TestReplaceText_SingleRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("The answer is TODO for now.")

	result, err := ReplaceText(doc, "TODO", "42")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 || result.Skipped != 0 {
		t.Errorf("result = %+v, want 1 replaced / 0 skipped", result)
	}
	if got := para.Text(); got != "The answer is 42 for now." {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_SplitAcrossIdenticalRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	// Word-style fragmentation: same formatting, split mid-word.
	for _, chunk := range []string{"Not ", "Applic", "able"} {
		r, _ := para.AddRun()
		r.SetText(chunk)
	}

	result, err := ReplaceText(doc, "Not Applicable", "Yes")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
	if got := para.Text(); got != "Yes" {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_SpanningDifferentFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r1, _ := para.AddRun()
	r1.SetText("Status: ")
	r2, _ := para.AddRun()
	r2.SetText("PENDING")
	r2.SetBold(true)
	r3, _ := para.AddRun()
	r3.SetText(" (review)")

	// Match spans the normal/bold boundary; consolidation cannot merge these.
	result, err := ReplaceText(doc, ": PENDING (", ": DONE (")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
	if got := para.Text(); got != "Status: DONE (review)" {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_MultipleOccurrences(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("aaa bbb aaa bbb aaa")

	result, err := ReplaceText(doc, "aaa", "X")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 3 {
		t.Errorf("replaced = %d, want 3", result.Replaced)
	}
	if got := para.Text(); got != "X bbb X bbb X" {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_ReplacementContainsFind(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("go go")

	result, err := ReplaceText(doc, "go", "go faster")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 2 {
		t.Errorf("replaced = %d, want 2", result.Replaced)
	}
	if got := para.Text(); got != "go faster go faster" {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_InTableCell(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 2)
	cell, _ := table.Rows()[0].Cell(1)
	para, _ := cell.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{answer pending}")

	result, err := ReplaceText(doc, "{answer pending}", "SOC 2 Type II")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
	if got := para.Text(); got != "SOC 2 Type II" {
		t.Errorf("cell paragraph text = %q", got)
	}
}

func TestReplaceText_SkipsSpanOverNonTextRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r1, _ := para.AddRun()
	r1.SetText("before")
	r1.SetBold(true) // prevent consolidation with r2
	r2, _ := para.AddRun()
	r2.SetText("after")
	r2.AddBreak(domain.BreakTypeLine)

	result, err := ReplaceText(doc, "beforeafter", "X")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 1 {
		t.Errorf("result = %+v, want 0 replaced / 1 skipped", result)
	}
	if got := para.Text(); got != "beforeafter" {
		t.Errorf("paragraph text = %q, want unchanged", got)
	}
}

func TestReplaceText_EmptyFindErrors(t *testing.T) {
	doc := core.NewDocument()
	if _, err := ReplaceText(doc, "", "x"); err == nil {
		t.Fatal("expected error for empty find")
	}
}

func TestReplaceText_DeleteWithEmptyReplacement(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Remove [DRAFT] marker")

	result, err := ReplaceText(doc, "[DRAFT] ", "")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
	if got := para.Text(); got != "Remove marker" {
		t.Errorf("paragraph text = %q", got)
	}
}

func TestReplaceText_NoMatch(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("nothing to see")

	result, err := ReplaceText(doc, "absent", "x")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 0 {
		t.Errorf("result = %+v, want zero result", result)
	}
	if got := para.Text(); got != "nothing to see" {
		t.Errorf("paragraph text = %q, want unchanged", got)
	}
}

// TestReplaceText_NoMatchDoesNotMutate guards against ReplaceText
// restructuring runs it never needed to touch. ConsolidateRuns only moves
// run boundaries — it never changes a paragraph's concatenated text — so a
// paragraph containing no occurrence of find must come out with exactly the
// same runs it went in with, in the body, in a table cell, and in a header.
func TestReplaceText_NoMatchDoesNotMutate(t *testing.T) {
	doc := core.NewDocument()

	bodyPara, _ := doc.AddParagraph()
	for _, chunk := range []string{"Not ", "Applic", "able"} {
		r, _ := bodyPara.AddRun()
		r.SetText(chunk)
	}

	table, _ := doc.AddTable(1, 1)
	cell, _ := table.Rows()[0].Cell(0)
	cellPara, _ := cell.AddParagraph()
	for _, chunk := range []string{"Cell ", "Text"} {
		r, _ := cellPara.AddRun()
		r.SetText(chunk)
	}

	section := doc.Sections()[0]
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	headerPara, _ := header.AddParagraph()
	for _, chunk := range []string{"Header ", "Text"} {
		r, _ := headerPara.AddRun()
		r.SetText(chunk)
	}

	wantBodyRuns := len(bodyPara.Runs())
	wantCellRuns := len(cellPara.Runs())
	wantHeaderRuns := len(headerPara.Runs())

	result, err := ReplaceText(doc, "absent from every paragraph", "x")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 0 {
		t.Errorf("result = %+v, want zero result", result)
	}

	if got := len(bodyPara.Runs()); got != wantBodyRuns {
		t.Errorf("body paragraph run count = %d, want %d (unchanged)", got, wantBodyRuns)
	}
	if got := len(cellPara.Runs()); got != wantCellRuns {
		t.Errorf("table cell paragraph run count = %d, want %d (unchanged)", got, wantCellRuns)
	}
	if got := len(headerPara.Runs()); got != wantHeaderRuns {
		t.Errorf("header paragraph run count = %d, want %d (unchanged)", got, wantHeaderRuns)
	}
}

// TestReplaceText_SkipsFieldInSingleRun pins the fix for a match confined to
// a single run that also carries a field. Rewriting that run's text with
// SetText would leave the field markup pointing at text Word no longer sees,
// and Word re-renders the field's own result on open — silently discarding
// the edit — so a match here must be skipped like any other non-text run.
func TestReplaceText_SkipsFieldInSingleRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	if err := r.AddField(core.NewPageNumberField()); err != nil {
		t.Fatalf("AddField: %v", err)
	}
	if err := r.SetText("Page 1 of 10"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	result, err := ReplaceText(doc, "Page 1 of 10", "Page 1 of Y")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 1 {
		t.Errorf("result = %+v, want 0 replaced / 1 skipped", result)
	}
	if got := para.Text(); got != "Page 1 of 10" {
		t.Errorf("paragraph text = %q, want unchanged", got)
	}
	if len(r.Fields()) != 1 {
		t.Errorf("field was dropped from the run")
	}
}

// TestReplaceText_DeletionMatchesStdlibSemantics fixes the deletion
// semantics as intentional, matching strings.ReplaceAll's non-overlapping
// left-to-right scan: a match created by an earlier deletion, starting
// before the point the scan resumed from, is not retroactively found.
func TestReplaceText_DeletionMatchesStdlibSemantics(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("aabb")

	result, err := ReplaceText(doc, "ab", "")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}

	want := strings.ReplaceAll("aabb", "ab", "")
	if got := para.Text(); got != want {
		t.Errorf("paragraph text = %q, want %q (matches strings.ReplaceAll)", got, want)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
}
