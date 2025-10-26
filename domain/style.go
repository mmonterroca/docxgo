/*
MIT License

Copyright (c) 2025 Misael Montero
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

// Built-in paragraph style IDs (OOXML standard)
const (
	StyleIDNormal           = "Normal"
	StyleIDHeading1         = "Heading1"
	StyleIDHeading2         = "Heading2"
	StyleIDHeading3         = "Heading3"
	StyleIDHeading4         = "Heading4"
	StyleIDHeading5         = "Heading5"
	StyleIDHeading6         = "Heading6"
	StyleIDHeading7         = "Heading7"
	StyleIDHeading8         = "Heading8"
	StyleIDHeading9         = "Heading9"
	StyleIDTitle            = "Title"
	StyleIDSubtitle         = "Subtitle"
	StyleIDQuote            = "Quote"
	StyleIDIntenseQuote     = "IntenseQuote"
	StyleIDListParagraph    = "ListParagraph"
	StyleIDCaption          = "Caption"
	StyleIDTOC1             = "TOC1"
	StyleIDTOC2             = "TOC2"
	StyleIDTOC3             = "TOC3"
	StyleIDTOC4             = "TOC4"
	StyleIDTOC5             = "TOC5"
	StyleIDTOC6             = "TOC6"
	StyleIDTOC7             = "TOC7"
	StyleIDTOC8             = "TOC8"
	StyleIDTOC9             = "TOC9"
	StyleIDHeader           = "Header"
	StyleIDFooter           = "Footer"
	StyleIDFootnoteText     = "FootnoteText"
	StyleIDEndnoteText      = "EndnoteText"
	StyleIDBodyText         = "BodyText"
	StyleIDBodyTextIndent   = "BodyTextIndent"
	StyleIDNoSpacing        = "NoSpacing"
)

// Built-in character style IDs
const (
	StyleIDDefaultParagraphFont = "DefaultParagraphFont"
	StyleIDEmphasis             = "Emphasis"
	StyleIDStrong               = "Strong"
	StyleIDSubtle               = "Subtle"
	StyleIDIntenseEmphasis      = "IntenseEmphasis"
	StyleIDIntenseReference     = "IntenseReference"
	StyleIDBookTitle            = "BookTitle"
	StyleIDHyperlink            = "Hyperlink"
	StyleIDFollowedHyperlink    = "FollowedHyperlink"
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

// Indentation represents paragraph indentation in twips.
type Indentation struct {
	Left      int // Left indent
	Right     int // Right indent
	FirstLine int // First line indent (positive) or hanging (negative)
}
