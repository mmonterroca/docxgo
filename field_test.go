/*
   Copyright (c) 2025 SlideLang Enhanced Fork

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
	"os"
	"strings"
	"testing"
)

func TestFldCharBasic(t *testing.T) {
	// Test FldChar creation
	fldChar := NewFldChar("begin")
	if fldChar.FldCharType != "begin" {
		t.Errorf("Expected fldCharType 'begin', got '%s'", fldChar.FldCharType)
	}

	// Test adding FldChar to Run
	w := New()
	run := &Run{file: w, RunProperties: &RunProperties{}}

	fldCharAdded := run.AddFldChar("separate")
	if fldCharAdded.FldCharType != "separate" {
		t.Errorf("Expected fldCharType 'separate', got '%s'", fldCharAdded.FldCharType)
	}
}

func TestInstrTextBasic(t *testing.T) {
	// Test InstrText creation
	instrText := NewInstrText("TOC \\o \"1-3\"")
	if instrText.Text != "TOC \\o \"1-3\"" {
		t.Errorf("Expected instrText 'TOC \\o \"1-3\"', got '%s'", instrText.Text)
	}

	// Test adding InstrText to Run
	w := New()
	run := &Run{file: w, RunProperties: &RunProperties{}}

	instrTextAdded := run.AddInstrText("PAGEREF _Toc123")
	if instrTextAdded.Text != "PAGEREF _Toc123" {
		t.Errorf("Expected instrText 'PAGEREF _Toc123', got '%s'", instrTextAdded.Text)
	}
}

func TestBasicField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	// Add a basic field
	field := p.AddField("PAGE", "1")

	// Verify field structure
	if field == nil {
		t.Fatal("Field was not created")
	}

	// Verify all components exist
	if field.Begin == nil || field.InstrText == nil || field.Separate == nil || field.Result == nil || field.End == nil {
		t.Error("Field structure is incomplete")
	}

	// Verify the paragraph has 5 runs (begin, instrText, separate, result, end)
	if len(p.Children) != 5 {
		t.Errorf("Expected 5 runs in paragraph, got %d", len(p.Children))
	}
}

func TestTOCField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	// Add TOC field with various options
	tocField := p.AddTOCField(3, true, true)

	if tocField == nil {
		t.Fatal("TOC field was not created")
	}

	// Check that instrText contains expected TOC instruction
	// Find the instrText run
	var instrTextContent string
	for _, child := range p.Children {
		if run, ok := child.(*Run); ok {
			for _, runChild := range run.Children {
				if instrText, ok := runChild.(*InstrText); ok {
					instrTextContent = instrText.Text
					break
				}
			}
		}
	}

	expectedInstr := "TOC \\o \"1-3\" \\h \\z \\u"
	if instrTextContent != expectedInstr {
		t.Errorf("Expected TOC instruction '%s', got '%s'", expectedInstr, instrTextContent)
	}
}

func TestPageRefField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	// Add PAGEREF field
	pageRefField := p.AddPageRefField("_Toc123456789", true)

	if pageRefField == nil {
		t.Fatal("PAGEREF field was not created")
	}

	// Check instrText content
	var instrTextContent string
	for _, child := range p.Children {
		if run, ok := child.(*Run); ok {
			for _, runChild := range run.Children {
				if instrText, ok := runChild.(*InstrText); ok {
					instrTextContent = instrText.Text
					break
				}
			}
		}
	}

	expectedInstr := "PAGEREF _Toc123456789 \\h"
	if instrTextContent != expectedInstr {
		t.Errorf("Expected PAGEREF instruction '%s', got '%s'", expectedInstr, instrTextContent)
	}
}

func TestPageField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	pageField := p.AddPageField()
	if pageField == nil {
		t.Fatal("PAGE field was not created")
	}

	// Verify instruction text
	var instrTextContent string
	for _, child := range p.Children {
		if run, ok := child.(*Run); ok {
			for _, runChild := range run.Children {
				if instrText, ok := runChild.(*InstrText); ok {
					instrTextContent = instrText.Text
					break
				}
			}
		}
	}

	if instrTextContent != "PAGE" {
		t.Errorf("Expected PAGE instruction, got '%s'", instrTextContent)
	}
}

func TestNumPagesField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	numPagesField := p.AddNumPagesField()
	if numPagesField == nil {
		t.Fatal("NUMPAGES field was not created")
	}
}

func TestRefField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	refField := p.AddRefField("intro_bookmark", false)
	if refField == nil {
		t.Fatal("REF field was not created")
	}

	// Check instruction content
	var instrTextContent string
	for _, child := range p.Children {
		if run, ok := child.(*Run); ok {
			for _, runChild := range run.Children {
				if instrText, ok := runChild.(*InstrText); ok {
					instrTextContent = instrText.Text
					break
				}
			}
		}
	}

	expectedInstr := "REF intro_bookmark"
	if instrTextContent != expectedInstr {
		t.Errorf("Expected REF instruction '%s', got '%s'", expectedInstr, instrTextContent)
	}
}

func TestSeqField(t *testing.T) {
	w := New()
	p := w.AddParagraph()

	seqField := p.AddSeqField("figure", "ARABIC")
	if seqField == nil {
		t.Fatal("SEQ field was not created")
	}

	// Check instruction content
	var instrTextContent string
	for _, child := range p.Children {
		if run, ok := child.(*Run); ok {
			for _, runChild := range run.Children {
				if instrText, ok := runChild.(*InstrText); ok {
					instrTextContent = instrText.Text
					break
				}
			}
		}
	}

	expectedInstr := "SEQ figure \\* ARABIC"
	if instrTextContent != expectedInstr {
		t.Errorf("Expected SEQ instruction '%s', got '%s'", expectedInstr, instrTextContent)
	}
}

func TestComplexDocumentWithFields(t *testing.T) {
	w := New()

	// Add TOC
	tocPara := w.AddParagraph()
	tocPara.AddText("Table of Contents").Bold()
	tocPara.AddTOCField(3, true, true)

	// Add some headings with bookmarks and references
	h1 := w.AddParagraph()
	h1.AddText("1. Introduction")
	h1.AddTOCBookmark("1. Introduction", 1)

	p1 := w.AddParagraph()
	p1.AddText("This is the introduction. See section ")
	p1.AddRefField("_Toc000000002", true)
	p1.AddText(" for more details.")

	h2 := w.AddParagraph()
	h2.AddText("2. Background")
	h2.AddTOCBookmark("2. Background", 2)

	p2 := w.AddParagraph()
	p2.AddText("This is on page ")
	p2.AddPageField()
	p2.AddText(" of ")
	p2.AddNumPagesField()
	p2.AddText(".")

	// Add figure caption
	figPara := w.AddParagraph()
	figPara.AddText("Figure ")
	figPara.AddSeqField("Figure", "ARABIC")
	figPara.AddText(": Sample diagram")

	// Save document
	f, err := os.Create("test_fields.docx")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()
	defer os.Remove("test_fields.docx") // cleanup

	_, err = w.WriteTo(f)
	if err != nil {
		t.Fatalf("Failed to write document: %v", err)
	}

	// Basic validation - check that paragraph contains expected text
	paraString := p1.String()
	if !strings.Contains(paraString, "This is the introduction") {
		t.Error("Paragraph should contain introduction text")
	}
}
