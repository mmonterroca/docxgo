// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// docxField implements the Field interface.
type docxField struct {
	mu         sync.RWMutex
	fieldType  domain.FieldType
	code       string
	result     string
	isDirty    bool // Indicates if field needs recalculation
	properties map[string]string
	// err records a rejection from one of the New*Field constructors — a
	// caller-supplied value that would have broken out of the field code's
	// quoted argument (e.g. a `"` in a hyperlink URL or a STYLEREF style
	// name). The field is left with its safe default code; run.AddField
	// checks this via ValidationError and refuses to attach the field.
	err error
}

// NewField creates a new field with the specified type.
func NewField(fieldType domain.FieldType) domain.Field {
	field := &docxField{
		fieldType:  fieldType,
		isDirty:    true,
		properties: make(map[string]string),
	}

	// Set default code based on type
	field.code = field.getDefaultCode()

	return field
}

// NewPageNumberField creates a field for page numbering.
func NewPageNumberField() domain.Field {
	field := NewField(domain.FieldTypePageNumber)
	return field
}

// NewPageCountField creates a field for total page count.
func NewPageCountField() domain.Field {
	field := NewField(domain.FieldTypePageCount)
	return field
}

// NewTOCField creates a Table of Contents field with options.
func NewTOCField(switches map[string]string) domain.Field {
	field := NewField(domain.FieldTypeTOC).(*docxField)

	// Apply switches
	for key, value := range switches {
		field.properties[key] = value
	}

	// Rebuild code with switches
	code, err := field.buildTOCCode()
	if err != nil {
		field.err = err
		return field
	}
	field.code = code

	return field
}

// NewHyperlinkField creates a hyperlink field.
//
// url must not contain a double quote: it is interpolated inside a quoted
// field-code argument, and Word's field-code syntax has no escape the
// docxgo reader can round-trip (see extractQuotedStrings in
// internal/reader/reconstruct.go). A url containing one is rejected — the
// field is left with a safe default code and no url/display property, and
// run.AddField refuses to attach it — rather than producing a corrupted or
// injectable field code.
func NewHyperlinkField(url, displayText string) domain.Field {
	field := NewField(domain.FieldTypeHyperlink).(*docxField)
	safe := strings.ReplaceAll(url, `"`, "")
	if safe != url {
		field.err = errors.NewValidationError("NewHyperlinkField", "url", url,
			`url must not contain a double quote`)
		return field
	}
	field.properties["url"] = url
	field.properties["display"] = displayText
	// A freshly created internal (#anchor) link defaults to w:history="1" --
	// matching this package's long-standing output for that case. This must
	// be decided here, at construction, not in the serializer: hydrateHyperlink
	// (internal/reader/reconstruct.go) builds a field for a link read back from
	// a source document without ever calling this constructor, so a property
	// set here can never leak onto a hydrated field whose source omitted
	// w:history -- which is exactly what must round-trip as omitted, not "1".
	// External (non-#) links keep no default; that asymmetry predates this
	// comment and is out of scope for it.
	if strings.HasPrefix(url, "#") {
		field.properties["history"] = "1"
	}
	field.code = fmt.Sprintf(`HYPERLINK "%s"`, safe)
	field.result = displayText
	field.isDirty = false
	return field
}

// NewStyleRefField creates a STYLEREF field. styleName is subject to the
// same quoting constraint as NewHyperlinkField's url.
func NewStyleRefField(styleName string) domain.Field {
	field := NewField(domain.FieldTypeStyleRef).(*docxField)
	safe := strings.ReplaceAll(styleName, `"`, "")
	if safe != styleName {
		field.err = errors.NewValidationError("NewStyleRefField", "styleName", styleName,
			`styleName must not contain a double quote`)
		return field
	}
	field.properties["style"] = styleName
	field.code = fmt.Sprintf(`STYLEREF "%s"`, safe)
	return field
}

// ValidationError reports a rejection recorded by NewHyperlinkField,
// NewStyleRefField, or NewTOCField. run.AddField checks this via a type
// assertion and refuses to attach a field that failed construction.
func (f *docxField) ValidationError() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.err
}

// Type returns the field type.
func (f *docxField) Type() domain.FieldType {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fieldType
}

// Code returns the field code.
func (f *docxField) Code() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.code
}

// SetCode sets the field code. code must be a complete, trusted field
// instruction: SetCode performs no quote-sanitization the way the New*Field
// constructors do (see docxField.err), because a well-formed instruction
// legitimately contains balanced quotes (e.g. `STYLEREF "Heading 1"`) that a
// naive quote check can't tell apart from an injection attempt. Untrusted,
// caller-supplied fragments (a URL, a style name, ...) must go through one of
// the New*Field constructors instead, which sanitize the fragment before
// building the code around it.
//
// The one thing SetCode does reject is a control character or line break:
// <w:instrText> holds a single-line instruction, and Word's field-code parser
// has no defined behavior for one containing a raw control byte.
func (f *docxField) SetCode(code string) error {
	if strings.TrimSpace(code) == "" {
		return errors.NewValidationError(
			"SetCode",
			"code",
			code,
			"field code cannot be empty",
		)
	}
	for _, r := range code {
		if r < 0x20 {
			return errors.NewValidationError(
				"SetCode",
				"code",
				code,
				"field code must not contain control characters or line breaks",
			)
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.code = code
	f.isDirty = true

	return nil
}

// SetCodeRaw hydrates the field's code directly, bypassing SetCode's
// control-character guard. Used exclusively by the reader package during
// OpenDocument to reflect a field instruction already present in the source
// file's <w:instrText> — a real .docx from any producer is the ground truth
// for what a "valid" instruction looks like, not this package's own
// validation, so OpenDocument must not fail on one merely because it
// contains a byte SetCode wouldn't have accepted from a caller. Not part of
// domain.Field, reached only via the reader's own type-assertion (mirrors
// SetLanguageRaw / SetPreservedStylesPart).
func (f *docxField) SetCodeRaw(code string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.code = code
	f.isDirty = true
}

// Result returns the field result (calculated value).
func (f *docxField) Result() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.result
}

// Update recalculates the field result.
func (f *docxField) Update() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isDirty {
		return nil // No update needed
	}

	// Update based on field type
	switch f.fieldType {
	case domain.FieldTypePageNumber:
		f.result = "1" // Placeholder - actual value determined by Word
	case domain.FieldTypePageCount:
		f.result = "1" // Placeholder - actual value determined by Word
	case domain.FieldTypeNumPages:
		f.result = "1" // Placeholder - actual value determined by Word
	case domain.FieldTypeTOC:
		f.result = "Table of Contents" // Placeholder
	case domain.FieldTypeStyleRef:
		f.result = "" // Placeholder - populated by Word
	case domain.FieldTypeHyperlink:
		// Result is the display text
		if display, ok := f.properties["display"]; ok {
			f.result = display
		}
	case domain.FieldTypeDate:
		f.result = "1/1/2025" // Placeholder
	case domain.FieldTypeTime:
		f.result = "12:00:00" // Placeholder
	case domain.FieldTypeSeq:
		f.result = "1" // Placeholder
	case domain.FieldTypeRef:
		f.result = "" // Placeholder - populated by Word
	case domain.FieldTypeCustom:
		f.result = "" // Custom fields have user-defined results
	default:
		f.result = "" // Unknown field type
	}

	f.isDirty = false
	return nil
}

// SetResult allows callers to inject a pre-rendered field result.
// This is primarily useful for examples so the initial document shows
// friendly placeholder text before Word updates the field values.
func (f *docxField) SetResult(result string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.result = result
	f.isDirty = false
}

// getDefaultCode returns the default field code for the field type.
func (f *docxField) getDefaultCode() string {
	switch f.fieldType {
	case domain.FieldTypePageNumber:
		return constants.FieldCodePageNumber
	case domain.FieldTypePageCount:
		return constants.FieldCodeNumPages
	case domain.FieldTypeNumPages:
		return constants.FieldCodeNumPages
	case domain.FieldTypeTOC:
		return constants.FieldCodeTOC + ` \o "1-3" \h \z \u`
	case domain.FieldTypeDate:
		return constants.FieldCodeDate
	case domain.FieldTypeTime:
		return constants.FieldCodeTime
	case domain.FieldTypeStyleRef:
		return constants.FieldCodeStyleRef + ` "Heading 1"`
	case domain.FieldTypeSeq:
		return constants.FieldCodeSeq + ` Figure`
	case domain.FieldTypeRef:
		return constants.FieldCodeRef
	case domain.FieldTypeHyperlink:
		return "HYPERLINK" // Hyperlink fields use HYPERLINK code
	case domain.FieldTypeCustom:
		return "" // Custom fields require user-defined codes
	default:
		return ""
	}
}

// buildTOCCode builds a TOC field code with switches. levels is subject to
// the same quoting constraint as NewHyperlinkField's url.
func (f *docxField) buildTOCCode() (string, error) {
	code := constants.FieldCodeTOC

	// Heading levels (default 1-3)
	if levels, ok := f.properties["levels"]; ok {
		safe := strings.ReplaceAll(levels, `"`, "")
		if safe != levels {
			return "", errors.NewValidationError("NewTOCField", "levels", levels,
				`levels must not contain a double quote`)
		}
		code += fmt.Sprintf(` \o "%s"`, safe)
	} else {
		code += ` \o "1-3"`
	}

	// Hyperlinks (always included by default)
	code += ` \h`

	// Hide page numbers (only if explicitly set to "true")
	if hidePN, ok := f.properties["hidePageNumbers"]; ok && hidePN == "true" {
		code += ` \n`
	}

	// Hide tab leader (only if explicitly set to "true")
	if hideTab, ok := f.properties["hideTabLeader"]; ok && hideTab == "true" {
		code += ` \p`
	}

	// Preserve tab entries
	code += ` \z`

	// Use styles
	code += ` \u`

	return code, nil
}

// SetProperty sets a field property (for advanced customization).
func (f *docxField) SetProperty(key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.properties[key] = value
	f.isDirty = true
}

// GetProperty gets a field property.
func (f *docxField) GetProperty(key string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	value, ok := f.properties[key]
	return value, ok
}

// IsDirty returns whether the field needs updating.
func (f *docxField) IsDirty() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isDirty
}

// MarkDirty marks the field as needing an update.
func (f *docxField) MarkDirty() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.isDirty = true
}
