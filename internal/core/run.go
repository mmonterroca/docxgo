// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package core provides core implementations of domain interfaces.
package core

import (
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// run implements the domain.Run interface.
type run struct {
	id         string
	text       string
	image      domain.Image
	font       domain.Font
	color      domain.Color
	size       int // in half-points
	bold       bool
	italic     bool
	underline  domain.UnderlineStyle
	strike     bool
	caps       bool
	highlight  domain.HighlightColor
	language   *domain.Language
	fields     []domain.Field     // Fields embedded in this run
	breaks     []domain.BreakType // Breaks in this run
	relManager *manager.RelationshipManager
}

// NewRun creates a new Run.
func NewRun(id string, relManager *manager.RelationshipManager) domain.Run {
	return &run{
		id:         id,
		font:       domain.Font{Name: constants.DefaultFontName},
		color:      domain.ColorBlack,
		size:       constants.DefaultFontSize,
		underline:  domain.UnderlineNone,
		highlight:  domain.HighlightNone,
		relManager: relManager,
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

// Image returns the image associated with this run, if any.
func (r *run) Image() domain.Image {
	return r.image
}

// setImage attaches an image to the run for serialization.
func (r *run) setImage(img domain.Image) {
	r.image = img
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

// Caps returns whether the text is displayed in all capitals.
func (r *run) Caps() bool {
	return r.caps
}

// SetCaps sets whether the text is displayed in all capitals.
func (r *run) SetCaps(caps bool) error {
	r.caps = caps
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

// Language returns a copy of the run's language override, or nil if unset.
// Mutating the returned value has no effect on the run.
func (r *run) Language() *domain.Language {
	if r.language == nil {
		return nil
	}
	langCopy := *r.language
	return &langCopy
}

// SetLanguage sets a per-run language override.
func (r *run) SetLanguage(lang *domain.Language) error {
	if lang != nil && lang.Val == "" && lang.EastAsia == "" && lang.Bidi == "" {
		return errors.InvalidArgument("Run.SetLanguage", "lang", lang, "at least one of Val, EastAsia, or Bidi must be set")
	}
	if lang == nil {
		r.language = nil
		return nil
	}
	langCopy := *lang
	r.language = &langCopy
	return nil
}

// AddText is a convenience method that appends text to the run.
func (r *run) AddText(text string) error {
	r.text += text
	return nil
}

// AddBreak adds a break to the run (page, column, or line break).
func (r *run) AddBreak(breakType domain.BreakType) error {
	if breakType < domain.BreakTypePage || breakType > domain.BreakTypeLine {
		return errors.InvalidArgument("Run.AddBreak", "breakType", breakType, "invalid break type")
	}

	if r.breaks == nil {
		r.breaks = make([]domain.BreakType, 0, 1)
	}

	r.breaks = append(r.breaks, breakType)
	return nil
}

// Breaks returns all breaks in this run.
func (r *run) Breaks() []domain.BreakType {
	if r.breaks == nil {
		return nil
	}
	// Return a defensive copy
	result := make([]domain.BreakType, len(r.breaks))
	copy(result, r.breaks)
	return result
}

// AddField adds a field to this run (e.g., page number, TOC, hyperlink).
func (r *run) AddField(field domain.Field) error {
	if field == nil {
		return errors.InvalidArgument("Run.AddField", "field", nil, "field cannot be nil")
	}

	if validator, ok := field.(interface{ ValidationError() error }); ok {
		if err := validator.ValidationError(); err != nil {
			return errors.Wrap(err, "Run.AddField")
		}
	}

	if field.Type() == domain.FieldTypeHyperlink {
		accessor, ok := field.(interface {
			GetProperty(string) (string, bool)
			SetProperty(string, string)
		})
		if !ok {
			return errors.InvalidArgument("Run.AddField", "field", field, "hyperlink field must support property access")
		}

		// Check if this hyperlink already has a relationship ID (preserved from read)
		// If so, skip creating a new one to preserve original document references
		existingRelID, hasExistingRelID := accessor.GetProperty("relationshipID")
		url, hasURL := accessor.GetProperty("url")
		isAnchor := hasURL && strings.HasPrefix(url, "#")

		switch {
		case hasExistingRelID && existingRelID != "":
			// Already has a relationship ID, skip creating new one.
		case isAnchor:
			// An internal link (w:anchor) has no relationship of its own --
			// the serializer resolves it from the "url"/"anchor" property
			// directly (see expandRunWithFields). Minting an External
			// relationship whose target is "#anchor" would only leave an
			// orphaned entry in the .rels part.
		default:
			// No existing relationship ID and not an internal anchor: create one.
			if r.relManager == nil {
				return errors.InvalidState("Run.AddField", "hyperlink relationship manager not initialized")
			}

			if !hasURL || url == "" {
				return errors.InvalidArgument("Run.AddField", "url", url, "hyperlink URL cannot be empty")
			}

			relID, err := r.relManager.AddHyperlink(url)
			if err != nil {
				return errors.Wrap(err, "Run.AddField")
			}

			accessor.SetProperty("relationshipID", relID)
		}
	}

	if r.fields == nil {
		r.fields = make([]domain.Field, 0, 2)
	}

	r.fields = append(r.fields, field)
	return nil
}

// ClearFields removes all fields from this run.
func (r *run) ClearFields() {
	r.fields = nil
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
