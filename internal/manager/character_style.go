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



package manager

import (
	"sync"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// characterStyle implements domain.CharacterStyle.
type characterStyle struct {
	mu        sync.RWMutex
	id        string
	name      string
	basedOn   string
	font      domain.Font
	isDefault bool
	isBuiltIn bool
	bold      bool
	italic    bool
	underline domain.UnderlineStyle
	color     domain.Color
	size      int // in half-points
}

// newCharacterStyle creates a new character style.
func newCharacterStyle(id, name string, builtIn bool) *characterStyle {
	return &characterStyle{
		id:        id,
		name:      name,
		isBuiltIn: builtIn,
		font:      domain.Font{Name: constants.DefaultFontName},
		color:     domain.ColorBlack,
		size:      constants.DefaultFontSize,
		underline: domain.UnderlineNone,
	}
}

// ID returns the unique style identifier.
func (cs *characterStyle) ID() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.id
}

// Name returns the style display name.
func (cs *characterStyle) Name() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.name
}

// Type returns the style type.
func (cs *characterStyle) Type() domain.StyleType {
	return domain.StyleTypeCharacter
}

// BasedOn returns the style ID this style is based on.
func (cs *characterStyle) BasedOn() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.basedOn
}

// SetBasedOn sets the parent style.
func (cs *characterStyle) SetBasedOn(styleID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.basedOn = styleID
	return nil
}

// Next returns the style ID for next paragraph (not applicable for character styles).
func (cs *characterStyle) Next() string {
	return ""
}

// SetNext sets next paragraph style (not applicable for character styles).
func (cs *characterStyle) SetNext(styleID string) error {
	return errors.NewValidationError(
		"CharacterStyle.SetNext",
		"styleID",
		styleID,
		"Next property not applicable for character styles",
	)
}

// Font returns the font settings.
func (cs *characterStyle) Font() domain.Font {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.font
}

// SetFont sets the font settings.
func (cs *characterStyle) SetFont(font domain.Font) error {
	if font.Name == "" {
		return errors.NewValidationError(
			"CharacterStyle.SetFont",
			"font.Name",
			"",
			"font name cannot be empty",
		)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.font = font
	return nil
}

// IsDefault returns whether this is a default style.
func (cs *characterStyle) IsDefault() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.isDefault
}

// SetDefault marks this style as default.
func (cs *characterStyle) SetDefault(isDefault bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.isDefault = isDefault
	return nil
}

// IsCustom returns whether this is a custom style.
func (cs *characterStyle) IsCustom() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return !cs.isBuiltIn
}

// Bold returns whether the text is bold.
func (cs *characterStyle) Bold() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.bold
}

// SetBold sets whether the text is bold.
func (cs *characterStyle) SetBold(bold bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.bold = bold
	return nil
}

// Italic returns whether the text is italic.
func (cs *characterStyle) Italic() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.italic
}

// SetItalic sets whether the text is italic.
func (cs *characterStyle) SetItalic(italic bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.italic = italic
	return nil
}

// Underline returns the underline style.
func (cs *characterStyle) Underline() domain.UnderlineStyle {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.underline
}

// SetUnderline sets the underline style.
func (cs *characterStyle) SetUnderline(style domain.UnderlineStyle) error {
	if style < domain.UnderlineNone || style > domain.UnderlineWave {
		return errors.NewValidationError(
			"CharacterStyle.SetUnderline",
			"style",
			style,
			"invalid underline style",
		)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.underline = style
	return nil
}

// Color returns the text color.
func (cs *characterStyle) Color() domain.Color {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.color
}

// SetColor sets the text color.
func (cs *characterStyle) SetColor(color domain.Color) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.color = color
	return nil
}

// Size returns the font size in half-points.
func (cs *characterStyle) Size() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.size
}

// SetSize sets the font size in half-points.
func (cs *characterStyle) SetSize(halfPoints int) error {
	if halfPoints < constants.MinFontSize || halfPoints > constants.MaxFontSize {
		return errors.NewValidationError(
			"CharacterStyle.SetSize",
			"halfPoints",
			halfPoints,
			"font size must be between 2 and 3276 half-points (1pt - 1638pt)",
		)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.size = halfPoints
	return nil
}
