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
	LineSpacingAuto LineSpacingRule = iota // Auto (value = 240 = single spacing)
	LineSpacingExact                       // Exact (value in twips)
	LineSpacingAtLeast                     // At least (value in twips)
)

// FieldType represents different field types in Word.
type FieldType int

const (
	FieldTypeTOC FieldType = iota
	FieldTypePageNumber
	FieldTypeNumPages
	FieldTypeDate
	FieldTypeTime
	FieldTypeStyleRef
	FieldTypeRef
	FieldTypeSeq
)
