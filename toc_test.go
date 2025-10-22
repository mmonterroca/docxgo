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

func TestDefaultTOCOptions(t *testing.T) {
	opts := DefaultTOCOptions()

	if opts.Title != "Table of Contents" {
		t.Errorf("Expected default title 'Table of Contents', got '%s'", opts.Title)
	}

	if opts.Depth != 3 {
		t.Errorf("Expected default depth 3, got %d", opts.Depth)
	}

	if !opts.PageNumbers {
		t.Error("Expected page numbers to be enabled by default")
	}

	if !opts.Hyperlinks {
		t.Error("Expected hyperlinks to be enabled by default")
	}
}

func TestAddBasicTOC(t *testing.T) {
	w := New()

	// Add basic TOC
	opts := DefaultTOCOptions()
	err := w.AddTOC(opts)
	if err != nil {
		t.Fatalf("Failed to add TOC: %v", err)
	}

	// Check that document has at least 2 paragraphs (title + TOC field)
	if len(w.Document.Body.Items) < 2 {
		t.Errorf("Expected at least 2 items in document, got %d", len(w.Document.Body.Items))
	}

	// Check that first paragraph contains the title
	if para, ok := w.Document.Body.Items[0].(*Paragraph); ok {
		paraText := para.String()
		if !strings.Contains(paraText, opts.Title) {
			t.Errorf("First paragraph should contain TOC title '%s', got '%s'", opts.Title, paraText)
		}
	}
}

func TestTOCEntry(t *testing.T) {
	entry := TOCEntry{
		Level:      1,
		Text:       "Introduction",
		BookmarkID: "_Toc000000001",
		PageNumber: "1",
	}

	if entry.Level != 1 {
		t.Errorf("Expected level 1, got %d", entry.Level)
	}

	if entry.Text != "Introduction" {
		t.Errorf("Expected text 'Introduction', got '%s'", entry.Text)
	}
}

func TestGenerateHeadingBookmark(t *testing.T) {
	bookmark1 := GenerateHeadingBookmark("Introduction", 1, 1)
	expected1 := "_Toc000000001"
	if bookmark1 != expected1 {
		t.Errorf("Expected bookmark '%s', got '%s'", expected1, bookmark1)
	}

	bookmark2 := GenerateHeadingBookmark("Background", 2, 123)
	expected2 := "_Toc000000123"
	if bookmark2 != expected2 {
		t.Errorf("Expected bookmark '%s', got '%s'", expected2, bookmark2)
	}
}

func TestAddHeadingWithTOC(t *testing.T) {
	w := New()

	// Add heading with TOC bookmark
	heading := w.AddHeadingWithTOC("Introduction", 1, 1)

	if heading == nil {
		t.Fatal("Failed to create heading")
	}

	// Check that heading contains the text
	headingText := heading.String()
	if !strings.Contains(headingText, "Introduction") {
		t.Errorf("Heading should contain 'Introduction', got '%s'", headingText)
	}

	// Check that heading has TOC bookmark
	if !heading.HasBookmark("_Toc000000001") {
		t.Error("Heading should have TOC bookmark")
	}
}

func TestAddTOCWithEntries(t *testing.T) {
	w := New()

	// Create sample entries
	entries := []TOCEntry{
		{Level: 1, Text: "Chapter 1: Introduction", BookmarkID: "_Toc000000001", PageNumber: "1"},
		{Level: 2, Text: "1.1 Overview", BookmarkID: "_Toc000000002", PageNumber: "1"},
		{Level: 2, Text: "1.2 Objectives", BookmarkID: "_Toc000000003", PageNumber: "2"},
		{Level: 1, Text: "Chapter 2: Background", BookmarkID: "_Toc000000004", PageNumber: "3"},
	}

	opts := DefaultTOCOptions()
	err := w.AddTOCWithEntries(opts, entries)
	if err != nil {
		t.Fatalf("Failed to add TOC with entries: %v", err)
	}

	// Check that document has appropriate number of items
	// Title + TOC field + entries = 2 + 4 = 6
	expectedItems := 2 + len(entries)
	if len(w.Document.Body.Items) != expectedItems {
		t.Errorf("Expected %d items in document, got %d", expectedItems, len(w.Document.Body.Items))
	}
}

func TestScanForHeadings(t *testing.T) {
	w := New()

	// Add some content with headings
	h1 := w.AddParagraph()
	h1.AddText("1. Introduction").Bold()
	h1.Style("Heading1")

	p1 := w.AddParagraph()
	p1.AddText("This is some content.")

	h2 := w.AddParagraph()
	h2.AddText("2. Background").Bold()
	h2.Style("Heading2")

	p2 := w.AddParagraph()
	p2.AddText("More content here.")

	// Scan for headings
	entries := w.ScanForHeadings(3)

	if len(entries) != 2 {
		t.Errorf("Expected 2 headings, found %d", len(entries))
	}

	// Check first entry
	if entries[0].Level != 1 {
		t.Errorf("First heading should be level 1, got %d", entries[0].Level)
	}

	if !strings.Contains(entries[0].Text, "Introduction") {
		t.Errorf("First heading should contain 'Introduction', got '%s'", entries[0].Text)
	}
}

func TestAddSmartHeading(t *testing.T) {
	w := New()

	heading := w.AddSmartHeading("Smart Introduction", 1)

	if heading == nil {
		t.Fatal("Failed to create smart heading")
	}

	// Check that heading has style
	if heading.Properties == nil || heading.Properties.Style == nil {
		t.Error("Smart heading should have style properties")
	} else if heading.Properties.Style.Val != "Heading1" {
		t.Errorf("Expected style 'Heading1', got '%s'", heading.Properties.Style.Val)
	}

	// Check text content
	headingText := heading.String()
	if !strings.Contains(headingText, "Smart Introduction") {
		t.Errorf("Heading should contain text 'Smart Introduction', got '%s'", headingText)
	}
}

func TestCompleteDocument(t *testing.T) {
	w := New()

	// Add TOC
	opts := DefaultTOCOptions()
	opts.Title = "Contents"
	err := w.AddTOC(opts)
	if err != nil {
		t.Fatalf("Failed to add TOC: %v", err)
	}

	// Add content with headings
	h1 := w.AddHeadingWithTOC("1. Introduction", 1, 1)
	h1.Style("Heading1")

	p1 := w.AddParagraph()
	p1.AddText("Welcome to the introduction section. This document demonstrates the TOC functionality.")

	h2 := w.AddHeadingWithTOC("1.1 Purpose", 2, 2)
	h2.Style("Heading2")

	p2 := w.AddParagraph()
	p2.AddText("The purpose of this section is to explain the goals.")

	h3 := w.AddHeadingWithTOC("2. Implementation", 1, 3)
	h3.Style("Heading1")

	p3 := w.AddParagraph()
	p3.AddText("This section covers the technical implementation details.")

	// Save document
	f, err := os.Create("test_toc_complete.docx")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()
	defer os.Remove("test_toc_complete.docx") // cleanup

	_, err = w.WriteTo(f)
	if err != nil {
		t.Fatalf("Failed to write document: %v", err)
	}

	// Verify document structure
	if len(w.Document.Body.Items) < 8 {
		t.Errorf("Expected at least 8 items in document (TOC + headings + paragraphs), got %d", len(w.Document.Body.Items))
	}
}

func TestTOCEntryFormatting(t *testing.T) {
	w := New()

	p := w.AddParagraph()
	opts := DefaultTOCOptions()

	// Test TOC entry with different levels
	p.AddTOCEntry(1, "Level 1 Heading", "_Toc001", "5", opts)

	// Check indentation was applied
	if p.Properties == nil || p.Properties.Ind == nil {
		t.Error("TOC entry should have indentation properties")
	}

	// Level 1 should have no indentation (level-1)*360 = 0
	expectedIndent := 0
	if p.Properties.Ind.Left != expectedIndent {
		t.Errorf("Expected indent %d for level 1, got %d", expectedIndent, p.Properties.Ind.Left)
	}

	// Test level 2
	p2 := w.AddParagraph()
	p2.AddTOCEntry(2, "Level 2 Heading", "_Toc002", "7", opts)

	expectedIndent2 := 360 // (2-1)*360
	if p2.Properties == nil || p2.Properties.Ind == nil || p2.Properties.Ind.Left != expectedIndent2 {
		t.Errorf("Expected indent %d for level 2, got %d", expectedIndent2, p2.Properties.Ind.Left)
	}
}
