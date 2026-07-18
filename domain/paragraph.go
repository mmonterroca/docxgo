// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

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

	// AddImageFromBytes adds an image from raw byte data.
	// The format parameter specifies the image type (e.g., ImageFormatPNG).
	AddImageFromBytes(data []byte, format ImageFormat) (Image, error)

	// AddImageFromBytesWithSize adds an image from byte data with custom dimensions.
	AddImageFromBytesWithSize(data []byte, format ImageFormat, size ImageSize) (Image, error)

	// AddImageFromBytesWithPosition adds an image from byte data with custom positioning.
	AddImageFromBytesWithPosition(data []byte, format ImageFormat, size ImageSize, pos ImagePosition) (Image, error)

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

	// Numbering returns the numbering reference applied to this paragraph, if any.
	Numbering() (NumberingReference, bool)

	// SetNumbering applies a numbering reference (list id + level) to the paragraph.
	SetNumbering(ref NumberingReference) error

	// ClearNumbering removes any numbering reference from the paragraph.
	ClearNumbering()

	// Borders returns the paragraph borders.
	Borders() ParagraphBorders

	// SetBorders sets all paragraph borders at once.
	SetBorders(borders ParagraphBorders) error

	// SetBorderTop sets the top border.
	SetBorderTop(border BorderStyle) error

	// SetBorderBottom sets the bottom border.
	SetBorderBottom(border BorderStyle) error

	// SetBorderLeft sets the left border.
	SetBorderLeft(border BorderStyle) error

	// SetBorderRight sets the right border.
	SetBorderRight(border BorderStyle) error

	// ClearRuns removes all runs from the paragraph.
	ClearRuns()

	// RemoveRun removes the run at the given index.
	RemoveRun(index int) error

	// InsertRunAt inserts a new empty run at the given index and returns it.
	// Index must be in [0, len(Runs())].
	InsertRunAt(index int) (Run, error)
}

// ParagraphBorders represents borders for a paragraph.
type ParagraphBorders struct {
	Top    BorderStyle
	Bottom BorderStyle
	Left   BorderStyle
	Right  BorderStyle
}

// Alignment represents horizontal alignment options for paragraphs.
type Alignment int

// Paragraph alignment constants.
const (
	AlignmentLeft       Alignment = iota // Left-aligned (default)
	AlignmentCenter                      // Center-aligned
	AlignmentRight                       // Right-aligned
	AlignmentJustify                     // Justified (both left and right)
	AlignmentDistribute                  // Distributed evenly
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

// NumberingReference identifies a numbering instance and level for a paragraph.
type NumberingReference struct {
	ID    int // numId value from numbering.xml
	Level int // ilvl value (0-8)
}

// Numbering level bounds supported by Word numbering definitions.
const (
	NumberingLevelMin = 0
	NumberingLevelMax = 8
)

// LineSpacingRule defines how line spacing is calculated.
type LineSpacingRule int

// Line spacing rule constants.
const (
	LineSpacingAuto    LineSpacingRule = iota // Auto spacing (value = 240 = single spacing)
	LineSpacingExact                          // Exact spacing (value in twips)
	LineSpacingAtLeast                        // At least spacing (value in twips, can expand)
)

// FieldType represents different field types in Word documents.
type FieldType int

// Field type constants for dynamic document fields.
const (
	FieldTypeTOC        FieldType = iota // Table of Contents
	FieldTypePageNumber                  // Current page number
	FieldTypeNumPages                    // Total number of pages
	FieldTypePageCount                   // Alias for NumPages
	FieldTypeDate                        // Current date
	FieldTypeTime                        // Current time
	FieldTypeStyleRef                    // Style reference
	FieldTypeRef                         // Cross-reference
	FieldTypeSeq                         // Sequence number
	FieldTypeHyperlink                   // Hyperlink field
	FieldTypeCustom                      // Custom field with user-defined code
)
