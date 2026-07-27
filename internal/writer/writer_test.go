// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package writer

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"regexp"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/internal/serializer"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestZipWriter_WriteDocument(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)

	// Create minimal document
	doc := &xmlstructs.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body: &xmlstructs.Body{
			Content: []interface{}{
				&xmlstructs.Paragraph{
					Elements: []interface{}{
						&xmlstructs.Run{
							Text: &xmlstructs.Text{Content: "Hello, World!"},
						},
					},
				},
			},
		},
	}

	rels := &xmlstructs.Relationships{
		Xmlns:         constants.NamespacePackageRels,
		Relationships: []*xmlstructs.Relationship{},
	}

	err := zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("WriteDocument failed: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify ZIP structure
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	expectedFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"docProps/core.xml",
		"docProps/app.xml",
		"word/styles.xml",
		"word/fontTable.xml",
		"word/theme/theme1.xml",
		"word/settings.xml",
		"word/webSettings.xml",
	}

	fileMap := make(map[string]bool)
	for _, f := range zipReader.File {
		fileMap[f.Name] = true
	}

	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("Expected file %s not found in ZIP", expected)
		}
	}
}

func TestZipWriter_ContentTypes(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)

	doc := &xmlstructs.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body:   &xmlstructs.Body{},
	}

	rels := &xmlstructs.Relationships{
		Xmlns:         constants.NamespacePackageRels,
		Relationships: []*xmlstructs.Relationship{},
	}

	zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	zw.Close()

	// Read and verify [Content_Types].xml
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	for _, f := range zipReader.File {
		if f.Name == "[Content_Types].xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open Content_Types: %v", err)
			}
			defer rc.Close()

			var ct xmlstructs.ContentTypes
			if err := xml.NewDecoder(rc).Decode(&ct); err != nil {
				t.Fatalf("Failed to decode Content_Types: %v", err)
			}

			if ct.Xmlns != constants.NamespaceContentTypes {
				t.Errorf("Wrong namespace: got %s, want %s", ct.Xmlns, constants.NamespaceContentTypes)
			}

			if len(ct.Defaults) != 2 {
				t.Errorf("Wrong number of defaults: got %d, want 2", len(ct.Defaults))
			}

			if len(ct.Overrides) != 8 {
				t.Errorf("Wrong number of overrides: got %d, want 8", len(ct.Overrides))
			}

			return
		}
	}

	t.Error("[Content_Types].xml not found in ZIP")
}

func TestZipWriter_DocumentXML(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)

	doc := &xmlstructs.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body: &xmlstructs.Body{
			Content: []interface{}{
				&xmlstructs.Paragraph{
					Elements: []interface{}{
						&xmlstructs.Run{
							Text: &xmlstructs.Text{Content: "Test paragraph"},
						},
					},
				},
			},
		},
	}

	rels := &xmlstructs.Relationships{
		Xmlns:         constants.NamespacePackageRels,
		Relationships: []*xmlstructs.Relationship{},
	}

	zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	zw.Close()

	// Read and verify word/document.xml exists
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	found := false
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			found = true

			// Verify file has content
			if f.UncompressedSize64 == 0 {
				t.Error("document.xml is empty")
			}

			break
		}
	}

	if !found {
		t.Error("word/document.xml not found in ZIP")
	}
}

// readZipFile returns the bytes of the named entry in a .docx zip buffer.
func readZipFile(t *testing.T, buf []byte, name string) []byte {
	t.Helper()
	zipReader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}
	for _, f := range zipReader.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer rc.Close()
		var out bytes.Buffer
		if _, err := out.ReadFrom(rc); err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return out.Bytes()
	}
	t.Fatalf("%s not found in ZIP", name)
	return nil
}

// pPrDefaultAttr extracts one attribute (by local name, ignoring the "w:"
// prefix) from the <w:pPrDefault>...<w:spacing .../> subtree of a
// word/styles.xml document. Matching on the raw attribute text rather than
// decoding through encoding/xml sidesteps namespace-prefix handling, which
// isn't the point of this test: both code paths must agree on the actual
// spacing values, not on identical XML encoding.
func pPrDefaultAttr(t *testing.T, stylesXML []byte, attr string) (string, bool) {
	t.Helper()
	start := bytes.Index(stylesXML, []byte("<w:pPrDefault"))
	if start < 0 {
		t.Fatal("no <w:pPrDefault> element found in styles.xml")
	}
	end := bytes.Index(stylesXML[start:], []byte("</w:pPrDefault>"))
	if end < 0 {
		t.Fatal("no closing </w:pPrDefault> found in styles.xml")
	}
	fragment := stylesXML[start : start+end]

	re := regexp.MustCompile(`w:` + attr + `="([^"]*)"`)
	m := re.FindSubmatch(fragment)
	if m == nil {
		return "", false
	}
	return string(m[1]), true
}

// TestWriteDefaultStyles_MatchesSerializedStyleManagerDefaults pins that the
// w:pPrDefault emitted by the no-style-manager fallback (writeDefaultStyles,
// hand-written XML string) and the one emitted for a document that does
// carry a style manager (DocumentSerializer.SerializeStyles, marshaled
// structs) agree on the actual spacing values. If they drifted, a document
// would render with different default paragraph spacing purely depending on
// whether a style manager happened to be attached.
func TestWriteDefaultStyles_MatchesSerializedStyleManagerDefaults(t *testing.T) {
	// Path A: no style manager -> writeDefaultStyles' raw XML fallback.
	var bufA bytes.Buffer
	zwA := NewZipWriter(&bufA)
	doc := &xmlstructs.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body:   &xmlstructs.Body{},
	}
	rels := &xmlstructs.Relationships{Xmlns: constants.NamespacePackageRels}
	if err := zwA.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Fatalf("WriteDocument (no styles): %v", err)
	}
	if err := zwA.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	stylesA := readZipFile(t, bufA.Bytes(), "word/styles.xml")

	// Path B: with a style manager -> SerializeStyles' marshaled structs,
	// written through the exact same encoder writeStyles uses.
	sm := manager.NewStyleManager()
	xmlStyles := serializer.NewDocumentSerializer().SerializeStyles(sm, nil)

	var bufB bytes.Buffer
	if _, err := bufB.WriteString(xml.Header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	enc := xml.NewEncoder(&bufB)
	enc.Indent("", "  ")
	if err := enc.Encode(xmlStyles); err != nil {
		t.Fatalf("encode styles: %v", err)
	}
	stylesB := bufB.Bytes()

	for _, attr := range []string{"before", "after", "line", "lineRule"} {
		valA, okA := pPrDefaultAttr(t, stylesA, attr)
		valB, okB := pPrDefaultAttr(t, stylesB, attr)
		if okA != okB || valA != valB {
			t.Errorf("w:%s mismatch between paths: no-style-manager=%q(present=%v) vs with-style-manager=%q(present=%v)",
				attr, valA, okA, valB, okB)
		}
	}
}

// TestSerializeStyles_PPrDefaultPresentRegardlessOfLang pins that
// SerializeStyles emits w:pPrDefault whether or not a language is set.
// Before this fix, DocDefaults (and therefore ParaDefaults) was only built
// when lang != nil, so any document without an explicit language silently
// lost the 0/0/240 paragraph spacing default.
func TestSerializeStyles_PPrDefaultPresentRegardlessOfLang(t *testing.T) {
	sm := manager.NewStyleManager()
	ser := serializer.NewDocumentSerializer()

	withoutLang := ser.SerializeStyles(sm, nil)
	if withoutLang.DocDefaults == nil || withoutLang.DocDefaults.ParaDefaults == nil {
		t.Fatal("expected ParaDefaults to be set even without a language")
	}

	withLang := ser.SerializeStyles(sm, &domain.Language{Val: "en-US"})
	if withLang.DocDefaults == nil || withLang.DocDefaults.ParaDefaults == nil {
		t.Fatal("expected ParaDefaults to be set alongside a language")
	}
	if withLang.DocDefaults.RunDefaults == nil {
		t.Error("expected RunDefaults (lang) to still be set when a language is provided")
	}
}
