/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)
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

package docx

import (
	"errors"
	"testing"
)

func TestValidateIndent(t *testing.T) {
	tests := []struct {
		name        string
		left        int
		firstLine   int
		hanging     int
		expectError bool
		errorType   error
	}{
		{
			name:        "valid left indent",
			left:        720,
			firstLine:   0,
			hanging:     0,
			expectError: false,
		},
		{
			name:        "valid with firstLine",
			left:        720,
			firstLine:   360,
			hanging:     0,
			expectError: false,
		},
		{
			name:        "valid with hanging",
			left:        720,
			firstLine:   0,
			hanging:     360,
			expectError: false,
		},
		{
			name:        "left too large",
			left:        MaxIndentTwips + 1,
			firstLine:   0,
			hanging:     0,
			expectError: true,
		},
		{
			name:        "left too small",
			left:        MinIndentTwips - 1,
			firstLine:   0,
			hanging:     0,
			expectError: true,
		},
		{
			name:        "both firstLine and hanging",
			left:        720,
			firstLine:   360,
			hanging:     360,
			expectError: true,
			errorType:   ErrConflictingIndent,
		},
		{
			name:        "negative firstLine",
			left:        720,
			firstLine:   -100,
			hanging:     0,
			expectError: true,
		},
		{
			name:        "negative hanging",
			left:        720,
			firstLine:   0,
			hanging:     -100,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIndent(tt.left, tt.firstLine, tt.hanging)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("expected error %v but got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateJustification(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
	}{
		{"valid start", AlignLeft, false},
		{"valid center", AlignCenter, false},
		{"valid end", AlignRight, false},
		{"valid both", AlignBoth, false},
		{"valid distribute", AlignDistribute, false},
		{"invalid value", "invalid", true},
		{"empty value", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJustification(tt.value)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUnderline(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
	}{
		{"valid none", UnderlineNone, false},
		{"valid single", UnderlineSingle, false},
		{"valid words", UnderlineWords, false},
		{"valid double", UnderlineDouble, false},
		{"valid thick", UnderlineThick, false},
		{"valid dotted", UnderlineDotted, false},
		{"valid dash", UnderlineDash, false},
		{"valid dotDash", UnderlineDotDash, false},
		{"valid dotDotDash", UnderlineDotDotDash, false},
		{"valid wave", UnderlineWave, false},
		{"valid dashLong", UnderlineDashLong, false},
		{"valid wavyDouble", UnderlineWavyDouble, false},
		{"invalid value", "invalid", true},
		{"empty value", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUnderline(tt.value)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateColor(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
	}{
		{"valid hex", "#FF0000", false},
		{"valid hex lowercase", "#ff0000", false},
		{"valid name", "red", false},
		{"invalid hex short", "#FFF", true},
		{"invalid hex long", "#FF00000", true},
		{"empty value", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColor(tt.value)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDocxError(t *testing.T) {
	err := &DocxError{
		Op:      "AddParagraph",
		Err:     errors.New("something went wrong"),
		Context: "nil document",
	}

	expected := "docx: AddParagraph: nil document: something went wrong"
	if err.Error() != expected {
		t.Errorf("expected %q but got %q", expected, err.Error())
	}

	// Test Unwrap
	if !errors.Is(err, err.Err) {
		t.Errorf("Unwrap should return the underlying error")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:      "left",
		Value:      40000,
		Constraint: "must be between -31680 and 31680 twips",
	}

	if err.Error() == "" {
		t.Errorf("ValidationError.Error() should not be empty")
	}

	t.Logf("ValidationError: %v", err)
}

func TestIsValidationError(t *testing.T) {
	err := newValidationError("test", 123, "constraint")
	if !IsValidationError(err) {
		t.Errorf("IsValidationError should return true for ValidationError")
	}

	otherErr := errors.New("not a validation error")
	if IsValidationError(otherErr) {
		t.Errorf("IsValidationError should return false for non-ValidationError")
	}
}

func TestIsDocxError(t *testing.T) {
	err := newDocxError("test", errors.New("test"), "context")
	if !IsDocxError(err) {
		t.Errorf("IsDocxError should return true for DocxError")
	}

	otherErr := errors.New("not a docx error")
	if IsDocxError(otherErr) {
		t.Errorf("IsDocxError should return false for non-DocxError")
	}
}

func TestConstants(t *testing.T) {
	// Verify twips constants
	if TwipsPerInch != 1440 {
		t.Errorf("TwipsPerInch should be 1440, got %d", TwipsPerInch)
	}

	if TwipsPerHalfInch != 720 {
		t.Errorf("TwipsPerHalfInch should be 720, got %d", TwipsPerHalfInch)
	}

	if TwipsPerQuarterInch != 360 {
		t.Errorf("TwipsPerQuarterInch should be 360, got %d", TwipsPerQuarterInch)
	}

	if TwipsPerPoint != 20 {
		t.Errorf("TwipsPerPoint should be 20, got %d", TwipsPerPoint)
	}

	// Verify indent limits
	if MaxIndentTwips != 31680 {
		t.Errorf("MaxIndentTwips should be 31680, got %d", MaxIndentTwips)
	}

	if MinIndentTwips != -31680 {
		t.Errorf("MinIndentTwips should be -31680, got %d", MinIndentTwips)
	}

	// Verify capacity constants
	if DefaultParagraphCapacity != 64 {
		t.Errorf("DefaultParagraphCapacity should be 64, got %d", DefaultParagraphCapacity)
	}
}
