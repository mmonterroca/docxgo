// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
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
