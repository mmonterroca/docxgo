// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package domain

import "testing"

func TestAlignmentConstants(t *testing.T) {
	tests := []struct {
		name      string
		alignment Alignment
		value     int
	}{
		{"Left", AlignmentLeft, 0},
		{"Center", AlignmentCenter, 1},
		{"Right", AlignmentRight, 2},
		{"Justify", AlignmentJustify, 3},
		{"Distribute", AlignmentDistribute, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.alignment) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.alignment, tt.value)
			}
		})
	}
}

func TestLineSpacingRuleConstants(t *testing.T) {
	tests := []struct {
		name  string
		rule  LineSpacingRule
		value int
	}{
		{"Auto", LineSpacingAuto, 0},
		{"Exact", LineSpacingExact, 1},
		{"AtLeast", LineSpacingAtLeast, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.rule) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.rule, tt.value)
			}
		})
	}
}

func TestFieldTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		value     int
	}{
		{"TOC", FieldTypeTOC, 0},
		{"PageNumber", FieldTypePageNumber, 1},
		{"NumPages", FieldTypeNumPages, 2},
		{"PageCount", FieldTypePageCount, 3},
		{"Date", FieldTypeDate, 4},
		{"Time", FieldTypeTime, 5},
		{"StyleRef", FieldTypeStyleRef, 6},
		{"Ref", FieldTypeRef, 7},
		{"Seq", FieldTypeSeq, 8},
		{"Hyperlink", FieldTypeHyperlink, 9},
		{"Custom", FieldTypeCustom, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.fieldType) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.fieldType, tt.value)
			}
		})
	}
}
