package writer

/*
   Copyright (c) 2025 Misael Monterroca

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"testing"

	xmlstructs "github.com/mmonterroca/docxgo/internal/xml"
	"github.com/mmonterroca/docxgo/pkg/constants"
)

func TestZipWriter_WriteDocument(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)

	// Create minimal document
	doc := &xmlstructs.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body: &xmlstructs.Body{
			Paragraphs: []*xmlstructs.Paragraph{
				{
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

	err := zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil)
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

	zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil)
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

			if len(ct.Overrides) != 6 {
				t.Errorf("Wrong number of overrides: got %d, want 6", len(ct.Overrides))
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
			Paragraphs: []*xmlstructs.Paragraph{
				{
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

	zw.WriteDocument(doc, rels, nil, nil, nil, nil, nil, nil)
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
