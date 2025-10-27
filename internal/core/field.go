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



package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// docxField implements the Field interface.
type docxField struct {
	mu         sync.RWMutex
	fieldType  domain.FieldType
	code       string
	result     string
	isDirty    bool // Indicates if field needs recalculation
	properties map[string]string
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
	field.code = field.buildTOCCode()
	
	return field
}

// NewHyperlinkField creates a hyperlink field.
func NewHyperlinkField(url, displayText string) domain.Field {
	field := NewField(domain.FieldTypeHyperlink).(*docxField)
	field.properties["url"] = url
	field.properties["display"] = displayText
	field.code = fmt.Sprintf(`HYPERLINK "%s"`, url)
	field.result = displayText
	field.isDirty = false
	return field
}

// NewStyleRefField creates a STYLEREF field.
func NewStyleRefField(styleName string) domain.Field {
	field := NewField(domain.FieldTypeStyleRef).(*docxField)
	field.properties["style"] = styleName
	field.code = fmt.Sprintf(`STYLEREF "%s"`, styleName)
	return field
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

// SetCode sets the field code.
func (f *docxField) SetCode(code string) error {
	if strings.TrimSpace(code) == "" {
		return errors.NewValidationError(
			"SetCode",
			"code",
			code,
			"field code cannot be empty",
		)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.code = code
	f.isDirty = true
	
	return nil
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
		return constants.FieldCodeTOC + ` \\o "1-3" \\h \\z \\u`
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

// buildTOCCode builds a TOC field code with switches.
func (f *docxField) buildTOCCode() string {
	code := constants.FieldCodeTOC

	// Heading levels (default 1-3)
	if levels, ok := f.properties["levels"]; ok {
		code += fmt.Sprintf(` \\o "%s"`, levels)
	} else {
		code += ` \\o "1-3"`
	}

	// Hyperlinks (always included by default)
	code += ` \\h`

	// Hide page numbers
	if _, ok := f.properties["hidePageNumbers"]; ok {
		code += ` \\n`
	}

	// Hide tab leader
	if _, ok := f.properties["hideTabLeader"]; ok {
		code += ` \\p`
	}

	// Preserve tab entries
	code += ` \\z`

	// Use styles
	code += ` \\u`

	return code
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
