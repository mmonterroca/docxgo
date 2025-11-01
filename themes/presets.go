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

import "github.com/mmonterroca/docxgo/domain"

// Preset themes available out of the box.
// These themes provide professional, ready-to-use document styles.
var (
	// Corporate is a professional theme with navy blue and red accents.
	// Ideal for business reports, proposals, and corporate documentation.
	// Inspired by traditional corporate branding with modern touches.
	Corporate = newCorporateTheme()

	// Startup is an energetic, modern theme with vibrant colors.
	// Perfect for pitch decks, business plans, and innovative proposals.
	// Features bold typography and dynamic color palette.
	Startup = newStartupTheme()

	// Modern is a clean, minimalist theme with contemporary styling.
	// Great for technical documentation, white papers, and modern reports.
	// Emphasizes readability with generous white space.
	Modern = newModernTheme()

	// Fintech is a professional theme with financial industry aesthetics.
	// Ideal for financial reports, investment documents, and banking materials.
	// Uses trustworthy blues and greens.
	Fintech = newFintechTheme()

	// Academic is a traditional theme for scholarly and educational documents.
	// Perfect for research papers, academic reports, and thesis documents.
	// Uses classic serif fonts and conservative styling.
	Academic = newAcademicTheme()
)

// newCorporateTheme creates the Corporate theme.
// Color palette inspired by the AliatUniversidades document:
// - Navy blue (#2F5496) for primary headings
// - Light blue (#4F81BD) for secondary elements
// - Red accent (#C00000) for highlights and emphasis
// - Clean, professional look suitable for business documents
func newCorporateTheme() Theme {
	theme := &baseTheme{
		name:        "corporate",
		displayName: "Corporate",
		description: "Professional business theme with navy blue and red accents",
		colors: ThemeColors{
			Primary:    domain.Color{R: 47, G: 84, B: 150},   // Navy blue (#2F5496)
			Secondary:  domain.Color{R: 79, G: 129, B: 189},  // Light blue (#4F81BD)
			Accent:     domain.Color{R: 192, G: 0, B: 0},     // Red (#C00000)
			Background: domain.Color{R: 255, G: 255, B: 255}, // White
			Text:       domain.Color{R: 0, G: 0, B: 0},       // Black
			TextLight:  domain.Color{R: 89, G: 89, B: 89},    // Dark gray
			Heading:    domain.Color{R: 47, G: 84, B: 150},   // Navy blue (same as primary)
			Muted:      domain.Color{R: 217, G: 217, B: 217}, // Light gray
			Success:    domain.Color{R: 0, G: 176, B: 80},    // Green
			Warning:    domain.Color{R: 255, G: 192, B: 0},   // Orange
			Error:      domain.Color{R: 192, G: 0, B: 0},     // Red (same as accent)
		},
		fonts: ThemeFonts{
			Body:      "Calibri",
			Heading:   "Calibri",
			Monospace: "Courier New",
			BodySize:  22, // 11pt
			SmallSize: 18, // 9pt
		},
		spacing: ThemeSpacing{
			ParagraphBefore: 0,
			ParagraphAfter:  160, // 8pt - comfortable reading
			LineSpacing:     240, // Single spacing
			HeadingBefore:   240, // 12pt - clear section separation
			HeadingAfter:    120, // 6pt
			SectionSpacing:  480, // 24pt - major section breaks
		},
		headings: ThemeHeadings{
			H1Size:      32, // 16pt - strong visual hierarchy
			H2Size:      26, // 13pt
			H3Size:      24, // 12pt
			H1Bold:      true,
			H2Bold:      true,
			H3Bold:      true,
			H1Uppercase: false,
			UseColor:    true, // Use navy blue for headings
		},
	}
	return theme
}

// newStartupTheme creates the Startup theme.
// Modern, energetic design with vibrant colors.
// Perfect for innovation-focused documents.
func newStartupTheme() Theme {
	theme := &baseTheme{
		name:        "startup",
		displayName: "Startup",
		description: "Energetic modern theme with vibrant colors for innovative documents",
		colors: ThemeColors{
			Primary:    domain.Color{R: 106, G: 90, B: 205},  // Slate blue (#6A5ACD)
			Secondary:  domain.Color{R: 72, G: 209, B: 204},  // Turquoise (#48D1CC)
			Accent:     domain.Color{R: 255, G: 99, B: 71},   // Tomato red (#FF6347)
			Background: domain.Color{R: 255, G: 255, B: 255}, // White
			Text:       domain.Color{R: 33, G: 33, B: 33},    // Near black
			TextLight:  domain.Color{R: 117, G: 117, B: 117}, // Medium gray
			Heading:    domain.Color{R: 106, G: 90, B: 205},  // Slate blue
			Muted:      domain.Color{R: 229, G: 229, B: 229}, // Light gray
			Success:    domain.Color{R: 46, G: 204, B: 113},  // Emerald
			Warning:    domain.Color{R: 241, G: 196, B: 15},  // Sun flower
			Error:      domain.Color{R: 231, G: 76, B: 60},   // Alizarin
		},
		fonts: ThemeFonts{
			Body:      "Calibri",
			Heading:   "Calibri Light",
			Monospace: "Consolas",
			BodySize:  22, // 11pt
			SmallSize: 18, // 9pt
		},
		spacing: ThemeSpacing{
			ParagraphBefore: 0,
			ParagraphAfter:  200, // 10pt - airy feel
			LineSpacing:     280, // 1.4x - modern line height
			HeadingBefore:   320, // 16pt - prominent sections
			HeadingAfter:    160, // 8pt
			SectionSpacing:  560, // 28pt - bold section breaks
		},
		headings: ThemeHeadings{
			H1Size:      36, // 18pt - bold statements
			H2Size:      28, // 14pt
			H3Size:      24, // 12pt
			H1Bold:      true,
			H2Bold:      true,
			H3Bold:      false, // H3 lighter for contrast
			H1Uppercase: false,
			UseColor:    true,
		},
	}
	return theme
}

// newModernTheme creates the Modern theme.
// Clean, minimalist design with emphasis on readability.
func newModernTheme() Theme {
	theme := &baseTheme{
		name:        "modern",
		displayName: "Modern",
		description: "Clean minimalist theme with contemporary styling",
		colors: ThemeColors{
			Primary:    domain.Color{R: 52, G: 73, B: 94},    // Wet asphalt (#34495E)
			Secondary:  domain.Color{R: 149, G: 165, B: 166}, // Concrete (#95A5A6)
			Accent:     domain.Color{R: 41, G: 128, B: 185},  // Peter river blue (#2980B9)
			Background: domain.Color{R: 255, G: 255, B: 255}, // White
			Text:       domain.Color{R: 44, G: 62, B: 80},    // Midnight blue (#2C3E50)
			TextLight:  domain.Color{R: 127, G: 140, B: 141}, // Asbestos (#7F8C8D)
			Heading:    domain.Color{R: 52, G: 73, B: 94},    // Wet asphalt
			Muted:      domain.Color{R: 236, G: 240, B: 241}, // Clouds (#ECF0F1)
			Success:    domain.Color{R: 39, G: 174, B: 96},   // Nephritis (#27AE60)
			Warning:    domain.Color{R: 243, G: 156, B: 18},  // Orange (#F39C12)
			Error:      domain.Color{R: 192, G: 57, B: 43},   // Pomegranate (#C0392B)
		},
		fonts: ThemeFonts{
			Body:      "Segoe UI",
			Heading:   "Segoe UI Light",
			Monospace: "Consolas",
			BodySize:  22, // 11pt
			SmallSize: 18, // 9pt
		},
		spacing: ThemeSpacing{
			ParagraphBefore: 0,
			ParagraphAfter:  240, // 12pt - generous spacing
			LineSpacing:     300, // 1.5x - very readable
			HeadingBefore:   400, // 20pt - clear hierarchy
			HeadingAfter:    200, // 10pt
			SectionSpacing:  640, // 32pt - distinct sections
		},
		headings: ThemeHeadings{
			H1Size:      34,    // 17pt
			H2Size:      28,    // 14pt
			H3Size:      24,    // 12pt
			H1Bold:      false, // Light weight for modern look
			H2Bold:      false,
			H3Bold:      true, // Only H3 is bold for contrast
			H1Uppercase: false,
			UseColor:    true,
		},
	}
	return theme
}

// newFintechTheme creates the Fintech theme.
// Professional financial industry aesthetic with trustworthy colors.
func newFintechTheme() Theme {
	theme := &baseTheme{
		name:        "fintech",
		displayName: "Fintech",
		description: "Professional financial theme with trustworthy blues and greens",
		colors: ThemeColors{
			Primary:    domain.Color{R: 0, G: 82, B: 136},    // Dark cerulean (#005288)
			Secondary:  domain.Color{R: 0, G: 123, B: 167},   // Blue NCS (#007BA7)
			Accent:     domain.Color{R: 0, G: 150, B: 136},   // Teal (#009688)
			Background: domain.Color{R: 255, G: 255, B: 255}, // White
			Text:       domain.Color{R: 33, G: 33, B: 33},    // Jet black (#212121)
			TextLight:  domain.Color{R: 97, G: 97, B: 97},    // Sonic silver (#616161)
			Heading:    domain.Color{R: 0, G: 82, B: 136},    // Dark cerulean
			Muted:      domain.Color{R: 224, G: 224, B: 224}, // Gainsboro (#E0E0E0)
			Success:    domain.Color{R: 76, G: 175, B: 80},   // Medium sea green (#4CAF50)
			Warning:    domain.Color{R: 255, G: 152, B: 0},   // Orange peel (#FF9800)
			Error:      domain.Color{R: 211, G: 47, B: 47},   // Jasper (#D32F2F)
		},
		fonts: ThemeFonts{
			Body:      "Arial",
			Heading:   "Arial",
			Monospace: "Courier New",
			BodySize:  22, // 11pt
			SmallSize: 18, // 9pt
		},
		spacing: ThemeSpacing{
			ParagraphBefore: 0,
			ParagraphAfter:  140, // 7pt - compact, professional
			LineSpacing:     240, // Single spacing - efficient
			HeadingBefore:   280, // 14pt - clear but compact
			HeadingAfter:    140, // 7pt
			SectionSpacing:  400, // 20pt - moderate breaks
		},
		headings: ThemeHeadings{
			H1Size:      30, // 15pt - professional, not flashy
			H2Size:      26, // 13pt
			H3Size:      22, // 11pt
			H1Bold:      true,
			H2Bold:      true,
			H3Bold:      true,
			H1Uppercase: false,
			UseColor:    true,
		},
	}
	return theme
}

// newAcademicTheme creates the Academic theme.
// Traditional scholarly styling with conservative design.
func newAcademicTheme() Theme {
	theme := &baseTheme{
		name:        "academic",
		displayName: "Academic",
		description: "Traditional scholarly theme for academic and research documents",
		colors: ThemeColors{
			Primary:    domain.Color{R: 0, G: 0, B: 0},       // Black
			Secondary:  domain.Color{R: 64, G: 64, B: 64},    // Dark gray
			Accent:     domain.Color{R: 128, G: 0, B: 0},     // Maroon (#800000)
			Background: domain.Color{R: 255, G: 255, B: 255}, // White
			Text:       domain.Color{R: 0, G: 0, B: 0},       // Black
			TextLight:  domain.Color{R: 96, G: 96, B: 96},    // Gray
			Heading:    domain.Color{R: 0, G: 0, B: 0},       // Black
			Muted:      domain.Color{R: 192, G: 192, B: 192}, // Silver
			Success:    domain.Color{R: 0, G: 128, B: 0},     // Green
			Warning:    domain.Color{R: 255, G: 165, B: 0},   // Orange
			Error:      domain.Color{R: 139, G: 0, B: 0},     // Dark red
		},
		fonts: ThemeFonts{
			Body:      "Times New Roman",
			Heading:   "Times New Roman",
			Monospace: "Courier New",
			BodySize:  24, // 12pt - traditional academic size
			SmallSize: 20, // 10pt
		},
		spacing: ThemeSpacing{
			ParagraphBefore: 0,
			ParagraphAfter:  0,   // No spacing - traditional format
			LineSpacing:     480, // Double spacing - academic standard
			HeadingBefore:   240, // 12pt
			HeadingAfter:    120, // 6pt
			SectionSpacing:  480, // 24pt
		},
		headings: ThemeHeadings{
			H1Size:      28, // 14pt - conservative sizing
			H2Size:      26, // 13pt
			H3Size:      24, // 12pt (same as body)
			H1Bold:      true,
			H2Bold:      true,
			H3Bold:      true,
			H1Uppercase: false,
			UseColor:    false, // Black only - traditional
		},
	}
	return theme
}

// AllThemes returns all available preset themes.
func AllThemes() []Theme {
	return []Theme{
		Corporate,
		Startup,
		Modern,
		Fintech,
		Academic,
	}
}

// GetTheme returns a theme by name.
// Returns nil if the theme is not found.
func GetTheme(name string) Theme {
	for _, theme := range AllThemes() {
		if theme.Name() == name {
			return theme
		}
	}
	return nil
}

// ThemeNames returns the names of all available themes.
func ThemeNames() []string {
	themes := AllThemes()
	names := make([]string, len(themes))
	for i, theme := range themes {
		names[i] = theme.Name()
	}
	return names
}
