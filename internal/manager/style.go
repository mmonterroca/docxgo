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



package manager

import (
	"sync"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// styleManager implements domain.StyleManager.
type styleManager struct {
	mu              sync.RWMutex
	styles          map[string]domain.Style
	defaultStyles   map[domain.StyleType]string
	builtInStyleIDs map[string]bool
}

// NewStyleManager creates a new StyleManager with built-in styles.
func NewStyleManager() domain.StyleManager {
	sm := &styleManager{
		styles:          make(map[string]domain.Style),
		defaultStyles:   make(map[domain.StyleType]string),
		builtInStyleIDs: make(map[string]bool),
	}

	// Initialize built-in styles
	sm.initializeBuiltInStyles()

	return sm
}

// initializeBuiltInStyles creates all built-in OOXML styles.
func (sm *styleManager) initializeBuiltInStyles() {
	// Mark built-in paragraph style IDs
	builtInParagraphStyles := []string{
		domain.StyleIDNormal,
		domain.StyleIDHeading1, domain.StyleIDHeading2, domain.StyleIDHeading3,
		domain.StyleIDHeading4, domain.StyleIDHeading5, domain.StyleIDHeading6,
		domain.StyleIDHeading7, domain.StyleIDHeading8, domain.StyleIDHeading9,
		domain.StyleIDTitle, domain.StyleIDSubtitle,
		domain.StyleIDQuote, domain.StyleIDIntenseQuote,
		domain.StyleIDListParagraph, domain.StyleIDCaption,
		domain.StyleIDTOC1, domain.StyleIDTOC2, domain.StyleIDTOC3,
		domain.StyleIDTOC4, domain.StyleIDTOC5, domain.StyleIDTOC6,
		domain.StyleIDTOC7, domain.StyleIDTOC8, domain.StyleIDTOC9,
		domain.StyleIDHeader, domain.StyleIDFooter,
		domain.StyleIDFootnoteText, domain.StyleIDEndnoteText,
		domain.StyleIDBodyText, domain.StyleIDBodyTextIndent,
		domain.StyleIDNoSpacing,
	}

	for _, id := range builtInParagraphStyles {
		sm.builtInStyleIDs[id] = true
	}

	// Mark built-in character style IDs
	builtInCharacterStyles := []string{
		domain.StyleIDDefaultParagraphFont,
		domain.StyleIDEmphasis, domain.StyleIDStrong, domain.StyleIDSubtle,
		domain.StyleIDIntenseEmphasis, domain.StyleIDIntenseReference,
		domain.StyleIDBookTitle,
		domain.StyleIDHyperlink, domain.StyleIDFollowedHyperlink,
	}

	for _, id := range builtInCharacterStyles {
		sm.builtInStyleIDs[id] = true
	}

	// Create and add built-in styles
	sm.createBuiltInParagraphStyles()
	sm.createBuiltInCharacterStyles()

	// Set default styles
	sm.defaultStyles[domain.StyleTypeParagraph] = domain.StyleIDNormal
	sm.defaultStyles[domain.StyleTypeCharacter] = domain.StyleIDDefaultParagraphFont
}

// createBuiltInParagraphStyles creates all built-in paragraph styles.
// Note: Error returns are intentionally ignored for built-in styles as they
// use only valid, hardcoded values that cannot fail under normal circumstances.
//
//nolint:errcheck // Built-in styles use hardcoded valid values
func (sm *styleManager) createBuiltInParagraphStyles() {
	// Normal style (base for all paragraphs)
	normal := newParagraphStyle(domain.StyleIDNormal, "Normal", true)
	normal.SetAlignment(domain.AlignmentLeft)
	normal.SetDefault(true)
	sm.styles[domain.StyleIDNormal] = normal

	// Heading styles
	headings := []struct {
		id           string
		name         string
		outlineLevel int
		size         int // in half-points
		bold         bool
	}{
		{domain.StyleIDHeading1, "Heading 1", 1, 32, true},  // 16pt
		{domain.StyleIDHeading2, "Heading 2", 2, 26, true},  // 13pt
		{domain.StyleIDHeading3, "Heading 3", 3, 24, true},  // 12pt
		{domain.StyleIDHeading4, "Heading 4", 4, 24, true},  // 12pt
		{domain.StyleIDHeading5, "Heading 5", 5, 22, true},  // 11pt
		{domain.StyleIDHeading6, "Heading 6", 6, 22, false}, // 11pt
		{domain.StyleIDHeading7, "Heading 7", 7, 22, false}, // 11pt
		{domain.StyleIDHeading8, "Heading 8", 8, 22, false}, // 11pt
		{domain.StyleIDHeading9, "Heading 9", 9, 22, false}, // 11pt
	}

	for _, h := range headings {
		style := newParagraphStyle(h.id, h.name, true)
		style.SetBasedOn(domain.StyleIDNormal)
		style.SetOutlineLevel(h.outlineLevel)
		style.SetSpacingBefore(240) // 240 twips = 12pt
		style.SetSpacingAfter(120)  // 120 twips = 6pt
		style.SetKeepNext(true)
		style.SetKeepLines(true)
		
		// Set font through the base Style interface
		font := domain.Font{Name: "Calibri Light"}
		style.SetFont(font)
		
		sm.styles[h.id] = style
	}

	// Title and Subtitle
	title := newParagraphStyle(domain.StyleIDTitle, "Title", true)
	title.SetBasedOn(domain.StyleIDNormal)
	title.SetSpacingAfter(180)
	sm.styles[domain.StyleIDTitle] = title

	subtitle := newParagraphStyle(domain.StyleIDSubtitle, "Subtitle", true)
	subtitle.SetBasedOn(domain.StyleIDNormal)
	sm.styles[domain.StyleIDSubtitle] = subtitle

	// Quote styles
	quote := newParagraphStyle(domain.StyleIDQuote, "Quote", true)
	quote.SetBasedOn(domain.StyleIDNormal)
	quote.SetIndentation(domain.Indentation{Left: 720, Right: 720}) // 0.5 inch
	sm.styles[domain.StyleIDQuote] = quote

	// List Paragraph
	listPara := newParagraphStyle(domain.StyleIDListParagraph, "List Paragraph", true)
	listPara.SetBasedOn(domain.StyleIDNormal)
	listPara.SetIndentation(domain.Indentation{Left: 720})
	sm.styles[domain.StyleIDListParagraph] = listPara

	// TOC styles
	for i := 1; i <= 9; i++ {
		id := ""
		name := ""
		switch i {
		case 1:
			id, name = domain.StyleIDTOC1, "TOC 1"
		case 2:
			id, name = domain.StyleIDTOC2, "TOC 2"
		case 3:
			id, name = domain.StyleIDTOC3, "TOC 3"
		case 4:
			id, name = domain.StyleIDTOC4, "TOC 4"
		case 5:
			id, name = domain.StyleIDTOC5, "TOC 5"
		case 6:
			id, name = domain.StyleIDTOC6, "TOC 6"
		case 7:
			id, name = domain.StyleIDTOC7, "TOC 7"
		case 8:
			id, name = domain.StyleIDTOC8, "TOC 8"
		case 9:
			id, name = domain.StyleIDTOC9, "TOC 9"
		}
		tocStyle := newParagraphStyle(id, name, true)
		tocStyle.SetBasedOn(domain.StyleIDNormal)
		tocStyle.SetIndentation(domain.Indentation{Left: (i - 1) * 220})
		sm.styles[id] = tocStyle
	}

	// Header and Footer
	header := newParagraphStyle(domain.StyleIDHeader, "Header", true)
	header.SetBasedOn(domain.StyleIDNormal)
	sm.styles[domain.StyleIDHeader] = header

	footer := newParagraphStyle(domain.StyleIDFooter, "Footer", true)
	footer.SetBasedOn(domain.StyleIDNormal)
	sm.styles[domain.StyleIDFooter] = footer

	// Body Text variants
	bodyText := newParagraphStyle(domain.StyleIDBodyText, "Body Text", true)
	bodyText.SetBasedOn(domain.StyleIDNormal)
	sm.styles[domain.StyleIDBodyText] = bodyText

	noSpacing := newParagraphStyle(domain.StyleIDNoSpacing, "No Spacing", true)
	noSpacing.SetBasedOn(domain.StyleIDNormal)
	noSpacing.SetSpacingBefore(0)
	noSpacing.SetSpacingAfter(0)
	sm.styles[domain.StyleIDNoSpacing] = noSpacing
}

// createBuiltInCharacterStyles creates all built-in character styles.
// Note: Error returns are intentionally ignored for built-in styles as they
// use only valid, hardcoded values that cannot fail under normal circumstances.
//
//nolint:errcheck // Built-in styles use hardcoded valid values
func (sm *styleManager) createBuiltInCharacterStyles() {
	// Default Paragraph Font (base for all character styles)
	defaultFont := newCharacterStyle(domain.StyleIDDefaultParagraphFont, "Default Paragraph Font", true)
	defaultFont.SetDefault(true)
	sm.styles[domain.StyleIDDefaultParagraphFont] = defaultFont

	// Emphasis (italic)
	emphasis := newCharacterStyle(domain.StyleIDEmphasis, "Emphasis", true)
	emphasis.SetBasedOn(domain.StyleIDDefaultParagraphFont)
	emphasis.SetItalic(true)
	sm.styles[domain.StyleIDEmphasis] = emphasis

	// Strong (bold)
	strong := newCharacterStyle(domain.StyleIDStrong, "Strong", true)
	strong.SetBasedOn(domain.StyleIDDefaultParagraphFont)
	strong.SetBold(true)
	sm.styles[domain.StyleIDStrong] = strong

	// Hyperlink (blue, underlined)
	hyperlink := newCharacterStyle(domain.StyleIDHyperlink, "Hyperlink", true)
	hyperlink.SetBasedOn(domain.StyleIDDefaultParagraphFont)
	hyperlink.SetColor(domain.Color{R: 0, G: 0, B: 255}) // Blue
	hyperlink.SetUnderline(domain.UnderlineSingle)
	sm.styles[domain.StyleIDHyperlink] = hyperlink

	// Followed Hyperlink (purple, underlined)
	followedHyperlink := newCharacterStyle(domain.StyleIDFollowedHyperlink, "Followed Hyperlink", true)
	followedHyperlink.SetBasedOn(domain.StyleIDDefaultParagraphFont)
	followedHyperlink.SetColor(domain.Color{R: 128, G: 0, B: 128}) // Purple
	followedHyperlink.SetUnderline(domain.UnderlineSingle)
	sm.styles[domain.StyleIDFollowedHyperlink] = followedHyperlink
}

// GetStyle retrieves a style by ID.
func (sm *styleManager) GetStyle(styleID string) (domain.Style, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	style, exists := sm.styles[styleID]
	if !exists {
		return nil, errors.NewNotFoundError(
			"StyleManager.GetStyle",
			"style",
			styleID,
			"style not found",
		)
	}

	return style, nil
}

// AddStyle adds a new custom style.
func (sm *styleManager) AddStyle(style domain.Style) error {
	if style == nil {
		return errors.NewValidationError(
			"StyleManager.AddStyle",
			"style",
			nil,
			"style cannot be nil",
		)
	}

	styleID := style.ID()
	if styleID == "" {
		return errors.NewValidationError(
			"StyleManager.AddStyle",
			"style.ID",
			"",
			"style ID cannot be empty",
		)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if it's a built-in style
	if sm.builtInStyleIDs[styleID] {
		return errors.NewValidationError(
			"StyleManager.AddStyle",
			"styleID",
			styleID,
			"cannot override built-in style",
		)
	}

	// Check if already exists
	if _, exists := sm.styles[styleID]; exists {
		return errors.NewValidationError(
			"StyleManager.AddStyle",
			"styleID",
			styleID,
			"style already exists",
		)
	}

	sm.styles[styleID] = style
	return nil
}

// RemoveStyle removes a custom style.
func (sm *styleManager) RemoveStyle(styleID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Cannot remove built-in styles
	if sm.builtInStyleIDs[styleID] {
		return errors.NewValidationError(
			"StyleManager.RemoveStyle",
			"styleID",
			styleID,
			"cannot remove built-in style",
		)
	}

	if _, exists := sm.styles[styleID]; !exists {
		return errors.NewNotFoundError(
			"StyleManager.RemoveStyle",
			"style",
			styleID,
			"style not found",
		)
	}

	delete(sm.styles, styleID)
	return nil
}

// ListStyles returns all available styles.
func (sm *styleManager) ListStyles() []domain.Style {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	styles := make([]domain.Style, 0, len(sm.styles))
	for _, style := range sm.styles {
		styles = append(styles, style)
	}

	return styles
}

// ListStylesByType returns all styles of a specific type.
func (sm *styleManager) ListStylesByType(styleType domain.StyleType) []domain.Style {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	styles := make([]domain.Style, 0)
	for _, style := range sm.styles {
		if style.Type() == styleType {
			styles = append(styles, style)
		}
	}

	return styles
}

// DefaultStyle returns the default style for a type.
func (sm *styleManager) DefaultStyle(styleType domain.StyleType) (domain.Style, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	styleID, exists := sm.defaultStyles[styleType]
	if !exists {
		return nil, errors.NewNotFoundError(
			"StyleManager.DefaultStyle",
			"default style",
			styleType,
			"no default style set for type",
		)
	}

	return sm.styles[styleID], nil
}

// SetDefaultStyle sets the default style for a type.
func (sm *styleManager) SetDefaultStyle(styleType domain.StyleType, styleID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Verify style exists
	style, exists := sm.styles[styleID]
	if !exists {
		return errors.NewNotFoundError(
			"StyleManager.SetDefaultStyle",
			"style",
			styleID,
			"style not found",
		)
	}

	// Verify style type matches
	if style.Type() != styleType {
		return errors.NewValidationError(
			"StyleManager.SetDefaultStyle",
			"styleType",
			styleType,
			"style type mismatch",
		)
	}

	sm.defaultStyles[styleType] = styleID
	return nil
}

// HasStyle checks if a style exists.
func (sm *styleManager) HasStyle(styleID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	_, exists := sm.styles[styleID]
	return exists
}

// IsBuiltIn checks if a style is built-in (not custom).
func (sm *styleManager) IsBuiltIn(styleID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.builtInStyleIDs[styleID]
}
