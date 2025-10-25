/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)

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

package docx

import (
	"bytes"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	// Create a simple docx file first
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Hello World")

	// Write to buffer
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write docx: %v", err)
	}

	// Parse from buffer
	parsedDoc, err := Parse(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to parse docx: %v", err)
	}

	if parsedDoc == nil {
		t.Fatal("Parsed document should not be nil")
	}
}

func TestWriteTo(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Test content")

	var buf bytes.Buffer
	n, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if n != 0 {
		// WriteTo returns 0 per implementation, but should not error
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty buffer")
	}
}

func TestRead(t *testing.T) {
	doc := New().WithDefaultTheme()
	b := make([]byte, 10)
	n, err := doc.Read(b)

	if n != 0 {
		t.Errorf("Expected Read to return 0, got %d", n)
	}
	if err != os.ErrInvalid {
		t.Errorf("Expected ErrInvalid, got %v", err)
	}
}

func TestLoadBodyItems(t *testing.T) {
	items := []interface{}{
		&Paragraph{
			Children: []interface{}{
				&Run{
					Children: []interface{}{
						&Text{Text: "Test"},
					},
				},
			},
		},
	}

	media := []Media{
		{Name: "image1.jpg", Data: []byte("fake image data")},
	}

	doc := LoadBodyItems(items, media)

	if doc == nil {
		t.Fatal("LoadBodyItems returned nil")
	}
	if len(doc.Document.Body.Items) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(doc.Document.Body.Items))
	}
	if len(doc.media) != len(media) {
		t.Errorf("Expected %d media, got %d", len(media), len(doc.media))
	}
	if len(doc.mediaNameIdx) != len(media) {
		t.Errorf("Expected %d mediaNameIdx entries, got %d", len(media), len(doc.mediaNameIdx))
	}
}
