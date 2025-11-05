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

func TestNewImageSize(t *testing.T) {
	tests := []struct {
		name     string
		widthPx  int
		heightPx int
		wantEMUW int
		wantEMUH int
	}{
		{
			name:     "100x100 pixels",
			widthPx:  100,
			heightPx: 100,
			wantEMUW: 952500, // 100 * 9525
			wantEMUH: 952500,
		},
		{
			name:     "640x480 pixels",
			widthPx:  640,
			heightPx: 480,
			wantEMUW: 6096000, // 640 * 9525
			wantEMUH: 4572000, // 480 * 9525
		},
		{
			name:     "zero dimensions",
			widthPx:  0,
			heightPx: 0,
			wantEMUW: 0,
			wantEMUH: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := NewImageSize(tt.widthPx, tt.heightPx)

			if size.WidthPx != tt.widthPx {
				t.Errorf("WidthPx = %d; want %d", size.WidthPx, tt.widthPx)
			}
			if size.HeightPx != tt.heightPx {
				t.Errorf("HeightPx = %d; want %d", size.HeightPx, tt.heightPx)
			}
			if size.WidthEMU != tt.wantEMUW {
				t.Errorf("WidthEMU = %d; want %d", size.WidthEMU, tt.wantEMUW)
			}
			if size.HeightEMU != tt.wantEMUH {
				t.Errorf("HeightEMU = %d; want %d", size.HeightEMU, tt.wantEMUH)
			}
		})
	}
}

func TestNewImageSizeInches(t *testing.T) {
	tests := []struct {
		name          string
		widthInches   float64
		heightInches  float64
		wantWidthPx   int
		wantHeightPx  int
		wantWidthEMU  int
		wantHeightEMU int
	}{
		{
			name:          "1x1 inches",
			widthInches:   1.0,
			heightInches:  1.0,
			wantWidthPx:   96, // 1 * 96 DPI
			wantHeightPx:  96,
			wantWidthEMU:  914400, // 1 * 914400 EMU/inch
			wantHeightEMU: 914400,
		},
		{
			name:          "2.5x3.5 inches",
			widthInches:   2.5,
			heightInches:  3.5,
			wantWidthPx:   240,     // 2.5 * 96
			wantHeightPx:  336,     // 3.5 * 96
			wantWidthEMU:  2286000, // 2.5 * 914400
			wantHeightEMU: 3200400, // 3.5 * 914400
		},
		{
			name:          "zero dimensions",
			widthInches:   0.0,
			heightInches:  0.0,
			wantWidthPx:   0,
			wantHeightPx:  0,
			wantWidthEMU:  0,
			wantHeightEMU: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := NewImageSizeInches(tt.widthInches, tt.heightInches)

			if size.WidthPx != tt.wantWidthPx {
				t.Errorf("WidthPx = %d; want %d", size.WidthPx, tt.wantWidthPx)
			}
			if size.HeightPx != tt.wantHeightPx {
				t.Errorf("HeightPx = %d; want %d", size.HeightPx, tt.wantHeightPx)
			}
			if size.WidthEMU != tt.wantWidthEMU {
				t.Errorf("WidthEMU = %d; want %d", size.WidthEMU, tt.wantWidthEMU)
			}
			if size.HeightEMU != tt.wantHeightEMU {
				t.Errorf("HeightEMU = %d; want %d", size.HeightEMU, tt.wantHeightEMU)
			}
		})
	}
}

func TestDefaultImagePosition(t *testing.T) {
	pos := DefaultImagePosition()

	if pos.Type != ImagePositionInline {
		t.Errorf("Type = %v; want %v", pos.Type, ImagePositionInline)
	}
	if pos.WrapText != WrapNone {
		t.Errorf("WrapText = %v; want %v", pos.WrapText, WrapNone)
	}
	if pos.ZOrder != 0 {
		t.Errorf("ZOrder = %d; want 0", pos.ZOrder)
	}
	if pos.BehindText {
		t.Error("BehindText = true; want false")
	}
}

func TestImageFormatConstants(t *testing.T) {
	tests := []struct {
		name   string
		format ImageFormat
		value  string
	}{
		{"PNG", ImageFormatPNG, "png"},
		{"JPEG", ImageFormatJPEG, "jpeg"},
		{"JPG", ImageFormatJPG, "jpg"},
		{"GIF", ImageFormatGIF, "gif"},
		{"BMP", ImageFormatBMP, "bmp"},
		{"TIFF", ImageFormatTIFF, "tiff"},
		{"TIF", ImageFormatTIF, "tif"},
		{"SVG", ImageFormatSVG, "svg"},
		{"WEBP", ImageFormatWEBP, "webp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.value {
				t.Errorf("%s format = %s; want %s", tt.name, tt.format, tt.value)
			}
		})
	}
}

func TestImagePositionTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		pos   ImagePositionType
		value string
	}{
		{"Inline", ImagePositionInline, "inline"},
		{"Floating", ImagePositionFloating, "floating"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.pos) != tt.value {
				t.Errorf("%s = %s; want %s", tt.name, tt.pos, tt.value)
			}
		})
	}
}

func TestHorizontalAlignConstants(t *testing.T) {
	tests := []struct {
		name  string
		align HorizontalAlign
		value string
	}{
		{"Left", HAlignLeft, "left"},
		{"Center", HAlignCenter, "center"},
		{"Right", HAlignRight, "right"},
		{"Inside", HAlignInside, "inside"},
		{"Outside", HAlignOutside, "outside"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.align) != tt.value {
				t.Errorf("%s = %s; want %s", tt.name, tt.align, tt.value)
			}
		})
	}
}

func TestVerticalAlignConstants(t *testing.T) {
	tests := []struct {
		name  string
		align VerticalAlign
		value string
	}{
		{"Top", VAlignTop, "top"},
		{"Center", VAlignCenter, "center"},
		{"Bottom", VAlignBottom, "bottom"},
		{"Inside", VAlignInside, "inside"},
		{"Outside", VAlignOutside, "outside"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.align) != tt.value {
				t.Errorf("%s = %s; want %s", tt.name, tt.align, tt.value)
			}
		})
	}
}

func TestTextWrapTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		wrap  TextWrapType
		value string
	}{
		{"None", WrapNone, "none"},
		{"Square", WrapSquare, "square"},
		{"Tight", WrapTight, "tight"},
		{"Through", WrapThrough, "through"},
		{"TopBottom", WrapTopBottom, "topBottom"},
		{"BehindText", WrapBehindText, "behindText"},
		{"InFrontText", WrapInFrontText, "inFrontText"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.wrap) != tt.value {
				t.Errorf("%s = %s; want %s", tt.name, tt.wrap, tt.value)
			}
		})
	}
}
