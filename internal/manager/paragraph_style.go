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
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// paragraphStyle implements domain.ParagraphStyle.
type paragraphStyle struct {
	mu              sync.RWMutex
	id              string
	name            string
	basedOn         string
	next            string
	font            domain.Font
	isDefault       bool
	isBuiltIn       bool
	alignment       domain.Alignment
	spacingBefore   int
	spacingAfter    int
	lineSpacing     int
	indentation     domain.Indentation
	keepNext        bool
	keepLines       bool
	pageBreakBefore bool
	outlineLevel    int
	runBold         bool
	runItalic       bool
	runUnderline    domain.UnderlineStyle
	runColor        domain.Color
	runSize         int
}

// newParagraphStyle creates a new paragraph style.
// Note: builtIn parameter is used in tests to create custom styles.
//
//nolint:unparam // builtIn=true in production, false in tests
func newParagraphStyle(id, name string, builtIn bool) *paragraphStyle {
	return &paragraphStyle{
		id:            id,
		name:          name,
		isBuiltIn:     builtIn,
		font:          domain.Font{Name: constants.DefaultFontName},
		alignment:     domain.AlignmentLeft,
		spacingBefore: 0,
		spacingAfter:  0,
		lineSpacing:   240, // Single spacing (240 = 1.0)
		outlineLevel:  0,   // Body text level
		runUnderline:  domain.UnderlineNone,
		runColor:      domain.ColorBlack,
		runSize:       constants.DefaultFontSize,
	}
}

// ID returns the unique style identifier.
func (ps *paragraphStyle) ID() string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.id
}

// Name returns the style display name.
func (ps *paragraphStyle) Name() string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.name
}

// Type returns the style type.
func (ps *paragraphStyle) Type() domain.StyleType {
	return domain.StyleTypeParagraph
}

// BasedOn returns the style ID this style is based on.
func (ps *paragraphStyle) BasedOn() string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.basedOn
}

// SetBasedOn sets the parent style.
func (ps *paragraphStyle) SetBasedOn(styleID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.basedOn = styleID
	return nil
}

// Next returns the style ID to use for the next paragraph.
func (ps *paragraphStyle) Next() string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.next
}

// SetNext sets the next paragraph style.
func (ps *paragraphStyle) SetNext(styleID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.next = styleID
	return nil
}

// Font returns the font settings.
func (ps *paragraphStyle) Font() domain.Font {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.font
}

// SetFont sets the font settings.
func (ps *paragraphStyle) SetFont(font domain.Font) error {
	if font.Name == "" {
		return errors.NewValidationError(
			"ParagraphStyle.SetFont",
			"font.Name",
			"",
			"font name cannot be empty",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.font = font
	return nil
}

// IsDefault returns whether this is a default style.
func (ps *paragraphStyle) IsDefault() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.isDefault
}

// SetDefault marks this style as default.
func (ps *paragraphStyle) SetDefault(isDefault bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.isDefault = isDefault
	return nil
}

// IsCustom returns whether this is a custom style.
func (ps *paragraphStyle) IsCustom() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return !ps.isBuiltIn
}

// Alignment returns the paragraph alignment.
func (ps *paragraphStyle) Alignment() domain.Alignment {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.alignment
}

// SetAlignment sets the paragraph alignment.
func (ps *paragraphStyle) SetAlignment(align domain.Alignment) error {
	if align < domain.AlignmentLeft || align > domain.AlignmentDistribute {
		return errors.NewValidationError(
			"ParagraphStyle.SetAlignment",
			"alignment",
			align,
			"invalid alignment value",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.alignment = align
	return nil
}

// SpacingBefore returns spacing before paragraph in twips.
func (ps *paragraphStyle) SpacingBefore() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.spacingBefore
}

// SetSpacingBefore sets spacing before paragraph.
func (ps *paragraphStyle) SetSpacingBefore(twips int) error {
	if twips < 0 {
		return errors.NewValidationError(
			"ParagraphStyle.SetSpacingBefore",
			"twips",
			twips,
			"spacing cannot be negative",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.spacingBefore = twips
	return nil
}

// SpacingAfter returns spacing after paragraph in twips.
func (ps *paragraphStyle) SpacingAfter() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.spacingAfter
}

// SetSpacingAfter sets spacing after paragraph.
func (ps *paragraphStyle) SetSpacingAfter(twips int) error {
	if twips < 0 {
		return errors.NewValidationError(
			"ParagraphStyle.SetSpacingAfter",
			"twips",
			twips,
			"spacing cannot be negative",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.spacingAfter = twips
	return nil
}

// LineSpacing returns the line spacing value.
func (ps *paragraphStyle) LineSpacing() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.lineSpacing
}

// SetLineSpacing sets the line spacing.
func (ps *paragraphStyle) SetLineSpacing(value int) error {
	if value < 0 {
		return errors.NewValidationError(
			"ParagraphStyle.SetLineSpacing",
			"value",
			value,
			"line spacing cannot be negative",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.lineSpacing = value
	return nil
}

// Indentation returns the paragraph indentation.
func (ps *paragraphStyle) Indentation() domain.Indentation {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.indentation
}

// SetIndentation sets the paragraph indentation.
func (ps *paragraphStyle) SetIndentation(indent domain.Indentation) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.indentation = indent
	return nil
}

// KeepNext returns whether to keep with next paragraph.
func (ps *paragraphStyle) KeepNext() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.keepNext
}

// SetKeepNext sets keep with next paragraph.
func (ps *paragraphStyle) SetKeepNext(keep bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.keepNext = keep
	return nil
}

// KeepLines returns whether to keep lines together.
func (ps *paragraphStyle) KeepLines() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.keepLines
}

// SetKeepLines sets keep lines together.
func (ps *paragraphStyle) SetKeepLines(keep bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.keepLines = keep
	return nil
}

// PageBreakBefore returns whether to insert page break before.
func (ps *paragraphStyle) PageBreakBefore() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.pageBreakBefore
}

// SetPageBreakBefore sets page break before.
func (ps *paragraphStyle) SetPageBreakBefore(breakBefore bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.pageBreakBefore = breakBefore
	return nil
}

// OutlineLevel returns the outline level.
func (ps *paragraphStyle) OutlineLevel() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.outlineLevel
}

// SetOutlineLevel sets the outline level.
func (ps *paragraphStyle) SetOutlineLevel(level int) error {
	if level < 0 || level > 9 {
		return errors.NewValidationError(
			"ParagraphStyle.SetOutlineLevel",
			"level",
			level,
			"outline level must be 0-9 (0=body text, 1-9=heading levels)",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.outlineLevel = level
	return nil
}

// Bold returns whether the default run is bold.
func (ps *paragraphStyle) Bold() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.runBold
}

// SetBold sets bold on the default run formatting.
func (ps *paragraphStyle) SetBold(bold bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.runBold = bold
	return nil
}

// Italic returns whether the default run is italic.
func (ps *paragraphStyle) Italic() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.runItalic
}

// SetItalic sets italic on the default run formatting.
func (ps *paragraphStyle) SetItalic(italic bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.runItalic = italic
	return nil
}

// Underline returns the underline style for the default run.
func (ps *paragraphStyle) Underline() domain.UnderlineStyle {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.runUnderline
}

// SetUnderline sets underline style for the default run.
func (ps *paragraphStyle) SetUnderline(style domain.UnderlineStyle) error {
	if style < domain.UnderlineNone || style > domain.UnderlineWave {
		return errors.NewValidationError(
			"ParagraphStyle.SetUnderline",
			"style",
			style,
			"invalid underline style",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.runUnderline = style
	return nil
}

// Color returns the default run color.
func (ps *paragraphStyle) Color() domain.Color {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.runColor
}

// SetColor sets the default run color.
func (ps *paragraphStyle) SetColor(color domain.Color) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.runColor = color
	return nil
}

// Size returns the default run font size in half-points.
func (ps *paragraphStyle) Size() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.runSize
}

// SetSize sets the default run font size in half-points.
func (ps *paragraphStyle) SetSize(halfPoints int) error {
	if halfPoints < constants.MinFontSize || halfPoints > constants.MaxFontSize {
		return errors.NewValidationError(
			"ParagraphStyle.SetSize",
			"halfPoints",
			halfPoints,
			"font size must be between 2 and 3276 half-points (1pt - 1638pt)",
		)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.runSize = halfPoints
	return nil
}
