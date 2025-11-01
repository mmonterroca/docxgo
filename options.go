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

package docx

import (
	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/constants"
)

// Config contains configuration options for document creation.
type Config struct {
	DefaultFont      string
	DefaultFontSize  int
	PageSize         PageSize
	Margins          Margins
	StrictValidation bool
	Metadata         *domain.Metadata
	Theme            interface{} // Theme to apply (using interface{} to avoid import cycle)
}

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

// WithTheme applies a theme to the document, configuring colors, fonts, and spacing.
// The theme parameter should be a themes.Theme interface.
//
// Example:
//
//	import "github.com/mmonterroca/docxgo/themes"
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithTheme(themes.Corporate),
//	)
func WithTheme(theme interface{}) Option {
	return func(c *Config) {
		c.Theme = theme
	}
}
