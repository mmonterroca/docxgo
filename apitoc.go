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
	"fmt"
	"strconv"
	"strings"
)

// TOCOptions contains configuration options for table of contents generation
type TOCOptions struct {
	Title         string   // "Table of Contents" 
	Depth         int      // 1-4 (niveles H1-H4)
	PageNumbers   bool     // Mostrar números de página
	Hyperlinks    bool     // Hyperlinks clicables
	RightAlign    bool     // Alinear números a la derecha
	TabLeader     string   // "dot", "hyphen", "underscore", "none"
}

// DefaultTOCOptions returns sensible defaults for TOC generation
func DefaultTOCOptions() TOCOptions {
	return TOCOptions{
		Title:       "Table of Contents",
		Depth:       3,
		PageNumbers: true,
		Hyperlinks:  true,
		RightAlign:  true,
		TabLeader:   "dot",
	}
}

// TOCEntry represents a single entry in the table of contents
type TOCEntry struct {
	Level      int    // 1, 2, 3, 4 (heading level)
	Text       string // Text content of the heading
	BookmarkID string // Bookmark ID for hyperlink
	PageNumber string // Page number (placeholder)
}

// AddTOC adds a table of contents to the document with the specified options
func (d *Docx) AddTOC(opts TOCOptions) error {
	// Create the TOC heading paragraph
	titlePara := d.AddParagraph()
	titlePara.AddText(opts.Title).Bold().Size("28")
	titlePara.Justification("center")
	
	// Create the TOC field paragraph
	tocPara := d.AddParagraph()
	tocPara.AddTOCField(opts.Depth, opts.Hyperlinks, opts.PageNumbers)
	
	return nil
}

// AddTOCWithEntries adds a TOC with pre-populated entries (useful for preview)
// Word will replace these with actual entries when the document is opened
func (d *Docx) AddTOCWithEntries(opts TOCOptions, entries []TOCEntry) error {
	// Add title
	if opts.Title != "" {
		titlePara := d.AddParagraph()
		titlePara.AddText(opts.Title).Bold().Size("28")
		titlePara.Justification("center")
	}
	
	// Add TOC field
	tocFieldPara := d.AddParagraph()
	tocFieldPara.AddTOCField(opts.Depth, opts.Hyperlinks, opts.PageNumbers)
	
	// Add sample entries (Word will regenerate these)
	for _, entry := range entries {
		entryPara := d.AddParagraph()
		entryPara.AddTOCEntry(entry.Level, entry.Text, entry.BookmarkID, entry.PageNumber, opts)
	}
	
	return nil
}

// AddTOCEntry adds a single TOC entry with proper formatting
func (p *Paragraph) AddTOCEntry(level int, text string, bookmarkID string, pageNum string, opts TOCOptions) {
	// Set indentation based on level (each level indented more)
	indentValue := (level - 1) * 360 // 0.25 inch per level in twentieths of a point
	
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}
	
	// Add indentation
	p.Properties.Ind = &Ind{
		Left: indentValue,
	}
	
	// Add the text with hyperlink to bookmark
	if opts.Hyperlinks && bookmarkID != "" {
		// Create hyperlink run
		textRun := &Run{
			RunProperties: &RunProperties{
				Color: &Color{Val: "0000FF"}, // Blue for hyperlinks
				Underline: &Underline{Val: "single"},
			},
			file: p.file,
		}
		textRun.AddText(text)
		p.Children = append(p.Children, textRun)
	} else {
		// Regular text
		p.AddText(text)
	}
	
	// Add tab with leader and page number
	if opts.PageNumbers {
		// Add tab
		tabRun := &Run{file: p.file}
		tabRun.AddTab()
		p.Children = append(p.Children, tabRun)
		
		// Add page number (or PAGEREF field)
		if bookmarkID != "" {
			pageRefRun := &Run{file: p.file}
			pageRefRun.Children = append(pageRefRun.Children, &Text{Text: pageNum})
			p.Children = append(p.Children, pageRefRun)
		} else {
			pageRun := &Run{file: p.file}
			pageRun.AddText(pageNum)
			p.Children = append(p.Children, pageRun)
		}
	}
}

// GenerateHeadingBookmark generates a bookmark name for a heading
func GenerateHeadingBookmark(headingText string, level int, index int) string {
	// TOC bookmarks follow the pattern "_Toc" + padded number
	return fmt.Sprintf("_Toc%09d", index)
}

// AddHeadingWithTOC adds a heading paragraph with automatic TOC bookmark
func (d *Docx) AddHeadingWithTOC(text string, level int, tocIndex int) *Paragraph {
	para := d.AddParagraph()
	para.AddText(text)
	
	// Generate bookmark for TOC
	para.AddTOCBookmark(text, tocIndex)
	
	return para
}

// ScanForHeadings scans the document and returns a list of TOC entries
// This is useful for regenerating TOC based on actual document content
func (d *Docx) ScanForHeadings(maxLevel int) []TOCEntry {
	entries := make([]TOCEntry, 0)
	tocIndex := 1
	
	for _, item := range d.Document.Body.Items {
		if para, ok := item.(*Paragraph); ok {
			// Check if paragraph has heading style or looks like a heading
			level := d.detectHeadingLevel(para)
			if level > 0 && level <= maxLevel {
				text := para.String()
				if text != "" {
					bookmarkID := GenerateHeadingBookmark(text, level, tocIndex)
					
					// Add bookmark if not present
					if !para.HasBookmark(bookmarkID) {
						para.AddTOCBookmark(text, tocIndex)
					}
					
					entries = append(entries, TOCEntry{
						Level:      level,
						Text:       text,
						BookmarkID: bookmarkID,
						PageNumber: strconv.Itoa(tocIndex), // Placeholder
					})
					tocIndex++
				}
			}
		}
	}
	
	return entries
}

// detectHeadingLevel tries to detect if a paragraph is a heading and what level
func (d *Docx) detectHeadingLevel(para *Paragraph) int {
	if para.Properties != nil && para.Properties.Style != nil {
		switch para.Properties.Style.Val {
		case "Heading1", "heading 1", "Heading 1":
			return 1
		case "Heading2", "heading 2", "Heading 2":
			return 2
		case "Heading3", "heading 3", "Heading 3":
			return 3
		case "Heading4", "heading 4", "Heading 4":
			return 4
		}
	}
	
	// Fallback: detect based on text formatting (bold, size, etc.)
	text := strings.TrimSpace(para.String())
	if text == "" {
		return 0
	}
	
	// Simple heuristics for heading detection
	if len(text) < 100 && // Short text
		!strings.Contains(text, ".") && // No sentences
		(strings.HasPrefix(text, "1.") || // Numbered
			strings.HasPrefix(text, "Chapter") || // Chapter
			strings.ToUpper(text[0:1]) == text[0:1]) { // Starts with capital
		
		// Check if paragraph has bold formatting
		if d.paragraphHasBoldFormatting(para) {
			return 1 // Default to level 1
		}
	}
	
	return 0
}

// paragraphHasBoldFormatting checks if a paragraph has bold text
func (d *Docx) paragraphHasBoldFormatting(para *Paragraph) bool {
	for _, child := range para.Children {
		if run, ok := child.(*Run); ok {
			if run.RunProperties != nil && run.RunProperties.Bold != nil {
				return true
			}
		}
	}
	return false
}

// RegenerateTOC regenerates the table of contents based on current document headings
func (d *Docx) RegenerateTOC(opts TOCOptions) error {
	// Scan document for headings
	entries := d.ScanForHeadings(opts.Depth)
	
	if len(entries) == 0 {
		return fmt.Errorf("no headings found in document")
	}
	
	// Find existing TOC and replace it, or add new one at the beginning
	// For now, just add at the beginning
	return d.AddTOCWithEntries(opts, entries)
}

// TOCConfiguration stores TOC settings for a document
type TOCConfiguration struct {
	AutoGenerate    bool       // Automatically generate TOC when adding headings
	Options         TOCOptions // TOC generation options
	UpdateOnSave    bool       // Update TOC when saving document
	InsertPosition  string     // "beginning", "end", "after_title"
}

// WithTOCConfiguration sets up automatic TOC generation for the document
func (d *Docx) WithTOCConfiguration(config TOCConfiguration) *Docx {
	// Store configuration in document (would need to extend Docx struct for this)
	// For now, just return the document
	return d
}

// AddSmartHeading adds a heading that automatically participates in TOC generation
func (d *Docx) AddSmartHeading(text string, level int) *Paragraph {
	para := d.AddParagraph()
	
	// Set heading style
	styleName := fmt.Sprintf("Heading%d", level)
	para.Style(styleName)
	
	// Add text
	para.AddText(text)
	
	// Add TOC bookmark
	// Generate a unique index - in a real implementation this would be tracked
	tocIndex := len(d.Document.Body.Items) // Simple approximation
	para.AddTOCBookmark(text, tocIndex)
	
	return para
}