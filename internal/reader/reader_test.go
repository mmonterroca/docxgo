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
	if got := strings.Count(written, "<w:hyperlink"); got != 1 {
		t.Errorf("resaved document.xml contains %d <w:hyperlink> elements, want 1 (reading flattens the source's single element into runs, and serializing merges consecutive same-target hyperlink runs back into one element)", got)
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

	if !strings.Contains(written, `w:history="1"`) {
		t.Errorf("resaved document.xml lost w:history=\"1\" from the source:\n%s", written)
	}
}

// TestReconstructHyperlink_ExplicitHistoryFalseRoundTrips pins the reason
// xml.Hyperlink.History is a *string rather than a bare string: w:history is
// ST_OnOff, so an explicit "0" is a real value distinct from the attribute
// being absent, and a bare string with omitempty could not tell them apart.
func TestReconstructHyperlink_ExplicitHistoryFalseRoundTrips(t *testing.T) {
	const url = "https://example.com/no-history"

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
<w:hyperlink r:id="rId2" w:history="0">
<w:r><w:t>No history</w:t></w:r>
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

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	if !strings.Contains(written, `w:history="0"`) {
		t.Errorf("resaved document.xml did not preserve explicit w:history=\"0\":\n%s", written)
	}
	if strings.Contains(written, `w:history="1"`) {
		t.Errorf("resaved document.xml turned an explicit w:history=\"0\" into \"1\":\n%s", written)
	}
}

// TestReconstructHyperlink_HydratedAnchorWithNoHistoryDoesNotInventOne covers
// an internal (w:anchor) hyperlink read from a source that never wrote
// w:history at all -- a real, common shape, not the explicit-"0" case above.
// The two hyperlink-emission branches in the serializer disagreed here: the
// external (r:id) branch already left History nil when the property was
// never set, but the internal (w:anchor) branch defaulted to "1"
// unconditionally, unable to tell "brand-new link, never touched" (which
// does default to "1", set at construction -- see
// TestAddHyperlink_InternalAnchorDefaultsToHistoryTrue in the top-level
// package) apart from "hydrated from a source that omitted the attribute"
// (which must round-trip as omitted).
func TestReconstructHyperlink_HydratedAnchorWithNoHistoryDoesNotInventOne(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p>
<w:hyperlink w:anchor="Chapter1">
<w:r><w:t>Chapter 1</w:t></w:r>
</w:hyperlink>
</w:p>`)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if strings.Contains(written, "w:history") {
		t.Errorf("resaved document.xml invented a w:history attribute the source never had:\n%s", written)
	}
}

// buildBookmarkPackage assembles a minimal single-part .docx around the
// given word/document.xml body, for the bookmark hydration tests below.
func buildBookmarkPackage(t *testing.T, bodyXML string) []byte {
	return buildBookmarkPackageWithPrefix(t, "w", bodyXML)
}

func buildBookmarkPackageWithPrefix(t *testing.T, prefix, bodyXML string) []byte {
	return buildBookmarkPackageWithPrefixAndBodyAttrs(t, prefix, "", bodyXML)
}

func buildBookmarkPackageWithPrefixAndBodyAttrs(t *testing.T, prefix, bodyAttrs, bodyXML string) []byte {
	return buildBookmarkPackageWithBindings(t, prefix, "", bodyAttrs, bodyXML)
}

func buildBookmarkPackageWithBindings(t *testing.T, prefix, rootAttrs, bodyAttrs, bodyXML string) []byte {
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
<` + prefix + `:document xmlns:` + prefix + `="` + constants.NamespaceMain + `"` + rootAttrs + `>
<` + prefix + `:body` + bodyAttrs + `>
` + bodyXML + `
<` + prefix + `:sectPr><` + prefix + `:pgSz ` + prefix + `:w="12240" ` + prefix + `:h="15840"/></` + prefix + `:sectPr>
</` + prefix + `:body>
</` + prefix + `:document>`

	return buildRawZipPackage(t, map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   mainDocumentXML,
	})
}

func reconstructFromBookmarkBody(t *testing.T, bodyXML string) domain.Document {
	t.Helper()
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, bodyXML))
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
	return doc
}

func TestDocumentNamespacesRetainsOriginalPrefixBindings(t *testing.T) {
	tree, err := parseXMLTree([]byte(`<x:document xmlns:x="` + constants.NamespaceMain + `" xmlns:rel="` + constants.NamespaceRelationships + `"><x:body xmlns:rel="urn:body"/></x:document>`))
	if err != nil {
		t.Fatalf("parseXMLTree: %v", err)
	}
	body := findChild(tree, "body")
	got := documentNamespaces(tree, body)
	if got["x"] != constants.NamespaceMain {
		t.Errorf("x namespace = %q, want %q", got["x"], constants.NamespaceMain)
	}
	if got["rel"] != "urn:body" {
		t.Errorf("rel namespace = %q, want body override", got["rel"])
	}
}

type capturedBookmarkIDs []string

func (ids *capturedBookmarkIDs) ObserveHydratedBookmarkID(id string) {
	*ids = append(*ids, id)
}

func TestSourceIDScansUseAttributeQName(t *testing.T) {
	tree, err := parseXMLTree([]byte(`<w:document xmlns:w="` + constants.NamespaceMain + `" xmlns:wp="` + constants.NamespaceWordprocessingDrawing + `" xmlns:x="urn:test"><w:bookmarkStart x:id="99" w:id="7"/><wp:docPr x:id="1" id="40"/></w:document>`))
	if err != nil {
		t.Fatalf("parseXMLTree: %v", err)
	}
	var ids capturedBookmarkIDs
	observeSourceBookmarkIDs(tree, &ids)
	if len(ids) != 1 || ids[0] != "7" {
		t.Fatalf("observed bookmark IDs = %v, want [7]", ids)
	}
	if got := highestSourceDrawingID(tree); got != 40 {
		t.Fatalf("highestSourceDrawingID = %d, want 40", got)
	}
}

func TestWordHydrationIgnoresForeignAttributesWithSameLocalName(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:pPr><w:sectPr><w:pgMar xmlns:x="urn:opaque" x:top="999" w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720"/></w:sectPr></w:pPr></w:p>`)
	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("len(Sections()) = %d, want 2", len(sections))
	}
	margins := sections[0].Margins()
	if margins.Top != 1440 {
		t.Fatalf("hydrated top margin = %d, want 1440", margins.Top)
	}
	margins.Right = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	for _, want := range []string{`x:top="999"`, `w:top="1440"`, `w:right="1500"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
}

func TestRoundTripMainDocument_ModifiesStrictOOXMLWithoutAddingTransitionalNamespace(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, `<w:p><w:r><w:t>Strict</w:t></w:r></w:p>`))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	pkg.MainDocument = bytes.ReplaceAll(pkg.MainDocument, []byte(constants.NamespaceMain), []byte(constants.NamespaceMainStrict))
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}
	sections := doc.Sections()
	if len(sections) != 1 {
		t.Fatalf("len(Sections()) = %d, want 1", len(sections))
	}
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	if !strings.Contains(written, constants.NamespaceMainStrict) || !strings.Contains(written, `w:top="1500"`) {
		t.Errorf("resaved strict document did not contain the requested change:\n%s", written)
	}
	if strings.Contains(written, constants.NamespaceMain+`"`) {
		t.Errorf("resaved strict document added the Transitional namespace:\n%s", written)
	}
	reopenedPackage, err := LoadPackageFromBytes(output.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(first write): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(first write): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(first write): %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate(first write): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second write): %v", err)
	}
	if secondXML := documentXML(t, second.Bytes()); secondXML != written {
		t.Errorf("second Strict write changed document.xml:\nfirst:  %s\nsecond: %s", written, secondXML)
	}
}

func TestRoundTripMainDocument_TransitionalRootWinsOverUnusedStrictDeclaration(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackageWithBindings(
		t,
		"w",
		` xmlns:s="`+constants.NamespaceMainStrict+`"`,
		"",
		`<w:p><w:r><w:t>Existing</w:t></w:r></w:p>`,
	))
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
	paragraph, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := paragraph.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Added"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	body := findWordChild(tree, "body")
	if body == nil {
		t.Fatalf("resaved document has no Transitional body:\n%s", written)
	}
	paragraphCount := 0
	for _, child := range body.Children {
		if child == nil || child.Name.Local != "p" {
			continue
		}
		paragraphCount++
		if child.Name.Space != constants.NamespaceMain {
			t.Errorf("paragraph %d namespace = %q, want Transitional %q:\n%s", paragraphCount, child.Name.Space, constants.NamespaceMain, written)
		}
	}
	if paragraphCount != 2 {
		t.Fatalf("resaved body has %d paragraphs, want 2:\n%s", paragraphCount, written)
	}
	reopenedPackage, err := LoadPackageFromBytes(output.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(output): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(output): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(output): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second): %v", err)
	}
	if secondXML := documentXML(t, second.Bytes()); secondXML != written {
		t.Errorf("second Transitional write changed document.xml:\nfirst:  %s\nsecond: %s", written, secondXML)
	}
}

func TestRoundTripMainDocument_StrictNewImageUsesStrictGraphicDataURI(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, `<w:p><w:r><w:t>Strict image</w:t></w:r></w:p>`))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	pkg.MainDocument = bytes.ReplaceAll(pkg.MainDocument, []byte(constants.NamespaceMain), []byte(constants.NamespaceMainStrict))
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}
	paragraph, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if _, err := paragraph.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	graphicData := findDrawingDescendant(tree, "graphicData")
	if graphicData == nil {
		t.Fatalf("resaved Strict document has no a:graphicData:\n%s", written)
	}
	if uri, _ := getUnqualifiedAttr(graphicData, "uri"); uri != constants.NamespacePictureStrict {
		t.Errorf("Strict graphicData URI = %q, want %q:\n%s", uri, constants.NamespacePictureStrict, written)
	}
	if strings.Contains(written, `uri="`+constants.NamespacePicture+`"`) {
		t.Errorf("Strict graphicData retained the Transitional picture URI:\n%s", written)
	}
	reopenedPackage, err := LoadPackageFromBytes(output.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(output): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(output): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(output): %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate(output): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second): %v", err)
	}
	if secondXML := documentXML(t, second.Bytes()); secondXML != written {
		t.Errorf("second Strict image write changed document.xml:\nfirst:  %s\nsecond: %s", written, secondXML)
	}
}

func TestRoundTripMainDocument_StrictNewHyperlinkUsesStrictRelationshipsNamespace(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, `<w:p><w:r><w:t>Strict</w:t></w:r></w:p>`))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	pkg.MainDocument = bytes.ReplaceAll(pkg.MainDocument, []byte(constants.NamespaceMain), []byte(constants.NamespaceMainStrict))
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}
	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paragraphs))
	}
	if _, err := paragraphs[0].AddHyperlink("https://example.com/strict", "Strict link"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if !strings.Contains(written, constants.NamespaceRelationshipsStrict) {
		t.Errorf("new Strict hyperlink is missing the Strict relationships namespace:\n%s", written)
	}
	if strings.Contains(written, constants.NamespaceRelationships+`"`) {
		t.Errorf("new Strict hyperlink introduced the Transitional relationships namespace:\n%s", written)
	}
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	reopenedPackage, err := LoadPackageFromBytes(output.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(output): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(output): %v", err)
	}
	if _, err := ReconstructDocument(reopenedParsed); err != nil {
		t.Fatalf("ReconstructDocument(output): %v", err)
	}
}

func TestRoundTripMainDocument_StrictChangedSectionCarrierHasOneParagraphPropertiesElement(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, `<w:p><w:pPr><w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720"/></w:sectPr></w:pPr><w:r><w:t>First section</w:t></w:r></w:p><w:p><w:r><w:t>Second section</w:t></w:r></w:p>`))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	pkg.MainDocument = bytes.ReplaceAll(pkg.MainDocument, []byte(constants.NamespaceMain), []byte(constants.NamespaceMainStrict))
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}
	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 2 {
		t.Fatalf("len(Paragraphs()) = %d, want 2", len(paragraphs))
	}
	if err := paragraphs[0].SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	sections := doc.Sections()
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	if count := strings.Count(written, `<w:pPr>`); count != 1 {
		t.Fatalf("first Strict section carrier produced %d pPr elements, want 1:\n%s", count, written)
	}
	if count := strings.Count(written, `<w:sectPr>`); count != 2 {
		t.Fatalf("Strict section count = %d, want one embedded and one final section:\n%s", count, written)
	}
	for _, want := range []string{`w:val="center"`, `w:top="1500"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved Strict section carrier lost %s:\n%s", want, written)
		}
	}
}

func TestRoundTripMainDocument_AddSectionMovesFinalSectionOpaquePropertiesWithOriginalSection(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, `<w:p><w:r><w:t>Original section</w:t></w:r></w:p>`))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	originalFinal := `<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>`
	customFinal := `<!--final-anchor--><w:sectPr w:rsidR="AAAA"><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720" w:gutter="72"/><w:docGrid w:linePitch="360"/></w:sectPr>`
	pkg.MainDocument = bytes.Replace(pkg.MainDocument, []byte(originalFinal), []byte(customFinal), 1)
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}
	newSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeContinuous)
	if err != nil {
		t.Fatalf("AddSectionWithBreak: %v", err)
	}
	newMargins := newSection.Margins()
	newMargins.Top = 1800
	if err := newSection.SetMargins(newMargins); err != nil {
		t.Fatalf("SetMargins(new section): %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	body := findWordChild(tree, "body")
	var embeddedSection, finalSection *Element
	for _, child := range body.Children {
		switch {
		case isWordElement(child, "p"):
			if props := findWordChild(child, "pPr"); props != nil {
				if section := findWordChild(props, "sectPr"); section != nil {
					embeddedSection = section
				}
			}
		case isWordElement(child, "sectPr"):
			finalSection = child
		}
	}
	if embeddedSection == nil || finalSection == nil {
		t.Fatalf("resaved document is missing embedded or final section:\n%s", written)
	}
	if rsid, _ := getWordAttr(embeddedSection, "rsidR"); rsid != "AAAA" {
		t.Errorf("embedded original section rsidR = %q, want AAAA", rsid)
	}
	if findWordChild(embeddedSection, "docGrid") == nil {
		t.Errorf("embedded original section lost docGrid:\n%s", written)
	}
	if margins := findWordChild(embeddedSection, "pgMar"); margins == nil {
		t.Errorf("embedded original section lost pgMar:\n%s", written)
	} else if gutter, _ := getWordAttr(margins, "gutter"); gutter != "72" {
		t.Errorf("embedded original section gutter = %q, want 72", gutter)
	}
	if sectionType := findWordChild(embeddedSection, "type"); sectionType == nil {
		t.Errorf("embedded original section lost its new break type:\n%s", written)
	} else if value, _ := getWordAttr(sectionType, "val"); value != "continuous" {
		t.Errorf("embedded section break type = %q, want continuous", value)
	}
	if _, ok := getWordAttr(finalSection, "rsidR"); ok || findWordChild(finalSection, "docGrid") != nil {
		t.Errorf("new final section inherited opaque properties from the original section:\n%s", written)
	}
	if margins := findWordChild(finalSection, "pgMar"); margins == nil {
		t.Errorf("new final section is missing pgMar:\n%s", written)
	} else {
		if top, _ := getWordAttr(margins, "top"); top != "1800" {
			t.Errorf("new final section top margin = %q, want 1800", top)
		}
		if _, ok := getWordAttr(margins, "gutter"); ok {
			t.Errorf("new final section inherited gutter from original section:\n%s", written)
		}
	}
	if anchor, section := strings.Index(written, `<!--final-anchor-->`), strings.Index(written, `<w:sectPr w:rsidR="AAAA"`); anchor < 0 || section < 0 || anchor > section {
		t.Errorf("opaque final-section prefix did not move with the original section:\n%s", written)
	}
	reopenedPackage, err := LoadPackageFromBytes(output.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(first write): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(first write): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(first write): %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate(first write): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second write): %v", err)
	}
	if secondXML := documentXML(t, second.Bytes()); secondXML != written {
		t.Errorf("second write changed moved-section document.xml:\nfirst:  %s\nsecond: %s", written, secondXML)
	}
}

func TestRoundTripMainDocument_RemovingEmptyFinalSectionMovesOpaquePropertiesToFinalSection(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:r><w:t>First section</w:t></w:r></w:p><w:p><w:pPr><w:sectPr w:rsidR="FIRST"><w:type w:val="continuous"/><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720" w:gutter="72"/><w:docGrid w:linePitch="360"/></w:sectPr></w:pPr></w:p>`)
	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("len(Sections()) = %d, want 2", len(sections))
	}
	remover, ok := doc.(interface {
		RemoveLastSection(domain.Section) bool
	})
	if !ok {
		t.Fatal("reconstructed document does not expose RemoveLastSection")
	}
	if !remover.RemoveLastSection(sections[1]) {
		t.Fatal("RemoveLastSection returned false for an empty final section")
	}

	var first bytes.Buffer
	if _, err := doc.WriteTo(&first); err != nil {
		t.Fatalf("WriteTo(first): %v", err)
	}
	written := documentXML(t, first.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	body := findWordChild(tree, "body")
	if body == nil {
		t.Fatalf("resaved document has no body:\n%s", written)
	}
	var finalSection *Element
	paragraphCount := 0
	for _, child := range body.Children {
		switch {
		case isWordElement(child, "p"):
			paragraphCount++
		case isWordElement(child, "sectPr"):
			if finalSection != nil {
				t.Fatalf("resaved body has more than one final sectPr:\n%s", written)
			}
			finalSection = child
		}
	}
	if paragraphCount != 1 || finalSection == nil {
		t.Fatalf("resaved body has %d paragraphs and final sectPr %v, want 1 and present:\n%s", paragraphCount, finalSection != nil, written)
	}
	if rsid, _ := getWordAttr(finalSection, "rsidR"); rsid != "FIRST" {
		t.Errorf("final section rsidR = %q, want FIRST", rsid)
	}
	if findWordChild(finalSection, "docGrid") == nil {
		t.Errorf("final section lost embedded docGrid:\n%s", written)
	}
	if margins := findWordChild(finalSection, "pgMar"); margins == nil {
		t.Errorf("final section lost embedded margins:\n%s", written)
	} else if gutter, _ := getWordAttr(margins, "gutter"); gutter != "72" {
		t.Errorf("final section gutter = %q, want 72", gutter)
	}
	if findWordChild(finalSection, "type") != nil {
		t.Errorf("final section retained the intermediate break type:\n%s", written)
	}

	reopenedPackage, err := LoadPackageFromBytes(first.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(first): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(first): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(first): %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate(first): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second): %v", err)
	}
	if secondXML := documentXML(t, second.Bytes()); secondXML != written {
		t.Errorf("second write changed document.xml after removing the final section:\nfirst:  %s\nsecond: %s", written, secondXML)
	}
}

func TestRoundTripMainDocument_ExpandsSelfClosingBodyBeforeAddingContent(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackage(t, ""))
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	original := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + constants.NamespaceMain + `"><w:body/></w:document>`
	pkg.MainDocument = []byte(original)
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}
	doc, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	var unchanged bytes.Buffer
	if _, err := doc.WriteTo(&unchanged); err != nil {
		t.Fatalf("WriteTo(unchanged): %v", err)
	}
	if got := documentXML(t, unchanged.Bytes()); got != original {
		t.Fatalf("unchanged self-closing body was rewritten:\ngot:  %s\nwant: %s", got, original)
	}
	paragraph, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := paragraph.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Added inside body"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var changed bytes.Buffer
	if _, err := doc.WriteTo(&changed); err != nil {
		t.Fatalf("WriteTo(changed): %v", err)
	}
	written := documentXML(t, changed.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	body := findWordChild(tree, "body")
	if body == nil || findWordChild(body, "p") == nil {
		t.Fatalf("added paragraph is not inside w:body:\n%s", written)
	}
	for _, child := range tree.Children {
		if isWordElement(child, "p") {
			t.Fatalf("added paragraph was written as a sibling of w:body:\n%s", written)
		}
	}
	reopenedPackage, err := LoadPackageFromBytes(changed.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(changed): %v", err)
	}
	reopenedParsed, err := ParsePackage(reopenedPackage)
	if err != nil {
		t.Fatalf("ParsePackage(changed): %v", err)
	}
	reopened, err := ReconstructDocument(reopenedParsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(changed): %v", err)
	}
	if len(reopened.Paragraphs()) != 1 {
		t.Fatalf("reopened paragraph count = %d, want 1", len(reopened.Paragraphs()))
	}
}

func TestRoundTripMainDocument_MergesSectionWithBodyScopedPrefix(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackageWithPrefixAndBodyAttrs(t, "w", ` xmlns:x="`+constants.NamespaceMain+`"`, `<x:p><x:pPr><x:sectPr><x:pgMar x:top="1440" x:right="1440" x:bottom="1440" x:left="1440" x:header="720" x:footer="720" x:gutter="36"/></x:sectPr></x:pPr></x:p>
<x:p><x:r><x:t>Next</x:t></x:r></x:p>`))
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
	margins := doc.Sections()[0].Margins()
	margins.Top = 1500
	if err := doc.Sections()[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	for _, want := range []string{`<x:sectPr>`, `x:top="1500"`, `x:gutter="36"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
	if strings.Contains(written, `x:top="1440"`) {
		t.Errorf("resaved document retained the old margin:\n%s", written)
	}
}

func TestRoundTripMainDocument_MergesSectionWithParagraphScopedPrefix(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<x:p xmlns:x="`+constants.NamespaceMain+`" xmlns:w="urn:not-wordprocessingml"><x:pPr><x:sectPr><x:pgMar x:top="1440" x:right="1440" x:bottom="1440" x:left="1440" x:header="720" x:footer="720" x:gutter="48"/></x:sectPr></x:pPr></x:p>
<w:p><w:r><w:t>Next</w:t></w:r></w:p>`)
	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("len(Sections()) = %d, want 2", len(sections))
	}
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	for _, want := range []string{`<x:sectPr>`, `x:top="1500"`, `x:gutter="48"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
	if strings.Contains(written, `x:top="1440"`) || strings.Contains(written, `<w:pgMar`) {
		t.Errorf("resaved section retained or duplicated the old margins:\n%s", written)
	}
}

func TestRoundTripMainDocument_MergesSectionUsingOriginalPrefix(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackageWithPrefix(t, "x", `<x:p><x:pPr><x:sectPr><x:pgMar x:top="1440" x:right="1440" x:bottom="1440" x:left="1440" x:header="720" x:footer="720" x:gutter="24"/></x:sectPr></x:pPr></x:p>
<x:p><x:r><x:t>Next</x:t></x:r></x:p>`))
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
	margins := doc.Sections()[0].Margins()
	margins.Top = 1500
	if err := doc.Sections()[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	for _, want := range []string{`<x:sectPr>`, `x:top="1500"`, `x:gutter="24"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
	if strings.Contains(written, `xmlns:w=`) {
		t.Errorf("resaved document added an unused w namespace:\n%s", written)
	}
}

func TestRoundTripMainDocument_AvoidsConflictingSerializerPrefixes(t *testing.T) {
	pkg, err := LoadPackageFromBytes(buildBookmarkPackageWithBindings(t, "x", ` xmlns:w="urn:not-wordprocessingml"`, "", `<x:p><x:pPr><x:sectPr><x:pgMar x:top="1440" x:right="1440" x:bottom="1440" x:left="1440" x:header="720" x:footer="720"/></x:sectPr></x:pPr></x:p>
<x:p><x:r><x:t>Existing</x:t></x:r></x:p>`))
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
	margins := doc.Sections()[0].Margins()
	margins.Top = 1500
	if err := doc.Sections()[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}
	paragraph, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := paragraph.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Added"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	for _, want := range []string{`x:top="1500"`, `<x:t>Added</x:t>`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
	if strings.Contains(written, `<w:p`) || strings.Contains(written, `w:top="1500"`) {
		t.Errorf("generated content used the conflicting w prefix:\n%s", written)
	}
}

type bookmarkedParagraph interface {
	BookmarkID() string
	BookmarkName() string
}

func startTagHas(docXML, element string, attrs ...string) bool {
	prefix := "<w:" + element
	for rest := docXML; ; {
		index := strings.Index(rest, prefix)
		if index < 0 {
			return false
		}
		rest = rest[index:]
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			return false
		}
		tag := rest[:end+1]
		matches := true
		for _, attr := range attrs {
			if !strings.Contains(tag, attr) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
		rest = rest[end+1:]
	}
}

// TestReconstructBookmark_SameParagraphHydrates covers populateParagraph's
// bookmarkStart/bookmarkEnd handling: a bookmark whose start and end both
// fall in one paragraph is the only shape core.paragraph's single (id, name)
// pair can represent, and it must survive a read-then-write round trip.
func TestReconstructBookmark_SameParagraphHydrates(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p xmlns:x="urn:test">
<w:bookmarkStart x:id="99" x:name="Opaque" w:id="3" w:name="MyBookmark"/>
<w:r><w:t>Hello</w:t></w:r>
<w:bookmarkEnd x:id="99" w:id="3"/>
</w:p>`)

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paras))
	}
	bp, ok := paras[0].(bookmarkedParagraph)
	if !ok {
		t.Fatal("paragraph does not expose BookmarkID/BookmarkName")
	}
	if bp.BookmarkID() != "3" || bp.BookmarkName() != "MyBookmark" {
		t.Errorf("BookmarkID/BookmarkName = %q/%q, want \"3\"/\"MyBookmark\"", bp.BookmarkID(), bp.BookmarkName())
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !startTagHas(written, "bookmarkStart", `w:id="3"`, `w:name="MyBookmark"`) {
		t.Errorf("resaved document.xml lost the bookmark start:\n%s", written)
	}
	if !startTagHas(written, "bookmarkEnd", `w:id="3"`) {
		t.Errorf("resaved document.xml lost the bookmark end:\n%s", written)
	}
}

// TestReconstructBookmark_SpanningParagraphsIsNotHydratedButIsPreserved pins the documented
// limit of the single-paragraph model: a bookmark whose end falls in a
// different paragraph than its start has nowhere to live in the domain model.
// Raw XML preservation nevertheless keeps the unedited source intact.
func TestReconstructBookmark_SpanningParagraphsIsNotHydratedButIsPreserved(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p xmlns:x="urn:test"><w:bookmarkStart x:id="999" w:id="1" w:name="Spans"/><w:r><w:t>A</w:t></w:r></w:p>
<w:p><w:r><w:t>B</w:t></w:r><w:bookmarkEnd w:id="1"/></w:p>`)

	for i, para := range doc.Paragraphs() {
		bp, ok := para.(bookmarkedParagraph)
		if !ok {
			t.Fatalf("paragraph %d does not expose BookmarkID/BookmarkName", i)
		}
		if bp.BookmarkID() != "" {
			t.Errorf("Paragraphs()[%d].BookmarkID() = %q, want \"\" (bookmark spans paragraphs, must be dropped)", i, bp.BookmarkID())
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:name="Spans"`) {
		t.Errorf("resaved document.xml lost the unmodeled spanning bookmark:\n%s", written)
	}

	if err := doc.Paragraphs()[0].SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	buf.Reset()
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo after edit: %v", err)
	}
	edited := documentXML(t, buf.Bytes())
	if strings.Contains(edited, "bookmarkStart") || strings.Contains(edited, "bookmarkEnd") {
		t.Errorf("editing one paragraph left an incomplete spanning bookmark:\n%s", edited)
	}
}

func TestRoundTripMainDocument_UnmatchedSourceRangeSurvivesUnrelatedEdit(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:bookmarkStart w:id="77" w:name="AlreadyUnmatched"/><w:r><w:t>Content</w:t></w:r></w:p>`)
	sections := doc.Sections()
	if len(sections) != 1 {
		t.Fatalf("len(Sections()) = %d, want 1", len(sections))
	}
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if !strings.Contains(written, `w:name="AlreadyUnmatched"`) {
		t.Errorf("unrelated section edit removed an unmatched source marker:\n%s", written)
	}
	if !strings.Contains(written, `w:top="1500"`) {
		t.Errorf("section edit was not written:\n%s", written)
	}
}

// TestReconstructBookmark_PartialRunIsNotHydratedButIsPreserved covers a bookmark that, in the
// source, wraps only "target" inside "prefix target suffix" -- all in one
// paragraph, so the same-paragraph check alone would accept it. But
// core.paragraph's single (id, name) pair has no notion of where within the
// paragraph a bookmark starts or ends: hydrating it would re-serialize as
// w:bookmarkStart at the very beginning of the paragraph and w:bookmarkEnd at
// the very end, silently widening it to wrap "prefix", "target", and
// "suffix" alike. A REF field pointed at this bookmark would then resolve
// the whole paragraph's text instead of just "target". Representing this
// correctly needs per-run position tracking that paragraph doesn't have, so
// the partial bookmark is not hydrated. Its raw XML remains intact until that
// paragraph is edited.
func TestReconstructBookmark_PartialRunIsNotHydratedButIsPreserved(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p>
<w:r><w:t xml:space="preserve">prefix </w:t></w:r>
<w:bookmarkStart w:id="4" w:name="Target"/>
<w:r><w:t>target</w:t></w:r>
<w:bookmarkEnd w:id="4"/>
<w:r><w:t xml:space="preserve"> suffix</w:t></w:r>
</w:p>`)

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paras))
	}
	bp, ok := paras[0].(bookmarkedParagraph)
	if !ok {
		t.Fatal("paragraph does not expose BookmarkID/BookmarkName")
	}
	if bp.BookmarkID() != "" {
		t.Errorf("BookmarkID() = %q, want \"\" (bookmark wraps only part of the paragraph's runs, must be dropped)", bp.BookmarkID())
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:name="Target"`) {
		t.Errorf("resaved document.xml lost the unmodeled partial bookmark:\n%s", written)
	}
	if err := paras[0].SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	buf.Reset()
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo after edit: %v", err)
	}
	edited := documentXML(t, buf.Bytes())
	if strings.Contains(edited, "Target") || strings.Contains(edited, "bookmarkStart") || strings.Contains(edited, "bookmarkEnd") {
		t.Errorf("edited paragraph re-emitted an unrepresentable partial bookmark:\n%s", edited)
	}
}

// TestReconstructBookmark_FullParagraphAcrossMultipleRunsHydrates is the
// counterpart to the partial-run case above: a bookmark that legitimately
// wraps a paragraph's entire content, even when that content is split across
// several differently-formatted runs, must still hydrate -- the fix for the
// partial case must key off position relative to the paragraph's content,
// not merely "does the paragraph have more than one run".
func TestReconstructBookmark_FullParagraphAcrossMultipleRunsHydrates(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p>
<w:bookmarkStart w:id="9" w:name="WholeThing"/>
<w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Bold </w:t></w:r>
<w:r><w:t>plain</w:t></w:r>
<w:bookmarkEnd w:id="9"/>
</w:p>`)

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paras))
	}
	bp, ok := paras[0].(bookmarkedParagraph)
	if !ok {
		t.Fatal("paragraph does not expose BookmarkID/BookmarkName")
	}
	if bp.BookmarkID() != "9" || bp.BookmarkName() != "WholeThing" {
		t.Errorf("BookmarkID/BookmarkName = %q/%q, want \"9\"/\"WholeThing\"", bp.BookmarkID(), bp.BookmarkName())
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:id="9" w:name="WholeThing"`) {
		t.Errorf("resaved document.xml lost a bookmark that legitimately wraps the whole paragraph:\n%s", written)
	}
}

// TestGenerateHeadingBookmarks_DoesNotClobberHydratedBookmark is the direct
// regression test for the bug generateHeadingBookmarks had before this
// change: it ran on every WriteTo, including a pure round trip, and
// unconditionally overwrote any bookmark on a Heading-styled paragraph.
func TestGenerateHeadingBookmarks_DoesNotClobberHydratedBookmark(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p>
<w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
<w:bookmarkStart w:id="7" w:name="_Ref123"/>
<w:r><w:t>Title</w:t></w:r>
<w:bookmarkEnd w:id="7"/>
</w:p>`)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	if !strings.Contains(written, `w:id="7" w:name="_Ref123"`) {
		t.Errorf("resaved document.xml lost the hydrated heading bookmark:\n%s", written)
	}
	if strings.Contains(written, `_Toc`) {
		t.Errorf("resaved document.xml assigned a generated _Toc bookmark to a paragraph that already had a hydrated one:\n%s", written)
	}
}

// TestGenerateHeadingBookmarks_StartsAboveHighestHydratedID covers the
// allocator: a newly added Heading-styled paragraph gets a bookmark whose
// numeric w:id starts above every bookmark ID hydrated from the source.
func TestGenerateHeadingBookmarks_StartsAboveHighestHydratedID(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:bookmarkStart w:id="5" w:name="SomeAnchor"/><w:r><w:t>Anchor</w:t></w:r><w:bookmarkEnd w:id="5"/></w:p>
<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Chapter One</w:t></w:r></w:p>`)
	newHeading, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := newHeading.SetStyle("Heading1"); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	run, err := newHeading.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Chapter Two"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())

	if !strings.Contains(written, `<w:bookmarkStart w:id="6" w:name="_Toc6"`) {
		t.Errorf("resaved document.xml did not start the generated heading bookmark at id 6 (above the hydrated id 5):\n%s", written)
	}
}

func TestGenerateHeadingBookmarks_ChangedOriginalParagraphStillGetsOne(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:r><w:t>Promote me</w:t></w:r></w:p>`)
	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paragraphs))
	}
	if err := paragraphs[0].SetStyle("Heading1"); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `<w:bookmarkStart w:id="0" w:name="_Toc0"`) {
		t.Errorf("edited original paragraph did not receive a heading bookmark:\n%s", written)
	}
}

func TestGenerateHeadingBookmarks_StartsAboveBookmarkInsideOpaqueContent(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:sdt><w:sdtContent><w:p>
<w:bookmarkStart w:id="0" w:name="Opaque"/><w:r><w:t>Opaque</w:t></w:r><w:bookmarkEnd w:id="0"/>
</w:p></w:sdtContent></w:sdt>`)
	heading, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := heading.SetStyle("Heading1"); err != nil {
		t.Fatalf("SetStyle: %v", err)
	}
	run, err := heading.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("New heading"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `<w:bookmarkStart w:id="1" w:name="_Toc1"`) {
		t.Errorf("generated heading bookmark collided with the opaque bookmark id 0:\n%s", written)
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

	// The source paragraph carries both content and sectPr. Raw preservation
	// keeps that valid shape instead of splitting it into a content paragraph
	// plus a synthetic empty section-break carrier.
	if got := len(paragraphOpenTagRE.FindAllString(written, -1)); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:p> elements, want 2", got)
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

	// The source shape is retained even when w:type is omitted and defaults to
	// nextPage.
	if got := len(paragraphOpenTagRE.FindAllString(written, -1)); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:p> elements, want 2", got)
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
	if !startTagHas(written, "tblStyle", `w:val="TableGrid"`) {
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
func TestParseMeasurementInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "integer", input: "9360", want: 9360},
		{name: "float formatted integer", input: "9360.0", want: 9360},
		{name: "fraction truncates", input: "9360.75", want: 9360},
		{name: "negative fraction truncates toward zero", input: "-12.75", want: -12},
		{name: "not a number", input: "broken", wantErr: true},
		{name: "nan", input: "NaN", wantErr: true},
		{name: "positive infinity", input: "+Inf", wantErr: true},
		{name: "overflow", input: "1e100", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMeasurementInt(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseMeasurementInt(%q) = %d, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseMeasurementInt(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseMeasurementInt(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

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

func TestRoundTripMainDocument_EditedParagraphKeepsOpaqueAdjacentBodyChild(t *testing.T) {
	opaque := `<w:sdt><w:sdtPr><w:tag w:val="keep-verbatim"/></w:sdtPr><w:sdtContent><w:p><w:r><w:t>Opaque content</w:t></w:r></w:p></w:sdtContent></w:sdt>`
	doc := reconstructFromBookmarkBody(t, opaque+`
<w:p><w:r><w:t>Editable</w:t></w:r></w:p>`)

	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1 modeled paragraph", len(paragraphs))
	}
	if err := paragraphs[0].SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, opaque) {
		t.Errorf("resaved document.xml changed the adjacent opaque body child:\n%s", written)
	}
	if !startTagHas(written, "jc", `w:val="center"`) {
		t.Errorf("resaved document.xml did not regenerate the edited paragraph:\n%s", written)
	}
}

func TestRoundTripMainDocument_RegeneratedFragmentDeclaresRelationshipNamespace(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:r><w:t>Source</w:t></w:r></w:p>`)
	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 1 {
		t.Fatalf("len(Paragraphs()) = %d, want 1", len(paragraphs))
	}
	if _, err := paragraphs[0].AddHyperlink("https://example.com", "link"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v", err)
	}
	hyperlink := findDescendant(tree, "hyperlink")
	if hyperlink == nil {
		t.Fatalf("resaved document.xml is missing the hyperlink:\n%s", written)
	}
	for _, attr := range hyperlink.Attr {
		if attr.Name.Local == "id" {
			if attr.Name.Space != constants.NamespaceRelationships {
				t.Fatalf("hyperlink id namespace = %q, want %q:\n%s", attr.Name.Space, constants.NamespaceRelationships, written)
			}
			return
		}
	}
	t.Fatalf("resaved hyperlink is missing r:id:\n%s", written)
}

func TestRoundTripMainDocument_NewDrawingStartsAboveOpaqueDrawingIDs(t *testing.T) {
	opaque := `<w:sdt><w:sdtContent><w:p><w:r><w:drawing>` +
		`<wp:inline xmlns:wp="` + constants.NamespaceWordprocessingDrawing + `">` +
		`<wp:docPr id="40" name="Opaque drawing"/>` +
		`</wp:inline></w:drawing></w:r></w:p></w:sdtContent></w:sdt>`
	doc := reconstructFromBookmarkBody(t, opaque)
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if _, err := para.AddImage(createTestPNG(t)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `<wp:docPr id="40" name="Opaque drawing"`) {
		t.Errorf("resaved document.xml lost the opaque drawing ID:\n%s", written)
	}
	if !strings.Contains(written, `<wp:docPr id="41"`) {
		t.Errorf("new drawing did not start above the opaque wp:docPr id 40:\n%s", written)
	}
}

func TestRoundTripMainDocument_EditedTableRowKeepsShellAndOtherRows(t *testing.T) {
	prefix := `<w:tbl><w:tblPr><w:tblpPr w:leftFromText="180" w:rightFromText="180"/></w:tblPr><w:tblGrid><w:gridCol w:w="2400"/></w:tblGrid>`
	firstRow := `<w:tr w:rsidR="00112233"><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>First</w:t></w:r></w:p></w:tc></w:tr>`
	betweenRows := `<w:bookmarkStart w:id="9" w:name="BetweenRows"/><w:bookmarkEnd w:id="9"/>`
	secondRow := `<w:tr w:rsidR="00445566"><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>Second</w:t></w:r></w:p></w:tc></w:tr>`
	table, doc := reconstructTableFromXML(t, prefix+firstRow+betweenRows+secondRow+`</w:tbl>`)

	row, err := table.Row(1)
	if err != nil {
		t.Fatalf("Row(1): %v", err)
	}
	if err := row.SetHeight(360); err != nil {
		t.Fatalf("SetHeight: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, prefix) {
		t.Errorf("resaved document.xml changed the table shell:\n%s", written)
	}
	if !strings.Contains(written, firstRow) {
		t.Errorf("resaved document.xml changed the untouched first row:\n%s", written)
	}
	if !strings.Contains(written, betweenRows) {
		t.Errorf("resaved document.xml lost the markup between rows:\n%s", written)
	}
	if strings.Contains(written, secondRow) {
		t.Errorf("resaved document.xml kept stale raw XML for the edited second row:\n%s", written)
	}
	if !startTagHas(written, "trHeight", `w:val="360"`) {
		t.Errorf("resaved document.xml did not serialize the edited row height:\n%s", written)
	}

	if err := table.DeleteRow(1); err != nil {
		t.Fatalf("DeleteRow(1): %v", err)
	}
	buf.Reset()
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo after DeleteRow: %v", err)
	}
	written = documentXML(t, buf.Bytes())
	if !strings.Contains(written, betweenRows) {
		t.Errorf("deleting the following row also deleted the markup before it:\n%s", written)
	}
}

func TestRoundTripMainDocument_EditedTableShellKeepsOpaquePropertiesAndGaps(t *testing.T) {
	floating := `<w:tblpPr w:leftFromText="180" w:rightFromText="180"/>`
	betweenRows := `<w:bookmarkStart w:id="9" w:name="BetweenRows"/><w:bookmarkEnd w:id="9"/>`
	tableXML := `<w:tbl><w:tblPr>` + floating + `<w:tblW w:w="4800" w:type="dxa"/></w:tblPr>` +
		`<w:tblGrid><w:gridCol w:w="2400"/></w:tblGrid>` +
		`<w:tr><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>First</w:t></w:r></w:p></w:tc></w:tr>` +
		betweenRows +
		`<w:tr><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>Second</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`
	table, doc := reconstructTableFromXML(t, tableXML)
	if err := table.SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, floating) {
		t.Errorf("changing table alignment deleted w:tblpPr:\n%s", written)
	}
	if !strings.Contains(written, betweenRows) {
		t.Errorf("changing table alignment deleted markup between rows:\n%s", written)
	}
	if !startTagHas(written, "jc", `w:val="center"`) {
		t.Errorf("resaved document.xml did not contain the changed table alignment:\n%s", written)
	}
}

func TestRoundTripMainDocument_EditedGridKeepsOpaqueGridChildren(t *testing.T) {
	gridChange := `<w:tblGridChange w:id="7"><w:tblGrid><w:gridCol w:w="1800"/></w:tblGrid></w:tblGridChange>`
	tableXML := `<w:tbl><w:tblPr><w:tblW w:w="2400" w:type="dxa"/></w:tblPr>` +
		`<w:tblGrid><w:gridCol w:w="2400"/>` + gridChange + `</w:tblGrid>` +
		`<w:tr><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>Cell</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`
	table, doc := reconstructTableFromXML(t, tableXML)
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if err := cell.SetWidth(3000); err != nil {
		t.Fatalf("SetWidth: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, gridChange) {
		t.Errorf("changing the current grid deleted w:tblGridChange:\n%s", written)
	}
	if !startTagHas(written, "gridCol", `w:w="3000"`) {
		t.Errorf("resaved document.xml did not contain the changed grid width:\n%s", written)
	}
}

func TestRoundTripMainDocument_EmbeddedSectionStaysInContentParagraph(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p w:rsidR="AAAA"><w:pPr><w:spacing w:after="120"/><w:sectPr w:rsidR="BBBB"><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720" w:gutter="72"/><w:docGrid w:linePitch="360"/></w:sectPr></w:pPr><w:r><w:t>First section</w:t></w:r></w:p>
<w:p><w:r><w:t>Second section</w:t></w:r></w:p>`)
	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("len(Sections()) = %d, want 2", len(sections))
	}
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var first bytes.Buffer
	if _, err := doc.WriteTo(&first); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, first.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v", err)
	}
	body := findDescendant(tree, "body")
	var paragraphs []*Element
	for _, child := range body.Children {
		if child != nil && child.Name.Local == "p" {
			paragraphs = append(paragraphs, child)
		}
	}
	if len(paragraphs) != 2 {
		t.Fatalf("resaved body has %d paragraphs, want 2:\n%s", len(paragraphs), written)
	}
	if findDescendant(paragraphs[0], "sectPr") == nil {
		t.Fatalf("first content paragraph lost its embedded sectPr:\n%s", written)
	}
	for _, want := range []string{`w:rsidR="AAAA"`, `w:rsidR="BBBB"`, `w:after="120"`, `w:top="1500"`, `w:gutter="72"`, `<w:docGrid w:linePitch="360"/>`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}

	pkg, err := LoadPackageFromBytes(first.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes(first write): %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage(first write): %v", err)
	}
	reopened, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument(first write): %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate(first write): %v", err)
	}
	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo(second write): %v", err)
	}
	if got := documentXML(t, second.Bytes()); got != written {
		t.Errorf("second write changed document.xml:\nfirst:  %s\nsecond: %s", written, got)
	}
}

func TestRoundTripMainDocument_SectionOnlyChangeKeepsHeadingParagraphRaw(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p w:rsidR="AAAA"><w:pPr><w:pStyle w:val="Heading1"/><w:kinsoku w:val="1"/><w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720"/></w:sectPr></w:pPr><w:r><w:t>Heading section</w:t></w:r></w:p>
<w:p><w:r><w:t>Second section</w:t></w:r></w:p>`)
	sections := doc.Sections()
	if len(sections) != 2 {
		t.Fatalf("len(Sections()) = %d, want 2", len(sections))
	}
	margins := sections[0].Margins()
	margins.Top = 1500
	if err := sections[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	for _, want := range []string{`w:rsidR="AAAA"`, `<w:kinsoku w:val="1"/>`, `w:top="1500"`} {
		if !strings.Contains(written, want) {
			t.Errorf("section-only change lost %s from the heading carrier: %s", want, written)
		}
	}
	if strings.Contains(written, `<w:bookmarkStart`) {
		t.Errorf("section-only change generated a bookmark in the unchanged heading: %s", written)
	}
}

func TestRoundTripMainDocument_ForeignSectPrDoesNotShadowFinalSection(t *testing.T) {
	foreign := `<x:sectPr xmlns:x="urn:opaque" x:marker="keep"/>`
	doc := reconstructFromBookmarkBody(t, foreign+`<w:p><w:r><w:t>Body</w:t></w:r></w:p>`)
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}
	margins := section.Margins()
	margins.Top = 1500
	if err := section.SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if !strings.Contains(written, foreign) {
		t.Errorf("resaved document changed the foreign sectPr: %s", written)
	}
	finalStart := strings.LastIndex(written, `<w:sectPr`)
	if finalStart < 0 {
		t.Fatalf("resaved document has no WordprocessingML final sectPr: %s", written)
	}
	if final := written[finalStart:]; !strings.Contains(final, `w:top="1500"`) {
		t.Errorf("final WordprocessingML sectPr did not receive the margin change: %s", written)
	}
}

func TestRoundTripMainDocument_ChangedParagraphKeepsItsChangedSection(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p w:rsidR="AAAA"><w:pPr><w:spacing w:after="120"/><w:sectPr w:rsidR="BBBB"><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720" w:gutter="72"/></w:sectPr></w:pPr><w:r><w:t>First section</w:t></w:r></w:p>
<w:p><w:r><w:t>Second section</w:t></w:r></w:p>`)
	paragraphs := doc.Paragraphs()
	if len(paragraphs) != 2 {
		t.Fatalf("len(Paragraphs()) = %d, want 2", len(paragraphs))
	}
	if err := paragraphs[0].SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	margins := doc.Sections()[0].Margins()
	margins.Top = 1500
	if err := doc.Sections()[0].SetMargins(margins); err != nil {
		t.Fatalf("SetMargins: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v", err)
	}
	body := findDescendant(tree, "body")
	var bodyParagraphs []*Element
	for _, child := range body.Children {
		if child != nil && child.Name.Local == "p" {
			bodyParagraphs = append(bodyParagraphs, child)
		}
	}
	if len(bodyParagraphs) != 2 {
		t.Fatalf("resaved body has %d paragraphs, want 2:\n%s", len(bodyParagraphs), written)
	}
	first := bodyParagraphs[0]
	if findDescendant(first, "sectPr") == nil || findDescendant(first, "jc") == nil {
		t.Fatalf("changed paragraph did not retain both pPr changes and sectPr:\n%s", written)
	}
	for _, want := range []string{`w:val="center"`, `w:top="1500"`, `w:gutter="72"`} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document lost %s:\n%s", want, written)
		}
	}
}

func TestRoundTripMainDocument_ChangingColumnsDropsIncompatibleExplicitLayout(t *testing.T) {
	doc := reconstructFromBookmarkBody(t, `<w:p><w:pPr><w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:cols w:num="2" w:equalWidth="0" w:space="500" w:sep="1"><w:col w:w="3000" w:space="500"/><x:balance xmlns:x="urn:opaque" x:mode="keep"/><w:col w:w="7000" w:space="0"/></w:cols></w:sectPr></w:pPr></w:p>
<w:p><w:r><w:t>Next section</w:t></w:r></w:p>`)
	if err := doc.Sections()[0].SetColumns(3); err != nil {
		t.Fatalf("SetColumns: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v", err)
	}
	cols := findDescendant(tree, "cols")
	if cols == nil {
		t.Fatalf("resaved document has no w:cols:\n%s", written)
	}
	if got, _ := getAttr(cols, "num"); got != "3" {
		t.Errorf("w:cols/@w:num = %q, want 3", got)
	}
	if _, ok := getAttr(cols, "equalWidth"); ok {
		t.Errorf("resaved w:cols retained incompatible equalWidth=0:\n%s", written)
	}
	if got, _ := getAttr(cols, "space"); got != "500" {
		t.Errorf("w:cols/@w:space = %q, want preserved 500", got)
	}
	if got, _ := getAttr(cols, "sep"); got != "1" {
		t.Errorf("w:cols/@w:sep = %q, want preserved 1", got)
	}
	if !strings.Contains(written, `<x:balance xmlns:x="urn:opaque" x:mode="keep"/>`) {
		t.Errorf("resaved w:cols lost its compatible opaque child:\n%s", written)
	}
	for _, child := range cols.Children {
		if child != nil && child.Name.Local == "col" {
			t.Errorf("resaved w:cols retained an incompatible explicit w:col:\n%s", written)
		}
	}
}

func TestRoundTripMainDocument_ChangingBorderPreservesUnmodeledAttributesPerSide(t *testing.T) {
	tableXML := `<w:tbl><w:tblPr><w:tblW w:w="2400" w:type="dxa"/><w:tblBorders><w:top w:val="single" w:sz="8" w:space="12" w:color="FF0000"/><w:bottom w:val="single" w:sz="8" w:space="13" w:color="00FF00"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="2400"/></w:tblGrid><w:tr><w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr><w:p/></w:tc></w:tr></w:tbl>`
	table, doc := reconstructTableFromXML(t, tableXML)
	borders := table.Borders()
	borders.Top.Width = 16
	if err := table.SetBorders(borders); err != nil {
		t.Fatalf("SetBorders: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	tree, err := parseXMLTree([]byte(written))
	if err != nil {
		t.Fatalf("parse resaved document.xml: %v", err)
	}
	bordersElement := findDescendant(tree, "tblBorders")
	top := findChild(bordersElement, "top")
	bottom := findChild(bordersElement, "bottom")
	if got, _ := getAttr(top, "sz"); got != "16" {
		t.Errorf("top border size = %q, want 16", got)
	}
	if got, _ := getAttr(top, "space"); got != "12" {
		t.Errorf("top border space = %q, want preserved 12", got)
	}
	if got, _ := getAttr(bottom, "space"); got != "13" {
		t.Errorf("bottom border space = %q, want preserved 13", got)
	}
}

func TestRoundTripMainDocument_NewRowAvoidsTableLocalPrefixCollision(t *testing.T) {
	tableXML := `<x:tbl xmlns:x="` + constants.NamespaceMain + `" xmlns:w="urn:not-wordprocessingml"><x:tblPr><x:tblW x:w="2400" x:type="dxa"/></x:tblPr><x:tblGrid><x:gridCol x:w="2400"/></x:tblGrid><x:tr><x:tc><x:tcPr><x:tcW x:w="2400" x:type="dxa"/></x:tcPr><x:p/></x:tc></x:tr></x:tbl>`
	table, doc := reconstructTableFromXML(t, tableXML)
	if _, err := table.AddRow(); err != nil {
		t.Fatalf("AddRow: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	if got := strings.Count(written, `<x:tr>`); got != 2 {
		t.Fatalf("x:tr count = %d, want 2:\n%s", got, written)
	}
	if strings.Contains(written, `<w:tr`) {
		t.Errorf("new row used the table-local conflicting w prefix:\n%s", written)
	}
}

func TestRoundTripMainDocument_EditedTablePropertiesUseTableLocalNamespace(t *testing.T) {
	tableXML := `<x:tbl xmlns:x="` + constants.NamespaceMain + `" xmlns:w="urn:not-wordprocessingml"><x:tblPr><x:tblW x:w="2400" x:type="dxa"/></x:tblPr><x:tblGrid><x:gridCol x:w="2400"/></x:tblGrid><x:tr><x:tc><x:tcPr><x:tcW x:w="2400" x:type="dxa"/></x:tcPr><x:p/></x:tc></x:tr></x:tbl>`
	table, doc := reconstructTableFromXML(t, tableXML)
	if err := table.SetWidth(domain.TableWidth{Type: domain.WidthDXA, Value: 3000}); err != nil {
		t.Fatalf("SetWidth: %v", err)
	}

	var output bytes.Buffer
	if _, err := doc.WriteTo(&output); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, output.Bytes())
	if _, err := parseXMLTree([]byte(written)); err != nil {
		t.Fatalf("parse resaved document.xml: %v\n%s", err, written)
	}
	if !strings.Contains(written, `<x:tblW x:w="3000"`) {
		t.Errorf("edited table width did not use the table-local Word namespace:\n%s", written)
	}
	if strings.Contains(written, `<w:tblW`) {
		t.Errorf("edited table width used the conflicting w prefix:\n%s", written)
	}
}

func TestReconstructTableMeasurements_AcceptsFloatFormattedValues(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblPr><w:tblW w:w="9360.0" w:type="dxa"/></w:tblPr>
<w:tblGrid><w:gridCol w:w="4680"/></w:tblGrid>
<w:tr>
<w:trPr><w:trHeight w:val="240.9"/></w:trPr>
<w:tc><w:tcPr><w:tcW w:w="4680.75" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc>
</w:tr>
</w:tbl>`

	table, doc := reconstructTableFromXML(t, tableXML)
	if got := table.Width(); got.Type != domain.WidthDXA || got.Value != 9360 {
		t.Fatalf("Width() = %+v, want {WidthDXA 9360}", got)
	}
	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	if got := row.Height(); got != 240 {
		t.Fatalf("Height() = %d, want 240", got)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if got := cell.Width(); got != 4680 {
		t.Fatalf("cell Width() = %d, want 4680", got)
	}
	// Change each hydrated measurement so the writer path emits canonical
	// integer lexical values. A pure round trip deliberately keeps the source
	// float spelling byte-for-byte.
	if err := table.SetWidth(domain.TableWidth{Type: domain.WidthDXA, Value: 9361}); err != nil {
		t.Fatalf("SetWidth: %v", err)
	}
	if err := table.SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	if err := row.SetHeight(241); err != nil {
		t.Fatalf("SetHeight: %v", err)
	}
	if err := cell.SetWidth(4681); err != nil {
		t.Fatalf("cell SetWidth: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	for _, want := range []struct {
		element string
		attrs   []string
	}{
		{element: "tblW", attrs: []string{`w:type="dxa"`, `w:w="9361"`}},
		{element: "trHeight", attrs: []string{`w:val="241"`, `w:hRule="atLeast"`}},
		{element: "tcW", attrs: []string{`w:type="dxa"`, `w:w="4681"`}},
	} {
		if !startTagHas(written, want.element, want.attrs...) {
			t.Errorf("resaved document.xml is missing w:%s with %v:\n%s", want.element, want.attrs, written)
		}
	}
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

	for _, want := range []struct {
		element string
		attrs   []string
	}{
		{element: "tblW", attrs: []string{`w:type="dxa"`, `w:w="5000"`}},
		{element: "jc", attrs: []string{`w:val="center"`}},
		{element: "tblBorders"},
		{element: "insideV", attrs: []string{`w:val="thick"`, `w:sz="18"`, `w:color="ABCDEF"`}},
		{element: "trHeight", attrs: []string{`w:val="567"`}},
		{element: "tcW", attrs: []string{`w:type="dxa"`, `w:w="918"`}},
		{element: "shd", attrs: []string{`w:val="clear"`, `w:fill="DDEEFF"`}},
		{element: "vAlign", attrs: []string{`w:val="center"`}},
		{element: "gridCol", attrs: []string{`w:w="918"`}},
		{element: "gridCol", attrs: []string{`w:w="1111"`}},
	} {
		if !startTagHas(written, want.element, want.attrs...) {
			t.Errorf("resaved document.xml is missing w:%s with %v:\n%s", want.element, want.attrs, written)
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

// TestReconstructTableCellShading_ThemedFillKeepsItsCachedColourAndTheLink
// pins a themed cell's round trip: both the resolved colour a
// theme-unaware consumer needs (w:fill) and the theme reference itself
// (w:themeFill, and its tint/shade) survive, because a producer writes the
// resolved colour into w:fill precisely so consumers that do not resolve
// themes still render correctly, while w:themeFill is what a theme-aware
// one actually resolves the colour from. Neither on its own is the whole
// picture, and a table is always rebuilt from the model on save, so
// declining to hydrate the link would not have preserved it -- it would
// have deleted the shading and returned the cell white.
//
// A themeFill with no w:fill alongside it still hydrates and round-trips the
// link, just with no cached colour riding along -- see
// TestReconstructTableCellShading_ThemeFillOnlyRoundTripsWithoutFabricatingFill.
func TestReconstructTableCellShading_ThemedFillKeepsItsCachedColourAndTheLink(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblGrid><w:gridCol/></w:tblGrid>
<w:tr><w:tc><w:tcPr><w:shd w:val="clear" w:fill="FF0000" w:themeFill="accent1" w:themeFillTint="80"/></w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc></w:tr>
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
	want := domain.Color{R: 0xFF, G: 0x00, B: 0x00}
	if got := cell.Shading(); got != want {
		t.Errorf("Cell(0).Shading() = %+v, want %+v (the cached w:fill, not the untouched default)", got, want)
	}

	themed, ok := cell.(interface {
		ThemeFill() (string, string, string)
	})
	if !ok {
		t.Fatal("cell does not expose ThemeFill")
	}
	if fill, tint, shade := themed.ThemeFill(); fill != "accent1" || tint != "80" || shade != "" {
		t.Errorf("ThemeFill() = (%q, %q, %q), want (\"accent1\", \"80\", \"\")", fill, tint, shade)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:themeFill="accent1"`) {
		t.Errorf("resaved document.xml lost the theme link:\n%s", written)
	}
	if !strings.Contains(written, `w:themeFillTint="80"`) {
		t.Errorf("resaved document.xml lost the theme tint:\n%s", written)
	}
	if !strings.Contains(written, `w:fill="FF0000"`) {
		t.Errorf("resaved document.xml lost the cached fallback colour:\n%s", written)
	}
}

// TestReconstructTableCellShading_ThemeFillOnlyRoundTripsWithoutFabricatingFill
// covers a producer that writes only w:themeFill, with no cached w:fill
// alongside it -- MS-OE376 makes the theme reference the primary colour and
// the cached fallback optional (used only when the reference is absent), so
// this is valid input, not a malformed one. It used to be dropped entirely:
// applyCellShading returned before ever reading w:themeFill because its
// "is there a concrete colour" guard ran first. The regression to watch for
// on the fix's other side is fabricating a fallback that was never there --
// Shading() defaults to ColorWhite for an untouched cell, and blindly
// emitting that as w:fill would claim the source said "white" when it
// said nothing about a fallback colour at all.
func TestReconstructTableCellShading_ThemeFillOnlyRoundTripsWithoutFabricatingFill(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblGrid><w:gridCol/></w:tblGrid>
<w:tr><w:tc><w:tcPr><w:shd w:val="clear" w:color="auto" w:themeFill="accent1" w:themeFillTint="80"/></w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc></w:tr>
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
	if got := cell.Shading(); got != domain.ColorWhite {
		t.Errorf("Cell(0).Shading() = %+v, want the untouched default %+v (no cached colour was ever present to hydrate)", got, domain.ColorWhite)
	}

	themed, ok := cell.(interface {
		ThemeFill() (string, string, string)
	})
	if !ok {
		t.Fatal("cell does not expose ThemeFill")
	}
	if fill, tint, shade := themed.ThemeFill(); fill != "accent1" || tint != "80" || shade != "" {
		t.Errorf("ThemeFill() = (%q, %q, %q), want (\"accent1\", \"80\", \"\")", fill, tint, shade)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:themeFill="accent1"`) {
		t.Errorf("resaved document.xml lost the theme-only link:\n%s", written)
	}
	if !strings.Contains(written, `w:themeFillTint="80"`) {
		t.Errorf("resaved document.xml lost the theme tint:\n%s", written)
	}
	if strings.Contains(written, `w:fill=`) {
		t.Errorf("resaved document.xml fabricated a w:fill that was never in the source:\n%s", written)
	}
}

// TestReconstructTableCellShading_SolidThemeColourNormalizesToThemeFill pins
// the same normalization the plain-colour case already has (see
// TestReconstructTableCellShading_ColourSourceDependsOnThePattern): a
// w:val="solid" shading's visible colour is w:color, so its theme link is
// w:themeColor -- but the domain caches only one colour and one link
// regardless of pattern, and the writer always re-emits as clear+w:fill.
// clear+w:themeFill and solid+w:themeColor resolve to the same visible
// colour (clear draws no pattern, so w:fill is fully visible; solid's
// foreground covers the background completely), so collapsing a source
// w:themeColor onto the cell's single w:themeFill slot is exact, not lossy.
func TestReconstructTableCellShading_SolidThemeColourNormalizesToThemeFill(t *testing.T) {
	tableXML := `<w:tbl>
<w:tblGrid><w:gridCol/></w:tblGrid>
<w:tr><w:tc><w:tcPr><w:shd w:val="solid" w:color="0000FF" w:themeColor="accent2" w:themeShade="40"/></w:tcPr><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc></w:tr>
</w:tbl>`

	table, doc := reconstructTableFromXML(t, tableXML)
	row, _ := table.Row(0)
	cell, _ := row.Cell(0)

	themed, ok := cell.(interface {
		ThemeFill() (string, string, string)
	})
	if !ok {
		t.Fatal("cell does not expose ThemeFill")
	}
	if fill, tint, shade := themed.ThemeFill(); fill != "accent2" || tint != "" || shade != "40" {
		t.Errorf("ThemeFill() = (%q, %q, %q), want (\"accent2\", \"\", \"40\")", fill, tint, shade)
	}
	// Regenerate this row so the assertion below exercises the modeled
	// normalization path rather than unchanged raw-fragment preservation.
	if err := cell.SetVerticalAlignment(domain.VerticalAlignCenter); err != nil {
		t.Fatalf("SetVerticalAlignment: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	written := documentXML(t, buf.Bytes())
	if !strings.Contains(written, `w:themeFill="accent2"`) {
		t.Errorf("resaved document.xml did not normalize w:themeColor onto w:themeFill:\n%s", written)
	}
	if !strings.Contains(written, `w:themeFillShade="40"`) {
		t.Errorf("resaved document.xml lost the theme shade:\n%s", written)
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
		// A solid pattern takes its colour from w:color, so an "auto" one is
		// exactly as unhydratable as an "auto" w:fill is under "clear" --
		// even though this shd does carry a concrete w:fill, which a solid
		// pattern covers completely.
		{"solid with auto colour", `<w:shd w:val="solid" w:color="auto" w:fill="DDEEFF"/>`},
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

// TestReconstructTableCellShading_ColourSourceDependsOnThePattern pins which
// attribute the visible colour comes from. w:shd paints a w:val pattern in
// w:color over a w:fill background: "clear" draws no pattern so w:fill shows,
// while "solid" is a 100% foreground fill that hides w:fill entirely so
// w:color shows. Both cases below carry *both* attributes with different
// colours, so reading the wrong one is a wrong colour rather than no colour --
// which is why a test that only ever set one attribute could not catch it.
func TestReconstructTableCellShading_ColourSourceDependsOnThePattern(t *testing.T) {
	cases := []struct {
		name string
		shd  string
		want domain.Color
	}{
		{
			name: "clear takes the fill",
			shd:  `<w:shd w:val="clear" w:color="FF0000" w:fill="0000FF"/>`,
			want: domain.Color{R: 0x00, G: 0x00, B: 0xFF},
		},
		{
			name: "solid takes the colour",
			shd:  `<w:shd w:val="solid" w:color="FF0000" w:fill="0000FF"/>`,
			want: domain.Color{R: 0xFF, G: 0x00, B: 0x00},
		},
		{
			// No w:val at all defaults to "clear" per ST_Shd.
			name: "absent val takes the fill",
			shd:  `<w:shd w:fill="00FF00"/>`,
			want: domain.Color{R: 0x00, G: 0xFF, B: 0x00},
		},
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
			if got := cell.Shading(); got != tc.want {
				t.Errorf("Cell(0).Shading() = %+v, want %+v", got, tc.want)
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

	if startTagHas(written, "gridCol", `w:w="2000"`) {
		t.Errorf("resaved document.xml split the merged title cell into two equal columns:\n%s", written)
	}
	for _, width := range []string{"900", "3100"} {
		if !startTagHas(written, "gridCol", `w:w="`+width+`"`) {
			t.Errorf("resaved document.xml is missing gridCol width %s:\n%s", width, written)
		}
	}
}
