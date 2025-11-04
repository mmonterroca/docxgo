/*
MIT License

Copyright (c) 2025 Misael Monterroca

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

package themes

import "github.com/mmonterroca/docxgo/v2/domain"

// Theme defines a complete visual style for a document including colors,
// fonts, spacing, and formatting rules. Themes provide a consistent look
// and feel across documents while maintaining flexibility for customization.
type Theme interface {
	// Name returns the unique identifier for the theme.
	Name() string

	// DisplayName returns the human-readable name for the theme.
	DisplayName() string

	// Description returns a brief description of the theme's purpose and style.
	Description() string

	// Colors returns the color palette for the theme.
	Colors() ThemeColors

	// Fonts returns the font configuration for the theme.
	Fonts() ThemeFonts

	// Spacing returns the spacing configuration for the theme.
	Spacing() ThemeSpacing

	// Headings returns the heading styles configuration for the theme.
	Headings() ThemeHeadings

	// ApplyTo applies the theme to a document, configuring all styles.
	ApplyTo(doc domain.Document) error

	// Clone creates a copy of the theme that can be customized.
	Clone() Theme

	// WithColors creates a new theme with custom colors.
	WithColors(colors ThemeColors) Theme

	// WithFonts creates a new theme with custom fonts.
	WithFonts(fonts ThemeFonts) Theme

	// WithSpacing creates a new theme with custom spacing.
	WithSpacing(spacing ThemeSpacing) Theme
}

// ThemeColors defines the color palette for a theme.
// All colors follow the RGB model with values from 0-255.
type ThemeColors struct {
	// Primary is the main brand color, used for major headings and key elements.
	Primary domain.Color

	// Secondary is the supporting brand color, used for subheadings and accents.
	Secondary domain.Color

	// Accent is used for highlights, links, and call-to-action elements.
	Accent domain.Color

	// Background is the default page background color (usually white or light).
	Background domain.Color

	// Text is the default text color for body content.
	Text domain.Color

	// TextLight is a lighter text color for secondary information.
	TextLight domain.Color

	// Heading is the color for headings (can be same as Primary).
	Heading domain.Color

	// Muted is a subtle color for borders, dividers, and disabled elements.
	Muted domain.Color

	// Success is used for positive messages and confirmations.
	Success domain.Color

	// Warning is used for caution messages.
	Warning domain.Color

	// Error is used for error messages and critical alerts.
	Error domain.Color
}

// ThemeFonts defines the font configuration for a theme.
type ThemeFonts struct {
	// Body is the font family for body text (e.g., "Calibri", "Arial").
	Body string

	// Heading is the font family for headings.
	Heading string

	// Monospace is the font family for code and technical content.
	Monospace string

	// BodySize is the default body text size in half-points (e.g., 22 = 11pt).
	BodySize int

	// SmallSize is the size for small text like footnotes (in half-points).
	SmallSize int
}

// ThemeSpacing defines the spacing configuration for a theme.
// All measurements are in twips (1/1440 inch), where 20 twips = 1pt.
type ThemeSpacing struct {
	// ParagraphBefore is the space before regular paragraphs.
	ParagraphBefore int

	// ParagraphAfter is the space after regular paragraphs.
	ParagraphAfter int

	// LineSpacing is the line height multiplier (240 = single, 360 = 1.5, 480 = double).
	LineSpacing int

	// HeadingBefore is the space before headings.
	HeadingBefore int

	// HeadingAfter is the space after headings.
	HeadingAfter int

	// SectionSpacing is additional space between major sections.
	SectionSpacing int
}

// ThemeHeadings defines heading-specific styling configuration.
type ThemeHeadings struct {
	// H1Size is the font size for Heading 1 in half-points.
	H1Size int

	// H2Size is the font size for Heading 2 in half-points.
	H2Size int

	// H3Size is the font size for Heading 3 in half-points.
	H3Size int

	// H1Bold indicates if Heading 1 should be bold.
	H1Bold bool

	// H2Bold indicates if Heading 2 should be bold.
	H2Bold bool

	// H3Bold indicates if Heading 3 should be bold.
	H3Bold bool

	// H1Uppercase indicates if Heading 1 should be uppercase.
	H1Uppercase bool

	// UseColor indicates if headings should use the theme color.
	UseColor bool
}

// DefaultThemeColors returns a neutral color palette suitable for most documents.
func DefaultThemeColors() ThemeColors {
	return ThemeColors{
		Primary:    domain.Color{R: 47, G: 84, B: 150},   // Professional blue
		Secondary:  domain.Color{R: 79, G: 129, B: 189},  // Light blue
		Accent:     domain.Color{R: 192, G: 0, B: 0},     // Red accent
		Background: domain.Color{R: 255, G: 255, B: 255}, // White
		Text:       domain.Color{R: 0, G: 0, B: 0},       // Black
		TextLight:  domain.Color{R: 89, G: 89, B: 89},    // Dark gray
		Heading:    domain.Color{R: 31, G: 56, B: 100},   // Dark blue
		Muted:      domain.Color{R: 217, G: 217, B: 217}, // Light gray
		Success:    domain.Color{R: 0, G: 176, B: 80},    // Green
		Warning:    domain.Color{R: 255, G: 192, B: 0},   // Orange
		Error:      domain.Color{R: 192, G: 0, B: 0},     // Red
	}
}

// DefaultThemeFonts returns a standard font configuration.
func DefaultThemeFonts() ThemeFonts {
	return ThemeFonts{
		Body:      "Calibri",
		Heading:   "Calibri",
		Monospace: "Courier New",
		BodySize:  22, // 11pt
		SmallSize: 18, // 9pt
	}
}

// DefaultThemeSpacing returns standard spacing suitable for professional documents.
func DefaultThemeSpacing() ThemeSpacing {
	return ThemeSpacing{
		ParagraphBefore: 0,
		ParagraphAfter:  160, // 8pt
		LineSpacing:     240, // Single spacing
		HeadingBefore:   240, // 12pt
		HeadingAfter:    120, // 6pt
		SectionSpacing:  480, // 24pt
	}
}

// DefaultThemeHeadings returns standard heading sizes and formatting.
func DefaultThemeHeadings() ThemeHeadings {
	return ThemeHeadings{
		H1Size:      32, // 16pt
		H2Size:      26, // 13pt
		H3Size:      24, // 12pt
		H1Bold:      true,
		H2Bold:      true,
		H3Bold:      true,
		H1Uppercase: false,
		UseColor:    true,
	}
}
