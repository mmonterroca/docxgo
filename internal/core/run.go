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

// Package core provides core implementations of domain interfaces.
package core

import (
	"github.com/SlideLang/go-docx/domain"
	"github.com/SlideLang/go-docx/pkg/constants"
	"github.com/SlideLang/go-docx/pkg/errors"
)

// run implements the domain.Run interface.
type run struct {
	id        string
	text      string
	font      domain.Font
	color     domain.Color
	size      int // in half-points
	bold      bool
	italic    bool
	underline domain.UnderlineStyle
	strike    bool
	highlight domain.HighlightColor
	fields    []domain.Field // Fields embedded in this run
}

// NewRun creates a new Run.
func NewRun(id string) domain.Run {
	return &run{
		id:        id,
		font:      domain.Font{Name: constants.DefaultFontName},
		color:     domain.ColorBlack,
		size:      constants.DefaultFontSize,
		underline: domain.UnderlineNone,
		highlight: domain.HighlightNone,
	}
}

// Text returns the text content of this run.
func (r *run) Text() string {
	return r.text
}

// SetText sets the text content of this run.
func (r *run) SetText(text string) error {
	r.text = text
	return nil
}

// Font returns the font settings for this run.
func (r *run) Font() domain.Font {
	return r.font
}

// SetFont sets the font for this run.
func (r *run) SetFont(font domain.Font) error {
	if font.Name == "" {
		return errors.InvalidArgument("Run.SetFont", "font.Name", font.Name, "font name cannot be empty")
	}
	r.font = font
	return nil
}

// Color returns the text color.
func (r *run) Color() domain.Color {
	return r.color
}

// SetColor sets the text color.
func (r *run) SetColor(color domain.Color) error {
	// Color validation is implicit via uint8 type (0-255)
	r.color = color
	return nil
}

// Size returns the font size in half-points.
func (r *run) Size() int {
	return r.size
}

// SetSize sets the font size in half-points.
func (r *run) SetSize(halfPoints int) error {
	if halfPoints < constants.MinFontSize || halfPoints > constants.MaxFontSize {
		return errors.InvalidArgument("Run.SetSize", "halfPoints", halfPoints,
			"font size must be between 2 and 3276 half-points (1pt - 1638pt)")
	}
	r.size = halfPoints
	return nil
}

// Bold returns whether the text is bold.
func (r *run) Bold() bool {
	return r.bold
}

// SetBold sets whether the text is bold.
func (r *run) SetBold(bold bool) error {
	r.bold = bold
	return nil
}

// Italic returns whether the text is italic.
func (r *run) Italic() bool {
	return r.italic
}

// SetItalic sets whether the text is italic.
func (r *run) SetItalic(italic bool) error {
	r.italic = italic
	return nil
}

// Underline returns the underline style.
func (r *run) Underline() domain.UnderlineStyle {
	return r.underline
}

// SetUnderline sets the underline style.
func (r *run) SetUnderline(style domain.UnderlineStyle) error {
	// Validate underline style
	if style < domain.UnderlineNone || style > domain.UnderlineWave {
		return errors.InvalidArgument("Run.SetUnderline", "style", style,
			"invalid underline style")
	}
	r.underline = style
	return nil
}

// Strike returns whether the text is struck through.
func (r *run) Strike() bool {
	return r.strike
}

// SetStrike sets whether the text is struck through.
func (r *run) SetStrike(strike bool) error {
	r.strike = strike
	return nil
}

// Highlight returns the highlight color.
func (r *run) Highlight() domain.HighlightColor {
	return r.highlight
}

// SetHighlight sets the highlight color.
func (r *run) SetHighlight(color domain.HighlightColor) error {
	// Validate highlight color
	if color < domain.HighlightNone || color > domain.HighlightLightGray {
		return errors.InvalidArgument("Run.SetHighlight", "color", color,
			"invalid highlight color")
	}
	r.highlight = color
	return nil
}

// AddText is a convenience method that appends text to the run.
func (r *run) AddText(text string) error {
	r.text += text
	return nil
}

// AddField adds a field to this run (e.g., page number, TOC, hyperlink).
func (r *run) AddField(field domain.Field) error {
	if field == nil {
		return errors.InvalidArgument("Run.AddField", "field", nil, "field cannot be nil")
	}
	
	if r.fields == nil {
		r.fields = make([]domain.Field, 0, 2)
	}
	
	r.fields = append(r.fields, field)
	return nil
}

// Fields returns all fields in this run.
func (r *run) Fields() []domain.Field {
	if r.fields == nil {
		return nil
	}
	// Return a defensive copy
	result := make([]domain.Field, len(r.fields))
	copy(result, r.fields)
	return result
}
