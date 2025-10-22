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

func TestPackAndUnpack(t *testing.T) {
	// Create a document with various elements
	doc := New().WithDefaultTheme()
	
	// Add paragraphs with various styles
	para1 := doc.AddParagraph()
	para1.AddText("Hello World").Bold().Size("24")
	
	para2 := doc.AddParagraph()
	para2.AddText("This is a test document")
	para2.AddLink("Google", "https://www.google.com")
	
	// Add a table
	table := doc.AddTable(2, 2, 5000, nil)
	table.TableRows[0].TableCells[0].AddParagraph().AddText("Cell 1")
	table.TableRows[0].TableCells[1].AddParagraph().AddText("Cell 2")
	table.TableRows[1].TableCells[0].AddParagraph().AddText("Cell 3")
	table.TableRows[1].TableCells[1].AddParagraph().AddText("Cell 4")
	
	// Add an image
	para3 := doc.AddParagraph()
	_, err := para3.AddInlineDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add image: %v", err)
	}
	
	// Write to buffer
	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write docx: %v", err)
	}
	
	if buf.Len() == 0 {
		t.Fatal("Written buffer is empty")
	}
	
	// Parse the document back
	parsedDoc, err := Parse(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to parse docx: %v", err)
	}
	
	if parsedDoc == nil {
		t.Fatal("Parsed document is nil")
	}
	
	// Verify structure
	if len(parsedDoc.Document.Body.Items) == 0 {
		t.Error("Expected non-empty body items")
	}
	
	// Test writing to file
	tmpFile, err := os.CreateTemp("", "test-docx-*.docx")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	_, err = doc.WriteTo(tmpFile)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
}

func TestParseWithMedia(t *testing.T) {
	// Create a document with image
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	_, err := para.AddInlineDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add image: %v", err)
	}
	
	// Write and parse back
	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write docx: %v", err)
	}
	
	parsedDoc, err := Parse(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to parse docx: %v", err)
	}
	
	// Check media was parsed
	if len(parsedDoc.media) == 0 {
		t.Error("Expected media to be parsed")
	}
}

func TestPackRead(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	// Test Read function (should return error)
	buf := make([]byte, 10)
	n, err := doc.Read(buf)
	if err != os.ErrInvalid {
		t.Errorf("Expected ErrInvalid, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected n=0, got %d", n)
	}
}
