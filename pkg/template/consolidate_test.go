// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
)

// createTestPNG writes a minimal PNG to a temp file and returns its path, for
// tests that need a real image to attach via Paragraph.AddImage.
func createTestPNG(t *testing.T) string {
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

// countImageRuns returns how many of the paragraph's runs carry an image.
func countImageRuns(para domain.Paragraph) int {
	n := 0
	for _, r := range para.Runs() {
		if r.Image() != nil {
			n++
		}
	}
	return n
}

// extractDocumentXML reads word/document.xml out of a serialized .docx
// (a zip archive) for tests that need to assert on the raw markup.
func extractDocumentXML(t *testing.T, docxBytes []byte) []byte {
	t.Helper()
	return extractZipPart(t, docxBytes, "word/document.xml")
}

// extractZipPart reads a single named entry out of a serialized .docx (a zip
// archive) for tests that need to assert on the raw markup of a part other
// than word/document.xml (e.g. a header or footer).
func extractZipPart(t *testing.T, docxBytes []byte, partName string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	for _, f := range zr.File {
		if f.Name != partName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", partName, err)
		}
		defer rc.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			t.Fatalf("read %s: %v", partName, err)
		}
		return buf.Bytes()
	}
	t.Fatalf("%s not found in archive", partName)
	return nil
}

func TestConsolidateRuns_EmptyParagraph(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	if err := ConsolidateRuns(para); err != nil { // should not panic
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

func TestConsolidateRuns_DifferentLanguage(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("hello ")

	r2, _ := para.AddRun()
	r2.SetText("bonjour")
	if err := r2.SetLanguage(&domain.Language{Val: "fr"}); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}

	r3, _ := para.AddRun()
	r3.SetText(" world")

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

	// Visually identical formatting (no bold/italic/color/etc on any of the
	// three), but r2's language override must keep it from merging with its
	// neighbors — a merge would either lose the "fr" tag on r2's text or
	// wrongly stamp it onto "hello "/" world" too.
	runs := para.Runs()
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs (no merge across a language boundary), got %d", len(runs))
	}
	if runs[0].Text() != "hello " || runs[0].Language() != nil {
		t.Errorf("run[0]: expected unset language 'hello ', got %q lang=%+v", runs[0].Text(), runs[0].Language())
	}
	lang := runs[1].Language()
	if runs[1].Text() != "bonjour" || lang == nil || lang.Val != "fr" {
		t.Errorf("run[1]: expected fr 'bonjour', got %q lang=%+v", runs[1].Text(), lang)
	}
	if runs[2].Text() != " world" || runs[2].Language() != nil {
		t.Errorf("run[2]: expected unset language ' world', got %q lang=%+v", runs[2].Text(), runs[2].Language())
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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

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

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}
	if err := ConsolidateRuns(para); err != nil { // second call should be a no-op
		t.Fatalf("ConsolidateRuns (second call): %v", err)
	}

	runs := para.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run after double consolidation, got %d", len(runs))
	}
	if runs[0].Text() != "{{name}}" {
		t.Errorf("expected '{{name}}', got %q", runs[0].Text())
	}
}

// ─── Image preservation regression tests ───────────────────────────────────
//
// ConsolidateRuns used to rebuild every run via AddRun + the domain.Run
// setters, which cannot copy a run's image (the interface doesn't expose
// one). A paragraph mixing mergeable text runs with an image run would
// silently lose the image the moment consolidation ran. These tests pin the
// fix: an untouched run (including an image run) is re-attached verbatim,
// never rebuilt.

func TestConsolidateRuns_PreservesImageWhenOtherRunsMerge(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello ")
	r2, _ := para.AddRun()
	r2.SetText("world") // same formatting as r1: these two should merge

	if _, err := para.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	if before := countImageRuns(para); before != 1 {
		t.Fatalf("setup: expected 1 image run before consolidation, got %d", before)
	}

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

	if got := countImageRuns(para); got != 1 {
		t.Errorf("image run lost during consolidation: got %d image runs, want 1", got)
	}
	if got := para.Text(); got != "Hello world" {
		t.Errorf("paragraph text = %q, want merged text preserved", got)
	}
}

func TestMergeTemplate_PreservesImage(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{na")
	r2, _ := para.AddRun()
	r2.SetText("me}}") // split placeholder: consolidation must merge these

	if _, err := para.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	if err := MergeTemplate(doc, MergeData{"name": "World"}); err != nil {
		t.Fatalf("MergeTemplate: %v", err)
	}

	if got := countImageRuns(para); got != 1 {
		t.Errorf("image run lost via MergeTemplate: got %d image runs, want 1", got)
	}
	if got := para.Text(); got != "Hello World" {
		t.Errorf("paragraph text = %q, want %q", got, "Hello World")
	}
}

func TestFindPlaceholders_PreservesImage(t *testing.T) {
	// FindPlaceholders is read-only: it must not mutate the paragraph at all
	// (see #68), so an unrelated image is untouched by construction, not by
	// surviving a consolidation the read-only path used to perform.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{na")
	r2, _ := para.AddRun()
	r2.SetText("me}}")

	if _, err := para.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	runsBefore := len(para.Runs())

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	if got := countImageRuns(para); got != 1 {
		t.Errorf("image run lost via FindPlaceholders: got %d image runs, want 1", got)
	}
	if got := len(para.Runs()); got != runsBefore {
		t.Errorf("FindPlaceholders mutated run count: got %d runs, want %d (unchanged)", got, runsBefore)
	}
}

// TestFindPlaceholders_DoesNotMergeRuns pins that scanning for placeholders
// never rewrites the paragraph's runs, even when it finds one split across
// several of them — the mutation ConsolidateRuns performs is reserved for
// MergeTemplate, which actually needs the merge to write replacement text.
func TestFindPlaceholders_DoesNotMergeRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{")
	r2, _ := para.AddRun()
	r2.SetText("Name")
	r3, _ := para.AddRun()
	r3.SetText("}}!")

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}
	if placeholders[0].Name != "Name" {
		t.Errorf("expected 'Name', got %q", placeholders[0].Name)
	}

	if got := len(para.Runs()); got != 3 {
		t.Fatalf("FindPlaceholders merged runs: got %d runs, want 3 (unchanged)", got)
	}
	if got := para.Runs()[0].Text(); got != "Hello {{" {
		t.Errorf("run 0 text changed: got %q, want %q", got, "Hello {{")
	}
	if got := para.Runs()[1].Text(); got != "Name" {
		t.Errorf("run 1 text changed: got %q, want %q", got, "Name")
	}
	if got := para.Runs()[2].Text(); got != "}}!" {
		t.Errorf("run 2 text changed: got %q, want %q", got, "}}!")
	}
}

// TestFindPlaceholders_LocationAcrossSplitRuns pins the Location reported for
// a placeholder split across runs: RunIndex/StartOffset point at the run and
// offset where the match starts, EndRunIndex/EndOffset at where it ends, both
// relative to the paragraph's own unmodified runs.
func TestFindPlaceholders_LocationAcrossSplitRuns(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{") // 8 chars, "{{" at offset 6-8
	r2, _ := para.AddRun()
	r2.SetText("Name") // 4 chars
	r3, _ := para.AddRun()
	r3.SetText("}}!") // "}}" at offset 0-2

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	loc := placeholders[0].Location
	if loc.RunIndex != 0 || loc.StartOffset != 6 {
		t.Errorf("start = (run %d, offset %d), want (run 0, offset 6)", loc.RunIndex, loc.StartOffset)
	}
	if loc.EndRunIndex != 2 || loc.EndOffset != 2 {
		t.Errorf("end = (run %d, offset %d), want (run 2, offset 2)", loc.EndRunIndex, loc.EndOffset)
	}
}

// TestFindPlaceholders_LocationNotSplitAtRunBoundary pins the off-by-one bug
// found in code review of #75/#68: a match that is NOT split — it starts
// exactly at the boundary between two mergeable runs — must resolve its
// start to the run that actually contains it (the second run), not to the
// end of the run before it. Before the fix, RunIndex/EndRunIndex disagreed
// (falsely signaling a split) and runs[loc.RunIndex].Text()[loc.StartOffset:loc.EndOffset]
// panicked with a slice-bounds error, since StartOffset came from the wrong
// run's length.
func TestFindPlaceholders_LocationNotSplitAtRunBoundary(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello ") // 6 chars, same formatting as r2 -> mergeable
	r2, _ := para.AddRun()
	r2.SetText("{{Name}}") // whole placeholder starts right at the boundary

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	loc := placeholders[0].Location
	if loc.RunIndex != loc.EndRunIndex {
		t.Fatalf("match is not split across runs, but RunIndex=%d != EndRunIndex=%d", loc.RunIndex, loc.EndRunIndex)
	}
	if loc.RunIndex != 1 || loc.StartOffset != 0 {
		t.Errorf("start = (run %d, offset %d), want (run 1, offset 0)", loc.RunIndex, loc.StartOffset)
	}
	if loc.EndOffset != 8 {
		t.Errorf("EndOffset = %d, want 8", loc.EndOffset)
	}

	runs := para.Runs()
	got := runs[loc.RunIndex].Text()[loc.StartOffset:loc.EndOffset]
	if got != "{{Name}}" {
		t.Errorf("slicing runs[RunIndex].Text()[StartOffset:EndOffset] = %q, want %q", got, "{{Name}}")
	}
}

// TestFindPlaceholders_LocationSkipsLeadingEmptyRun confirms a placeholder
// starting right after a zero-length run resolves to the run that actually
// holds the text, not the empty one.
func TestFindPlaceholders_LocationSkipsLeadingEmptyRun(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("") // empty, same formatting as r2 -> mergeable group
	r2, _ := para.AddRun()
	r2.SetText("{{Name}}")

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	loc := placeholders[0].Location
	if loc.RunIndex != 1 || loc.StartOffset != 0 {
		t.Errorf("start = (run %d, offset %d), want (run 1, offset 0) — skipping the empty run", loc.RunIndex, loc.StartOffset)
	}
}

// TestFindPlaceholders_HeaderNotMutated confirms the read-only guarantee also
// holds for header paragraphs, where logos live alongside mergeable text runs.
func TestFindPlaceholders_HeaderNotMutated(t *testing.T) {
	doc := core.NewDocument()
	if _, err := doc.AddSection(); err != nil {
		t.Fatalf("AddSection: %v", err)
	}
	section := doc.Sections()[0]
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	para, _ := header.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{")
	r2, _ := para.AddRun()
	r2.SetText("Name}}")

	placeholders := FindPlaceholders(doc)
	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}
	if got := len(para.Runs()); got != 2 {
		t.Errorf("FindPlaceholders merged header runs: got %d runs, want 2 (unchanged)", got)
	}
}

// TestValidateTemplate_DoesNotMutate confirms ValidateTemplate, which scans
// via FindPlaceholders internally, does not merge runs either.
func TestValidateTemplate_DoesNotMutate(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{")
	r2, _ := para.AddRun()
	r2.SetText("Name}}")

	_ = ValidateTemplate(doc, MergeData{"Name": "World"})

	if got := len(para.Runs()); got != 2 {
		t.Errorf("ValidateTemplate merged runs: got %d runs, want 2 (unchanged)", got)
	}
}

func TestConsolidateRuns_PreservesImageInHeader(t *testing.T) {
	// Logos live in headers. A replace/merge touching only the body must not
	// disturb an image sitting in a header paragraph that happens to have
	// mergeable text runs alongside it.
	doc := core.NewDocument()
	if _, err := doc.AddSection(); err != nil {
		t.Fatalf("AddSection: %v", err)
	}
	section := doc.Sections()[0]
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	hp, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("header AddParagraph: %v", err)
	}

	r1, _ := hp.AddRun()
	r1.SetText("ACME ")
	r2, _ := hp.AddRun()
	r2.SetText("Inc.") // same formatting as r1: should merge

	if _, err := hp.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("header AddImage: %v", err)
	}

	if err := ConsolidateRuns(hp); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

	if got := countImageRuns(hp); got != 1 {
		t.Errorf("header logo lost during consolidation: got %d image runs, want 1", got)
	}
	if got := hp.Text(); got != "ACME Inc." {
		t.Errorf("header text = %q, want %q", got, "ACME Inc.")
	}
}

// decoratedParagraph is the plain embedding decorator a consumer would write
// to add logging, instrumentation, or a test double over a paragraph. It is a
// domain.Paragraph but not internal/core's concrete type — consolidation must
// work on it through the exported interface alone, with no type assertion on
// an unexported implementation.
type decoratedParagraph struct{ domain.Paragraph }

func TestConsolidateRuns_WorksThroughDomainInterfaceOnly(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{na")
	r2, _ := para.AddRun()
	r2.SetText("me}}")

	if err := ConsolidateRuns(decoratedParagraph{para}); err != nil {
		t.Fatalf("ConsolidateRuns on a wrapped domain.Paragraph: %v", err)
	}

	if got := para.Text(); got != "Hello {{name}}" {
		t.Errorf("paragraph text = %q, want %q", got, "Hello {{name}}")
	}
	if got := len(para.Runs()); got != 1 {
		t.Errorf("run count = %d, want 1 (the split placeholder should have merged)", got)
	}
}

func TestMergeTemplate_WorksThroughDomainInterfaceOnly(t *testing.T) {
	// The failure this guards against is silent: a consolidation error inside
	// MergeTemplate aborts substitution, so the output keeps the raw
	// {{placeholder}} instead of the merged value.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Hello {{na")
	r2, _ := para.AddRun()
	r2.SetText("me}}")

	if err := ConsolidateRuns(decoratedParagraph{para}); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}
	if err := MergeTemplate(doc, MergeData{"name": "World"}); err != nil {
		t.Fatalf("MergeTemplate: %v", err)
	}

	if got := para.Text(); got != "Hello World" {
		t.Errorf("paragraph text = %q, want %q", got, "Hello World")
	}
}

func TestConsolidateRuns_ImageSurvivesSerialization(t *testing.T) {
	// End-to-end: the image must still be present in the serialized
	// document.xml (as a <w:drawing>), not just in the in-memory run list.
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	r1, _ := para.AddRun()
	r1.SetText("Vendor: ")
	r2, _ := para.AddRun()
	r2.SetText("ACME Corp")

	if _, err := para.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	if err := ConsolidateRuns(para); err != nil {
		t.Fatalf("ConsolidateRuns: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	xmlBytes := extractDocumentXML(t, buf.Bytes())
	if got := strings.Count(string(xmlBytes), "<w:drawing>"); got != 1 {
		t.Errorf("serialized document.xml has %d <w:drawing> elements, want 1", got)
	}
}
