// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func documentXML(t *testing.T, doc domain.Document) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open document.xml: %v", err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read document.xml: %v", err)
		}
		return string(data)
	}
	t.Fatal("word/document.xml not found")
	return ""
}

// TestParagraphShading_SchemaOrder pins where w:shd and w:keepNext/
// w:keepLines land inside w:pPr. CT_PPrBase is a sequence, so Word rejects
// the file ("unreadable content") when they are out of order: keepNext and
// keepLines right after pStyle, shd right after pBdr and before spacing.
func TestParagraphShading_SchemaOrder(t *testing.T) {
	doc := NewDocument()
	p, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	_ = p.SetStyle(domain.StyleIDNormal)
	_ = p.SetBorderLeft(domain.BorderStyle{Style: domain.BorderSingle, Width: 4, Color: domain.Color{R: 0x33, G: 0x33, B: 0x33}})
	if err := p.SetShading(domain.Color{R: 0x1E, G: 0x1E, B: 0x1E}); err != nil {
		t.Fatalf("SetShading: %v", err)
	}
	_ = p.SetKeepNext(true)
	_ = p.SetKeepLines(true)
	_ = p.SetSpacingAfter(0)
	_ = p.SetIndent(domain.Indentation{Left: 360})
	r, _ := p.AddRun()
	_ = r.SetText("code")

	xml := documentXML(t, doc)
	pPr := regexp.MustCompile(`(?s)<w:pPr>.*?</w:pPr>`).FindString(xml)
	if pPr == "" {
		t.Fatalf("no w:pPr in document.xml:\n%s", xml)
	}
	if !strings.Contains(pPr, `<w:shd w:val="clear" w:color="auto" w:fill="1E1E1E"></w:shd>`) {
		t.Errorf("w:shd missing or malformed in %s", pPr)
	}
	order := []string{"<w:pStyle", "<w:keepNext", "<w:keepLines", "<w:pBdr", "<w:shd", "<w:spacing", "<w:ind"}
	last := -1
	for _, tag := range order {
		idx := strings.Index(pPr, tag)
		if idx < 0 {
			t.Fatalf("%s missing from %s", tag, pPr)
		}
		if idx < last {
			t.Errorf("%s is out of CT_PPrBase order in %s", tag, pPr)
		}
		last = idx
	}
}

// TestParagraphShading_BlackIsNotUnset guards the reason for the bool in
// Shading(): Color's zero value is black, so an explicit black fill must
// still be written, and an untouched paragraph must not get one.
func TestParagraphShading_BlackIsNotUnset(t *testing.T) {
	doc := NewDocument()
	plain, _ := doc.AddParagraph()
	if _, ok := plain.Shading(); ok {
		t.Fatal("new paragraph reports shading set")
	}
	black, _ := doc.AddParagraph()
	_ = black.SetShading(domain.Color{})

	xml := documentXML(t, doc)
	if got := strings.Count(xml, "<w:shd"); got != 1 {
		t.Fatalf("want exactly one w:shd (the black one), got %d:\n%s", got, xml)
	}
	if !strings.Contains(xml, `w:fill="000000"`) {
		t.Errorf("explicit black fill not written:\n%s", xml)
	}

	black.ClearShading()
	if _, ok := black.Shading(); ok {
		t.Error("ClearShading left shading set")
	}
	if strings.Contains(documentXML(t, doc), "<w:shd") {
		t.Error("w:shd still written after ClearShading")
	}
}

// TestParagraphKeepNext_ExplicitFalse checks that SetKeepNext(false) is
// written as an explicit off (so it can override a style), while a
// paragraph that never called it writes nothing.
func TestParagraphKeepNext_ExplicitFalse(t *testing.T) {
	doc := NewDocument()
	p, _ := doc.AddParagraph()
	_ = p.SetKeepNext(false)
	_, _ = doc.AddParagraph()

	xml := documentXML(t, doc)
	if got := strings.Count(xml, "<w:keepNext"); got != 1 {
		t.Fatalf("want one w:keepNext, got %d:\n%s", got, xml)
	}
	if !strings.Contains(xml, `<w:keepNext w:val="false">`) {
		t.Errorf("explicit false not written as w:val=\"false\":\n%s", xml)
	}
}

// TestParagraphShading_RoundTrip writes, reopens and rewrites a document:
// shading and keep flags must be hydrated by the reader and written again,
// once each.
func TestParagraphShading_RoundTrip(t *testing.T) {
	doc := NewDocument()
	p, _ := doc.AddParagraph()
	_ = p.SetShading(domain.Color{R: 0xF5, G: 0xF5, B: 0xF5})
	_ = p.SetKeepNext(true)
	_ = p.SetKeepLines(false)
	r, _ := p.AddRun()
	_ = r.SetText("shaded")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	var got domain.Paragraph
	for _, para := range doc2.Paragraphs() {
		if para.Text() == "shaded" {
			got = para
		}
	}
	if got == nil {
		t.Fatal("shaded paragraph not found after reopen")
	}
	fill, ok := got.Shading()
	if !ok || fill != (domain.Color{R: 0xF5, G: 0xF5, B: 0xF5}) {
		t.Errorf("Shading() after reopen = %v, %v; want F5F5F5, true", fill, ok)
	}
	if !got.KeepNext() {
		t.Error("KeepNext lost on reopen")
	}
	if got.KeepLines() {
		t.Error("KeepLines should stay false")
	}

	xml := documentXML(t, doc2)
	if n := strings.Count(xml, `w:fill="F5F5F5"`); n != 1 {
		t.Errorf("want one F5F5F5 fill after rewrite, got %d:\n%s", n, xml)
	}
	if !strings.Contains(xml, `<w:keepLines w:val="false">`) {
		t.Errorf("explicit keepLines=false lost on rewrite:\n%s", xml)
	}

	// Editing the shading of an opened paragraph must count as an edit for
	// the lossless round-trip merge, not be masked by the source bytes.
	_ = got.SetShading(domain.Color{R: 0x1E, G: 0x1E, B: 0x1E})
	xml = documentXML(t, doc2)
	if !strings.Contains(xml, `w:fill="1E1E1E"`) || strings.Contains(xml, `w:fill="F5F5F5"`) {
		t.Errorf("edited shading not written after reopen:\n%s", xml)
	}
}

// TestParagraphShading_Builder covers the fluent builder entry points.
func TestParagraphShading_Builder(t *testing.T) {
	builder := NewDocumentBuilder()
	builder.AddParagraph().Text("x").Shading(domain.Color{R: 1, G: 2, B: 3}).KeepNext().End()
	doc, err := builder.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	xml := documentXML(t, doc)
	if !strings.Contains(xml, `w:fill="010203"`) || !strings.Contains(xml, "<w:keepNext></w:keepNext>") {
		t.Errorf("builder did not write shading/keepNext:\n%s", xml)
	}
}
