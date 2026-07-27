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
	"reflect"
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

// pPrDefaultProperties returns every element and attribute inside the
// <w:pPrDefault> subtree of a word/styles.xml document, keyed as
// "element/attribute" (e.g. "spacing/before"). Decoding the subtree rather
// than comparing raw bytes means the two producers are compared on the
// properties they describe, not on incidental encoding differences
// (self-closing vs paired tags, attribute order, indentation). Any property
// present in one producer and absent in the other shows up as a key
// difference, so this catches drift the test was written to prevent even for
// defaults nobody thought to enumerate.
func pPrDefaultProperties(t *testing.T, stylesXML []byte) map[string]string {
	t.Helper()

	start := bytes.Index(stylesXML, []byte("<w:pPrDefault"))
	if start < 0 {
		t.Fatal("no <w:pPrDefault> element found in styles.xml")
	}
	end := bytes.Index(stylesXML[start:], []byte("</w:pPrDefault>"))
	if end < 0 {
		t.Fatal("no closing </w:pPrDefault> found in styles.xml")
	}
	fragment := stylesXML[start : start+end+len("</w:pPrDefault>")]

	props := make(map[string]string)
	dec := xml.NewDecoder(bytes.NewReader(fragment))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		// Skip the wrapper elements themselves; record their contents.
		if se.Name.Local == "pPrDefault" || se.Name.Local == "pPr" {
			continue
		}
		if len(se.Attr) == 0 {
			// A valueless element (e.g. <w:keepNext/>) is itself a property.
			props[se.Name.Local] = ""
			continue
		}
		for _, a := range se.Attr {
			props[se.Name.Local+"/"+a.Name.Local] = a.Value
		}
	}

	if len(props) == 0 {
		t.Fatalf("no properties decoded from w:pPrDefault subtree: %s", fragment)
	}
	return props
}

// TestWriteDefaultStyles_MatchesSerializedStyleManagerDefaults pins that the
// w:pPrDefault written by the no-style-manager fallback (writeDefaultStyles'
// hand-written XML string) and the one written for a document that does carry
// a style manager (DocumentSerializer.SerializeStyles' marshaled structs)
// describe the same default paragraph properties. If they drifted, a document
// would render with different default spacing purely depending on whether a
// style manager happened to be attached.
//
// Both sides go through the real ZipWriter.WriteDocument and are read back out
// of the resulting archive, so neither is a hand-rolled imitation that could
// keep passing while the production paths diverge. The whole w:pPrDefault
// subtree is compared, not a chosen list of attributes, so a default added to
// one producer and not the other fails here rather than shipping.
func TestWriteDefaultStyles_MatchesSerializedStyleManagerDefaults(t *testing.T) {
	writeStylesXML := func(t *testing.T, styles *xmlstructs.Styles) []byte {
		t.Helper()
		var buf bytes.Buffer
		zw := NewZipWriter(&buf)
		doc := &xmlstructs.Document{
			XMLnsW: constants.NamespaceMain,
			XMLnsR: constants.NamespaceRelationships,
			Body:   &xmlstructs.Body{},
		}
		rels := &xmlstructs.Relationships{Xmlns: constants.NamespacePackageRels}
		if err := zw.WriteDocument(doc, rels, nil, nil, styles, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("WriteDocument: %v", err)
		}
		if err := zw.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
		return readZipFile(t, buf.Bytes(), "word/styles.xml")
	}

	// Path A: nil styles -> writeDefaultStyles' raw XML fallback.
	stylesA := writeStylesXML(t, nil)

	// Path B: a real style manager -> SerializeStyles' marshaled structs.
	sm := manager.NewStyleManager()
	stylesB := writeStylesXML(t, serializer.NewDocumentSerializer().SerializeStyles(sm, nil))

	defaultsA := pPrDefaultProperties(t, stylesA)
	defaultsB := pPrDefaultProperties(t, stylesB)

	if !reflect.DeepEqual(defaultsA, defaultsB) {
		t.Errorf("w:pPrDefault differs between the two styles.xml producers:\n  no-style-manager  = %+v\n  with-style-manager= %+v",
			defaultsA, defaultsB)
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
