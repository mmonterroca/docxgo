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
	"regexp"
	"testing"
)

func TestBodyKeepElements(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	// Add some elements
	doc.AddParagraph().AddText("Paragraph 1")
	doc.AddTable(2, 2, 5000, nil)
	doc.AddParagraph().AddText("Paragraph 2")
	doc.AddTable(1, 1, 1000, nil)
	
	initialCount := len(doc.Document.Body.Items)
	if initialCount < 4 {
		t.Fatalf("Expected at least 4 items, got %d", initialCount)
	}
	
	// Keep only paragraphs
	doc.Document.Body.KeepElements("*docx.Paragraph")
	
	finalCount := len(doc.Document.Body.Items)
	if finalCount >= initialCount {
		t.Error("Expected fewer items after KeepElements")
	}
	
	// All remaining items should be paragraphs
	for _, item := range doc.Document.Body.Items {
		if _, ok := item.(*Paragraph); !ok {
			t.Error("Expected only Paragraph items")
			break
		}
	}
}

func TestBodyDropDrawingOf(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Before image")
	
	_, err := para.AddInlineDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add drawing: %v", err)
	}
	
	para.AddText("After image")
	
	// Drop nil pictures
	doc.Document.Body.DropDrawingOf("NilPicture")
	
	// The function should execute without error
	// We just verify it can be called
}

func TestSplitDocxByPlainTextRegex(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	// Add some paragraphs with a separator pattern
	doc.AddParagraph().AddText("Section 1 content")
	doc.AddParagraph().AddText("--- SEPARATOR ---")
	doc.AddParagraph().AddText("Section 2 content")
	doc.AddParagraph().AddText("--- SEPARATOR ---")
	doc.AddParagraph().AddText("Section 3 content")
	
	// Split by separator
	re := regexp.MustCompile("--- SEPARATOR ---")
	rule := SplitDocxByPlainTextRegex(re)
	
	docs := doc.SplitByParagraph(rule)
	
	if len(docs) == 0 {
		t.Error("Expected at least one document from split")
	}
}

func TestParagraphDropCanvas(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Test content")
	
	// Call drop methods
	para.DropCanvas()
	para.DropShape()
	para.DropGroup()
	para.DropShapeAndCanvas()
	para.DropShapeAndCanvasAndGroup()
	para.DropNilPicture()
	
	// Just verify they can be called without panic
}

func TestParagraphKeepElements(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Test")
	
	// Try to add various children
	para.AddTab()
	
	// Keep only certain elements
	para.KeepElements("*docx.Run", "*docx.Text")
	
	// Verify it executes without error
}

func TestRunKeepElements(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("Test")
	
	// Keep text elements
	run.KeepElements("*docx.Text")
	
	// Verify it executes without error
}

func TestAppendFile(t *testing.T) {
	doc1 := New().WithDefaultTheme()
	doc1.AddParagraph().AddText("Document 1")
	
	doc2 := New().WithDefaultTheme()
	doc2.AddParagraph().AddText("Document 2")
	
	initialLen := len(doc1.Document.Body.Items)
	
	// Append doc2 to doc1
	doc1.AppendFile(doc2)
	
	finalLen := len(doc1.Document.Body.Items)
	if finalLen <= initialLen {
		t.Error("Expected more items after AppendFile")
	}
}
