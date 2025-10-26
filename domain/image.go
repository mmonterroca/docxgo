/*
MIT License

Copyright (c) 2025 Misael Montero

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

// ImageFormat represents supported image formats.
type ImageFormat string

const (
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatJPG  ImageFormat = "jpg"
	ImageFormatGIF  ImageFormat = "gif"
	ImageFormatBMP  ImageFormat = "bmp"
	ImageFormatTIFF ImageFormat = "tiff"
	ImageFormatTIF  ImageFormat = "tif"
	ImageFormatSVG  ImageFormat = "svg"
	ImageFormatWEBP ImageFormat = "webp"
)

// ImageSize represents image dimensions.
type ImageSize struct {
	WidthPx  int // Width in pixels
	HeightPx int // Height in pixels
	WidthEMU int // Width in EMUs (English Metric Units, 914400 EMU = 1 inch)
	HeightEMU int // Height in EMUs
}

// NewImageSize creates an ImageSize from pixel dimensions.
// Converts pixels to EMUs assuming 96 DPI.
func NewImageSize(widthPx, heightPx int) ImageSize {
	const emuPerPixel = 9525 // 914400 EMU per inch / 96 DPI
	return ImageSize{
		WidthPx:   widthPx,
		HeightPx:  heightPx,
		WidthEMU:  widthPx * emuPerPixel,
		HeightEMU: heightPx * emuPerPixel,
	}
}

// NewImageSizeInches creates an ImageSize from inch dimensions.
func NewImageSizeInches(widthInches, heightInches float64) ImageSize {
	const emuPerInch = 914400
	const pixelsPerInch = 96
	return ImageSize{
		WidthPx:   int(widthInches * pixelsPerInch),
		HeightPx:  int(heightInches * pixelsPerInch),
		WidthEMU:  int(widthInches * emuPerInch),
		HeightEMU: int(heightInches * emuPerInch),
	}
}

// Image represents an embedded image in a document.
type Image interface {
	// ID returns the unique image ID.
	ID() string

	// Format returns the image format.
	Format() ImageFormat

	// Size returns the image dimensions.
	Size() ImageSize

	// SetSize sets custom dimensions for the image.
	// If width or height is 0, maintains aspect ratio.
	SetSize(size ImageSize) error

	// Data returns the raw image data.
	Data() []byte

	// RelationshipID returns the relationship ID for this image.
	RelationshipID() string

	// Target returns the target path in the .docx package (e.g., "media/image1.png").
	Target() string

	// Description returns the alt text description.
	Description() string

	// SetDescription sets the alt text description.
	SetDescription(desc string) error

	// Position returns the image position settings.
	Position() ImagePosition
}

// ImagePosition represents image positioning options.
type ImagePosition struct {
	Type       ImagePositionType // Inline or Floating
	HAlign     HorizontalAlign   // Horizontal alignment (for floating)
	VAlign     VerticalAlign     // Vertical alignment (for floating)
	OffsetX    int               // Horizontal offset in EMUs
	OffsetY    int               // Vertical offset in EMUs
	WrapText   TextWrapType      // Text wrapping style
	ZOrder     int               // Z-order for layering
	BehindText bool              // Whether image is behind text
}

// ImagePositionType defines how an image is positioned.
type ImagePositionType string

const (
	ImagePositionInline   ImagePositionType = "inline"   // Inline with text
	ImagePositionFloating ImagePositionType = "floating" // Floating/absolute position
)

// HorizontalAlign defines horizontal alignment for floating images.
type HorizontalAlign string

const (
	HAlignLeft   HorizontalAlign = "left"
	HAlignCenter HorizontalAlign = "center"
	HAlignRight  HorizontalAlign = "right"
	HAlignInside HorizontalAlign = "inside"  // Inside margin
	HAlignOutside HorizontalAlign = "outside" // Outside margin
)

// VerticalAlign defines vertical alignment for floating images.
type VerticalAlign string

const (
	VAlignTop    VerticalAlign = "top"
	VAlignCenter VerticalAlign = "center"
	VAlignBottom VerticalAlign = "bottom"
	VAlignInside VerticalAlign = "inside"
	VAlignOutside VerticalAlign = "outside"
)

// TextWrapType defines how text wraps around an image.
type TextWrapType string

// Text wrapping constants for floating images.
const (
	WrapNone        TextWrapType = "none"        // No wrapping
	WrapSquare      TextWrapType = "square"      // Square wrapping
	WrapTight       TextWrapType = "tight"       // Tight wrapping
	WrapThrough     TextWrapType = "through"     // Through wrapping
	WrapTopBottom   TextWrapType = "topBottom"   // Top and bottom only
	WrapBehindText  TextWrapType = "behindText"  // Behind text
	WrapInFrontText TextWrapType = "inFrontText" // In front of text
)

// DefaultImagePosition returns default inline position.
func DefaultImagePosition() ImagePosition {
	return ImagePosition{
		Type:       ImagePositionInline,
		WrapText:   WrapNone,
		ZOrder:     0,
		BehindText: false,
	}
}
