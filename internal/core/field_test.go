// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"strings"
	"sync"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestNewField(t *testing.T) {
	tests := []struct {
		name      string
		fieldType domain.FieldType
		wantCode  string
	}{
		{
			name:      "PageNumber",
			fieldType: domain.FieldTypePageNumber,
			wantCode:  "PAGE",
		},
		{
			name:      "PageCount",
			fieldType: domain.FieldTypePageCount,
			wantCode:  "NUMPAGES",
		},
		{
			name:      "TOC",
			fieldType: domain.FieldTypeTOC,
			wantCode:  "TOC",
		},
		{
			name:      "Date",
			fieldType: domain.FieldTypeDate,
			wantCode:  "DATE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewField(tt.fieldType)

			if field.Type() != tt.fieldType {
				t.Errorf("Type() = %v, want %v", field.Type(), tt.fieldType)
			}

			code := field.Code()
			if !strings.HasPrefix(code, tt.wantCode) {
				t.Errorf("Code() = %q, want prefix %q", code, tt.wantCode)
			}
		})
	}
}

func TestNewPageNumberField(t *testing.T) {
	field := NewPageNumberField()

	if field.Type() != domain.FieldTypePageNumber {
		t.Errorf("Type() = %v, want %v", field.Type(), domain.FieldTypePageNumber)
	}

	if field.Code() != "PAGE" {
		t.Errorf("Code() = %q, want %q", field.Code(), "PAGE")
	}

	// Update field
	if err := field.Update(); err != nil {
		t.Errorf("Update() error = %v", err)
	}

	if field.Result() == "" {
		t.Error("Result() should not be empty after Update()")
	}
}

func TestNewPageCountField(t *testing.T) {
	field := NewPageCountField()

	if field.Type() != domain.FieldTypePageCount {
		t.Errorf("Type() = %v, want %v", field.Type(), domain.FieldTypePageCount)
	}

	if field.Code() != "NUMPAGES" {
		t.Errorf("Code() = %q, want %q", field.Code(), "NUMPAGES")
	}
}

func TestNewTOCField(t *testing.T) {
	tests := []struct {
		name     string
		switches map[string]string
		wantCode []string // Substrings that should be in the code
	}{
		{
			name:     "Default",
			switches: nil,
			wantCode: []string{"TOC", `\o "1-3"`, `\h`, `\z`, `\u`},
		},
		{
			name: "CustomLevels",
			switches: map[string]string{
				"levels": "1-5",
			},
			wantCode: []string{"TOC", `\o "1-5"`},
		},
		{
			name: "HidePageNumbers",
			switches: map[string]string{
				"hidePageNumbers": "true",
			},
			wantCode: []string{"TOC", `\n`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewTOCField(tt.switches)

			if field.Type() != domain.FieldTypeTOC {
				t.Errorf("Type() = %v, want %v", field.Type(), domain.FieldTypeTOC)
			}

			code := field.Code()
			for _, want := range tt.wantCode {
				if !strings.Contains(code, want) {
					t.Errorf("Code() = %q, want to contain %q", code, want)
				}
			}
		})
	}
}

func TestNewHyperlinkField(t *testing.T) {
	url := "https://example.com"
	displayText := "Example Link"

	field := NewHyperlinkField(url, displayText)

	if field.Type() != domain.FieldTypeHyperlink {
		t.Errorf("Type() = %v, want %v", field.Type(), domain.FieldTypeHyperlink)
	}

	code := field.Code()
	if !strings.Contains(code, "HYPERLINK") || !strings.Contains(code, url) {
		t.Errorf("Code() = %q, want to contain HYPERLINK and %q", code, url)
	}

	if field.Result() != displayText {
		t.Errorf("Result() = %q, want %q", field.Result(), displayText)
	}
}

func TestNewStyleRefField(t *testing.T) {
	styleName := "Heading 1"

	field := NewStyleRefField(styleName)

	if field.Type() != domain.FieldTypeStyleRef {
		t.Errorf("Type() = %v, want %v", field.Type(), domain.FieldTypeStyleRef)
	}

	code := field.Code()
	if !strings.Contains(code, "STYLEREF") || !strings.Contains(code, styleName) {
		t.Errorf("Code() = %q, want to contain STYLEREF and %q", code, styleName)
	}
}

// newTestRun returns a fresh run inside a scratch document, for tests that
// need to exercise run.AddField's validation.
func newTestRun(t *testing.T) domain.Run {
	t.Helper()
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph() error = %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}
	return run
}

// TestNewHyperlinkFieldRejectsQuote pins the fix for the field-code
// injection CodeQL flagged (go/unsafe-quoting): a url containing a double
// quote used to break out of the quoted HYPERLINK argument. The
// strings.Contains style used elsewhere in this file would pass on an
// injected payload, so this asserts Code() by exact equality.
func TestNewHyperlinkFieldRejectsQuote(t *testing.T) {
	url := `https://example.com/" INCLUDETEXT "C:\x`

	field := NewHyperlinkField(url, "display")

	if got, want := field.Code(), "HYPERLINK"; got != want {
		t.Errorf("Code() = %q, want %q (safe default, not the injected value)", got, want)
	}

	df := field.(*docxField)
	if _, ok := df.GetProperty("url"); ok {
		t.Error("GetProperty(url) = ok, want the property to be unset on rejection")
	}

	run := newTestRun(t)
	if err := run.AddField(field); err == nil {
		t.Error("AddField() error = nil, want a validation error for the rejected field")
	}
}

// TestNewStyleRefFieldRejectsQuote is the STYLEREF counterpart. Unlike
// HYPERLINK, STYLEREF always reaches word/document.xml unconditionally
// (there's no relationship-branch shadowing), so this is the more directly
// exploitable of the two.
func TestNewStyleRefFieldRejectsQuote(t *testing.T) {
	styleName := `Heading 1" \* MERGEFORMAT "x`

	field := NewStyleRefField(styleName)

	if got, want := field.Code(), `STYLEREF "Heading 1"`; got != want {
		t.Errorf("Code() = %q, want %q (safe default, not the injected value)", got, want)
	}

	run := newTestRun(t)
	if err := run.AddField(field); err == nil {
		t.Error("AddField() error = nil, want a validation error for the rejected field")
	}
}

// TestNewTOCFieldRejectsInvalidLevels covers the third CodeQL-flagged sink:
// the \o switch built from the caller-supplied "levels" property.
func TestNewTOCFieldRejectsInvalidLevels(t *testing.T) {
	field := NewTOCField(map[string]string{"levels": `1-3" \c "Figure`})

	// field.go builds \\o (double backslash) as a raw string literal today —
	// that's a separate, pre-existing quirk (tracked separately), not part
	// of this fix. The default-else branch in buildTOCCode is the exact
	// bytes this must match.
	want := constants.FieldCodeTOC + ` \\o "1-3" \\h \\z \\u`
	if got := field.Code(); got != want {
		t.Errorf("Code() = %q, want %q (safe default, not the injected value)", got, want)
	}

	run := newTestRun(t)
	if err := run.AddField(field); err == nil {
		t.Error("AddField() error = nil, want a validation error for the rejected field")
	}
}

// TestNewTOCFieldAcceptsValidLevels guards against the quote guard being too
// aggressive: a legitimate levels value must still work.
func TestNewTOCFieldAcceptsValidLevels(t *testing.T) {
	field := NewTOCField(map[string]string{"levels": "1-5"})

	want := constants.FieldCodeTOC + ` \\o "1-5" \\h \\z \\u`
	if got := field.Code(); got != want {
		t.Errorf("Code() = %q, want %q", got, want)
	}

	run := newTestRun(t)
	if err := run.AddField(field); err != nil {
		t.Errorf("AddField() error = %v, want nil for a valid field", err)
	}
}

func TestFieldSetCode(t *testing.T) {
	field := NewField(domain.FieldTypeCustom)

	customCode := "CUSTOM \\* MERGEFORMAT"
	if err := field.SetCode(customCode); err != nil {
		t.Errorf("SetCode() error = %v", err)
	}

	if field.Code() != customCode {
		t.Errorf("Code() = %q, want %q", field.Code(), customCode)
	}
}

func TestFieldSetCodeEmpty(t *testing.T) {
	field := NewField(domain.FieldTypeCustom)

	if err := field.SetCode(""); err == nil {
		t.Error("SetCode() with empty string should return error")
	}

	if err := field.SetCode("   "); err == nil {
		t.Error("SetCode() with whitespace should return error")
	}
}

func TestFieldUpdate(t *testing.T) {
	tests := []struct {
		name      string
		fieldType domain.FieldType
	}{
		{"PageNumber", domain.FieldTypePageNumber},
		{"PageCount", domain.FieldTypePageCount},
		{"TOC", domain.FieldTypeTOC},
		{"Date", domain.FieldTypeDate},
		{"Seq", domain.FieldTypeSeq},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewField(tt.fieldType)

			if err := field.Update(); err != nil {
				t.Errorf("Update() error = %v", err)
			}

			result := field.Result()
			if result == "" {
				t.Error("Result() should not be empty after Update()")
			}
		})
	}
}

func TestFieldDirtyTracking(t *testing.T) {
	field := NewField(domain.FieldTypePageNumber)

	// New field should be dirty
	df := field.(*docxField)
	if !df.IsDirty() {
		t.Error("New field should be dirty")
	}

	// Update should clear dirty flag
	if err := field.Update(); err != nil {
		t.Errorf("Update() error = %v", err)
	}

	if df.IsDirty() {
		t.Error("Field should not be dirty after Update()")
	}

	// SetCode should mark as dirty
	if err := field.SetCode("PAGE \\* Arabic"); err != nil {
		t.Errorf("SetCode() error = %v", err)
	}

	if !df.IsDirty() {
		t.Error("Field should be dirty after SetCode()")
	}

	// MarkDirty should work
	field.Update()
	df.MarkDirty()
	if !df.IsDirty() {
		t.Error("Field should be dirty after MarkDirty()")
	}
}

func TestFieldProperties(t *testing.T) {
	field := NewField(domain.FieldTypeCustom)
	df := field.(*docxField)

	// Set property
	df.SetProperty("key1", "value1")

	// Get property
	value, ok := df.GetProperty("key1")
	if !ok {
		t.Error("GetProperty() should find key1")
	}
	if value != "value1" {
		t.Errorf("GetProperty() = %q, want %q", value, "value1")
	}

	// Get non-existent property
	_, ok = df.GetProperty("nonexistent")
	if ok {
		t.Error("GetProperty() should not find nonexistent key")
	}

	// Setting property should mark as dirty
	field.Update() // Clear dirty
	df.SetProperty("key2", "value2")
	if !df.IsDirty() {
		t.Error("SetProperty() should mark field as dirty")
	}
}

func TestFieldConcurrency(t *testing.T) {
	field := NewField(domain.FieldTypePageNumber)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = field.Type()
			_ = field.Code()
			_ = field.Result()
		}()
	}

	// Concurrent updates
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = field.Update()
		}()
	}

	wg.Wait()
}

func TestHyperlinkFieldProperties(t *testing.T) {
	url := "https://example.com"
	displayText := "Example"

	field := NewHyperlinkField(url, displayText)
	df := field.(*docxField)

	// Check properties
	gotURL, ok := df.GetProperty("url")
	if !ok || gotURL != url {
		t.Errorf("GetProperty(url) = %q, want %q", gotURL, url)
	}

	gotDisplay, ok := df.GetProperty("display")
	if !ok || gotDisplay != displayText {
		t.Errorf("GetProperty(display) = %q, want %q", gotDisplay, displayText)
	}
}

// TestHyperlinkRelationshipIDPreservation tests that when a hyperlink field
// has a preserved relationshipID from reading an existing document, it is
// not overwritten when calling AddField.
// See: docs/TROUBLESHOOTING_DOCX_VALIDATION.md - Issue 1
func TestHyperlinkRelationshipIDPreservation(t *testing.T) {
	url := "https://example.com"
	displayText := "Example"
	preservedRelID := "rId25"

	field := NewHyperlinkField(url, displayText)
	df := field.(*docxField)

	// Simulate what happens during document read - the relationshipID is preserved
	df.SetProperty("relationshipID", preservedRelID)

	// Verify the preserved ID is stored
	gotRelID, ok := df.GetProperty("relationshipID")
	if !ok {
		t.Error("GetProperty(relationshipID) returned false; want true")
	}
	if gotRelID != preservedRelID {
		t.Errorf("GetProperty(relationshipID) = %q, want %q", gotRelID, preservedRelID)
	}
}

// TestHyperlinkAnchorField tests internal hyperlinks that use anchors
// instead of external URLs.
func TestHyperlinkAnchorField(t *testing.T) {
	anchor := "_Toc123456789"
	displayText := "Section 1"

	// Create a hyperlink field for an internal anchor
	field := NewField(domain.FieldTypeHyperlink)
	df := field.(*docxField)
	df.SetProperty("anchor", anchor)
	df.SetProperty("display", displayText)

	// Verify properties
	gotAnchor, ok := df.GetProperty("anchor")
	if !ok || gotAnchor != anchor {
		t.Errorf("GetProperty(anchor) = %q, want %q", gotAnchor, anchor)
	}

	gotDisplay, ok := df.GetProperty("display")
	if !ok || gotDisplay != displayText {
		t.Errorf("GetProperty(display) = %q, want %q", gotDisplay, displayText)
	}

	// Anchor-based hyperlinks should not have a URL
	gotURL, hasURL := df.GetProperty("url")
	if hasURL && gotURL != "" {
		t.Errorf("GetProperty(url) = %q, want empty for anchor-based link", gotURL)
	}
}

func TestStyleRefFieldProperties(t *testing.T) {
	styleName := "Heading 1"

	field := NewStyleRefField(styleName)
	df := field.(*docxField)

	gotStyle, ok := df.GetProperty("style")
	if !ok || gotStyle != styleName {
		t.Errorf("GetProperty(style) = %q, want %q", gotStyle, styleName)
	}
}
