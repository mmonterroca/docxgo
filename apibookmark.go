/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)
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
	"encoding/xml"
	"fmt"
)

// BookmarkStart represents a w:bookmarkStart element in Word XML
type BookmarkStart struct {
	XMLName xml.Name `xml:"w:bookmarkStart"`
	ID      string   `xml:"w:id,attr"`
	Name    string   `xml:"w:name,attr"`
}

// BookmarkEnd represents a w:bookmarkEnd element in Word XML
type BookmarkEnd struct {
	XMLName xml.Name `xml:"w:bookmarkEnd"`
	ID      string   `xml:"w:id,attr"`
}

// Bookmark represents a complete bookmark with start and end
type Bookmark struct {
	Start *BookmarkStart
	End   *BookmarkEnd
}

// NewBookmark creates a new bookmark with start and end elements
func NewBookmark(name string, id string) *Bookmark {
	if id == "" {
		id = generateBookmarkID()
	}
	return &Bookmark{
		Start: &BookmarkStart{
			ID:   id,
			Name: name,
		},
		End: &BookmarkEnd{
			ID: id,
		},
	}
}

// generateBookmarkID generates a unique ID for bookmarks
func generateBookmarkID() string {
	// Generate a simple incremental ID
	// In a real implementation, this should use a counter in the Docx struct
	return fmt.Sprintf("bookmark_%d", 12345) // Placeholder - should be improved to use document counter
}

// AddBookmark adds a bookmark to a paragraph
// This will add both bookmarkStart and bookmarkEnd elements to the paragraph
func (p *Paragraph) AddBookmark(name string) *Bookmark {
	bookmark := NewBookmark(name, "")

	// Add bookmark start to the beginning of paragraph children
	p.Children = append([]interface{}{bookmark.Start}, p.Children...)

	// Add bookmark end to the end of paragraph children
	p.Children = append(p.Children, bookmark.End)

	return bookmark
}

// AddBookmarkWithID adds a bookmark with a specific ID to a paragraph
func (p *Paragraph) AddBookmarkWithID(name string, id string) *Bookmark {
	bookmark := NewBookmark(name, id)

	// Add bookmark start to the beginning of paragraph children
	p.Children = append([]interface{}{bookmark.Start}, p.Children...)

	// Add bookmark end to the end of paragraph children
	p.Children = append(p.Children, bookmark.End)

	return bookmark
}

// AddTOCBookmark adds a TOC-style bookmark to a paragraph (used for headings)
// TOC bookmarks follow the pattern "_Toc" + sequential number
func (p *Paragraph) AddTOCBookmark(heading string, tocNumber int) *Bookmark {
	bookmarkName := fmt.Sprintf("_Toc%09d", tocNumber)
	return p.AddBookmarkWithID(bookmarkName, fmt.Sprintf("toc_%d", tocNumber))
}

// GetBookmarkByName finds a bookmark in the paragraph by name
func (p *Paragraph) GetBookmarkByName(name string) *BookmarkStart {
	for _, child := range p.Children {
		if bookmark, ok := child.(*BookmarkStart); ok {
			if bookmark.Name == name {
				return bookmark
			}
		}
	}
	return nil
}

// HasBookmark checks if the paragraph contains a bookmark with the given name
func (p *Paragraph) HasBookmark(name string) bool {
	return p.GetBookmarkByName(name) != nil
}
