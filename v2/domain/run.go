/*
   Copyright (c) 2025 SlideLang

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

// Hex returns the color as a hex string (e.g., "FF0000" for red).
func (c Color) Hex() string {
	return sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

// NewColorFromHex creates a color from a hex string.
func NewColorFromHex(hex string) (Color, error) {
	// Implementation would parse hex string
	return Color{}, nil
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

// Helper function placeholder (would be in a separate package)
func sprintf(format string, args ...interface{}) string {
	// Real implementation would use fmt.Sprintf
	return ""
}
