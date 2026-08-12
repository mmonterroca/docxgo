// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package domain

// Run represents a run of formatted text within a paragraph.
// A run is the smallest unit of text with consistent formatting.
type Run interface {
	// Text returns the text content of this run.
	Text() string

	// SetText sets the text content of this run.
	SetText(text string) error

	// Font returns the font settings for this run.
	Font() Font

	// SetFont sets the font for this run.
	SetFont(font Font) error

	// Color returns the text color.
	Color() Color

	// SetColor sets the text color.
	SetColor(color Color) error

	// Size returns the font size in half-points (e.g., 24 = 12pt).
	Size() int

	// SetSize sets the font size in half-points.
	SetSize(halfPoints int) error

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

	// Strike returns whether the text is struck through.
	Strike() bool

	// SetStrike sets whether the text is struck through.
	SetStrike(strike bool) error

	// Caps returns whether the text is displayed in all capitals. This is a
	// display-only override (w:caps): it does not change the run's stored
	// text, only how it renders — the same distinction as Word's own "All
	// Caps" character formatting versus actually typing in capitals.
	Caps() bool

	// SetCaps sets whether the text is displayed in all capitals.
	SetCaps(caps bool) error

	// Highlight returns the highlight color.
	Highlight() HighlightColor

	// SetHighlight sets the highlight color.
	SetHighlight(color HighlightColor) error

	// Language returns the run's language override, or nil if unset — in
	// which case the run inherits the document's default proofing language
	// (see Document.SetLanguage). Mutating the returned value has no effect
	// on the run.
	Language() *Language

	// SetLanguage sets a per-run language override, used by Word for
	// spell-checking, grammar-checking, and hyphenation of just this run —
	// e.g. a foreign-language phrase inside an otherwise single-language
	// paragraph. Pass nil to clear it and fall back to the document default.
	// Returns an error if lang is non-nil but has no tag set (Val, EastAsia,
	// and Bidi all empty).
	SetLanguage(lang *Language) error

	// AddText is a convenience method that appends text to the run.
	AddText(text string) error

	// AddBreak adds a break to the run (page, column, or line break).
	AddBreak(breakType BreakType) error

	// AddField adds a field to this run (e.g., page number, TOC, hyperlink).
	AddField(field Field) error

	// ClearFields removes all fields from this run.
	// This is used when replacing MERGEFIELD placeholders to prevent the
	// serializer from writing the field structure alongside the replaced text.
	ClearFields()

	// Fields returns the fields embedded in this run.
	Fields() []Field

	// Breaks returns the breaks in this run.
	Breaks() []BreakType

	// Image returns the image associated with this run, if any.
	Image() Image
}

// Font represents font settings.
type Font struct {
	Name     string
	EastAsia string
	CS       string // Complex script
}

// Color represents an RGB color.
type Color struct {
	R uint8
	G uint8
	B uint8
}

// Common colors
var (
	ColorBlack = Color{0, 0, 0}
	ColorWhite = Color{255, 255, 255}
	ColorRed   = Color{255, 0, 0}
	ColorGreen = Color{0, 255, 0}
	ColorBlue  = Color{0, 0, 255}
)

// BreakType represents the type of break.
type BreakType int

// Break type constants.
const (
	BreakTypePage   BreakType = iota // Page break
	BreakTypeColumn                  // Column break
	BreakTypeLine                    // Text wrapping break
)

// UnderlineStyle represents text underline styles.
type UnderlineStyle int

// Underline style constants.
const (
	UnderlineNone   UnderlineStyle = iota // No underline
	UnderlineSingle                       // Single line underline
	UnderlineDouble                       // Double line underline
	UnderlineThick                        // Thick line underline
	UnderlineDotted                       // Dotted line underline
	UnderlineDashed                       // Dashed line underline
	UnderlineWave                         // Wavy line underline
)

// HighlightColor represents text highlight/background colors.
type HighlightColor int

// Highlight color constants.
const (
	HighlightNone        HighlightColor = iota // No highlight
	HighlightYellow                            // Yellow highlight
	HighlightGreen                             // Green highlight
	HighlightCyan                              // Cyan highlight
	HighlightMagenta                           // Magenta highlight
	HighlightBlue                              // Blue highlight
	HighlightRed                               // Red highlight
	HighlightDarkBlue                          // Dark blue highlight
	HighlightDarkCyan                          // Dark cyan highlight
	HighlightDarkGreen                         // Dark green highlight
	HighlightDarkMagenta                       // Dark magenta highlight
	HighlightDarkRed                           // Dark red highlight
	HighlightDarkYellow                        // Dark yellow highlight
	HighlightDarkGray                          // Dark gray highlight
	HighlightLightGray                         // Light gray highlight
)
