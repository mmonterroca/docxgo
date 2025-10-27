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

package core

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/domain"
)

func createTestImage(t *testing.T, width, height int) string {
	t.Helper()

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with a gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			})
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_image.png")
	
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	return filePath
}

func TestNewImage(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		wantFormat  domain.ImageFormat
		expectError bool
	}{
		{
			name:        "valid PNG 100x100",
			width:       100,
			height:      100,
			wantFormat:  domain.ImageFormatPNG,
			expectError: false,
		},
		{
			name:        "valid PNG 800x600",
			width:       800,
			height:      600,
			wantFormat:  domain.ImageFormatPNG,
			expectError: false,
		},
		{
			name:        "small image 10x10",
			width:       10,
			height:      10,
			wantFormat:  domain.ImageFormatPNG,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imgPath := createTestImage(t, tt.width, tt.height)

			img, err := NewImage("img1", imgPath)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewImage() error = %v", err)
			}

			if img == nil {
				t.Fatal("NewImage() returned nil")
			}

			// Check ID
			if img.ID() != "img1" {
				t.Errorf("ID() = %v, want %v", img.ID(), "img1")
			}

			// Check format
			if img.Format() != tt.wantFormat {
				t.Errorf("Format() = %v, want %v", img.Format(), tt.wantFormat)
			}

			// Check size
			size := img.Size()
			if size.WidthPx != tt.width {
				t.Errorf("Size().WidthPx = %v, want %v", size.WidthPx, tt.width)
			}
			if size.HeightPx != tt.height {
				t.Errorf("Size().HeightPx = %v, want %v", size.HeightPx, tt.height)
			}

			// Check EMU conversion (914400 EMUs per inch, assuming 96 DPI)
			expectedWidthEMU := tt.width * 9525
			expectedHeightEMU := tt.height * 9525
			if size.WidthEMU != expectedWidthEMU {
				t.Errorf("Size().WidthEMU = %v, want %v", size.WidthEMU, expectedWidthEMU)
			}
			if size.HeightEMU != expectedHeightEMU {
				t.Errorf("Size().HeightEMU = %v, want %v", size.HeightEMU, expectedHeightEMU)
			}

			// Check data is loaded
			data := img.Data()
			if len(data) == 0 {
				t.Error("Data() returned empty slice")
			}

			// Check target path contains image ID
			target := img.Target()
			if !strings.Contains(target, "image"+img.ID()) {
				t.Errorf("Target() = %v, should contain 'image%s'", target, img.ID())
			}
		})
	}
}

func TestNewImageWithSize(t *testing.T) {
	imgPath := createTestImage(t, 800, 600)
	
	customSize := domain.NewImageSize(400, 300)
	img, err := NewImageWithSize("img2", imgPath, customSize)
	if err != nil {
		t.Fatalf("NewImageWithSize() error = %v", err)
	}

	size := img.Size()
	if size.WidthPx != 400 {
		t.Errorf("Size().WidthPx = %v, want 400", size.WidthPx)
	}
	if size.HeightPx != 300 {
		t.Errorf("Size().HeightPx = %v, want 300", size.HeightPx)
	}
}

func TestNewImageWithPosition(t *testing.T) {
	imgPath := createTestImage(t, 200, 150)
	
	size := domain.NewImageSize(200, 150)
	pos := domain.ImagePosition{
		Type:       domain.ImagePositionFloating,
		HAlign:     domain.HAlignCenter,
		VAlign:     domain.VAlignTop,
		WrapText:   domain.WrapSquare,
		BehindText: false,
	}

	img, err := NewImageWithPosition("img3", imgPath, size, pos)
	if err != nil {
		t.Fatalf("NewImageWithPosition() error = %v", err)
	}

	imgPos := img.Position()
	if imgPos.Type != domain.ImagePositionFloating {
		t.Errorf("Position().Type = %v, want %v", imgPos.Type, domain.ImagePositionFloating)
	}
	if imgPos.HAlign != domain.HAlignCenter {
		t.Errorf("Position().HAlign = %v, want %v", imgPos.HAlign, domain.HAlignCenter)
	}
	if imgPos.VAlign != domain.VAlignTop {
		t.Errorf("Position().VAlign = %v, want %v", imgPos.VAlign, domain.VAlignTop)
	}
}

func TestReadImageFromReader(t *testing.T) {
	// Create test image in memory
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Encode to buffer
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	// Read from buffer
	docxImg, err := ReadImageFromReader("img4", buf, domain.ImageFormatPNG)
	if err != nil {
		t.Fatalf("ReadImageFromReader() error = %v", err)
	}

	if docxImg.Format() != domain.ImageFormatPNG {
		t.Errorf("Format() = %v, want %v", docxImg.Format(), domain.ImageFormatPNG)
	}

	size := docxImg.Size()
	if size.WidthPx != 100 || size.HeightPx != 100 {
		t.Errorf("Size() = %dx%d, want 100x100", size.WidthPx, size.HeightPx)
	}
}

func TestImageSetSize(t *testing.T) {
	imgPath := createTestImage(t, 400, 300)
	img, err := NewImage("img5", imgPath)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}

	// Change size
	newSize := domain.NewImageSize(200, 150)
	if err := img.SetSize(newSize); err != nil {
		t.Fatalf("SetSize() error = %v", err)
	}

	size := img.Size()
	if size.WidthPx != 200 {
		t.Errorf("Size().WidthPx = %v, want 200", size.WidthPx)
	}
	if size.HeightPx != 150 {
		t.Errorf("Size().HeightPx = %v, want 150", size.HeightPx)
	}
}

func TestImageDescription(t *testing.T) {
	imgPath := createTestImage(t, 100, 100)
	img, err := NewImage("img6", imgPath)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}

	// Set description
	img.SetDescription("Test image description")
	if img.Description() != "Test image description" {
		t.Errorf("Description() = %v, want 'Test image description'", img.Description())
	}
}

func TestImageRelationshipID(t *testing.T) {
	imgPath := createTestImage(t, 100, 100)
	img, err := NewImage("img7", imgPath)
	if err != nil {
		t.Fatalf("NewImage() error = %v", err)
	}

	// Cast to concrete type to test internal methods
	docxImg, ok := img.(*docxImage)
	if !ok {
		t.Fatal("Failed to cast to *docxImage")
	}

	// Initially empty
	if docxImg.RelationshipID() != "" {
		t.Errorf("RelationshipID() = %v, want empty", docxImg.RelationshipID())
	}

	// Set relationship ID
	docxImg.relationshipID = "rId123"
	if docxImg.RelationshipID() != "rId123" {
		t.Errorf("RelationshipID() = %v, want 'rId123'", docxImg.RelationshipID())
	}
}

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     domain.ImageFormat
	}{
		{"PNG", "image.png", domain.ImageFormatPNG},
		{"PNG uppercase", "IMAGE.PNG", domain.ImageFormatPNG},
		{"JPEG", "photo.jpeg", domain.ImageFormatJPEG},
		{"JPG normalizes to JPEG", "photo.jpg", domain.ImageFormatJPEG}, // .jpg normalizes to .jpeg
		{"GIF", "animation.gif", domain.ImageFormatGIF},
		{"BMP", "bitmap.bmp", domain.ImageFormatBMP},
		{"TIFF", "scan.tiff", domain.ImageFormatTIFF},
		{"TIF normalizes to TIFF", "scan.tif", domain.ImageFormatTIFF}, // .tif normalizes to .tiff
		{"SVG", "vector.svg", domain.ImageFormatSVG},
		{"WEBP", "modern.webp", domain.ImageFormatWEBP},
		{"Unknown defaults to PNG", "file.xyz", domain.ImageFormatPNG}, // defaults to PNG
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create image with test format detection
			imgPath := createTestImage(t, 50, 50)
			
			// Rename to test extension
			tmpDir := filepath.Dir(imgPath)
			newPath := filepath.Join(tmpDir, tt.filename)
			if err := os.Rename(imgPath, newPath); err != nil {
				t.Fatalf("Failed to rename test file: %v", err)
			}

			// Create image and check format detection
			img, err := NewImage("test", newPath)
			if err != nil && tt.filename != "file.xyz" {
				// Only .xyz might fail, others should work
				t.Fatalf("NewImage() error = %v", err)
			}

			if err == nil && img.Format() != tt.want {
				t.Errorf("Format() = %v, want %v", img.Format(), tt.want)
			}
		})
	}
}

func TestImageSizeConversions(t *testing.T) {
	tests := []struct {
		name         string
		widthPx      int
		heightPx     int
		expectedEMUW int
		expectedEMUH int
	}{
		{
			name:         "100x100 pixels",
			widthPx:      100,
			heightPx:     100,
			expectedEMUW: 952500,
			expectedEMUH: 952500,
		},
		{
			name:         "800x600 pixels",
			widthPx:      800,
			heightPx:     600,
			expectedEMUW: 7620000,
			expectedEMUH: 5715000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := domain.NewImageSize(tt.widthPx, tt.heightPx)

			if size.WidthEMU != tt.expectedEMUW {
				t.Errorf("WidthEMU = %v, want %v", size.WidthEMU, tt.expectedEMUW)
			}
			if size.HeightEMU != tt.expectedEMUH {
				t.Errorf("HeightEMU = %v, want %v", size.HeightEMU, tt.expectedEMUH)
			}
		})
	}
}

func TestNewImageSizeInches(t *testing.T) {
	size := domain.NewImageSizeInches(3.0, 2.0)

	// 3 inches = 3 * 914400 = 2743200 EMUs
	// 2 inches = 2 * 914400 = 1828800 EMUs
	expectedWidthEMU := 2743200
	expectedHeightEMU := 1828800

	if size.WidthEMU != expectedWidthEMU {
		t.Errorf("WidthEMU = %v, want %v", size.WidthEMU, expectedWidthEMU)
	}
	if size.HeightEMU != expectedHeightEMU {
		t.Errorf("HeightEMU = %v, want %v", size.HeightEMU, expectedHeightEMU)
	}
}

func TestDefaultImagePosition(t *testing.T) {
	pos := domain.DefaultImagePosition()

	if pos.Type != domain.ImagePositionInline {
		t.Errorf("Type = %v, want %v", pos.Type, domain.ImagePositionInline)
	}
	if pos.WrapText != domain.WrapNone {
		t.Errorf("WrapText = %v, want %v", pos.WrapText, domain.WrapNone)
	}
}
