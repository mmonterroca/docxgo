// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package reader

import (
	"archive/zip"
	"bytes"
	"fmt"
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

// documentXML unzips a .docx buffer and returns the raw word/document.xml
// bytes, so tests can assert on what was actually serialized rather than on
// in-memory accessor values that can diverge from it.
func documentXML(t *testing.T, docxBytes []byte) string {
	t.Helper()
	return zipPartXML(t, docxBytes, "word/document.xml")
}

// zipPartXML unzips a .docx buffer and returns the raw bytes of the named
// part, as a string.
func zipPartXML(t *testing.T, docxBytes []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(data)
	}
	t.Fatalf("%s not found in package", name)
	return ""
}

// paragraphOpenTagRE matches a <w:p> opening tag, with or without attributes
// or self-closing -- but not <w:pPr>, <w:pStyle>, etc, whose 4th character
// isn't a space, '>', or '/'.
var paragraphOpenTagRE = regexp.MustCompile(`<w:p( |>|/)`)

// buildRawZipPackage assembles a .docx archive byte-for-byte from the given
// part contents. Unlike every other fixture in this file, it does not go
// through docxgo's own writer -- see
// TestReconstructHyperlink_AgainstHandAuthoredPackage for why that matters.
func buildRawZipPackage(t *testing.T, parts map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}
	return buf.Bytes()
}

// TestReconstructHyperlink_AgainstHandAuthoredPackage is the answer to issue
// #101's own diagnosis: every other fixture in this test suite is written by
// docxgo and read back by docxgo, so a writer gap and a reader gap that
// happen to cancel out are indistinguishable from correctness (a document
// whose writer never emits <w:hyperlink> round-trips as "a document with no
// links", which looks identical to a document that genuinely has none).
//
// This builds a minimal OOXML package as raw XML bytes -- not through
// core.NewDocument()/WriteTo -- shaped the way real Word writes a hyperlink:
// a <w:hyperlink r:id="..." w:history="1"> wrapping two separate <w:r>
// children (Word splits a hyperlink's display text across runs whenever
// formatting changes mid-link), each styled with rStyle="Hyperlink". It
// exercises hydrateHyperlink and hydrateRun directly against input the
// writer side of this package had no part in producing.
func TestReconstructHyperlink_AgainstHandAuthoredPackage(t *testing.T) {
	const url = "https://example.com/from-word"

	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	docRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId2" Type="` + constants.RelTypeHyperlink + `" Target="` + url + `" TargetMode="External"/>
</Relationships>`

	mainDocumentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:body>
<w:p>
<w:hyperlink r:id="rId2" w:history="1">
<w:r><w:rPr><w:rStyle w:val="Hyperlink"/></w:rPr><w:t xml:space="preserve">Example </w:t></w:r>
<w:r><w:rPr><w:rStyle w:val="Hyperlink"/><w:b/></w:rPr><w:t>Site</w:t></w:r>
</w:hyperlink>
</w:p>
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body>
</w:document>`

	docxBytes := buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml":          contentTypesXML,
		"_rels/.rels":                  rootRelsXML,
		"word/document.xml":            mainDocumentXML,
		"word/_rels/document.xml.rels": docRelsXML,
	})

	pkg, err := LoadPackageFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paras))
	}
	runs := paras[0].Runs()
	if len(runs) != 2 {
		t.Fatalf("len(Runs()) = %d, want 2 (one <w:hyperlink> flattens its children onto the paragraph)", len(runs))
	}

	for i, wantText := range []string{"Example ", "Site"} {
		run := runs[i]
		if got := run.Text(); got != wantText {
			t.Errorf("Runs()[%d].Text() = %q, want %q", i, got, wantText)
		}

		fields := run.Fields()
		if len(fields) != 1 {
			t.Fatalf("len(Runs()[%d].Fields()) = %d, want 1", i, len(fields))
		}
		if fields[0].Type() != domain.FieldTypeHyperlink {
			t.Errorf("Runs()[%d].Fields()[0].Type() = %v, want %v", i, fields[0].Type(), domain.FieldTypeHyperlink)
		}

		accessor, ok := fields[0].(interface {
			GetProperty(string) (string, bool)
		})
		if !ok {
			t.Fatalf("Runs()[%d] field does not support GetProperty", i)
		}
		if gotURL, ok := accessor.GetProperty("url"); !ok || gotURL != url {
			t.Errorf("Runs()[%d] field GetProperty(url) = %q, ok=%v, want %q", i, gotURL, ok, url)
		}
		if gotRelID, ok := accessor.GetProperty("relationshipID"); !ok || gotRelID != "rId2" {
			t.Errorf("Runs()[%d] field GetProperty(relationshipID) = %q, ok=%v, want %q", i, gotRelID, ok, "rId2")
		}
	}

	// Writing the reconstructed document back out must not mint a second
	// relationship for the preserved rId2 (see run.AddField's
	// "already has a relationship ID" branch).
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if got := strings.Count(written, "<w:hyperlink"); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:hyperlink> elements, want 2 (one per run, since reading flattens the source's single element)", got)
	}

	// Regression for a defect found while planning #101's follow-ups: this
	// hydration path (hydrateHyperlink -> hydrateRun -> run.AddField) predates
	// issue #101's fix and was never covered by a rendered-text assertion --
	// only the <w:hyperlink> element count above, which stays correct whether
	// or not the run's text is also duplicated as a trailing plain <w:r>.
	// Reading ANY hyperlink-bearing document (not just ones built through
	// AddHyperlink) and resaving it without modification hit this.
	for _, wantText := range []string{"Example ", "Site"} {
		if got := strings.Count(written, wantText); got != 1 {
			t.Errorf("resaved document.xml contains %q %d times, want 1 (duplicated trailing run)", wantText, got)
		}
	}
}

// TestReconstructHyperlink_AnchorDoesNotMintRelationship pins the other half
// of the issue #101 fix: an internal hyperlink read from a real document
// (w:anchor, no r:id, and no relationship at all in the source's own rels)
// must not cause docxgo to invent an External relationship whose target is
// the literal "#Chapter1" string when the document is written back out. See
// run.AddField's isAnchor branch in internal/core/run.go.
func TestReconstructHyperlink_AnchorDoesNotMintRelationship(t *testing.T) {
	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	// Deliberately no word/_rels/document.xml.rels part: a <w:hyperlink
	// w:anchor> has no relationship of its own in a real Word document.
	mainDocumentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:body>
<w:p>
<w:hyperlink w:anchor="Chapter1" w:history="1">
<w:r><w:rPr><w:rStyle w:val="Hyperlink"/></w:rPr><w:t>See chapter 1</w:t></w:r>
</w:hyperlink>
</w:p>
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body>
</w:document>`

	docxBytes := buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   mainDocumentXML,
	})

	pkg, err := LoadPackageFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 1 || len(paras[0].Runs()) != 1 {
		t.Fatalf("unexpected shape: %d paragraphs", len(paras))
	}
	run := paras[0].Runs()[0]
	fields := run.Fields()
	if len(fields) != 1 || fields[0].Type() != domain.FieldTypeHyperlink {
		t.Fatalf("expected a single hyperlink field, got %+v", fields)
	}
	accessor, ok := fields[0].(interface {
		GetProperty(string) (string, bool)
	})
	if !ok {
		t.Fatal("field does not support GetProperty")
	}
	if relID, ok := accessor.GetProperty("relationshipID"); ok && relID != "" {
		t.Errorf("GetProperty(relationshipID) = %q, want unset for an internal anchor read from a source document", relID)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if got := documentXML(t, buf.Bytes()); !strings.Contains(got, `w:anchor="Chapter1"`) {
		t.Errorf("resaved document.xml lost the anchor: %s", got)
	}

	relsXML := zipPartXML(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if strings.Contains(relsXML, constants.RelTypeHyperlink) {
		t.Errorf("resaved document.xml.rels contains an orphaned hyperlink relationship for an internal anchor:\n%s", relsXML)
	}
}

// TestReconstructMidBodySectionBreak_AgainstHandAuthoredPackage is a
// regression for issue #102: a mid-document section break (a <w:sectPr>
// embedded in a paragraph's own pPr, rather than the body's last child) was
// silently dropped on read, collapsing a two-section document into one.
//
// The paragraph here matches the real, Word-authored shape that exposed
// this (see testdata/word/issue-102-input.docx): no runs, and its pPr's only
// child is sectPr -- Word's usual way of ending a section without any other
// content on that paragraph. Deliberately omits <w:type>: its schema default
// is "nextPage", not "no section break", so the fix must not treat an absent
// w:type as "don't start a new section" -- see applySectionProperties's
// embedded parameter in reconstruct.go.
func TestReconstructMidBodySectionBreak_AgainstHandAuthoredPackage(t *testing.T) {
	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	mainDocumentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:body>
<w:p><w:r><w:t>First section</w:t></w:r></w:p>
<w:p><w:pPr><w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr></w:pPr></w:p>
<w:p><w:r><w:t>Second section</w:t></w:r></w:p>
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body>
</w:document>`

	docxBytes := buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   mainDocumentXML,
	})

	pkg, err := LoadPackageFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	if got := len(doc.Sections()); got != 2 {
		t.Fatalf("len(Sections()) = %d, want 2 (the mid-body sectPr must start a new section)", got)
	}

	// The bare paragraph that carried the sectPr must not survive as its own
	// domain.Paragraph -- it exists in the source purely to hold the break,
	// and the writer already synthesizes a paragraph like it for every
	// section break (see DocumentSerializer.SerializeBody). Keeping both
	// would double it on the next write.
	paras := doc.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("len(Paragraphs()) = %d, want 2 (the bare section-break paragraph is not a separate domain.Paragraph)", len(paras))
	}
	if got := paras[0].Runs()[0].Text(); got != "First section" {
		t.Errorf("Paragraphs()[0].Runs()[0].Text() = %q, want %q", got, "First section")
	}
	if got := paras[1].Runs()[0].Text(); got != "Second section" {
		t.Errorf("Paragraphs()[1].Runs()[0].Text() = %q, want %q", got, "Second section")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	if got := strings.Count(written, "<w:sectPr"); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:sectPr> elements, want 2 (the mid-body break plus the final section)", got)
	}
	if got := len(paragraphOpenTagRE.FindAllString(written, -1)); got != 3 {
		t.Errorf("resaved document.xml contains %d <w:p> elements, want 3 (2 content paragraphs + 1 synthetic section-break carrier -- not 4)", got)
	}
}

func TestLoadPackageFromBytes(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Hello, reader!"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	if len(pkg.MainDocument) == 0 {
		t.Fatalf("expected main document data")
	}
	if pkg.ContentTypes == nil {
		t.Fatalf("expected content types")
	}
	if pkg.RootRelationships == nil {
		t.Fatalf("expected root relationships part")
	}
	if pkg.DocumentRelationships == nil {
		t.Fatalf("expected document relationships part")
	}
	if got := pkg.contentTypeFor("word/document.xml"); got == "" {
		t.Fatalf("expected content type for main document")
	}
	if pkg.PackageSize == 0 {
		t.Fatalf("expected package size to be recorded")
	}
}

func TestParsePackage(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Parse! "); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.SetBold(true); err != nil {
		t.Fatalf("SetBold: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	if parsed.DocumentTree == nil {
		t.Fatalf("expected parsed document tree")
	}
	if parsed.DocumentTree.Name.Local != "document" {
		t.Fatalf("unexpected document local name: %s", parsed.DocumentTree.Name.Local)
	}
	if parsed.DocumentTree.Name.Space != constants.NamespaceMain {
		t.Fatalf("unexpected document namespace: %s", parsed.DocumentTree.Name.Space)
	}
	if parsed.DocumentRelationships == nil {
		t.Fatalf("expected document relationships")
	}
	if parsed.RootRelationships == nil {
		t.Fatalf("expected root relationships")
	}
	if parsed.Package == nil {
		t.Fatalf("expected parsed package to retain raw package reference")
	}
	if parsed.CorePropertiesTree == nil {
		t.Fatalf("expected core properties to be parsed")
	}
	if parsed.AppPropertiesTree == nil {
		t.Fatalf("expected app properties to be parsed")
	}
}

func TestParsePackageNil(t *testing.T) {
	if _, err := ParsePackage(nil); err == nil {
		t.Fatalf("expected error when parsing nil package")
	}
}

func TestReconstructDocumentFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := para.SetSpacingBefore(360); err != nil {
		t.Fatalf("SetSpacingBefore: %v", err)
	}
	if err := para.SetSpacingAfter(120); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}
	if err := para.SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	indent := domain.Indentation{Left: 720, Right: 360, FirstLine: 240}
	if err := para.SetIndent(indent); err != nil {
		t.Fatalf("SetIndent: %v", err)
	}
	if err := para.SetLineSpacing(domain.LineSpacing{Rule: domain.LineSpacingExact, Value: 480}); err != nil {
		t.Fatalf("SetLineSpacing: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Spacing sample"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.SetBold(true); err != nil {
		t.Fatalf("SetBold: %v", err)
	}
	if err := run.SetItalic(true); err != nil {
		t.Fatalf("SetItalic: %v", err)
	}
	if err := run.SetUnderline(domain.UnderlineDouble); err != nil {
		t.Fatalf("SetUnderline: %v", err)
	}
	if err := run.SetStrike(true); err != nil {
		t.Fatalf("SetStrike: %v", err)
	}
	if err := run.SetHighlight(domain.HighlightGreen); err != nil {
		t.Fatalf("SetHighlight: %v", err)
	}
	expectedColor := domain.Color{R: 0x12, G: 0x34, B: 0x56}
	if err := run.SetColor(expectedColor); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	if err := run.SetSize(48); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	if err := run.SetFont(domain.Font{Name: "Times New Roman", EastAsia: "SimSun", CS: "Arial"}); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if err := run.SetLanguage(&domain.Language{Val: "fr", EastAsia: "ja", Bidi: "ar"}); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}

	got := paras[0]
	if got.SpacingBefore() != 360 {
		t.Fatalf("SpacingBefore mismatch: got %d", got.SpacingBefore())
	}
	if got.SpacingAfter() != 120 {
		t.Fatalf("SpacingAfter mismatch: got %d", got.SpacingAfter())
	}
	if got.Alignment() != domain.AlignmentCenter {
		t.Fatalf("Alignment mismatch: got %v", got.Alignment())
	}
	recoveredIndent := got.Indent()
	if recoveredIndent.Left != 720 || recoveredIndent.Right != 360 || recoveredIndent.FirstLine != 240 {
		t.Fatalf("Indent mismatch: %+v", recoveredIndent)
	}
	lineSpacing := got.LineSpacing()
	if lineSpacing.Rule != domain.LineSpacingExact {
		t.Fatalf("unexpected line spacing rule: %v", lineSpacing.Rule)
	}
	if lineSpacing.Value != 480 {
		t.Fatalf("unexpected line spacing value: %d", lineSpacing.Value)
	}
	if got.Text() != "Spacing sample" {
		t.Fatalf("unexpected paragraph text: %q", got.Text())
	}
	runs := got.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	recoveredRun := runs[0]
	if !recoveredRun.Bold() {
		t.Fatalf("expected run to be bold")
	}
	if !recoveredRun.Italic() {
		t.Fatalf("expected run to be italic")
	}
	if !recoveredRun.Strike() {
		t.Fatalf("expected run to be strike")
	}
	if recoveredRun.Underline() != domain.UnderlineDouble {
		t.Fatalf("unexpected underline: %v", recoveredRun.Underline())
	}
	if recoveredRun.Size() != 48 {
		t.Fatalf("unexpected size: %d", recoveredRun.Size())
	}
	if recoveredRun.Color() != expectedColor {
		t.Fatalf("unexpected color: %+v", recoveredRun.Color())
	}
	if recoveredRun.Highlight() != domain.HighlightGreen {
		t.Fatalf("unexpected highlight: %v", recoveredRun.Highlight())
	}
	font := recoveredRun.Font()
	if font.Name != "Times New Roman" {
		t.Fatalf("unexpected font name: %s", font.Name)
	}
	if font.EastAsia != "SimSun" {
		t.Fatalf("unexpected east asia font: %s", font.EastAsia)
	}
	if font.CS != "Arial" {
		t.Fatalf("unexpected complex script font: %s", font.CS)
	}
	recoveredLang := recoveredRun.Language()
	if recoveredLang == nil || recoveredLang.Val != "fr" || recoveredLang.EastAsia != "ja" || recoveredLang.Bidi != "ar" {
		t.Fatalf("unexpected run language: %+v", recoveredLang)
	}
}

func TestReconstructRunContentExtensions(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}

	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Intro"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.AddText("\tTab"); err != nil {
		t.Fatalf("AddText: %v", err)
	}
	if err := run.AddBreak(domain.BreakTypeLine); err != nil {
		t.Fatalf("AddBreak: %v", err)
	}

	fieldRun, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun field: %v", err)
	}
	pageField := core.NewField(domain.FieldTypePageNumber)
	if err := fieldRun.AddField(pageField); err != nil {
		t.Fatalf("AddField: %v", err)
	}
	if err := fieldRun.SetText("1"); err != nil {
		t.Fatalf("SetText field: %v", err)
	}

	hyperlinkRun, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun hyperlink: %v", err)
	}
	hyperlinkField := core.NewField(domain.FieldTypeHyperlink)
	if accessor, ok := hyperlinkField.(interface{ SetProperty(string, string) }); ok {
		accessor.SetProperty("url", "https://example.com")
	}
	if err := hyperlinkField.SetCode(`HYPERLINK "https://example.com"`); err != nil {
		t.Fatalf("SetCode hyperlink: %v", err)
	}
	if err := hyperlinkRun.AddField(hyperlinkField); err != nil {
		t.Fatalf("AddField hyperlink: %v", err)
	}
	if err := hyperlinkRun.SetText("Example"); err != nil {
		t.Fatalf("SetText hyperlink: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs")
	}

	reconstructedRuns := paras[0].Runs()
	if len(reconstructedRuns) < 3 {
		t.Fatalf("expected at least 3 runs, got %d", len(reconstructedRuns))
	}

	firstRun := reconstructedRuns[0]
	if !strings.Contains(firstRun.Text(), "\t") {
		t.Fatalf("expected tab character in first run text: %q", firstRun.Text())
	}
	if breakAccessor, ok := firstRun.(interface{ Breaks() []domain.BreakType }); ok {
		if len(breakAccessor.Breaks()) == 0 {
			t.Fatalf("expected line break in first run")
		}
	}

	var (
		pageFieldFound bool
		hyperFieldURL  string
	)

	for _, candidate := range reconstructedRuns {
		runFields, ok := candidate.(interface{ Fields() []domain.Field })
		if !ok {
			continue
		}

		for _, f := range runFields.Fields() {
			switch f.Type() {
			case domain.FieldTypePageNumber:
				pageFieldFound = true
			case domain.FieldTypeHyperlink:
				if accessor, ok := f.(interface{ GetProperty(string) (string, bool) }); ok {
					if url, ok := accessor.GetProperty("url"); ok {
						hyperFieldURL = url
					}
				}
			}
		}
	}

	if !pageFieldFound {
		t.Fatalf("expected page number field to round-trip")
	}
	if hyperFieldURL != "https://example.com" {
		t.Fatalf("unexpected hyperlink url: %q", hyperFieldURL)
	}
}

func TestReconstructImageRuns(t *testing.T) {
	imgPath := createTestPNG(t)

	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}

	originalImg, err := para.AddImage(imgPath)
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	if err := originalImg.SetDescription("Sample image"); err != nil {
		t.Fatalf("SetDescription: %v", err)
	}

	originalData := originalImg.Data()
	originalTarget := originalImg.Target()
	originalSize := originalImg.Size()

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs")
	}

	var (
		foundImage   domain.Image
		imageRun     domain.Run
		runSummaries []string
	)

	for pi, p := range paras {
		for ri, run := range p.Runs() {
			provider, ok := run.(interface{ Image() domain.Image })
			hasImage := false
			if ok {
				if img := provider.Image(); img != nil {
					hasImage = true
				}
			}
			runSummaries = append(runSummaries, fmt.Sprintf("para[%d] run[%d]: text=%q, image=%t", pi, ri, run.Text(), hasImage))
			if !ok {
				continue
			}
			if img := provider.Image(); img != nil {
				foundImage = img
				imageRun = run
				break
			}
		}
		if foundImage != nil {
			break
		}
	}

	if foundImage == nil {
		t.Fatalf("expected hydrated image run; got runs: %s", strings.Join(runSummaries, "; "))
	}

	if imageRun.Text() != "" {
		t.Fatalf("expected image run text to be empty, got %q", imageRun.Text())
	}

	if foundImage.Description() != "Sample image" {
		t.Fatalf("unexpected image description: %q", foundImage.Description())
	}

	if foundImage.Target() != originalTarget {
		t.Fatalf("expected image target %q, got %q", originalTarget, foundImage.Target())
	}

	if size := foundImage.Size(); size.WidthEMU != originalSize.WidthEMU || size.HeightEMU != originalSize.HeightEMU {
		t.Fatalf("unexpected hydrated image size: %+v", size)
	}

	if gotData := foundImage.Data(); len(gotData) == 0 {
		t.Fatalf("expected image data")
	} else if !bytes.Equal(gotData, originalData) {
		t.Fatalf("hydrated image data mismatch")
	}

	if len(paras[0].Images()) == 0 {
		t.Fatalf("expected paragraph to register hydrated image")
	}

	var roundTrip bytes.Buffer
	if _, err := reconstructed.WriteTo(&roundTrip); err != nil {
		t.Fatalf("WriteTo after hydration: %v", err)
	}
}

func TestReconstructParagraphNumbering(t *testing.T) {
	numberingXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0">
      <w:start w:val="1"/>
      <w:numFmt w:val="bullet"/>
      <w:lvlText w:val="•"/>
      <w:lvlJc w:val="left"/>
    </w:lvl>
  </w:abstractNum>
  <w:num w:numId="1">
    <w:abstractNumId w:val="1"/>
  </w:num>
</w:numbering>`) // minimal numbering definition

	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := para.SetNumbering(domain.NumberingReference{ID: 1, Level: 0}); err != nil {
		t.Fatalf("SetNumbering: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("List item 1"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	config, ok := doc.(interface {
		SetNumberingPart([]byte, string)
		RegisterExistingRelationship(string, string, string, string) error
	})
	if !ok {
		t.Fatalf("document does not expose numbering configuration hooks")
	}
	config.SetNumberingPart(numberingXML, "numbering.xml")
	if err := config.RegisterExistingRelationship("rId50", constants.RelTypeNumbering, "numbering.xml", "Internal"); err != nil {
		t.Fatalf("registerExistingRelationship: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	if len(pkg.Numbering) == 0 {
		t.Fatalf("expected numbering part to be written")
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}

	ref, ok := paras[0].Numbering()
	if !ok {
		t.Fatalf("expected numbering reference to be hydrated")
	}
	if ref.ID != 1 || ref.Level != 0 {
		t.Fatalf("unexpected numbering ref: %+v", ref)
	}

	if accessor, ok := reconstructed.(interface{ NumberingPartInfo() ([]byte, string) }); ok {
		data, target := accessor.NumberingPartInfo()
		if target != "numbering.xml" {
			t.Fatalf("unexpected numbering target: %q", target)
		}
		if !bytes.Equal(data, numberingXML) {
			t.Fatalf("numbering part mismatch")
		}
	}

	var roundTrip bytes.Buffer
	if _, err := reconstructed.WriteTo(&roundTrip); err != nil {
		t.Fatalf("WriteTo (roundtrip): %v", err)
	}

	roundPkg, err := LoadPackageFromBytes(roundTrip.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes roundtrip: %v", err)
	}
	if len(roundPkg.Numbering) == 0 {
		t.Fatalf("expected numbering part to persist after roundtrip")
	}
	if !bytes.Equal(roundPkg.Numbering, numberingXML) {
		t.Fatalf("roundtrip numbering mismatch")
	}
}

func TestReconstructDocumentSections(t *testing.T) {
	doc := core.NewDocument()

	defaultSection, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}

	landscapeSize := domain.PageSize{Width: domain.PageSizeA4.Height, Height: domain.PageSizeA4.Width}
	if err := defaultSection.SetPageSize(landscapeSize); err != nil {
		t.Fatalf("SetPageSize default: %v", err)
	}
	if err := defaultSection.SetOrientation(domain.OrientationLandscape); err != nil {
		t.Fatalf("SetOrientation default: %v", err)
	}
	marginsDefault := domain.Margins{Top: 720, Right: 900, Bottom: 720, Left: 1440, Header: 480, Footer: 600}
	if err := defaultSection.SetMargins(marginsDefault); err != nil {
		t.Fatalf("SetMargins default: %v", err)
	}
	if err := defaultSection.SetColumns(2); err != nil {
		t.Fatalf("SetColumns default: %v", err)
	}

	headDefault, err := defaultSection.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header default: %v", err)
	}
	headPara, err := headDefault.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph header default: %v", err)
	}
	headRun, err := headPara.AddRun()
	if err != nil {
		t.Fatalf("AddRun header default: %v", err)
	}
	if err := headRun.SetText("Section 1 Header"); err != nil {
		t.Fatalf("SetText header default: %v", err)
	}

	bodyPara1, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph body default: %v", err)
	}
	bodyRun1, err := bodyPara1.AddRun()
	if err != nil {
		t.Fatalf("AddRun body default: %v", err)
	}
	if err := bodyRun1.SetText("Default section content"); err != nil {
		t.Fatalf("SetText body default: %v", err)
	}

	secondSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeEvenPage)
	if err != nil {
		t.Fatalf("AddSectionWithBreak: %v", err)
	}
	if err := secondSection.SetPageSize(domain.PageSizeLetter); err != nil {
		t.Fatalf("SetPageSize second: %v", err)
	}
	if err := secondSection.SetOrientation(domain.OrientationPortrait); err != nil {
		t.Fatalf("SetOrientation second: %v", err)
	}
	marginsSecond := domain.Margins{Top: 1440, Right: 1440, Bottom: 1440, Left: 1440, Header: 720, Footer: 960}
	if err := secondSection.SetMargins(marginsSecond); err != nil {
		t.Fatalf("SetMargins second: %v", err)
	}
	if err := secondSection.SetColumns(3); err != nil {
		t.Fatalf("SetColumns second: %v", err)
	}

	footDefault, err := secondSection.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer second: %v", err)
	}
	footPara, err := footDefault.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph footer: %v", err)
	}
	footRun, err := footPara.AddRun()
	if err != nil {
		t.Fatalf("AddRun footer: %v", err)
	}
	if err := footRun.SetText("Section 2 Footer"); err != nil {
		t.Fatalf("SetText footer: %v", err)
	}

	bodyPara2, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph body second: %v", err)
	}
	bodyRun2, err := bodyPara2.AddRun()
	if err != nil {
		t.Fatalf("AddRun body second: %v", err)
	}
	if err := bodyRun2.SetText("Second section content"); err != nil {
		t.Fatalf("SetText body second: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	sections := reconstructed.Sections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	rehydratedDefault := sections[0]
	if rehydratedDefault.Orientation() != domain.OrientationLandscape {
		t.Fatalf("expected default section landscape orientation, got %v", rehydratedDefault.Orientation())
	}
	if size := rehydratedDefault.PageSize(); size.Width <= size.Height {
		t.Fatalf("expected width > height for landscape page size, got %+v", size)
	}
	if cols := rehydratedDefault.Columns(); cols != 2 {
		t.Fatalf("expected default section columns=2, got %d", cols)
	}
	if gotMargins := rehydratedDefault.Margins(); gotMargins.Left != marginsDefault.Left || gotMargins.Right != marginsDefault.Right || gotMargins.Header != marginsDefault.Header {
		t.Fatalf("unexpected default section margins: %+v", gotMargins)
	}

	defaultHeader, err := rehydratedDefault.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(default) after hydration: %v", err)
	}
	defaultHeaderParas := defaultHeader.Paragraphs()
	if len(defaultHeaderParas) == 0 {
		t.Fatalf("expected default section header paragraphs")
	}
	if text := defaultHeaderParas[0].Text(); text != "Section 1 Header" {
		t.Fatalf("unexpected default header text: %q", text)
	}

	rehydratedSecond := sections[1]
	if rehydratedSecond.Orientation() != domain.OrientationPortrait {
		t.Fatalf("expected second section portrait orientation, got %v", rehydratedSecond.Orientation())
	}
	if cols := rehydratedSecond.Columns(); cols != 3 {
		t.Fatalf("expected second section columns=3, got %d", cols)
	}
	if gotMargins := rehydratedSecond.Margins(); gotMargins.Footer != marginsSecond.Footer {
		t.Fatalf("unexpected second section footer margin: %+v", gotMargins)
	}

	secondFooter, err := rehydratedSecond.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer(default) after hydration: %v", err)
	}
	footerParas := secondFooter.Paragraphs()
	if len(footerParas) == 0 {
		t.Fatalf("expected second section footer paragraphs")
	}
	if text := footerParas[0].Text(); text != "Section 2 Footer" {
		t.Fatalf("unexpected second footer text: %q", text)
	}

	foundBreak := false
	for _, block := range reconstructed.Blocks() {
		if block.SectionBreak == nil {
			continue
		}
		if block.SectionBreak.Type == domain.SectionBreakTypeEvenPage {
			foundBreak = true
			break
		}
	}
	if !foundBreak {
		t.Fatalf("expected section break block to be preserved")
	}
}

func createTestPNG(t *testing.T) string {
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

func TestReconstructDocumentTable(t *testing.T) {
	doc := core.NewDocument()
	table, err := doc.AddTable(1, 2)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}

	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row: %v", err)
	}

	for idx, text := range []string{"Cell 1", "Cell 2"} {
		cell, err := row.Cell(idx)
		if err != nil {
			t.Fatalf("Cell(%d): %v", idx, err)
		}
		para, err := cell.AddParagraph()
		if err != nil {
			t.Fatalf("AddParagraph cell %d: %v", idx, err)
		}
		run, err := para.AddRun()
		if err != nil {
			t.Fatalf("AddRun cell %d: %v", idx, err)
		}
		if err := run.SetText(text); err != nil {
			t.Fatalf("SetText cell %d: %v", idx, err)
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	tables := reconstructed.Tables()
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}

	r, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}

	for idx, expected := range []string{"Cell 1", "Cell 2"} {
		cell, err := r.Cell(idx)
		if err != nil {
			t.Fatalf("Cell(%d): %v", idx, err)
		}
		paras := cell.Paragraphs()
		if len(paras) == 0 {
			t.Fatalf("expected paragraphs in cell %d", idx)
		}
		if got := paras[0].Text(); got != expected {
			t.Fatalf("unexpected text in cell %d: %q", idx, got)
		}
	}
}

// indElementXML returns the self-closing <w:ind .../> tag from a
// word/document.xml string. Page margins (<w:pgMar>) also carry
// left/right/header/footer attributes, so callers must not grep the whole
// document for these attribute names — only this element's own text.
func indElementXML(t *testing.T, docXML string) string {
	t.Helper()
	idx := strings.Index(docXML, "<w:ind ")
	if idx == -1 {
		t.Fatalf("no <w:ind> element found in document.xml: %s", docXML)
	}
	end := strings.Index(docXML[idx:], ">")
	if end == -1 {
		t.Fatalf("malformed <w:ind> element in document.xml")
	}
	return docXML[idx : idx+end+1]
}

// TestApplyParagraphIndentation_ExplicitZeroSideSurvivesRoundTrip pins the
// defect #76 describes: a source <w:ind> that explicitly names a side (even
// with value 0, e.g. to override an inherited style) must come back out with
// that side explicit, not silently dropped because 0 looks the same as
// "never set". Before the per-side setters, applyParagraphIndentation read
// the attribute into a merged domain.Indentation and called SetIndent once,
// which cannot distinguish "right was 0 in the source" from "right was never
// mentioned" — both look like the zero value and only the explicit-set path
// added here tells them apart on serialization.
func TestApplyParagraphIndentation_ExplicitZeroSideSurvivesRoundTrip(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("indent sample"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	props, err := parseXMLTree([]byte(
		`<w:pPr xmlns:w="` + constants.NamespaceMain + `"><w:ind w:left="720" w:right="0"/></w:pPr>`,
	))
	if err != nil {
		t.Fatalf("parseXMLTree: %v", err)
	}
	if err := applyParagraphIndentation(para, props); err != nil {
		t.Fatalf("applyParagraphIndentation: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	indXML := indElementXML(t, documentXML(t, buf.Bytes()))
	if !strings.Contains(indXML, `w:left="720"`) {
		t.Fatalf("expected explicit left=720, got %s", indXML)
	}
	if !strings.Contains(indXML, `w:right="0"`) {
		t.Fatalf("expected explicit right=0 to survive round-trip, got %s", indXML)
	}
}

// TestApplyParagraphIndentation_OmittedSidesStayOmitted is the mirror guard:
// a source <w:ind> that names only "left" must not cause the other three
// sides to be written as explicit zero, which would clobber a style's own
// non-zero indentation on a side this element never touched.
func TestApplyParagraphIndentation_OmittedSidesStayOmitted(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("indent sample"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	props, err := parseXMLTree([]byte(
		`<w:pPr xmlns:w="` + constants.NamespaceMain + `"><w:ind w:left="720"/></w:pPr>`,
	))
	if err != nil {
		t.Fatalf("parseXMLTree: %v", err)
	}
	if err := applyParagraphIndentation(para, props); err != nil {
		t.Fatalf("applyParagraphIndentation: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	indXML := indElementXML(t, documentXML(t, buf.Bytes()))
	if !strings.Contains(indXML, `w:left="720"`) {
		t.Fatalf("expected explicit left=720, got %s", indXML)
	}
	for _, attr := range []string{"w:right=", "w:firstLine=", "w:hanging="} {
		if strings.Contains(indXML, attr) {
			t.Fatalf("expected %s to stay omitted, got %s", attr, indXML)
		}
	}
}

// TestApplyParagraphIndentation_AllowsFirstLineAndHangingBothPresent pins a
// deliberate, previously-untested behavior change: SetIndent rejects a
// struct with both FirstLine and Hanging set (they are mutually exclusive in
// Word's model), so a source <w:ind> naming both used to fail the whole
// document load. The per-side setters apply independently and carry no such
// cross-field check, so a document reader is now more tolerant of documents
// where both attributes are present in the source XML.
func TestApplyParagraphIndentation_AllowsFirstLineAndHangingBothPresent(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("indent sample"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	props, err := parseXMLTree([]byte(
		`<w:pPr xmlns:w="` + constants.NamespaceMain + `"><w:ind w:firstLine="360" w:hanging="720"/></w:pPr>`,
	))
	if err != nil {
		t.Fatalf("parseXMLTree: %v", err)
	}
	if err := applyParagraphIndentation(para, props); err != nil {
		t.Fatalf("applyParagraphIndentation: %v, expected both attributes to apply independently", err)
	}

	indent := para.Indent()
	if indent.FirstLine != 360 {
		t.Fatalf("expected FirstLine=360, got %d", indent.FirstLine)
	}
	if indent.Hanging != 720 {
		t.Fatalf("expected Hanging=720, got %d", indent.Hanging)
	}
}
