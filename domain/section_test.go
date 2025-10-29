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

func TestPageSizeConstants(t *testing.T) {
	tests := []struct {
		name   string
		size   PageSize
		width  int
		height int
	}{
		{"A4", PageSizeA4, 11906, 16838},
		{"Letter", PageSizeLetter, 12240, 15840},
		{"Legal", PageSizeLegal, 12240, 20160},
		{"A3", PageSizeA3, 16838, 23811},
		{"Tabloid", PageSizeTableid, 15840, 24480},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.size.Width != tt.width {
				t.Errorf("%s Width = %d; want %d", tt.name, tt.size.Width, tt.width)
			}
			if tt.size.Height != tt.height {
				t.Errorf("%s Height = %d; want %d", tt.name, tt.size.Height, tt.height)
			}
		})
	}
}

func TestDefaultMargins(t *testing.T) {
	if DefaultMargins.Top != 1440 {
		t.Errorf("DefaultMargins.Top = %d; want 1440", DefaultMargins.Top)
	}
	if DefaultMargins.Right != 1440 {
		t.Errorf("DefaultMargins.Right = %d; want 1440", DefaultMargins.Right)
	}
	if DefaultMargins.Bottom != 1440 {
		t.Errorf("DefaultMargins.Bottom = %d; want 1440", DefaultMargins.Bottom)
	}
	if DefaultMargins.Left != 1440 {
		t.Errorf("DefaultMargins.Left = %d; want 1440", DefaultMargins.Left)
	}
}

func TestSectionBreakTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		breakType SectionBreakType
		value     int
	}{
		{"NextPage", SectionBreakTypeNextPage, 0},
		{"Continuous", SectionBreakTypeContinuous, 1},
		{"EvenPage", SectionBreakTypeEvenPage, 2},
		{"OddPage", SectionBreakTypeOddPage, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.breakType) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.breakType, tt.value)
			}
		})
	}
}

func TestOrientationConstants(t *testing.T) {
	tests := []struct {
		name        string
		orientation Orientation
		value       int
	}{
		{"Portrait", OrientationPortrait, 0},
		{"Landscape", OrientationLandscape, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.orientation) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.orientation, tt.value)
			}
		})
	}
}

func TestHeaderTypeConstants(t *testing.T) {
	tests := []struct {
		name       string
		headerType HeaderType
		value      int
	}{
		{"Default", HeaderDefault, 0},
		{"First", HeaderFirst, 1},
		{"Even", HeaderEven, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.headerType) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.headerType, tt.value)
			}
		})
	}
}

func TestFooterTypeConstants(t *testing.T) {
	tests := []struct {
		name       string
		footerType FooterType
		value      int
	}{
		{"Default", FooterDefault, 0},
		{"First", FooterFirst, 1},
		{"Even", FooterEven, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.footerType) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, tt.footerType, tt.value)
			}
		})
	}
}
