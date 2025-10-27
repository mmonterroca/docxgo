/*
MIT License

Copyright (c) 2025 Misael Monterroca
Copyright (c) 2020-2023 fumiama (original go-docx)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/



package domain

// Built-in paragraph style IDs (OOXML standard).
// These style IDs are recognized by Microsoft Word and compatible applications.
const (
	StyleIDNormal           = "Normal"           // Default normal paragraph style
	StyleIDHeading1         = "Heading1"         // Heading level 1
	StyleIDHeading2         = "Heading2"         // Heading level 2
	StyleIDHeading3         = "Heading3"         // Heading level 3
	StyleIDHeading4         = "Heading4"         // Heading level 4
	StyleIDHeading5         = "Heading5"         // Heading level 5
	StyleIDHeading6         = "Heading6"         // Heading level 6
	StyleIDHeading7         = "Heading7"         // Heading level 7
	StyleIDHeading8         = "Heading8"         // Heading level 8
	StyleIDHeading9         = "Heading9"         // Heading level 9
	StyleIDTitle            = "Title"            // Document title style
	StyleIDSubtitle         = "Subtitle"         // Document subtitle style
	StyleIDQuote            = "Quote"            // Block quote style
	StyleIDIntenseQuote     = "IntenseQuote"     // Emphasized block quote style
	StyleIDListParagraph    = "ListParagraph"    // List paragraph style
	StyleIDCaption          = "Caption"          // Caption style (for figures/tables)
	StyleIDTOC1             = "TOC1"             // Table of contents level 1
	StyleIDTOC2             = "TOC2"             // Table of contents level 2
	StyleIDTOC3             = "TOC3"             // Table of contents level 3
	StyleIDTOC4             = "TOC4"             // Table of contents level 4
	StyleIDTOC5             = "TOC5"             // Table of contents level 5
	StyleIDTOC6             = "TOC6"             // Table of contents level 6
	StyleIDTOC7             = "TOC7"             // Table of contents level 7
	StyleIDTOC8             = "TOC8"             // Table of contents level 8
	StyleIDTOC9             = "TOC9"             // Table of contents level 9
	StyleIDHeader           = "Header"           // Header text style
	StyleIDFooter           = "Footer"           // Footer text style
	StyleIDFootnoteText     = "FootnoteText"     // Footnote text style
	StyleIDEndnoteText      = "EndnoteText"      // Endnote text style
	StyleIDBodyText         = "BodyText"         // Body text style
	StyleIDBodyTextIndent   = "BodyTextIndent"   // Body text with first-line indent
	StyleIDNoSpacing        = "NoSpacing"        // No spacing paragraph style
)

// Built-in character style IDs (OOXML standard).
// These apply formatting to individual runs of text within paragraphs.
const (
	StyleIDDefaultParagraphFont = "DefaultParagraphFont" // Default character formatting
	StyleIDEmphasis             = "Emphasis"             // Emphasis (typically italic)
	StyleIDStrong               = "Strong"               // Strong emphasis (typically bold)
	StyleIDSubtle               = "Subtle"               // Subtle text
	StyleIDIntenseEmphasis      = "IntenseEmphasis"      // Intense emphasis
	StyleIDIntenseReference     = "IntenseReference"     // Intense reference
	StyleIDBookTitle            = "BookTitle"            // Book title style
	StyleIDHyperlink            = "Hyperlink"            // Hyperlink text style
	StyleIDFollowedHyperlink    = "FollowedHyperlink"    // Visited hyperlink style
)

// StyleManager manages document styles.
type StyleManager interface {
	// GetStyle retrieves a style by ID.
	GetStyle(styleID string) (Style, error)

	// AddStyle adds a new custom style.
	AddStyle(style Style) error

	// RemoveStyle removes a custom style.
	RemoveStyle(styleID string) error

	// ListStyles returns all available styles.
	ListStyles() []Style

	// ListStylesByType returns all styles of a specific type.
	ListStylesByType(styleType StyleType) []Style

	// DefaultStyle returns the default style for a type.
	DefaultStyle(styleType StyleType) (Style, error)

	// SetDefaultStyle sets the default style for a type.
	SetDefaultStyle(styleType StyleType, styleID string) error

	// HasStyle checks if a style exists.
	HasStyle(styleID string) bool

	// IsBuiltIn checks if a style is built-in (not custom).
	IsBuiltIn(styleID string) bool
}

// ParagraphStyle extends Style with paragraph-specific properties.
type ParagraphStyle interface {
	Style

	// Alignment returns the paragraph alignment.
	Alignment() Alignment

	// SetAlignment sets the paragraph alignment.
	SetAlignment(align Alignment) error

	// SpacingBefore returns spacing before paragraph in twips.
	SpacingBefore() int

	// SetSpacingBefore sets spacing before paragraph.
	SetSpacingBefore(twips int) error

	// SpacingAfter returns spacing after paragraph in twips.
	SpacingAfter() int

	// SetSpacingAfter sets spacing after paragraph.
	SetSpacingAfter(twips int) error

	// LineSpacing returns the line spacing value.
	LineSpacing() int

	// SetLineSpacing sets the line spacing.
	SetLineSpacing(value int) error

	// Indentation returns the paragraph indentation.
	Indentation() Indentation

	// SetIndentation sets the paragraph indentation.
	SetIndentation(indent Indentation) error

	// KeepNext returns whether to keep with next paragraph.
	KeepNext() bool

	// SetKeepNext sets keep with next paragraph.
	SetKeepNext(keep bool) error

	// KeepLines returns whether to keep lines together.
	KeepLines() bool

	// SetKeepLines sets keep lines together.
	SetKeepLines(keep bool) error

	// PageBreakBefore returns whether to insert page break before.
	PageBreakBefore() bool

	// SetPageBreakBefore sets page break before.
	SetPageBreakBefore(breakBefore bool) error

	// OutlineLevel returns the outline level (0-9, 0=body text).
	OutlineLevel() int

	// SetOutlineLevel sets the outline level.
	SetOutlineLevel(level int) error
}

// CharacterStyle extends Style with character-specific properties.
type CharacterStyle interface {
	Style

	// Bold returns whether the text is bold.
	Bold() bool

	// SetBold sets whether the text is bold.
	SetBold(bold bool) error

	// Italic returns whether the text is italic.
	Italic() bool

	// SetItalic sets whether the text is italic.
	SetItalic(italic bool) error

	// Underline returns the underline style.
	Underline() UnderlineStyle

	// SetUnderline sets the underline style.
	SetUnderline(style UnderlineStyle) error

	// Color returns the text color.
	Color() Color

	// SetColor sets the text color.
	SetColor(color Color) error

	// Size returns the font size in half-points.
	Size() int

	// SetSize sets the font size in half-points.
	SetSize(halfPoints int) error
}
