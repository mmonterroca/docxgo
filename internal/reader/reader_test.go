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

// TestReconstructSectionBreakWithContent_AgainstHandAuthoredPackage pins a
// known, PRE-EXISTING limitation -- not something this fix introduces or
// widens. Unlike the bare carrier paragraph above (no runs, pPr's only child
// is sectPr), it is also legal OOXML for the paragraph that ends a section
// to carry real content alongside its sectPr (ECMA-376 17.6.17): Word does
// this whenever a section boundary falls on a paragraph that already has
// text, rather than on its own dedicated blank paragraph.
//
// applyRunProperties/populateParagraph hydrates that paragraph's content
// into one domain.Paragraph *and* calls applySectionProperties, which starts
// a new domain.Section as its own block. DocumentSerializer.SerializeBody
// (serializer.go) always synthesizes a fresh, separate <w:p> for every
// SectionBreak block -- it has no way to attach a section's sectPr onto an
// existing content paragraph -- so a single source paragraph carrying both
// content and a section break is split into two on resave: the original
// text, followed by an empty section-break carrier.
//
// Confirmed via the same repro run against master before this PR's changes:
// this exact shape -- an embedded sectPr with an explicit w:type -- already
// doubled the paragraph today, since master's own extractSectionBreakType
// check succeeds whenever w:type is present. This PR's bareSectionBreakSectPr
// check only *removes* the double-count for the no-content case, it does not
// add a new one for this with-content, explicit-w:type case.
//
// This is NOT true of the with-content, *absent*-w:type shape -- the actual
// shape of the real #102 fixture -- which behaves differently on master (no
// tripling, but the section boundary is silently dropped instead); see
// TestReconstructSectionBreakWithContent_NoExplicitType_AgainstHandAuthoredPackage
// below and the CHANGELOG's Known limitations entry for that distinction.
//
// Fixing this properly means teaching the writer to fold a SectionBreak
// block's sectPr into the immediately preceding content paragraph instead
// of always minting a new one -- a writer-wide behavior change (it would
// affect every docxgo-authored document, not just round-tripped ones) that
// belongs in its own PR, not bundled into this reader fix. See CHANGELOG.
func TestReconstructSectionBreakWithContent_AgainstHandAuthoredPackage(t *testing.T) {
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
<w:p><w:pPr><w:sectPr><w:type w:val="nextPage"/><w:pgSz w:w="12240" w:h="15840"/></w:sectPr></w:pPr><w:r><w:t>First section</w:t></w:r></w:p>
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
		t.Fatalf("len(Sections()) = %d, want 2 (the content-bearing sectPr must still start a new section)", got)
	}

	// The content survives, un-duplicated, on its own paragraph.
	paras := doc.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("len(Paragraphs()) = %d, want 2", len(paras))
	}
	if got := paras[0].Runs()[0].Text(); got != "First section" {
		t.Errorf("Paragraphs()[0].Runs()[0].Text() = %q, want %q", got, "First section")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	// Pins the known limitation: 2 source paragraphs resave as 3 -- the
	// original content paragraph (now without its own sectPr, which moved to
	// section level) plus a synthetic empty section-break carrier, plus the
	// second content paragraph. See the doc comment above: not a regression
	// from this PR, not fixed by it either.
	if got := len(paragraphOpenTagRE.FindAllString(written, -1)); got != 3 {
		t.Errorf("resaved document.xml contains %d <w:p> elements, want 3 (known limitation: content+sectPr paragraphs still split in two -- see test doc comment)", got)
	}
	if got := strings.Count(written, "<w:sectPr"); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:sectPr> elements, want 2", got)
	}
}

// TestReconstructSectionBreakWithContent_NoExplicitType_AgainstHandAuthoredPackage
// is the same pinned limitation as the test above, but with the embedded
// sectPr's optional <w:type> omitted entirely -- the exact shape of the real
// #102 fixture (Word commonly emits sectPr without w:type, relying on its
// schema default of "nextPage"). The test above alone left that specific
// shape unverified: extractSectionBreakType defaults to
// domain.SectionBreakTypeNextPage whether w:type is present-and-"nextPage" or
// absent, so the two shapes are handled identically -- confirmed here rather
// than assumed.
func TestReconstructSectionBreakWithContent_NoExplicitType_AgainstHandAuthoredPackage(t *testing.T) {
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
<w:p><w:pPr><w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr></w:pPr><w:r><w:t>First section</w:t></w:r></w:p>
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
		t.Fatalf("len(Sections()) = %d, want 2 (an embedded sectPr with no w:type must still default to nextPage and start a new section)", got)
	}

	paras := doc.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("len(Paragraphs()) = %d, want 2", len(paras))
	}
	if got := paras[0].Runs()[0].Text(); got != "First section" {
		t.Errorf("Paragraphs()[0].Runs()[0].Text() = %q, want %q", got, "First section")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	// Same pinned known limitation as the w:type="nextPage" variant above:
	// still 3 <w:p> elements, not a difference introduced by omitting w:type.
	if got := len(paragraphOpenTagRE.FindAllString(written, -1)); got != 3 {
		t.Errorf("resaved document.xml contains %d <w:p> elements, want 3 (known limitation: content+sectPr paragraphs still split in two -- see test doc comment)", got)
	}
	if got := strings.Count(written, "<w:sectPr"); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:sectPr> elements, want 2", got)
	}
}

// buildHeaderTableDocx assembles a hand-authored package (not through
// docxgo's own writer, for the same reason documented on
// TestReconstructHyperlink_AgainstHandAuthoredPackage) whose header1.xml
// contains a paragraph followed by a table with the given number of
// columns in its single row.
func buildHeaderTableDocx(t *testing.T, cols int) []byte {
	t.Helper()

	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	mainDocumentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:body>
<w:p><w:r><w:t>Body</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId7"/>
<w:pgSz w:w="12240" w:h="15840"/>
</w:sectPr>
</w:body>
</w:document>`

	docRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId7" Type="` + constants.RelTypeHeader + `" Target="header1.xml"/>
</Relationships>`

	var cells strings.Builder
	for i := 0; i < cols; i++ {
		fmt.Fprintf(&cells, "<w:tc><w:p><w:r><w:t>Cell%d</w:t></w:r></w:p></w:tc>", i)
	}

	headerXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:p><w:r><w:t>Before Table</w:t></w:r></w:p>
<w:tbl><w:tr>` + cells.String() + `</w:tr></w:tbl>
</w:hdr>`

	return buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml":          contentTypesXML,
		"_rels/.rels":                  rootRelsXML,
		"word/document.xml":            mainDocumentXML,
		"word/_rels/document.xml.rels": docRelsXML,
		"word/header1.xml":             headerXML,
	})
}

// TestReconstructDocument_HeaderTable pins that a table in a header (not
// reachable at all before PR 2b -- header.Paragraphs() returned 0 for a
// table-only header, see the plan for #101's follow-ups) hydrates into the
// domain model: header.Tables() gets an entry, header.Blocks() preserves
// insertion order, and cell text round-trips.
func TestReconstructDocument_HeaderTable(t *testing.T) {
	docxBytes := buildHeaderTableDocx(t, 2)

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

	sections := doc.Sections()
	if len(sections) == 0 {
		t.Fatal("no sections")
	}
	header, err := sections[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(): %v", err)
	}

	tables := header.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(header.Tables()) = %d, want 1", len(tables))
	}

	blocks := header.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("len(header.Blocks()) = %d, want 2 (paragraph, table)", len(blocks))
	}
	if blocks[0].Paragraph == nil {
		t.Errorf("Blocks()[0] is not a paragraph: %+v", blocks[0])
	}
	if blocks[1].Table == nil {
		t.Errorf("Blocks()[1] is not a table: %+v", blocks[1])
	}

	row, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell0, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	cellParas := cell0.Paragraphs()
	if len(cellParas) != 1 || cellParas[0].Text() != "Cell0" {
		t.Fatalf("cell 0 text = %+v, want [\"Cell0\"]", cellParas)
	}
}

// TestReconstructDocument_HeaderTableOversizedIsTolerated pins the
// best-effort choice documented on hydratePartBlocks: a header table whose
// grid exceeds constants.MaxTableCols must not fail OpenDocument the way an
// oversized body table would -- TestOpenDocument_MalformedHeaderRelsIsTolerated
// already establishes that header/footer parts tolerate malformed content the
// body doesn't, and hydratePartBlocks extends that same tolerance to a table
// it can't represent. The table itself is skipped, not partially added.
func TestReconstructDocument_HeaderTableOversizedIsTolerated(t *testing.T) {
	docxBytes := buildHeaderTableDocx(t, constants.MaxTableCols+1)

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
		t.Fatalf("ReconstructDocument: %v (an oversized header table must not fail the whole document)", err)
	}

	sections := doc.Sections()
	if len(sections) == 0 {
		t.Fatal("no sections")
	}
	header, err := sections[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(): %v", err)
	}

	if got := len(header.Tables()); got != 0 {
		t.Errorf("len(header.Tables()) = %d, want 0 (oversized table skipped)", got)
	}
	// The paragraph before the skipped table must still have hydrated.
	if got := len(header.Paragraphs()); got != 1 {
		t.Errorf("len(header.Paragraphs()) = %d, want 1", got)
	}
}

// TestReconstructDocument_HeaderTableWithBrokenCellIsNotPartiallyAdded pins
// that a header table which fails partway through hydration (a cell with an
// unparseable gridSpan) is skipped entirely -- not left in header.Tables()
// half-populated. AddTable attaches the table to the header before any row
// or cell is hydrated, so a failure on, say, the second row must roll that
// attach back too; otherwise the "best-effort skip" hydratePartBlocks
// documents would actually leave a corrupt, partially-hydrated table
// visible to callers instead of skipping it.
func TestReconstructDocument_HeaderTableWithBrokenCellIsNotPartiallyAdded(t *testing.T) {
	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	mainDocumentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:body>
<w:p><w:r><w:t>Body</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId7"/>
<w:pgSz w:w="12240" w:h="15840"/>
</w:sectPr>
</w:body>
</w:document>`

	docRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + constants.NamespacePackageRels + `">
<Relationship Id="rId7" Type="` + constants.RelTypeHeader + `" Target="header1.xml"/>
</Relationships>`

	// Row 1 hydrates fine. Row 2's second cell has a gridSpan that isn't a
	// number at all -- strconv.Atoi fails inside hydrateTableCell, after
	// AddTable already attached a 2x2 table to the header.
	headerXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="` + constants.NamespaceMain + `" xmlns:r="` + constants.NamespaceRelationships + `">
<w:p><w:r><w:t>Before Table</w:t></w:r></w:p>
<w:tbl>
<w:tr><w:tc><w:p><w:r><w:t>R1C1</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>R1C2</w:t></w:r></w:p></w:tc></w:tr>
<w:tr><w:tc><w:p><w:r><w:t>R2C1</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:gridSpan w:val="broken"/></w:tcPr><w:p><w:r><w:t>R2C2</w:t></w:r></w:p></w:tc></w:tr>
</w:tbl>
</w:hdr>`

	docxBytes := buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml":          contentTypesXML,
		"_rels/.rels":                  rootRelsXML,
		"word/document.xml":            mainDocumentXML,
		"word/_rels/document.xml.rels": docRelsXML,
		"word/header1.xml":             headerXML,
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
		t.Fatalf("ReconstructDocument: %v (a header table with a broken cell must not fail the whole document)", err)
	}

	sections := doc.Sections()
	if len(sections) == 0 {
		t.Fatal("no sections")
	}
	header, err := sections[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(): %v", err)
	}

	if got := len(header.Tables()); got != 0 {
		t.Errorf("len(header.Tables()) = %d, want 0 (the table must be skipped entirely, not left half-hydrated)", got)
	}
	for i, block := range header.Blocks() {
		if block.Table != nil {
			t.Errorf("Blocks()[%d] is a table; want it entirely absent alongside Tables()", i)
		}
	}
	// The paragraph before the skipped table must still have hydrated.
	if got := len(header.Paragraphs()); got != 1 {
		t.Errorf("len(header.Paragraphs()) = %d, want 1", got)
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
	if err := run.SetCaps(true); err != nil {
		t.Fatalf("SetCaps: %v", err)
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
	if !recoveredRun.Caps() {
		t.Fatalf("expected run to be caps")
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

// TestReconstructRunCaps_AgainstHandAuthoredPackage is a regression for one
// of the three losses reported in issue #102: a run's <w:caps> (Word's "All
// Caps" display formatting, distinct from the text actually being typed in
// capitals) was never hydrated at all -- absent from domain.Run entirely --
// so it silently disappeared on the next save. The run's stored text is
// untouched either way; only whether it renders forced-uppercase changes.
func TestReconstructRunCaps_AgainstHandAuthoredPackage(t *testing.T) {
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
<w:p><w:r><w:rPr><w:caps/></w:rPr><w:t>tItLe</w:t></w:r></w:p>
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
	if !run.Caps() {
		t.Error("Runs()[0].Caps() = false, want true")
	}
	// The stored text is untouched by w:caps -- it is a display override,
	// not a transformation of the characters.
	if got := run.Text(); got != "tItLe" {
		t.Errorf("Runs()[0].Text() = %q, want %q (w:caps must not alter stored text)", got, "tItLe")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, "<w:caps") {
		t.Errorf("resaved document.xml lost the run's w:caps:\n%s", written)
	}
	if !strings.Contains(written, "tItLe") {
		t.Errorf("resaved document.xml altered the run's stored text:\n%s", written)
	}
}

// TestReconstructRunCapsExplicitFalse_AgainstHandAuthoredPackage covers the
// other half of w:caps fidelity: a run that explicitly cancels All Caps with
// <w:caps w:val="false"/>, distinct from a run that simply never mentions
// w:caps at all. Both hydrate to Runs()[0].Caps() == false, but only the
// explicit one must round-trip back to an explicit w:val="false" -- omitting
// it entirely would be indistinguishable from "never set" on resave, and if
// a run/paragraph style further up sets All Caps, that omission would let
// the style's All Caps silently apply to text the source explicitly
// exempted from it.
func TestReconstructRunCapsExplicitFalse_AgainstHandAuthoredPackage(t *testing.T) {
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
<w:p><w:r><w:rPr><w:caps w:val="false"/></w:rPr><w:t>not shouted</w:t></w:r></w:p>
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
	if run.Caps() {
		t.Error("Runs()[0].Caps() = true, want false")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `<w:caps w:val="false"`) {
		t.Errorf("resaved document.xml lost the run's explicit w:caps w:val=\"false\" override:\n%s", written)
	}
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

// TestReconstructTableStyle_AgainstHandAuthoredPackage is a regression for
// one of the three losses reported in issue #102: a table's <w:tblStyle>
// reference (e.g. Word's built-in "TableGrid") was never hydrated, so it
// silently disappeared on the next save even though the style definition
// itself survived untouched in styles.xml. TableGrid's own definition
// carries the table's visible borders (<w:tblBorders>) -- there is no
// explicit <w:tblBorders> on the table itself, which is exactly why this
// was reported as "table borders removed" rather than "table style lost".
//
// Uses buildRawZipPackage rather than a docxgo-written-then-read fixture:
// TestReconstructDocumentTable never sets a style, so it can't catch a
// hydration gap that a docxgo round-trip wouldn't exercise either way.
func TestReconstructTableStyle_AgainstHandAuthoredPackage(t *testing.T) {
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
<w:tbl>
<w:tblPr><w:tblStyle w:val="TableGrid"/><w:tblW w:w="0" w:type="auto"/></w:tblPr>
<w:tblGrid><w:gridCol w:w="918"/></w:tblGrid>
<w:tr><w:tc><w:p><w:r><w:t>Cell</w:t></w:r></w:p></w:tc></w:tr>
</w:tbl>
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

	tables := doc.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(Tables()) = %d, want 1", len(tables))
	}
	if got := tables[0].Style().Name; got != "TableGrid" {
		t.Errorf("Tables()[0].Style().Name = %q, want %q", got, "TableGrid")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `<w:tblStyle w:val="TableGrid">`) {
		t.Errorf("resaved document.xml lost the table's tblStyle reference:\n%s", written)
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

// buildTableBodyDocx wraps a <w:tbl> fragment in a minimal, hand-authored
// OOXML package -- raw bytes, not routed through docxgo's writer, so a reader
// gap cannot be masked by a matching writer gap. See buildRawZipPackage.
func buildTableBodyDocx(t *testing.T, tableXML string) []byte {
	t.Helper()

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
` + tableXML + `
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body>
</w:document>`

	return buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   mainDocumentXML,
	})
}

// reconstructTableFromXML reads a hand-authored <w:tbl> back into the domain
// model and returns both the table and the document, so a caller can assert
// on the hydrated state and then resave to check what survives the writer.
func reconstructTableFromXML(t *testing.T, tableXML string) (domain.Table, domain.Document) {
	t.Helper()

	pkg, err := LoadPackageFromBytes(buildTableBodyDocx(t, tableXML))
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

	tables := doc.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(Tables()) = %d, want 1", len(tables))
	}
	return tables[0], doc
}

// TestReconstructTableProperties_AgainstHandAuthoredPackage covers the table,
// row and cell properties that the reader dropped on the floor: before this,
// hydration read <w:tblStyle> and the merge attributes and nothing else, so a
// table whose layout was drawn by hand -- explicit widths, an explicit height,
// hand-drawn borders, cell shading -- came back as an unstyled auto-width grid
// even though the domain model and the serializer had supported all of it for
// releases. Row properties were the starkest case: <w:trPr> was never read at
// all.
func TestReconstructTableProperties_AgainstHandAuthoredPackage(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblPr>
<w:tblStyle w:val="TableGrid"/>
<w:tblW w:w="5000" w:type="dxa"/>
<w:jc w:val="center"/>
<w:tblBorders>
<w:top w:val="single" w:sz="8" w:color="FF0000"/>
<w:left w:val="dashed" w:sz="4" w:color="00FF00"/>
<w:bottom w:val="double" w:sz="12" w:color="0000FF"/>
<w:right w:val="dotted" w:sz="6" w:color="123456"/>
<w:insideH w:val="single" w:sz="2" w:color="654321"/>
<w:insideV w:val="thick" w:sz="18" w:color="ABCDEF"/>
</w:tblBorders>
</w:tblPr>
<w:tblGrid><w:gridCol w:w="918"/><w:gridCol w:w="1111"/></w:tblGrid>
<w:tr>
<w:trPr><w:trHeight w:val="567"/></w:trPr>
<w:tc>
<w:tcPr>
<w:tcW w:w="918" w:type="dxa"/>
<w:tcBorders><w:top w:val="single" w:sz="4" w:color="112233"/></w:tcBorders>
<w:shd w:val="clear" w:color="auto" w:fill="DDEEFF"/>
<w:vAlign w:val="center"/>
</w:tcPr>
<w:p><w:r><w:t>A</w:t></w:r></w:p>
</w:tc>
<w:tc>
<w:tcPr><w:tcW w:w="1111" w:type="dxa"/></w:tcPr>
<w:p><w:r><w:t>B</w:t></w:r></w:p>
</w:tc>
</w:tr>
</w:tbl>`

	table, doc := reconstructTableFromXML(t, tableXML)

	if got := table.Width(); got.Type != domain.WidthDXA || got.Value != 5000 {
		t.Errorf("Width() = %+v, want {WidthDXA 5000}", got)
	}
	if got := table.Alignment(); got != domain.AlignmentCenter {
		t.Errorf("Alignment() = %v, want AlignmentCenter", got)
	}

	borders := table.Borders()
	if borders.Top.Style != domain.BorderSingle || borders.Top.Width != 8 ||
		borders.Top.Color != (domain.Color{R: 0xFF}) {
		t.Errorf("Borders().Top = %+v, want single/8/FF0000", borders.Top)
	}
	if borders.Left.Style != domain.BorderDashed {
		t.Errorf("Borders().Left.Style = %v, want BorderDashed", borders.Left.Style)
	}
	if borders.Bottom.Style != domain.BorderDouble {
		t.Errorf("Borders().Bottom.Style = %v, want BorderDouble", borders.Bottom.Style)
	}
	if borders.Right.Style != domain.BorderDotted {
		t.Errorf("Borders().Right.Style = %v, want BorderDotted", borders.Right.Style)
	}
	if borders.InsideH.Style != domain.BorderSingle {
		t.Errorf("Borders().InsideH.Style = %v, want BorderSingle", borders.InsideH.Style)
	}
	if borders.InsideV.Style != domain.BorderThick || borders.InsideV.Width != 18 {
		t.Errorf("Borders().InsideV = %+v, want thick/18", borders.InsideV)
	}

	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	if got := row.Height(); got != 567 {
		t.Errorf("Row(0).Height() = %d, want 567", got)
	}

	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if got := cell.Width(); got != 918 {
		t.Errorf("Cell(0).Width() = %d, want 918", got)
	}
	if got := cell.VerticalAlignment(); got != domain.VerticalAlignCenter {
		t.Errorf("Cell(0).VerticalAlignment() = %v, want VerticalAlignCenter", got)
	}
	if got := cell.Shading(); got != (domain.Color{R: 0xDD, G: 0xEE, B: 0xFF}) {
		t.Errorf("Cell(0).Shading() = %+v, want DDEEFF", got)
	}
	if got := cell.Borders().Top; got.Style != domain.BorderSingle || got.Width != 4 {
		t.Errorf("Cell(0).Borders().Top = %+v, want single/4", got)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	for _, want := range []string{
		`<w:tblW w:type="dxa" w:w="5000">`,
		`<w:jc w:val="center">`,
		`<w:tblBorders>`,
		`<w:insideV w:val="thick" w:sz="18" w:color="ABCDEF">`,
		`<w:trHeight w:val="567" w:hRule="atLeast">`,
		`<w:tcW w:type="dxa" w:w="918">`,
		`<w:shd w:val="clear" w:fill="DDEEFF">`,
		`<w:vAlign w:val="center">`,
		// The grid columns come from the row's cell widths, so they have to
		// match the source's <w:gridCol> values rather than being omitted.
		`<w:gridCol w:w="918">`,
		`<w:gridCol w:w="1111">`,
	} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document.xml is missing %s:\n%s", want, written)
		}
	}
}

// TestReconstructTableCellWidth_NonDxaIsNotHydrated pins a deliberate
// omission. domain.TableCell.SetWidth takes a bare twip count with no type,
// and the serializer writes w:type="dxa" for any positive width, so hydrating
// a percentage width would rewrite "half the table" as a fixed 2500 twips --
// wrong, not merely lossy. Leaving it alone degrades the cell to auto width,
// which still lays out sensibly.
func TestReconstructTableCellWidth_NonDxaIsNotHydrated(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblPr><w:tblW w:w="0" w:type="auto"/></w:tblPr>
<w:tblGrid><w:gridCol/></w:tblGrid>
<w:tr><w:tc>
<w:tcPr><w:tcW w:w="2500" w:type="pct"/></w:tcPr>
<w:p><w:r><w:t>A</w:t></w:r></w:p>
</w:tc></w:tr>
</w:tbl>`

	table, doc := reconstructTableFromXML(t, tableXML)

	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if got := cell.Width(); got != 0 {
		t.Errorf("Cell(0).Width() = %d, want 0 (a pct width must not be hydrated as twips)", got)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if strings.Contains(written, `<w:tcW w:type="dxa" w:w="2500">`) {
		t.Errorf("resaved document.xml turned a 50%% width into 2500 twips:\n%s", written)
	}
}

// TestReconstructTableCellShading_OnlySolidFillsAreHydrated pins the other
// deliberate omission: domain.TableCell.SetShading holds a single colour, so a
// pattern fill or w:fill="auto" has nothing to map onto, and picking a colour
// would paint a background the source never had.
func TestReconstructTableCellShading_OnlySolidFillsAreHydrated(t *testing.T) {
	cases := []struct {
		name string
		shd  string
	}{
		{"pattern fill", `<w:shd w:val="pct25" w:color="auto" w:fill="DDEEFF"/>`},
		{"auto fill", `<w:shd w:val="clear" w:color="auto" w:fill="auto"/>`},
		{"theme fill only", `<w:shd w:val="clear" w:color="auto" w:themeFill="accent1"/>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tableXML := `<w:tbl>
<w:tblGrid><w:gridCol/></w:tblGrid>
<w:tr><w:tc><w:tcPr>` + tc.shd + `</w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc></w:tr>
</w:tbl>`

			table, _ := reconstructTableFromXML(t, tableXML)
			row, err := table.Row(0)
			if err != nil {
				t.Fatalf("Row(0): %v", err)
			}
			cell, err := row.Cell(0)
			if err != nil {
				t.Fatalf("Cell(0): %v", err)
			}
			if got := cell.Shading(); got != domain.ColorWhite {
				t.Errorf("Cell(0).Shading() = %+v, want the untouched default (%+v)", got, domain.ColorWhite)
			}
		})
	}
}

// TestReconstructTableGrid_MergedFirstRowDoesNotInventColumnWidths covers the
// write side of hydration. serializeGrid derives w:gridCol widths from a row's
// cell widths; once cell widths are actually hydrated, a table whose first row
// is a single merged full-width title cell would have that one width split
// evenly across columns that are not equal, inventing a grid the source never
// described. The second row describes the grid honestly, so it is used
// instead.
func TestReconstructTableGrid_MergedFirstRowDoesNotInventColumnWidths(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblGrid><w:gridCol w:w="900"/><w:gridCol w:w="3100"/></w:tblGrid>
<w:tr><w:tc>
<w:tcPr><w:tcW w:w="4000" w:type="dxa"/><w:gridSpan w:val="2"/></w:tcPr>
<w:p><w:r><w:t>Title</w:t></w:r></w:p>
</w:tc></w:tr>
<w:tr>
<w:tc><w:tcPr><w:tcW w:w="900" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc>
<w:tc><w:tcPr><w:tcW w:w="3100" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>B</w:t></w:r></w:p></w:tc>
</w:tr>
</w:tbl>`

	_, doc := reconstructTableFromXML(t, tableXML)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	if strings.Contains(written, `<w:gridCol w:w="2000">`) {
		t.Errorf("resaved document.xml split the merged title cell into two equal columns:\n%s", written)
	}
	for _, want := range []string{`<w:gridCol w:w="900">`, `<w:gridCol w:w="3100">`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document.xml is missing %s:\n%s", want, written)
		}
	}
}
