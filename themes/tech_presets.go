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

// Tech preset themes for technical presentations and architecture documents.
var (
	// TechPresentation is a modern technical presentation theme with vibrant accents.
	// Perfect for architecture documents, technical specifications, and code documentation.
	TechPresentation = newTechPresentationTheme()

	// TechDarkMode is a dark theme optimized for technical content.
	// Features dark background with high contrast for code and diagrams.
	TechDarkMode = newTechDarkModeTheme()
)

// newTechPresentationTheme creates a modern technical presentation theme.
func newTechPresentationTheme() Theme {
	theme := NewTheme(
		"tech-presentation",
		"Tech Presentation",
		"Modern theme for technical presentations and architecture documents with vibrant accents",
	)

	// Color palette inspired by VS Code and modern dev tools
	colors := ThemeColors{
		Primary:    domain.Color{R: 0, G: 122, B: 204},   // VS Code Blue
		Secondary:  domain.Color{R: 80, G: 80, B: 80},    // Dark Gray
		Accent:     domain.Color{R: 237, G: 100, B: 166}, // Hot Pink (for highlights)
		Background: domain.Color{R: 255, G: 255, B: 255}, // White
		Text:       domain.Color{R: 33, G: 33, B: 33},    // Almost Black
		TextLight:  domain.Color{R: 108, G: 117, B: 125}, // Gray
		Heading:    domain.Color{R: 0, G: 122, B: 204},   // VS Code Blue
		Muted:      domain.Color{R: 238, G: 238, B: 238}, // Light Gray (for code backgrounds)
		Success:    domain.Color{R: 40, G: 167, B: 69},   // Green
		Warning:    domain.Color{R: 255, G: 193, B: 7},   // Amber
		Error:      domain.Color{R: 220, G: 53, B: 69},   // Red
	}
	theme = theme.WithColors(colors)

	// Modern tech fonts with excellent readability
	fonts := ThemeFonts{
		Body:      "Roboto",    // Modern, tech-focused
		Heading:   "Roboto",    // Consistent
		Monospace: "Fira Code", // Modern monospace for code
		BodySize:  22,          // 11pt
		SmallSize: 18,          // 9pt
	}
	theme = theme.WithFonts(fonts)

	// Tighter spacing for technical content
	spacing := ThemeSpacing{
		ParagraphBefore: 0,
		ParagraphAfter:  160, // 8pt
		LineSpacing:     260, // 1.3x
		HeadingBefore:   320, // 16pt
		HeadingAfter:    160, // 8pt
		SectionSpacing:  480, // 24pt
	}
	theme = theme.WithSpacing(spacing)

	return theme
}

// newTechDarkModeTheme creates a dark mode theme for technical content.
func newTechDarkModeTheme() Theme {
	theme := NewTheme(
		"tech-darkmode",
		"Tech Dark Mode",
		"Dark theme optimized for technical presentations with high contrast for code",
	)

	// Dark theme colors inspired by VS Code Dark+ and GitHub Dark
	colors := ThemeColors{
		Primary:    domain.Color{R: 79, G: 192, B: 255},  // Bright Blue
		Secondary:  domain.Color{R: 206, G: 145, B: 120}, // Peach/Orange
		Accent:     domain.Color{R: 255, G: 121, B: 198}, // Pink
		Background: domain.Color{R: 30, G: 30, B: 30},    // Dark Background
		Text:       domain.Color{R: 212, G: 212, B: 212}, // Light Gray Text
		TextLight:  domain.Color{R: 150, G: 150, B: 150}, // Medium Gray
		Heading:    domain.Color{R: 79, G: 192, B: 255},  // Bright Blue
		Muted:      domain.Color{R: 45, G: 45, B: 45},    // Slightly lighter background
		Success:    domain.Color{R: 73, G: 204, B: 144},  // Mint Green
		Warning:    domain.Color{R: 255, G: 203, B: 107}, // Golden
		Error:      domain.Color{R: 255, G: 85, B: 85},   // Bright Red
	}
	theme = theme.WithColors(colors)

	// Same tech fonts as light mode
	fonts := ThemeFonts{
		Body:      "Roboto",
		Heading:   "Roboto",
		Monospace: "Fira Code",
		BodySize:  22,
		SmallSize: 18,
	}
	theme = theme.WithFonts(fonts)

	// Same spacing
	spacing := ThemeSpacing{
		ParagraphBefore: 0,
		ParagraphAfter:  160,
		LineSpacing:     260,
		HeadingBefore:   320,
		HeadingAfter:    160,
		SectionSpacing:  480,
	}
	theme = theme.WithSpacing(spacing)

	return theme
}
