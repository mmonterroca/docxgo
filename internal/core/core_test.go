// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core_test

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"regexp"
	"strings"
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

// TestParagraph_AddHyperlink_AttachesField pins the issue #101 fix: the run
// AddHyperlink returns must carry a domain.FieldTypeHyperlink field, since
// that field is what makes the serializer emit a real <w:hyperlink r:id>
// wrapper instead of a plain styled run (see expandRunWithFields in
// internal/serializer/serializer.go). Before the fix, this run had zero
// fields.
func TestParagraph_AddHyperlink_AttachesField(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	run, err := para.AddHyperlink("https://example.com/policy", "See the policy")
	if err != nil {
		t.Fatalf("AddHyperlink() error = %v", err)
	}

	fields := run.Fields()
	if len(fields) != 1 {
		t.Fatalf("len(run.Fields()) = %d, want 1", len(fields))
	}
	if fields[0].Type() != domain.FieldTypeHyperlink {
		t.Errorf("Fields()[0].Type() = %v, want %v", fields[0].Type(), domain.FieldTypeHyperlink)
	}

	if got, want := run.Text(), "See the policy"; got != want {
		t.Errorf("run.Text() = %q, want %q", got, want)
	}
}

// TestParagraph_AddHyperlink_DefaultDisplayTextIsURL pins the existing
// "empty displayText falls back to the url" behavior, unchanged by the fix.
func TestParagraph_AddHyperlink_DefaultDisplayTextIsURL(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	run, err := para.AddHyperlink("https://example.com", "")
	if err != nil {
		t.Fatalf("AddHyperlink() error = %v", err)
	}
	if got, want := run.Text(), "https://example.com"; got != want {
		t.Errorf("run.Text() = %q, want %q", got, want)
	}
}

// TestParagraph_AddHyperlink_EmptyURL pins the existing empty-url guard.
func TestParagraph_AddHyperlink_EmptyURL(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	if _, err := para.AddHyperlink("", "text"); err == nil {
		t.Error("AddHyperlink(\"\", ...) error = nil, want an error")
	}
}

// TestParagraph_AddHyperlink_RejectsQuoteInURL is a Compatibility change from
// the issue #101 fix: routing through NewHyperlinkField means a url
// containing a double quote is now rejected (see
// TestNewHyperlinkFieldRejectsQuote), where it previously produced a broken
// link silently.
//
// It also pins a fix to the fix, found in the PR's own review: AddHyperlink
// used to build the field (and validate it) only after already calling
// AddRun/SetText/SetColor/SetUnderline, so a rejected url left a stray blue,
// underlined run behind the returned error. The field is now validated
// before the paragraph is touched at all, so a rejected call must leave the
// paragraph exactly as it found it.
func TestParagraph_AddHyperlink_RejectsQuoteInURL(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	if _, err := para.AddHyperlink(`https://example.com/?q="x"`, "text"); err == nil {
		t.Error("AddHyperlink() error = nil, want an error for a url containing a double quote")
	}

	if got := len(para.Runs()); got != 0 {
		t.Errorf("len(para.Runs()) = %d, want 0: a rejected AddHyperlink call must not leave a run behind", got)
	}
}

// TestParagraph_AddHyperlink_InternalAnchor pins the free win found while
// fixing issue #101: a "#anchor" url now produces a working internal link
// (via w:anchor) instead of a broken external relationship whose target is
// the literal "#anchor" string. See run.AddField's hyperlink branch.
func TestParagraph_AddHyperlink_InternalAnchor(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	run, err := para.AddHyperlink("#Chapter1", "See chapter 1")
	if err != nil {
		t.Fatalf("AddHyperlink() error = %v", err)
	}

	fields := run.Fields()
	if len(fields) != 1 {
		t.Fatalf("len(run.Fields()) = %d, want 1", len(fields))
	}

	accessor, ok := fields[0].(interface {
		GetProperty(string) (string, bool)
	})
	if !ok {
		t.Fatal("field does not support GetProperty")
	}
	if relID, ok := accessor.GetProperty("relationshipID"); ok && relID != "" {
		t.Errorf("GetProperty(relationshipID) = %q, want unset for an internal anchor link", relID)
	}
}

// TestParagraph_AddHyperlink_HeaderFooterUsesPartRels is the inverse of the
// guard that used to live here: a hyperlink on a header/footer paragraph is
// now supported, because the relationship is minted into that part's own
// manager and written to word/_rels/headerN.xml.rels.
//
// The assertion that matters is not merely that AddHyperlink succeeds -- it is
// *where* the relationship lands. A header cannot resolve an r:id declared in
// word/_rels/document.xml.rels, so a hyperlink whose relationship went there
// would still produce a package Word offers to repair, while looking fine to a
// test that only checked the error return.
func TestParagraph_AddHyperlink_HeaderFooterUsesPartRels(t *testing.T) {
	const headerURL = "https://example.com/header"
	const footerURL = "https://example.com/footer"

	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection() error = %v", err)
	}

	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header() error = %v", err)
	}
	headerPara, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("header.AddParagraph() error = %v", err)
	}
	if _, err := headerPara.AddHyperlink(headerURL, "header link"); err != nil {
		t.Fatalf("header paragraph AddHyperlink() error = %v, want nil", err)
	}

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer() error = %v", err)
	}
	footerPara, err := footer.AddParagraph()
	if err != nil {
		t.Fatalf("footer.AddParagraph() error = %v", err)
	}
	if _, err := footerPara.AddHyperlink(footerURL, "footer link"); err != nil {
		t.Fatalf("footer paragraph AddHyperlink() error = %v, want nil", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	saved := buf.Bytes()

	headerRels := zipPartText(t, saved, "word/_rels/header1.xml.rels")
	if !strings.Contains(headerRels, headerURL) {
		t.Errorf("word/_rels/header1.xml.rels is missing the hyperlink target %q:\n%s", headerURL, headerRels)
	}

	footerRels := zipPartText(t, saved, "word/_rels/footer1.xml.rels")
	if !strings.Contains(footerRels, footerURL) {
		t.Errorf("word/_rels/footer1.xml.rels is missing the hyperlink target %q:\n%s", footerURL, footerRels)
	}

	// Each part's rels must hold ONLY its own relationships. Asserting just
	// "the header's link is in the header's rels" would still pass if every
	// part shared one manager and each rels part listed everything -- which
	// is a real mis-scoping, not a harmless one: it puts targets in a part
	// that never references them and makes relationship IDs collide across
	// parts as soon as two parts each mint their own.
	if strings.Contains(headerRels, footerURL) {
		t.Errorf("word/_rels/header1.xml.rels contains the FOOTER's target %q -- the two parts are sharing a relationship manager:\n%s", footerURL, headerRels)
	}
	if strings.Contains(footerRels, headerURL) {
		t.Errorf("word/_rels/footer1.xml.rels contains the HEADER's target %q -- the two parts are sharing a relationship manager:\n%s", headerURL, footerRels)
	}

	// The whole point: neither relationship may leak into the document part.
	docRels := zipPartText(t, saved, "word/_rels/document.xml.rels")
	if strings.Contains(docRels, headerURL) {
		t.Errorf("header hyperlink leaked into word/_rels/document.xml.rels, which header1.xml cannot resolve:\n%s", docRels)
	}
	if strings.Contains(docRels, footerURL) {
		t.Errorf("footer hyperlink leaked into word/_rels/document.xml.rels, which footer1.xml cannot resolve:\n%s", docRels)
	}

	// The r:id the header part references must be one its own rels declares.
	headerXML := zipPartText(t, saved, "word/header1.xml")
	relID := hyperlinkRelIDFrom(t, headerXML)
	if !strings.Contains(headerRels, `Id="`+relID+`"`) {
		t.Errorf("header1.xml references r:id=%q but word/_rels/header1.xml.rels does not declare it:\n%s", relID, headerRels)
	}
}

// TestHeader_ImageProducesValidPart is the regression for the corruption bug
// that shipped unnoticed for as long as headers have existed: an image added
// to a header produced a package Word offers to repair, silently. Nothing
// covered it -- the only header-image test in the repo before this one
// (pkg/template's TestConsolidateRuns_PreservesImageInHeader) never calls
// WriteTo, so it could not see either half of the problem.
//
// Both halves are asserted here because either one alone breaks the file:
//
//  1. <w:hdr> must declare xmlns:wp. Every <w:drawing> wrapper element is
//     wp:-prefixed, so without the declaration the part is an undeclared-prefix
//     error. w:document has always declared it; w:hdr/w:ftr did not.
//  2. The image relationship must live in word/_rels/header1.xml.rels. Minted
//     into word/_rels/document.xml.rels -- as it was -- the header's r:embed
//     resolves to nothing.
//
// The image is placed in a header table cell rather than a plain paragraph on
// purpose: that is the path #109 opened up, and it exercises the longest
// relationship-manager chain.
func TestHeader_ImageProducesValidPart(t *testing.T) {
	imgPath := createTestPNGFile(t)

	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection() error = %v", err)
	}
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header() error = %v", err)
	}

	// Plain header paragraph.
	headerPara, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("header.AddParagraph() error = %v", err)
	}
	if _, err := headerPara.AddImage(imgPath); err != nil {
		t.Fatalf("header paragraph AddImage() error = %v", err)
	}

	// Header table cell -- the path #109 made reachable.
	headerTable, err := header.AddTable(1, 1)
	if err != nil {
		t.Fatalf("header.AddTable() error = %v", err)
	}
	row, err := headerTable.Row(0)
	if err != nil {
		t.Fatalf("Row(0) error = %v", err)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) error = %v", err)
	}
	cellPara, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("cell.AddParagraph() error = %v", err)
	}
	if _, err := cellPara.AddImage(imgPath); err != nil {
		t.Fatalf("header table cell AddImage() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	saved := buf.Bytes()

	headerXML := zipPartText(t, saved, "word/header1.xml")

	// Half 1: the namespace.
	if !strings.Contains(headerXML, `xmlns:wp="`) {
		t.Errorf("word/header1.xml does not declare xmlns:wp, so its wp:-prefixed drawing elements are an undeclared-prefix error:\n%s", headerXML)
	}
	if !strings.Contains(headerXML, "<wp:") {
		t.Fatalf("word/header1.xml contains no wp:-prefixed drawing element, so this test is not exercising the bug:\n%s", headerXML)
	}

	// Half 2: every r:embed the header references must be declared in the
	// header's own rels part, not the document's.
	headerRels := zipPartText(t, saved, "word/_rels/header1.xml.rels")
	embedIDs := embedRelIDsFrom(t, headerXML)
	if len(embedIDs) != 2 {
		t.Fatalf("found %d r:embed references in word/header1.xml, want 2 (paragraph image + table cell image):\n%s", len(embedIDs), headerXML)
	}
	for _, id := range embedIDs {
		if !strings.Contains(headerRels, `Id="`+id+`"`) {
			t.Errorf("header1.xml references r:embed=%q but word/_rels/header1.xml.rels does not declare it:\n%s", id, headerRels)
		}
	}
	// Match the Type attribute specifically -- a bare "/image" also matches
	// each relationship's "media/imageN.png" target and double-counts.
	if got := strings.Count(headerRels, `/relationships/image"`); got != 2 {
		t.Errorf("word/_rels/header1.xml.rels declares %d image relationships, want 2:\n%s", got, headerRels)
	}

	// Per-part relationship numbering: the header's own manager starts at
	// rId1 rather than continuing the document's sequence.
	if !strings.Contains(headerRels, `Id="rId1"`) {
		t.Errorf("word/_rels/header1.xml.rels does not start numbering at rId1, so the part is sharing the document's relationship counter:\n%s", headerRels)
	}

	// And the images must not also be declared document-side.
	docRels := zipPartText(t, saved, "word/_rels/document.xml.rels")
	if strings.Contains(docRels, `/relationships/image"`) {
		t.Errorf("the header's images leaked into word/_rels/document.xml.rels:\n%s", docRels)
	}
}

// embedRelIDsFrom returns every a:blip r:embed value in a part, in order.
func embedRelIDsFrom(t *testing.T, partXML string) []string {
	t.Helper()

	var ids []string
	for _, m := range embedRelIDRE.FindAllStringSubmatch(partXML, -1) {
		ids = append(ids, m[1])
	}
	return ids
}

var embedRelIDRE = regexp.MustCompile(`r:embed="([^"]+)"`)

// createTestPNGFile writes a tiny valid PNG to a temp file and returns its path.
func createTestPNGFile(t *testing.T) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(20 * x), G: uint8(20 * y), B: 200, A: 255})
		}
	}

	file, err := os.CreateTemp(t.TempDir(), "img-*.png")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return file.Name()
}

// hyperlinkRelIDFrom extracts the r:id of the first <w:hyperlink> element in a
// header/footer/document part.
func hyperlinkRelIDFrom(t *testing.T, partXML string) string {
	t.Helper()

	match := hyperlinkRelIDRE.FindStringSubmatch(partXML)
	if match == nil {
		t.Fatalf("no <w:hyperlink r:id=...> element found in part:\n%s", partXML)
	}
	return match[1]
}

var hyperlinkRelIDRE = regexp.MustCompile(`<w:hyperlink[^>]*r:id="([^"]+)"`)

// zipPartText reads a named part out of a saved .docx package as a string,
// failing the test if it is missing. io_test.go has an identical helper, but
// that file is in package core (internal tests) while this one is in
// package core_test, so it is not reachable from here.
func zipPartText(t *testing.T, docBytes []byte, name string) string {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(docBytes), int64(len(docBytes)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}
	for _, f := range zipReader.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("failed to open %s: %v", name, err)
		}
		defer rc.Close()
		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		return string(content)
	}
	t.Fatalf("%s not found in the saved package", name)
	return ""
}

// TestHeaderTableCell_AddHyperlinkUsesPartRels is the table-cell counterpart
// to TestParagraph_AddHyperlink_HeaderFooterUsesPartRels, covering the deeper
// path #109 opened up: a paragraph inside a header/footer table's cell, and
// inside a table nested one level further down.
//
// This is the path most likely to regress. The relationship manager reaches a
// nested cell paragraph through a long chain -- docxHeader -> NewTable ->
// NewTableRow -> NewTableCell -> AddParagraph -> NewRun, and again for the
// nested table -- so any link in it that reaches for the document's manager
// instead of the header's puts the relationship in a part the header cannot
// resolve.
func TestHeaderTableCell_AddHyperlinkUsesPartRels(t *testing.T) {
	const cellURL = "https://example.com/header-cell"
	const nestedURL = "https://example.com/header-nested-cell"
	const footerCellURL = "https://example.com/footer-cell"

	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection() error = %v", err)
	}

	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header() error = %v", err)
	}
	headerTable, err := header.AddTable(1, 1)
	if err != nil {
		t.Fatalf("header.AddTable() error = %v", err)
	}
	headerRow, err := headerTable.Row(0)
	if err != nil {
		t.Fatalf("headerTable.Row(0) error = %v", err)
	}
	headerCell, err := headerRow.Cell(0)
	if err != nil {
		t.Fatalf("headerRow.Cell(0) error = %v", err)
	}
	headerCellPara, err := headerCell.AddParagraph()
	if err != nil {
		t.Fatalf("headerCell.AddParagraph() error = %v", err)
	}
	if _, err := headerCellPara.AddHyperlink(cellURL, "cell link"); err != nil {
		t.Fatalf("header table cell paragraph AddHyperlink() error = %v, want nil", err)
	}

	// One level deeper: a table nested inside the header table's cell.
	nestedTable, err := headerCell.AddTable(1, 1)
	if err != nil {
		t.Fatalf("headerCell.AddTable() error = %v", err)
	}
	nestedRow, err := nestedTable.Row(0)
	if err != nil {
		t.Fatalf("nestedTable.Row(0) error = %v", err)
	}
	nestedCell, err := nestedRow.Cell(0)
	if err != nil {
		t.Fatalf("nestedRow.Cell(0) error = %v", err)
	}
	nestedCellPara, err := nestedCell.AddParagraph()
	if err != nil {
		t.Fatalf("nestedCell.AddParagraph() error = %v", err)
	}
	if _, err := nestedCellPara.AddHyperlink(nestedURL, "nested link"); err != nil {
		t.Fatalf("nested header table cell paragraph AddHyperlink() error = %v, want nil", err)
	}

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer() error = %v", err)
	}
	footerTable, err := footer.AddTable(1, 1)
	if err != nil {
		t.Fatalf("footer.AddTable() error = %v", err)
	}
	footerRow, err := footerTable.Row(0)
	if err != nil {
		t.Fatalf("footerTable.Row(0) error = %v", err)
	}
	footerCell, err := footerRow.Cell(0)
	if err != nil {
		t.Fatalf("footerRow.Cell(0) error = %v", err)
	}
	footerCellPara, err := footerCell.AddParagraph()
	if err != nil {
		t.Fatalf("footerCell.AddParagraph() error = %v", err)
	}
	if _, err := footerCellPara.AddHyperlink(footerCellURL, "footer cell link"); err != nil {
		t.Fatalf("footer table cell paragraph AddHyperlink() error = %v, want nil", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	saved := buf.Bytes()

	headerRels := zipPartText(t, saved, "word/_rels/header1.xml.rels")
	for _, url := range []string{cellURL, nestedURL} {
		if !strings.Contains(headerRels, url) {
			t.Errorf("word/_rels/header1.xml.rels is missing %q:\n%s", url, headerRels)
		}
	}

	footerRels := zipPartText(t, saved, "word/_rels/footer1.xml.rels")
	if !strings.Contains(footerRels, footerCellURL) {
		t.Errorf("word/_rels/footer1.xml.rels is missing %q:\n%s", footerCellURL, footerRels)
	}
	// Cross-contamination: see the same check in
	// TestParagraph_AddHyperlink_HeaderFooterUsesPartRels for why.
	if strings.Contains(headerRels, footerCellURL) {
		t.Errorf("word/_rels/header1.xml.rels contains the FOOTER's target %q -- the parts are sharing a relationship manager:\n%s", footerCellURL, headerRels)
	}
	for _, url := range []string{cellURL, nestedURL} {
		if strings.Contains(footerRels, url) {
			t.Errorf("word/_rels/footer1.xml.rels contains the HEADER's target %q -- the parts are sharing a relationship manager:\n%s", url, footerRels)
		}
	}

	docRels := zipPartText(t, saved, "word/_rels/document.xml.rels")
	for _, url := range []string{cellURL, nestedURL, footerCellURL} {
		if strings.Contains(docRels, url) {
			t.Errorf("%q leaked into word/_rels/document.xml.rels, which the header/footer part cannot resolve:\n%s", url, docRels)
		}
	}
}

// TestHeader_BlocksOrdering pins the Header/Footer Blocks/Paragraphs/Tables
// contract: Blocks preserves insertion order across mixed paragraph/table
// content, Paragraphs/Tables are top-level-only filters over the same
// content, and AddTable validates its bounds the same way Document.AddTable
// does.
func TestHeader_BlocksOrdering(t *testing.T) {
	doc := core.NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection() error = %v", err)
	}
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header() error = %v", err)
	}

	if _, err := header.AddParagraph(); err != nil {
		t.Fatalf("header.AddParagraph() [1] error = %v", err)
	}
	table, err := header.AddTable(1, 1)
	if err != nil {
		t.Fatalf("header.AddTable() error = %v", err)
	}
	if _, err := header.AddParagraph(); err != nil {
		t.Fatalf("header.AddParagraph() [2] error = %v", err)
	}

	blocks := header.Blocks()
	if len(blocks) != 3 {
		t.Fatalf("len(Blocks()) = %d, want 3", len(blocks))
	}
	if blocks[0].Paragraph == nil || blocks[0].Table != nil {
		t.Errorf("Blocks()[0] = %+v, want a paragraph block", blocks[0])
	}
	if blocks[1].Table == nil || blocks[1].Paragraph != nil {
		t.Errorf("Blocks()[1] = %+v, want a table block", blocks[1])
	}
	if blocks[2].Paragraph == nil || blocks[2].Table != nil {
		t.Errorf("Blocks()[2] = %+v, want a paragraph block", blocks[2])
	}

	if got := len(header.Paragraphs()); got != 2 {
		t.Errorf("len(Paragraphs()) = %d, want 2 (top-level only)", got)
	}
	if got := len(header.Tables()); got != 1 {
		t.Errorf("len(Tables()) = %d, want 1", got)
	}

	// A cell paragraph must not leak into the top-level Paragraphs() view.
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("table.Row(0) error = %v", err)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("row.Cell(0) error = %v", err)
	}
	if _, err := cell.AddParagraph(); err != nil {
		t.Fatalf("cell.AddParagraph() error = %v", err)
	}
	if got := len(header.Paragraphs()); got != 2 {
		t.Errorf("len(Paragraphs()) after cell.AddParagraph() = %d, want still 2", got)
	}

	if _, err := header.AddTable(0, 1); err == nil {
		t.Error("header.AddTable(0, 1) error = nil, want an InvalidArgument error")
	}
	if _, err := header.AddTable(1, 64); err == nil {
		t.Error("header.AddTable(1, 64) error = nil, want an InvalidArgument error")
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
