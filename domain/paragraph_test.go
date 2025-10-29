/*
MIT License

Copyright (c) 2025 Misael Monterroca

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
