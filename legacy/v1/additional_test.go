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
	"testing"
)

func TestComplexDocument(t *testing.T) {
	// Create a complex document to increase code coverage
	doc := New().WithDefaultTheme()
	
	// Add paragraph with numbered list
	para1 := doc.AddParagraph()
	para1.NumPr("1", "0").NumFont("Arial", "SimSun", "Arial", "eastAsia").NumSize("20")
	para1.AddText("First item")
	
	// Add paragraph with style
	para2 := doc.AddParagraph().Style("Heading1")
	run := para2.AddText("Heading Text")
	run.Bold().Italic().Underline("single").Highlight("yellow")
	run.Color("FF0000").Size("28").SizeCs("28")
	run.Spacing(100).Strike(true)
	run.Font("Times New Roman", "SimSun", "Times New Roman", "default")
	
	// Add tab
	run.AddTab()
	para2.AddText("After tab")
	
	// Add page break
	para3 := doc.AddParagraph()
	para3.AddPageBreaks()
	
	// Add table with more complex structure
	borderColors := &APITableBorderColors{
		Top:     "FF0000",
		Left:    "00FF00",
		Bottom:  "0000FF",
		Right:   "FFFF00",
		InsideH: "FF00FF",
		InsideV: "00FFFF",
	}
	
	table := doc.AddTable(3, 3, 8000, borderColors)
	table.Justification("center")
	
	// Fill table with content
	for i, row := range table.TableRows {
		row.Justification("center")
		for j, cell := range row.TableCells {
			p := cell.AddParagraph()
			p.AddText("Row " + string(rune('0'+i)) + " Col " + string(rune('0'+j)))
			
			// Add shade to some cells
			if (i+j)%2 == 0 {
				cell.Shade("clear", "auto", "E7E6E6")
			}
		}
	}
	
	// Test WriteTo
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write document: %v", err)
	}
	
	if buf.Len() == 0 {
		t.Fatal("Document is empty")
	}
}

func TestDrawingOperations(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	// Test inline drawing
	para1 := doc.AddParagraph()
	run1, err := para1.AddInlineDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add inline drawing: %v", err)
	}
	
	// Test inline drawing size
	if drawing, ok := run1.Children[0].(*Drawing); ok && drawing.Inline != nil {
		drawing.Inline.Size(2000000, 2000000)
	}
	
	// Test anchor drawing
	para2 := doc.AddParagraph()
	run2, err := para2.AddAnchorDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add anchor drawing: %v", err)
	}
	
	// Test anchor drawing size
	if drawing, ok := run2.Children[0].(*Drawing); ok && drawing.Anchor != nil {
		drawing.Anchor.Size(3000000, 3000000)
	}
	
	// Test shape operations
	para3 := doc.AddParagraph()
	para3.AddInlineShape(1000000, 1000000, "Shape1", "auto", "rect", nil)
	
	para4 := doc.AddParagraph()
	para4.AddAnchorShape(1000000, 1000000, "Shape2", "auto", "ellipse", nil)
	
	// Write to verify structure
	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write document with drawings: %v", err)
	}
}

func TestTableWithRowHeights(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	rowHeights := []int64{500, 600, 700, 800}
	colWidths := []int64{1000, 2000, 3000}
	
	table := doc.AddTableTwips(rowHeights, colWidths, 6000, nil)
	
	if table == nil {
		t.Fatal("AddTableTwips returned nil")
	}
	
	if len(table.TableRows) != len(rowHeights) {
		t.Errorf("Expected %d rows, got %d", len(rowHeights), len(table.TableRows))
	}
	
	// Fill table
	for i, row := range table.TableRows {
		for j, cell := range row.TableCells {
			p := cell.AddParagraph()
			p.AddText("Content")
			if i == 0 && j == 0 {
				// Add drawing to a cell
				_, err := p.AddInlineDrawingFrom("testdata/fumiama.JPG")
				if err != nil {
					t.Logf("Warning: could not add image to cell: %v", err)
				}
			}
		}
	}
	
	// Write to verify
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write table document: %v", err)
	}
}

func TestMultipleDocumentOperations(t *testing.T) {
	// Test creating multiple documents and operations
	doc1 := New().WithDefaultTheme().WithA4Page()
	doc1.AddParagraph().AddText("Document 1 with A4 page")
	
	doc2 := New().WithDefaultTheme().WithA3Page()
	doc2.AddParagraph().AddText("Document 2 with A3 page")
	
	doc3 := New().WithDefaultTheme()
	doc3.AddParagraph().AddText("Document 3")
	
	// Test LoadBodyItems
	items := []interface{}{
		&Paragraph{
			Children: []interface{}{
				&Run{
					RunProperties: &RunProperties{},
					Children: []interface{}{
						&Text{Text: "Loaded paragraph"},
					},
				},
			},
		},
	}
	
	media := []Media{
		{Name: "test.jpg", Data: []byte("fake image")},
		{Name: "test2.png", Data: []byte("fake image 2")},
	}
	
	doc4 := LoadBodyItems(items, media)
	if doc4 == nil {
		t.Fatal("LoadBodyItems returned nil")
	}
	
	if len(doc4.media) != len(media) {
		t.Errorf("Expected %d media items, got %d", len(media), len(doc4.media))
	}
	
	// Test media retrieval
	m := doc4.Media("test.jpg")
	if m == nil {
		t.Error("Media not found")
	} else if m.Name != "test.jpg" {
		t.Errorf("Expected media name 'test.jpg', got '%s'", m.Name)
	}
	
	// Test media string
	if m != nil {
		str := m.String()
		if str == "" {
			t.Error("Media String() returned empty")
		}
	}
}

func TestTextOperations(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	
	// Test AddTab on text
	para.AddTab()
	para.AddText("After tab")
	
	// Test various text operations
	run := para.AddText("Test text with various formatting")
	run.Shade("clear", "auto", "FFFF00")
	
	// Write document
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write text document: %v", err)
	}
}

func TestParagraphJustification(t *testing.T) {
	doc := New().WithDefaultTheme()
	
	// Test different justification values
	justifications := []string{"start", "center", "end", "both", "distribute"}
	
	for _, just := range justifications {
		para := doc.AddParagraph()
		para.Justification(just)
		para.AddText("Text with " + just + " justification")
		
		if para.Properties == nil || para.Properties.Justification == nil {
			t.Errorf("Justification not set for %s", just)
		}
	}
	
	// Write to verify
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write document with justifications: %v", err)
	}
}
