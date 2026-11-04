// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

// blockTexts describes each top-level block: a paragraph's text,
// "<table>", or "<section>".
func blockTexts(doc domain.Document) []string {
	var out []string
	for _, b := range doc.Blocks() {
		switch {
		case b.Paragraph != nil:
			out = append(out, b.Paragraph.Text())
		case b.Table != nil:
			out = append(out, "<table>")
		default:
			out = append(out, "<section>")
		}
	}
	return out
}

// addTextParagraph appends a paragraph with a single run of text.
func addTextParagraph(t *testing.T, doc domain.Document, text string) {
	t.Helper()
	p, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	r, err := p.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err = r.SetText(text); err != nil {
		t.Fatalf("SetText: %v", err)
	}
}

// reopen saves doc and opens the result again.
func reopen(t *testing.T, doc domain.Document) domain.Document {
	t.Helper()
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	opened, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	return opened
}

// openBody opens a document whose document.xml has the given body content.
func openBody(t *testing.T, body string) domain.Document {
	t.Helper()
	documentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>` + body + `</w:body>
</w:document>`

	var base bytes.Buffer
	if _, err := NewDocument().WriteTo(&base); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	doc, err := OpenDocumentFromBytes(rewriteZipEntry(t, base.Bytes(), constants.PathDocument, []byte(documentXML)))
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	return doc
}

func TestRemoveBlock(t *testing.T) {
	doc := NewDocument()
	addTextParagraph(t, doc, "A")
	if _, err := doc.AddTable(1, 1); err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	addTextParagraph(t, doc, "B")
	addTextParagraph(t, doc, "C")

	// Remove from a document opened from a file, so the round-trip writer
	// (not just the in-memory model) is exercised.
	opened := reopen(t, doc)
	for _, index := range []int{2, 1} { // "B", then the table
		if err := opened.RemoveBlock(index); err != nil {
			t.Fatalf("RemoveBlock(%d): %v", index, err)
		}
	}

	want := []string{"A", "C"}
	if got := blockTexts(opened); !slices.Equal(got, want) {
		t.Fatalf("blocks after RemoveBlock = %q, want %q", got, want)
	}
	if n := len(opened.Paragraphs()); n != 2 {
		t.Errorf("len(Paragraphs()) = %d, want 2", n)
	}
	if n := len(opened.Tables()); n != 0 {
		t.Errorf("len(Tables()) = %d, want 0", n)
	}
	if got := blockTexts(reopen(t, opened)); !slices.Equal(got, want) {
		t.Errorf("blocks after save = %q, want %q", got, want)
	}
}

func TestRemoveBlockErrors(t *testing.T) {
	doc := NewDocument()
	addTextParagraph(t, doc, "A")
	if _, err := doc.AddSection(); err != nil {
		t.Fatalf("AddSection: %v", err)
	}

	for _, index := range []int{-1, 2} {
		if err := doc.RemoveBlock(index); err == nil {
			t.Errorf("RemoveBlock(%d) = nil, want out-of-range error", index)
		}
	}
	if err := doc.RemoveBlock(1); err == nil {
		t.Error("RemoveBlock(section break) = nil, want error")
	}
	if got, want := blockTexts(doc), []string{"A", "<section>"}; !slices.Equal(got, want) {
		t.Errorf("blocks after failed removals = %q, want %q", got, want)
	}
}

// A paragraph with text whose pPr carries a mid-document sectPr, as Word
// writes it, can be removed without losing the section break.
func TestRemoveBlockKeepsEmbeddedSectionBreak(t *testing.T) {
	const (
		landscape = `<w:sectPr><w:pgSz w:w="15840" w:h="12240" w:orient="landscape"/></w:sectPr>`
		paraB     = `<w:p w:rsidR="00CD34EF"><w:r><w:t>B</w:t></w:r></w:p>`
	)
	doc := openBody(t, `
<w:p><w:r><w:t>A</w:t></w:r></w:p>
<w:p><w:pPr>`+landscape+`</w:pPr><w:r><w:t>carrier</w:t></w:r></w:p>
`+paraB+`
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
`)
	if got, want := blockTexts(doc), []string{"A", "carrier", "<section>", "B"}; !slices.Equal(got, want) {
		t.Fatalf("blocks = %q, want %q", got, want)
	}
	if err := doc.RemoveBlock(1); err != nil { // "carrier"
		t.Fatalf("RemoveBlock(1): %v", err)
	}

	written := readZipPart(t, doc, constants.PathDocument)
	if strings.Contains(written, "carrier") {
		t.Errorf("carrier text survived the save:\n%s", written)
	}
	for _, kept := range []string{landscape, paraB} {
		if !strings.Contains(written, kept) {
			t.Errorf("%s was not written back as it was:\n%s", kept, written)
		}
	}

	again := reopen(t, doc)
	if got, want := blockTexts(again), []string{"A", "<section>", "B"}; !slices.Equal(got, want) {
		t.Errorf("blocks after save = %q, want %q", got, want)
	}
	sections := again.Sections()
	if len(sections) != 2 {
		t.Fatalf("sections = %d, want 2", len(sections))
	}
	if o := sections[0].Orientation(); o != domain.OrientationLandscape {
		t.Errorf("first section orientation = %v, want landscape", o)
	}
	if o := sections[1].Orientation(); o == domain.OrientationLandscape {
		t.Error("second section orientation = landscape, want portrait")
	}
}

// A bookmark that starts in the removed paragraph and a comment range that
// ends in it lose their unmatched halves on save. A balanced bookmark in a
// kept paragraph and an untouched paragraph are written back as they were.
func TestRemoveBlockDropsUnmatchedRangeHalves(t *testing.T) {
	doc := openBody(t, `
<w:p><w:commentRangeStart w:id="2"/><w:bookmarkStart w:id="3" w:name="Kept"/><w:r><w:t>A</w:t></w:r><w:bookmarkEnd w:id="3"/></w:p>
<w:p><w:bookmarkStart w:id="1" w:name="Spans"/><w:r><w:t>B</w:t></w:r><w:commentRangeEnd w:id="2"/></w:p>
<w:p><w:r><w:t>C</w:t></w:r><w:bookmarkEnd w:id="1"/></w:p>
<w:p w:rsidR="00AB12CD"><w:r><w:t>D</w:t></w:r></w:p>
<w:sectPr/>
`)
	if err := doc.RemoveBlock(1); err != nil { // "B"
		t.Fatalf("RemoveBlock(1): %v", err)
	}

	written := readZipPart(t, doc, constants.PathDocument)
	for _, gone := range []string{`<w:bookmarkEnd w:id="1"/>`, `<w:commentRangeStart w:id="2"/>`} {
		if strings.Contains(written, gone) {
			t.Errorf("unmatched %s survived the save:\n%s", gone, written)
		}
	}
	for _, kept := range []string{
		`<w:bookmarkStart w:id="3" w:name="Kept"/>`,
		`<w:bookmarkEnd w:id="3"/>`,
		`<w:p w:rsidR="00AB12CD"><w:r><w:t>D</w:t></w:r></w:p>`,
	} {
		if !strings.Contains(written, kept) {
			t.Errorf("%s was not written back as it was:\n%s", kept, written)
		}
	}

	if got, want := blockTexts(reopen(t, doc)), []string{"A", "C", "D"}; !slices.Equal(got, want) {
		t.Errorf("blocks after save = %q, want %q", got, want)
	}
}
