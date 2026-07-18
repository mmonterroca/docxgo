// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"testing"

	"github.com/mmonterroca/docxgo/v2/internal/core"
)

func TestConsolidateRuns_EmptyParagraph(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	ConsolidateRuns(para) // should not panic

	if len(para.Runs()) != 0 {
		t.Errorf("expected 0 runs, got %d", len(para.Runs()))
	}
}

func TestConsolidateRuns_SingleRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("hello")
	r.SetBold(true)

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Text() != "hello" {
		t.Errorf("expected 'hello', got %q", runs[0].Text())
	}
	if !runs[0].Bold() {
		t.Error("expected bold to be preserved")
	}
}

func TestConsolidateRuns_IdenticalFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// 3 runs with same default formatting
	r1, _ := para.AddRun()
	r1.SetText("one ")
	r2, _ := para.AddRun()
	r2.SetText("two ")
	r3, _ := para.AddRun()
	r3.SetText("three")

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 merged run, got %d", len(runs))
	}
	if runs[0].Text() != "one two three" {
		t.Errorf("expected 'one two three', got %q", runs[0].Text())
	}
}

func TestConsolidateRuns_DifferentFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("bold")
	r1.SetBold(true)

	r2, _ := para.AddRun()
	r2.SetText("normal")

	r3, _ := para.AddRun()
	r3.SetText("bold again")
	r3.SetBold(true)

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs (no merge), got %d", len(runs))
	}
	if runs[0].Text() != "bold" || !runs[0].Bold() {
		t.Errorf("run[0]: expected bold 'bold', got %q bold=%v", runs[0].Text(), runs[0].Bold())
	}
	if runs[1].Text() != "normal" || runs[1].Bold() {
		t.Errorf("run[1]: expected non-bold 'normal', got %q bold=%v", runs[1].Text(), runs[1].Bold())
	}
	if runs[2].Text() != "bold again" || !runs[2].Bold() {
		t.Errorf("run[2]: expected bold 'bold again', got %q bold=%v", runs[2].Text(), runs[2].Bold())
	}
}

func TestConsolidateRuns_SplitPlaceholder(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Simulate Word splitting {{name}} across 3 runs
	r1, _ := para.AddRun()
	r1.SetText("{{")
	r1.SetBold(true)

	r2, _ := para.AddRun()
	r2.SetText("name")
	r2.SetBold(true)

	r3, _ := para.AddRun()
	r3.SetText("}}")
	r3.SetBold(true)

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 merged run, got %d", len(runs))
	}
	if runs[0].Text() != "{{name}}" {
		t.Errorf("expected '{{name}}', got %q", runs[0].Text())
	}
	if !runs[0].Bold() {
		t.Error("expected bold formatting to be preserved")
	}
}

func TestConsolidateRuns_PreservesFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("hello ")
	r1.SetBold(true)
	r1.SetItalic(true)
	r1.SetSize(28)

	r2, _ := para.AddRun()
	r2.SetText("world")
	r2.SetBold(true)
	r2.SetItalic(true)
	r2.SetSize(28)

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	r := runs[0]
	if r.Text() != "hello world" {
		t.Errorf("expected 'hello world', got %q", r.Text())
	}
	if !r.Bold() {
		t.Error("expected bold")
	}
	if !r.Italic() {
		t.Error("expected italic")
	}
	if r.Size() != 28 {
		t.Errorf("expected size 28, got %d", r.Size())
	}
}

func TestConsolidateRuns_MixedMergeable(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Runs: [same] [same] [different] [same] [same]
	r1, _ := para.AddRun()
	r1.SetText("A")

	r2, _ := para.AddRun()
	r2.SetText("B")

	r3, _ := para.AddRun()
	r3.SetText("C")
	r3.SetBold(true)

	r4, _ := para.AddRun()
	r4.SetText("D")

	r5, _ := para.AddRun()
	r5.SetText("E")

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	if runs[0].Text() != "AB" {
		t.Errorf("run[0]: expected 'AB', got %q", runs[0].Text())
	}
	if runs[1].Text() != "C" || !runs[1].Bold() {
		t.Errorf("run[1]: expected bold 'C', got %q bold=%v", runs[1].Text(), runs[1].Bold())
	}
	if runs[2].Text() != "DE" {
		t.Errorf("run[2]: expected 'DE', got %q", runs[2].Text())
	}
}

func TestConsolidateRuns_WithBreaks(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("before")

	r2, _ := para.AddRun()
	r2.SetText("") // run with a break should not merge
	r2.AddBreak(1) // page break

	r3, _ := para.AddRun()
	r3.SetText("after")

	ConsolidateRuns(para)

	runs := para.Runs()
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs (break prevents merge), got %d", len(runs))
	}
}

func TestConsolidateRuns_Idempotent(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("{{")

	r2, _ := para.AddRun()
	r2.SetText("name}}")

	ConsolidateRuns(para)
	ConsolidateRuns(para) // second call should be a no-op

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run after double consolidation, got %d", len(runs))
	}
	if runs[0].Text() != "{{name}}" {
		t.Errorf("expected '{{name}}', got %q", runs[0].Text())
	}
}
