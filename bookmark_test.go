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

func TestBookmarkBasic(t *testing.T) {
	w := New()
	
	// Create a paragraph with text and bookmark
	p := w.AddParagraph()
	p.AddText("Introduction")
	bookmark := p.AddBookmark("intro_section")
	
	// Verify bookmark was created
	if bookmark == nil {
		t.Fatal("Failed to create bookmark")
	}
	
	// Verify bookmark has correct name
	if bookmark.Start.Name != "intro_section" {
		t.Errorf("Expected bookmark name 'intro_section', got '%s'", bookmark.Start.Name)
	}
	
	// Verify bookmark start and end have same ID
	if bookmark.Start.ID != bookmark.End.ID {
		t.Errorf("Bookmark start and end IDs don't match: start=%s, end=%s", bookmark.Start.ID, bookmark.End.ID)
	}
	
	// Verify we can find the bookmark
	foundBookmark := p.GetBookmarkByName("intro_section")
	if foundBookmark == nil {
		t.Error("Could not find bookmark by name")
	}
	
	// Verify HasBookmark works
	if !p.HasBookmark("intro_section") {
		t.Error("HasBookmark returned false for existing bookmark")
	}
	
	if p.HasBookmark("non_existent") {
		t.Error("HasBookmark returned true for non-existent bookmark")
	}
}

func TestTOCBookmark(t *testing.T) {
	w := New()
	
	// Create a heading with TOC bookmark
	h1 := w.AddParagraph()
	h1.AddText("Chapter 1: Introduction")
	tocBookmark := h1.AddTOCBookmark("Chapter 1: Introduction", 1)
	
	// Verify TOC bookmark format
	expectedName := "_Toc000000001"
	if tocBookmark.Start.Name != expectedName {
		t.Errorf("Expected TOC bookmark name '%s', got '%s'", expectedName, tocBookmark.Start.Name)
	}
	
	// Test multiple TOC bookmarks
	h2 := w.AddParagraph()
	h2.AddText("Chapter 2: Background")
	tocBookmark2 := h2.AddTOCBookmark("Chapter 2: Background", 2)
	
	expectedName2 := "_Toc000000002"
	if tocBookmark2.Start.Name != expectedName2 {
		t.Errorf("Expected TOC bookmark name '%s', got '%s'", expectedName2, tocBookmark2.Start.Name)
	}
}

func TestBookmarkWithSpecificID(t *testing.T) {
	w := New()
	
	p := w.AddParagraph()
	p.AddText("Test paragraph")
	bookmark := p.AddBookmarkWithID("test_bookmark", "custom_123")
	
	// Verify custom ID was used
	if bookmark.Start.ID != "custom_123" {
		t.Errorf("Expected custom ID 'custom_123', got '%s'", bookmark.Start.ID)
	}
	
	if bookmark.End.ID != "custom_123" {
		t.Errorf("Expected custom ID 'custom_123' for end, got '%s'", bookmark.End.ID)
	}
}

func TestDocumentWithBookmarks(t *testing.T) {
	w := New()
	
	// Add a TOC placeholder
	tocPara := w.AddParagraph()
	tocPara.AddText("Table of Contents")
	
	// Add some content with bookmarks
	h1 := w.AddParagraph()
	h1.AddText("1. Introduction")
	h1.AddTOCBookmark("1. Introduction", 1)
	
	p1 := w.AddParagraph()
	p1.AddText("This is the introduction section.")
	
	h2 := w.AddParagraph()
	h2.AddText("2. Background")
	h2.AddTOCBookmark("2. Background", 2)
	
	p2 := w.AddParagraph()
	p2.AddText("This is the background section.")
	
	// Save and verify structure
	f, err := os.Create("test_bookmarks.docx")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()
	defer os.Remove("test_bookmarks.docx") // cleanup
	
	_, err = w.WriteTo(f)
	if err != nil {
		t.Fatalf("Failed to write document: %v", err)
	}
	
	// Basic verification - check paragraph string representation
	docString := h1.String()
	if strings.Contains(docString, "[BOOKMARK:") {
		t.Error("Paragraph string should not contain bookmark markers when they're disabled")
	}
}

func TestNewBookmark(t *testing.T) {
	// Test with provided ID
	bookmark1 := NewBookmark("test_name", "test_id")
	if bookmark1.Start.Name != "test_name" {
		t.Errorf("Expected name 'test_name', got '%s'", bookmark1.Start.Name)
	}
	if bookmark1.Start.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got '%s'", bookmark1.Start.ID)
	}
	
	// Test with auto-generated ID
	bookmark2 := NewBookmark("auto_name", "")
	if bookmark2.Start.Name != "auto_name" {
		t.Errorf("Expected name 'auto_name', got '%s'", bookmark2.Start.Name)
	}
	if bookmark2.Start.ID == "" {
		t.Error("Expected auto-generated ID, got empty string")
	}
}