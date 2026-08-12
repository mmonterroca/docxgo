// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"bytes"
	"strings"
	"testing"

	docx "github.com/mmonterroca/docxgo/v2"
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

// TestReplaceText_InHeaderTableCell pins that ReplaceText reaches a
// paragraph inside a header table's cell (via walkHeaderFooterTables) on a
// freshly built document -- as opposed to TestReplaceText_SkipsPreservedHeaderAndFooter,
// which pins the opposite for a document opened from an existing .docx.
func TestReplaceText_InHeaderTableCell(t *testing.T) {
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
	r.SetText("{company}")

	result, err := ReplaceText(doc, "{company}", "Acme Corp")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("replaced = %d, want 1", result.Replaced)
	}
	if got := para.Text(); got != "Acme Corp" {
		t.Errorf("header table cell paragraph text = %q, want %q", got, "Acme Corp")
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

// TestReplaceText_ReplacesInSingleRunCarryingABreak pins the other half of
// the run-content rule. Unlike a field, a break's run still has its text
// serialized normally, and SetText leaves the break itself alone — so a
// match confined to that one run is replaceable, and skipping it would lose
// a legitimate replacement rather than prevent a bad one.
func TestReplaceText_ReplacesInSingleRunCarryingABreak(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	if err := r.SetText("Line one and TODO"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := r.AddBreak(domain.BreakTypeLine); err != nil {
		t.Fatalf("AddBreak: %v", err)
	}

	result, err := ReplaceText(doc, "TODO", "done")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 1 || result.Skipped != 0 {
		t.Errorf("result = %+v, want 1 replaced / 0 skipped", result)
	}
	if got := para.Text(); got != "Line one and done" {
		t.Errorf("paragraph text = %q", got)
	}
	if len(r.Breaks()) != 1 {
		t.Errorf("break count = %d, want 1 (preserved)", len(r.Breaks()))
	}
}

// TestReplaceText_SkipsMergefieldCarriedOnAnEmptyRun covers the shape Word
// uses for a mail-merge field: an empty run holding the MERGEFIELD followed
// by a separate run holding its display text. The field run carries no match
// text of its own, so a naive overlap test never sees it — but replacing the
// display text while the field survives leaves the saved document with both
// the field and the replacement, and a field update in Word reverts the
// visible text while the replacement lingers. merge.go handles this shape
// explicitly; replace.go must at least refuse it.
func TestReplaceText_SkipsMergefieldCarriedOnAnEmptyRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	mergeField := core.NewField(domain.FieldTypeCustom)
	if err := mergeField.SetCode("MERGEFIELD Name"); err != nil {
		t.Fatalf("SetCode: %v", err)
	}
	fieldRun, _ := para.AddRun()
	if err := fieldRun.AddField(mergeField); err != nil {
		t.Fatalf("AddField: %v", err)
	}
	// The field run carries the field but no text of its own.
	if err := fieldRun.SetText(""); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	textRun, _ := para.AddRun()
	if err := textRun.SetText("«Name»"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	result, err := ReplaceText(doc, "«Name»", "Alice")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 1 {
		t.Errorf("result = %+v, want 0 replaced / 1 skipped", result)
	}
	if got := textRun.Text(); got != "«Name»" {
		t.Errorf("display text = %q, want unchanged", got)
	}
	if len(fieldRun.Fields()) != 1 {
		t.Errorf("field was dropped from the run")
	}
}

// TestReplaceText_SkipsBreakOnAnEmptyRunBetweenMatchedRuns covers the
// standalone <w:r><w:br/></w:r> shape the reader produces. The break run
// carries no match text, so it is not part of the spliced span — but it sits
// between two runs that are, and it survives the replacement stranded in the
// middle of the text that replaced them.
func TestReplaceText_SkipsBreakOnAnEmptyRunBetweenMatchedRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello")

	breakRun, _ := para.AddRun()
	if err := breakRun.AddBreak(domain.BreakTypeLine); err != nil {
		t.Fatalf("AddBreak: %v", err)
	}
	breakRun.SetText("")

	r3, _ := para.AddRun()
	r3.SetText("World")

	result, err := ReplaceText(doc, "HelloWorld", "Hi")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 0 || result.Skipped != 1 {
		t.Errorf("result = %+v, want 0 replaced / 1 skipped", result)
	}
	if got := para.Text(); got != "HelloWorld" {
		t.Errorf("paragraph text = %q, want unchanged", got)
	}
	if len(breakRun.Breaks()) != 1 {
		t.Errorf("break count = %d, want 1", len(breakRun.Breaks()))
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

// TestReplaceText_ReachesPreservedHeaderAndFooter is the end-to-end proof
// that an edit to a round-tripped header now survives the save. It builds a
// document with "FOO" in the body, a header, and a footer, saves it, reopens
// it (which preserves the header and footer bytes for round-trip), replaces
// "FOO" with "BAR", saves again, and inspects the raw zip entries: all three
// parts must change.
//
// This inverts TestReplaceText_SkipsPreservedHeaderAndFooter, which pinned
// the old behaviour -- ReplaceText skipped every header/footer match on a
// round-tripped document, because WriteTo wrote those parts back verbatim and
// the edit could never reach the file. WriteTo now regenerates a part the
// caller actually edited, so skipping would discard replacements that do
// land. See TestOpenDocument_UntouchedHeaderStaysByteIdentical for the other
// half of the contract: a part nobody touched is still written verbatim.
func TestReplaceText_ReachesPreservedHeaderAndFooter(t *testing.T) {
	doc := core.NewDocument()
	bodyPara, _ := doc.AddParagraph()
	bodyRun, _ := bodyPara.AddRun()
	bodyRun.SetText("FOO in body")

	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	headerPara, _ := header.AddParagraph()
	headerRun, _ := headerPara.AddRun()
	headerRun.SetText("FOO in header")

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer: %v", err)
	}
	footerPara, _ := footer.AddParagraph()
	footerRun, _ := footerPara.AddRun()
	footerRun.SetText("FOO in footer")

	templatePath := t.TempDir() + "/preserved_header_footer.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	opened, err := docx.OpenDocument(templatePath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	result, err := ReplaceText(opened, "FOO", "BAR")
	if err != nil {
		t.Fatalf("ReplaceText: %v", err)
	}
	if result.Replaced != 3 {
		t.Errorf("replaced = %d, want 3 (body + header + footer)", result.Replaced)
	}
	if result.Skipped != 0 {
		t.Errorf("skipped = %d, want 0", result.Skipped)
	}

	var buf bytes.Buffer
	if _, err := opened.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	savedBytes := buf.Bytes()

	docXML := string(extractDocumentXML(t, savedBytes))
	if !strings.Contains(docXML, "BAR in body") {
		t.Errorf("document.xml: expected replaced body text, got %s", docXML)
	}
	if strings.Contains(docXML, "FOO") {
		t.Errorf("document.xml: unexpected leftover FOO, got %s", docXML)
	}

	headerXML := string(extractZipPart(t, savedBytes, "word/header1.xml"))
	if !strings.Contains(headerXML, "BAR in header") {
		t.Errorf("header1.xml: expected the replacement to reach the saved header, got %s", headerXML)
	}
	if strings.Contains(headerXML, "FOO") {
		t.Errorf("header1.xml: unexpected leftover FOO, got %s", headerXML)
	}

	footerXML := string(extractZipPart(t, savedBytes, "word/footer1.xml"))
	if !strings.Contains(footerXML, "BAR in footer") {
		t.Errorf("footer1.xml: expected the replacement to reach the saved footer, got %s", footerXML)
	}
	if strings.Contains(footerXML, "FOO") {
		t.Errorf("footer1.xml: unexpected leftover FOO, got %s", footerXML)
	}
}
