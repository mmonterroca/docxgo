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

import (
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// baseTheme provides a concrete implementation of the Theme interface.
// It serves as the foundation for all preset themes.
type baseTheme struct {
	name        string
	displayName string
	description string
	colors      ThemeColors
	fonts       ThemeFonts
	spacing     ThemeSpacing
	headings    ThemeHeadings
}

// NewTheme creates a new theme with the given configuration.
func NewTheme(name, displayName, description string) Theme {
	return &baseTheme{
		name:        name,
		displayName: displayName,
		description: description,
		colors:      DefaultThemeColors(),
		fonts:       DefaultThemeFonts(),
		spacing:     DefaultThemeSpacing(),
		headings:    DefaultThemeHeadings(),
	}
}

func (t *baseTheme) Name() string {
	return t.name
}

func (t *baseTheme) DisplayName() string {
	return t.displayName
}

func (t *baseTheme) Description() string {
	return t.description
}

func (t *baseTheme) Colors() ThemeColors {
	return t.colors
}

func (t *baseTheme) Fonts() ThemeFonts {
	return t.fonts
}

func (t *baseTheme) Spacing() ThemeSpacing {
	return t.spacing
}

func (t *baseTheme) Headings() ThemeHeadings {
	return t.headings
}

func (t *baseTheme) Clone() Theme {
	return &baseTheme{
		name:        t.name,
		displayName: t.displayName,
		description: t.description,
		colors:      t.colors,
		fonts:       t.fonts,
		spacing:     t.spacing,
		headings:    t.headings,
	}
}

func (t *baseTheme) WithColors(colors ThemeColors) Theme {
	cloned := t.Clone().(*baseTheme)
	cloned.colors = colors
	return cloned
}

func (t *baseTheme) WithFonts(fonts ThemeFonts) Theme {
	cloned := t.Clone().(*baseTheme)
	cloned.fonts = fonts
	return cloned
}

func (t *baseTheme) WithSpacing(spacing ThemeSpacing) Theme {
	cloned := t.Clone().(*baseTheme)
	cloned.spacing = spacing
	return cloned
}

// ApplyTo applies the theme to a document by configuring all relevant styles.
func (t *baseTheme) ApplyTo(doc domain.Document) error {
	const op = "Theme.ApplyTo"

	if doc == nil {
		return errors.NewValidationError(op, "document", nil, "document cannot be nil")
	}

	styleMgr := doc.StyleManager()
	if styleMgr == nil {
		return errors.InvalidState(op, "style manager is nil")
	}

	if err := doc.SetBackgroundColor(t.colors.Background); err != nil {
		return errors.Wrap(err, op)
	}

	// Apply styles in order: Normal, Headings, Title, Quote, etc.
	if err := t.applyNormalStyle(styleMgr); err != nil {
		return errors.Wrap(err, op)
	}

	if err := t.applyHeadingStyles(styleMgr); err != nil {
		return errors.Wrap(err, op)
	}

	t.applyTitleStyles(styleMgr)
	t.applyQuoteStyles(styleMgr)
	t.applyListStyle(styleMgr)

	return nil
}

// applyNormalStyle configures the default paragraph style.
func (t *baseTheme) applyNormalStyle(styleMgr domain.StyleManager) error {
	style, err := styleMgr.GetStyle(domain.StyleIDNormal)
	if err != nil {
		return err
	}

	paraStyle, ok := style.(domain.ParagraphStyle)
	if !ok {
		return errors.InvalidState("applyNormalStyle", "style is not a paragraph style")
	}

	// Set font
	if err := paraStyle.SetFont(domain.Font{Name: t.fonts.Body}); err != nil {
		return err
	}

	// Set default body size and color so runs inherit theme values.
	if err := paraStyle.SetSize(t.fonts.BodySize); err != nil {
		return err
	}

	if err := paraStyle.SetColor(t.colors.Text); err != nil {
		return err
	}

	// Set spacing
	if err := paraStyle.SetSpacingBefore(t.spacing.ParagraphBefore); err != nil {
		return err
	}

	if err := paraStyle.SetSpacingAfter(t.spacing.ParagraphAfter); err != nil {
		return err
	}

	if err := paraStyle.SetLineSpacing(t.spacing.LineSpacing); err != nil {
		return err
	}

	// Note: Size and Color are properties of runs, not paragraph styles in domain.
	// These will be applied when creating actual paragraphs/runs.

	return nil
}

// applyHeadingStyles configures heading styles (H1-H3).
func (t *baseTheme) applyHeadingStyles(styleMgr domain.StyleManager) error {
	headingConfigs := []struct {
		styleID string
		size    int
		bold    bool
	}{
		{domain.StyleIDHeading1, t.headings.H1Size, t.headings.H1Bold},
		{domain.StyleIDHeading2, t.headings.H2Size, t.headings.H2Bold},
		{domain.StyleIDHeading3, t.headings.H3Size, t.headings.H3Bold},
	}

	for _, config := range headingConfigs {
		style, err := styleMgr.GetStyle(config.styleID)
		if err != nil {
			return err
		}

		paraStyle, ok := style.(domain.ParagraphStyle)
		if !ok {
			continue
		}

		// Set font
		if err := paraStyle.SetFont(domain.Font{Name: t.fonts.Heading}); err != nil {
			return err
		}

		// Set size
		if err := paraStyle.SetSize(config.size); err != nil {
			return err
		}

		// Set bold
		if err := paraStyle.SetBold(config.bold); err != nil {
			return err
		}

		// Set color if UseColor is enabled
		if t.headings.UseColor {
			if err := paraStyle.SetColor(t.colors.Heading); err != nil {
				return err
			}
		}

		// Set spacing
		if err := paraStyle.SetSpacingBefore(t.spacing.HeadingBefore); err != nil {
			return err
		}

		if err := paraStyle.SetSpacingAfter(t.spacing.HeadingAfter); err != nil {
			return err
		}

		// Set alignment to left
		if err := paraStyle.SetAlignment(domain.AlignmentLeft); err != nil {
			return err
		}
	}

	return nil
}

// applyTitleStyles configures Title and Subtitle styles.
func (t *baseTheme) applyTitleStyles(styleMgr domain.StyleManager) {
	// Title style
	if titleStyle, err := styleMgr.GetStyle(domain.StyleIDTitle); err == nil {
		if paraStyle, ok := titleStyle.(domain.ParagraphStyle); ok {
			_ = paraStyle.SetFont(domain.Font{Name: t.fonts.Heading})
			_ = paraStyle.SetSize(t.headings.H1Size + 8) // Slightly larger than H1
			_ = paraStyle.SetBold(true)
			if t.headings.UseColor {
				_ = paraStyle.SetColor(t.colors.Primary)
			}
			_ = paraStyle.SetAlignment(domain.AlignmentCenter)
			_ = paraStyle.SetSpacingAfter(t.spacing.HeadingAfter * 2)
		}
	}

	// Subtitle style
	if subtitleStyle, err := styleMgr.GetStyle(domain.StyleIDSubtitle); err == nil {
		if paraStyle, ok := subtitleStyle.(domain.ParagraphStyle); ok {
			_ = paraStyle.SetFont(domain.Font{Name: t.fonts.Body})
			_ = paraStyle.SetSize(t.fonts.BodySize + 4) // Slightly larger than body
			_ = paraStyle.SetColor(t.colors.TextLight)
			_ = paraStyle.SetAlignment(domain.AlignmentCenter)
			_ = paraStyle.SetSpacingAfter(t.spacing.SectionSpacing)
		}
	}
}

// applyQuoteStyles configures quote styles.
func (t *baseTheme) applyQuoteStyles(styleMgr domain.StyleManager) {
	// Quote style
	if quoteStyle, err := styleMgr.GetStyle(domain.StyleIDQuote); err == nil {
		if paraStyle, ok := quoteStyle.(domain.ParagraphStyle); ok {
			_ = paraStyle.SetFont(domain.Font{Name: t.fonts.Body})
			_ = paraStyle.SetColor(t.colors.TextLight)
			_ = paraStyle.SetItalic(true)
			_ = paraStyle.SetIndentation(domain.Indentation{Left: 720, Right: 720}) // 0.5 inch
		}
	}

	// Intense Quote style
	if intenseQuoteStyle, err := styleMgr.GetStyle(domain.StyleIDIntenseQuote); err == nil {
		if paraStyle, ok := intenseQuoteStyle.(domain.ParagraphStyle); ok {
			_ = paraStyle.SetFont(domain.Font{Name: t.fonts.Body})
			_ = paraStyle.SetSize(t.fonts.BodySize + 2)
			_ = paraStyle.SetColor(t.colors.Secondary)
			_ = paraStyle.SetBold(true)
			_ = paraStyle.SetAlignment(domain.AlignmentCenter)
		}
	}
}

// applyListStyle configures list paragraph style.
func (t *baseTheme) applyListStyle(styleMgr domain.StyleManager) {
	if listStyle, err := styleMgr.GetStyle(domain.StyleIDListParagraph); err == nil {
		if paraStyle, ok := listStyle.(domain.ParagraphStyle); ok {
			_ = paraStyle.SetFont(domain.Font{Name: t.fonts.Body})
			_ = paraStyle.SetSize(t.fonts.BodySize)
			_ = paraStyle.SetColor(t.colors.Text)
			_ = paraStyle.SetSpacingAfter(t.spacing.ParagraphAfter / 2) // Tighter spacing for lists
		}
	}
}
