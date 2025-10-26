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



// Package domain defines the core interfaces for go-docx v2.
package domain

// Paragraph represents a paragraph in a document.
// A paragraph contains one or more runs of formatted text.
type Paragraph interface {
	// AddRun adds a new text run to the paragraph.
	AddRun() (Run, error)

	// AddField adds a field to the paragraph (TOC, page number, etc.)
	AddField(fieldType FieldType) (Field, error)

	// AddHyperlink adds a hyperlink to the paragraph.
	AddHyperlink(url, displayText string) (Run, error)

	// AddImage adds an image to the paragraph from a file path.
	// Returns the created Image object.
	AddImage(path string) (Image, error)

	// AddImageWithSize adds an image with custom dimensions.
	// If width or height is 0, maintains aspect ratio.
	AddImageWithSize(path string, size ImageSize) (Image, error)

	// AddImageWithPosition adds an image with custom positioning.
	AddImageWithPosition(path string, size ImageSize, pos ImagePosition) (Image, error)

	// Images returns all images in this paragraph.
	Images() []Image

	// Runs returns all runs in this paragraph.
	Runs() []Run

	// Fields returns all fields in this paragraph.
	Fields() []Field

	// Text returns the plain text content of the paragraph.
	Text() string

	// Style returns the style applied to this paragraph.
	Style() Style

	// SetStyle applies a named style to the paragraph.
	SetStyle(styleName string) error

	// Alignment returns the paragraph's horizontal alignment.
	Alignment() Alignment

	// SetAlignment sets the paragraph's horizontal alignment.
	SetAlignment(align Alignment) error

	// Indent returns the paragraph's indentation settings.
	Indent() Indentation

	// SetIndent sets the paragraph's indentation.
	SetIndent(indent Indentation) error

	// SpacingBefore returns spacing before the paragraph (in twips).
	SpacingBefore() int

	// SetSpacingBefore sets spacing before the paragraph.
	SetSpacingBefore(twips int) error

	// SpacingAfter returns spacing after the paragraph (in twips).
	SpacingAfter() int

	// SetSpacingAfter sets spacing after the paragraph.
	SetSpacingAfter(twips int) error

	// LineSpacing returns the line spacing setting.
	LineSpacing() LineSpacing

	// SetLineSpacing sets the line spacing.
	SetLineSpacing(spacing LineSpacing) error
}

// Alignment represents horizontal alignment options.
type Alignment int

const (
	AlignmentLeft Alignment = iota
	AlignmentCenter
	AlignmentRight
	AlignmentJustify
	AlignmentDistribute
)

// Indentation represents paragraph indentation settings.
type Indentation struct {
	Left      int // Left indent in twips
	Right     int // Right indent in twips
	FirstLine int // First line indent in twips (positive)
	Hanging   int // Hanging indent in twips (positive)
}

// LineSpacing represents line spacing settings.
type LineSpacing struct {
	Rule  LineSpacingRule
	Value int // Meaning depends on Rule
}

// LineSpacingRule defines how line spacing is calculated.
type LineSpacingRule int

const (
	LineSpacingAuto    LineSpacingRule = iota // Auto (value = 240 = single spacing)
	LineSpacingExact                          // Exact (value in twips)
	LineSpacingAtLeast                        // At least (value in twips)
)

// FieldType represents different field types in Word.
type FieldType int

const (
	FieldTypeTOC FieldType = iota
	FieldTypePageNumber
	FieldTypeNumPages
	FieldTypePageCount // Alias for NumPages
	FieldTypeDate
	FieldTypeTime
	FieldTypeStyleRef
	FieldTypeRef
	FieldTypeSeq
	FieldTypeHyperlink
	FieldTypeCustom // Custom field with user-defined code
)
