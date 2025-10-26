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

	// Highlight returns the highlight color.
	Highlight() HighlightColor

	// SetHighlight sets the highlight color.
	SetHighlight(color HighlightColor) error

	// AddText is a convenience method that appends text to the run.
	AddText(text string) error

	// AddField adds a field to this run (e.g., page number, TOC, hyperlink).
	AddField(field Field) error
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

// UnderlineStyle represents underline styles.
type UnderlineStyle int

const (
	UnderlineNone UnderlineStyle = iota
	UnderlineSingle
	UnderlineDouble
	UnderlineThick
	UnderlineDotted
	UnderlineDashed
	UnderlineWave
)

// HighlightColor represents highlight colors.
type HighlightColor int

const (
	HighlightNone HighlightColor = iota
	HighlightYellow
	HighlightGreen
	HighlightCyan
	HighlightMagenta
	HighlightBlue
	HighlightRed
	HighlightDarkBlue
	HighlightDarkCyan
	HighlightDarkGreen
	HighlightDarkMagenta
	HighlightDarkRed
	HighlightDarkYellow
	HighlightDarkGray
	HighlightLightGray
)
