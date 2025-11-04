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

package color

import (
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func TestToHex(t *testing.T) {
	tests := []struct {
		name     string
		color    domain.Color
		expected string
	}{
		{
			name:     "black",
			color:    domain.Color{R: 0, G: 0, B: 0},
			expected: "000000",
		},
		{
			name:     "white",
			color:    domain.Color{R: 255, G: 255, B: 255},
			expected: "FFFFFF",
		},
		{
			name:     "red",
			color:    domain.Color{R: 255, G: 0, B: 0},
			expected: "FF0000",
		},
		{
			name:     "green",
			color:    domain.Color{R: 0, G: 128, B: 0},
			expected: "008000",
		},
		{
			name:     "blue",
			color:    domain.Color{R: 0, G: 0, B: 255},
			expected: "0000FF",
		},
		{
			name:     "orange",
			color:    domain.Color{R: 255, G: 165, B: 0},
			expected: "FFA500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToHex(tt.color)
			if result != tt.expected {
				t.Errorf("ToHex(%+v) = %s; want %s", tt.color, result, tt.expected)
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		name      string
		hex       string
		expected  domain.Color
		shouldErr bool
	}{
		{
			name:     "full form black",
			hex:      "000000",
			expected: domain.Color{R: 0, G: 0, B: 0},
		},
		{
			name:     "full form white",
			hex:      "FFFFFF",
			expected: domain.Color{R: 255, G: 255, B: 255},
		},
		{
			name:     "full form red",
			hex:      "FF0000",
			expected: domain.Color{R: 255, G: 0, B: 0},
		},
		{
			name:     "full form with hash prefix",
			hex:      "#FF0000",
			expected: domain.Color{R: 255, G: 0, B: 0},
		},
		{
			name:     "short form RGB",
			hex:      "F00",
			expected: domain.Color{R: 255, G: 0, B: 0},
		},
		{
			name:     "short form RGB with hash",
			hex:      "#F00",
			expected: domain.Color{R: 255, G: 0, B: 0},
		},
		{
			name:     "short form gray",
			hex:      "888",
			expected: domain.Color{R: 136, G: 136, B: 136},
		},
		{
			name:     "lowercase hex",
			hex:      "ff0000",
			expected: domain.Color{R: 255, G: 0, B: 0},
		},
		{
			name:      "invalid length",
			hex:       "FF00",
			shouldErr: true,
		},
		{
			name:      "invalid hex characters",
			hex:       "GGGGGG",
			shouldErr: true,
		},
		{
			name:      "empty string",
			hex:       "",
			shouldErr: true,
		},
		{
			name:      "too long",
			hex:       "FF0000FF",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromHex(tt.hex)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("FromHex(%s) expected error but got none", tt.hex)
				}
			} else {
				if err != nil {
					t.Errorf("FromHex(%s) unexpected error: %v", tt.hex, err)
				}
				if result != tt.expected {
					t.Errorf("FromHex(%s) = %+v; want %+v", tt.hex, result, tt.expected)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name  string
		color domain.Color
	}{
		{
			name:  "black",
			color: domain.Color{R: 0, G: 0, B: 0},
		},
		{
			name:  "white",
			color: domain.Color{R: 255, G: 255, B: 255},
		},
		{
			name:  "mid value",
			color: domain.Color{R: 128, G: 128, B: 128},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.color)
			if err != nil {
				t.Errorf("Validate(%+v) unexpected error: %v", tt.color, err)
			}
		})
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are properly defined
	tests := []struct {
		name     string
		color    domain.Color
		expected domain.Color
	}{
		{"Black", Black, domain.Color{R: 0, G: 0, B: 0}},
		{"White", White, domain.Color{R: 255, G: 255, B: 255}},
		{"Red", Red, domain.Color{R: 255, G: 0, B: 0}},
		{"Green", Green, domain.Color{R: 0, G: 128, B: 0}},
		{"Blue", Blue, domain.Color{R: 0, G: 0, B: 255}},
		{"Yellow", Yellow, domain.Color{R: 255, G: 255, B: 0}},
		{"Cyan", Cyan, domain.Color{R: 0, G: 255, B: 255}},
		{"Magenta", Magenta, domain.Color{R: 255, G: 0, B: 255}},
		{"Orange", Orange, domain.Color{R: 255, G: 165, B: 0}},
		{"Purple", Purple, domain.Color{R: 128, G: 0, B: 128}},
		{"Gray", Gray, domain.Color{R: 128, G: 128, B: 128}},
		{"Silver", Silver, domain.Color{R: 192, G: 192, B: 192}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != tt.expected {
				t.Errorf("%s = %+v; want %+v", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestToHexFromHexRoundTrip(t *testing.T) {
	tests := []domain.Color{
		{R: 0, G: 0, B: 0},
		{R: 255, G: 255, B: 255},
		{R: 255, G: 0, B: 0},
		{R: 0, G: 128, B: 0},
		{R: 0, G: 0, B: 255},
		{R: 128, G: 128, B: 128},
	}

	for _, color := range tests {
		t.Run(ToHex(color), func(t *testing.T) {
			hex := ToHex(color)
			result, err := FromHex(hex)
			if err != nil {
				t.Errorf("FromHex(%s) unexpected error: %v", hex, err)
			}
			if result != color {
				t.Errorf("Round trip failed: %+v -> %s -> %+v", color, hex, result)
			}
		})
	}
}
