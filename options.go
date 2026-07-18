// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

// Config contains configuration options for document creation.
type Config struct {
	DefaultFont      string
	DefaultFontSize  int
	PageSize         PageSize
	Margins          Margins
	StrictValidation bool
	Metadata         *domain.Metadata
	Language         *Language
	Theme            interface{} // Theme to apply (using interface{} to avoid import cycle)
}

// Language represents a document's default proofing language, expressed as
// BCP 47 language tags (e.g. "es-MX", "en-US"). See docx.WithLanguage and
// docx.WithLanguageEx.
type Language = domain.Language

// PageSize represents paper dimensions.
type PageSize struct {
	Width  int // in twips (1/1440 inch)
	Height int // in twips
}

// Margins represents page margins.
type Margins struct {
	Top    int // in twips
	Bottom int // in twips
	Left   int // in twips
	Right  int // in twips
}

// Common page sizes in twips (1/1440 inch)
var (
	// A4 is 210mm x 297mm
	A4 = PageSize{Width: 11906, Height: 16838}

	// Letter is 8.5" x 11"
	Letter = PageSize{Width: 12240, Height: 15840}

	// Legal is 8.5" x 14"
	Legal = PageSize{Width: 12240, Height: 20160}

	// A3 is 297mm x 420mm
	A3 = PageSize{Width: 16838, Height: 23811}

	// Tabloid is 11" x 17"
	Tabloid = PageSize{Width: 15840, Height: 24480}
)

// Common margins presets
var (
	// NormalMargins is 1 inch on all sides
	NormalMargins = Margins{
		Top:    1440,
		Bottom: 1440,
		Left:   1440,
		Right:  1440,
	}

	// NarrowMargins is 0.5 inch on all sides
	NarrowMargins = Margins{
		Top:    720,
		Bottom: 720,
		Left:   720,
		Right:  720,
	}

	// WideMargins is 1 inch top/bottom, 2 inch left/right
	WideMargins = Margins{
		Top:    1440,
		Bottom: 1440,
		Left:   2880,
		Right:  2880,
	}
)

// Option is a function that configures a Config.
type Option func(*Config)

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		DefaultFont:      constants.DefaultFontName,
		DefaultFontSize:  constants.DefaultFontSize,
		PageSize:         Letter,
		Margins:          NormalMargins,
		StrictValidation: false,
		Metadata:         &domain.Metadata{},
	}
}

// WithDefaultFont sets the default font for the document.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithDefaultFont("Arial"),
//	)
func WithDefaultFont(font string) Option {
	return func(c *Config) {
		c.DefaultFont = font
	}
}

// WithDefaultFontSize sets the default font size in half-points.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithDefaultFontSize(24), // 12pt
//	)
func WithDefaultFontSize(halfPoints int) Option {
	return func(c *Config) {
		c.DefaultFontSize = halfPoints
	}
}

// WithPageSize sets the page size for the document.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithPageSize(docx.A4),
//	)
func WithPageSize(size PageSize) Option {
	return func(c *Config) {
		c.PageSize = size
	}
}

// WithMargins sets the page margins for the document.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithMargins(docx.NarrowMargins),
//	)
func WithMargins(margins Margins) Option {
	return func(c *Config) {
		c.Margins = margins
	}
}

// WithStrictValidation enables strict validation of the document structure.
// When enabled, Build() will perform more rigorous checks.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithStrictValidation(),
//	)
func WithStrictValidation() Option {
	return func(c *Config) {
		c.StrictValidation = true
	}
}

// WithMetadata sets the document metadata.
//
// Example:
//
//	meta := &domain.Metadata{
//	    Title:   "My Document",
//	    Author:  "John Doe",
//	    Subject: "Report",
//	}
//	builder := docx.NewDocumentBuilder(
//	    docx.WithMetadata(meta),
//	)
func WithMetadata(meta *domain.Metadata) Option {
	return func(c *Config) {
		c.Metadata = meta
	}
}

// WithTitle is a convenience function to set the document title.
func WithTitle(title string) Option {
	return func(c *Config) {
		if c.Metadata == nil {
			c.Metadata = &domain.Metadata{}
		}
		c.Metadata.Title = title
	}
}

// WithAuthor is a convenience function to set the document author.
func WithAuthor(author string) Option {
	return func(c *Config) {
		if c.Metadata == nil {
			c.Metadata = &domain.Metadata{}
		}
		c.Metadata.Creator = author
	}
}

// WithSubject is a convenience function to set the document subject.
func WithSubject(subject string) Option {
	return func(c *Config) {
		if c.Metadata == nil {
			c.Metadata = &domain.Metadata{}
		}
		c.Metadata.Subject = subject
	}
}

// WithLanguage sets the document's default proofing language, using a BCP 47
// language tag. Word uses this to select the spell-checking and grammar
// dictionaries, and to apply the correct hyphenation rules.
//
// NewDocumentBuilder always starts from a new, empty document, so this always
// takes effect. To change the language of an existing .docx opened via
// OpenDocument, use its SetLanguage method instead — note that it errors on a
// document whose styles.xml/settings.xml were preserved for round-trip
// fidelity, since the language could never actually reach the saved file.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithLanguage("es-MX"),
//	)
func WithLanguage(lang string) Option {
	return func(c *Config) {
		c.Language = &Language{Val: lang}
	}
}

// WithLanguageEx sets the document's default proofing language, including
// optional language tags for East Asian (CJK) and right-to-left (bidi)
// scripts. Use this over WithLanguage when the document mixes scripts, e.g.
// Latin text with embedded Japanese or Arabic. At least one of Val, EastAsia,
// or Bidi must be set.
//
// See WithLanguage regarding existing documents opened via OpenDocument.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithLanguageEx(docx.Language{
//	        Val:      "en-US",
//	        EastAsia: "ja-JP",
//	        Bidi:     "ar-SA",
//	    }),
//	)
func WithLanguageEx(lang Language) Option {
	return func(c *Config) {
		// Copy so that reusing this Option across multiple builders doesn't
		// leave their Configs aliasing the same *Language.
		langCopy := lang
		c.Language = &langCopy
	}
}

// WithTheme applies a theme to the document, configuring colors, fonts, and spacing.
// The theme parameter should be a themes.Theme interface.
//
// Example:
//
//	import "github.com/mmonterroca/docxgo/v2/themes"
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithTheme(themes.Corporate),
//	)
func WithTheme(theme interface{}) Option {
	return func(c *Config) {
		c.Theme = theme
	}
}
