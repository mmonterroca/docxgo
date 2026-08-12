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

	err := zw.WriteDocument(doc, rels, nil, nil, xmlstructs.NewStyles(), nil, nil, nil, nil, nil, nil, nil)
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

	zw.WriteDocument(doc, rels, nil, nil, xmlstructs.NewStyles(), nil, nil, nil, nil, nil, nil, nil)
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

	zw.WriteDocument(doc, rels, nil, nil, xmlstructs.NewStyles(), nil, nil, nil, nil, nil, nil, nil)
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

// TestZipWriter_WriteStyles_NilStylesIsError pins that writeStyles requires a
// non-nil Styles rather than silently falling back to a hand-maintained
// default (the fallback was unreachable from any production caller and had
// already drifted from the serializer's own defaults; see #70).
func TestZipWriter_WriteStyles_NilStylesIsError(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)

	if err := zw.writeStyles(nil); err == nil {
		t.Fatal("expected an error for nil styles, got nil")
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

	withoutLang := ser.SerializeStyles(sm, nil, nil, nil)
	if withoutLang.DocDefaults == nil || withoutLang.DocDefaults.ParaDefaults == nil {
		t.Fatal("expected ParaDefaults to be set even without a language")
	}

	withLang := ser.SerializeStyles(sm, &domain.Language{Val: "en-US"}, nil, nil)
	if withLang.DocDefaults == nil || withLang.DocDefaults.ParaDefaults == nil {
		t.Fatal("expected ParaDefaults to be set alongside a language")
	}
	if withLang.DocDefaults.RunDefaults == nil {
		t.Error("expected RunDefaults (lang) to still be set when a language is provided")
	}
}
