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
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format decoder
	_ "image/jpeg" // Register JPEG format decoder
	_ "image/png"  // Register PNG format decoder
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// docxImage implements the domain.Image interface.
type docxImage struct {
	id             string
	format         domain.ImageFormat
	size           domain.ImageSize
	originalSize   domain.ImageSize
	data           []byte
	relationshipID string
	target         string
	description    string
	position       domain.ImagePosition
}

// NewImage creates a new image from a file path.
func NewImage(id, path string) (domain.Image, error) {
	// Read image file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, "NewImage")
	}

	// Detect format from extension
	format := detectImageFormat(path)
	if format == "" {
		return nil, errors.InvalidArgument("NewImage", "path", path, "unsupported image format")
	}

	// Get image dimensions
	size, err := getImageDimensions(data)
	if err != nil {
		return nil, errors.Wrap(err, "NewImage")
	}

	// Generate target path
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		ext = "." + string(format)
	}
	target := fmt.Sprintf("media/image%s%s", id, ext)

	return &docxImage{
		id:           id,
		format:       format,
		size:         size,
		originalSize: size,
		data:         data,
		target:       target,
		description:  "",
		position:     domain.DefaultImagePosition(),
	}, nil
}

// NewImageWithSize creates a new image with custom dimensions.
func NewImageWithSize(id, path string, size domain.ImageSize) (domain.Image, error) {
	img, err := NewImage(id, path)
	if err != nil {
		return nil, err
	}

	if err := img.SetSize(size); err != nil {
		return nil, err
	}

	return img, nil
}

// NewImageWithPosition creates a new image with custom positioning.
func NewImageWithPosition(id, path string, size domain.ImageSize, pos domain.ImagePosition) (domain.Image, error) {
	img, err := NewImageWithSize(id, path, size)
	if err != nil {
		return nil, err
	}

	docxImg := img.(*docxImage)
	docxImg.position = pos

	return img, nil
}

// NewImageFromPackage creates an image from raw package data and target.
func NewImageFromPackage(target string, data []byte, contentType string) (domain.Image, error) {
	if target == "" {
		return nil, errors.InvalidArgument("NewImageFromPackage", "target", target, "image target cannot be empty")
	}
	if len(data) == 0 {
		return nil, errors.InvalidArgument("NewImageFromPackage", "data", data, "image data cannot be empty")
	}

	normalized := strings.ReplaceAll(target, "\\", "/")
	normalized = strings.TrimSpace(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")

	for strings.HasPrefix(normalized, "../") {
		normalized = strings.TrimPrefix(normalized, "../")
	}

	relative := strings.TrimPrefix(normalized, "word/")

	base := filepath.Base(relative)
	if base == "" {
		return nil, errors.InvalidArgument("NewImageFromPackage", "target", target, "image target missing file name")
	}

	id := strings.TrimSuffix(base, filepath.Ext(base))
	if id == "" {
		id = base
	}

	format := detectImageFormat(relative)
	if format == "" {
		format = formatFromContentType(contentType)
	}
	if format == "" {
		return nil, errors.InvalidArgument("NewImageFromPackage", "target", target, "unsupported or unknown image format")
	}

	size, err := getImageDimensions(data)
	if err != nil {
		return nil, errors.Wrap(err, "NewImageFromPackage")
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	return &docxImage{
		id:           id,
		format:       format,
		size:         size,
		originalSize: size,
		data:         copyData,
		target:       relative,
		description:  "",
		position:     domain.DefaultImagePosition(),
	}, nil
}

// ID returns the unique image ID.
func (img *docxImage) ID() string {
	return img.id
}

// Format returns the image format.
func (img *docxImage) Format() domain.ImageFormat {
	return img.format
}

// Size returns the image dimensions.
func (img *docxImage) Size() domain.ImageSize {
	return img.size
}

// SetSize sets custom dimensions for the image.
func (img *docxImage) SetSize(size domain.ImageSize) error {
	// If width or height is 0, maintain aspect ratio
	if size.WidthPx == 0 && size.HeightPx == 0 {
		return errors.InvalidArgument("Image.SetSize", "size", size, "both width and height cannot be zero")
	}

	if size.WidthPx == 0 {
		// Calculate width from height maintaining aspect ratio
		ratio := float64(img.originalSize.WidthPx) / float64(img.originalSize.HeightPx)
		size.WidthPx = int(float64(size.HeightPx) * ratio)
		size.WidthEMU = size.WidthPx * 9525
	} else if size.HeightPx == 0 {
		// Calculate height from width maintaining aspect ratio
		ratio := float64(img.originalSize.HeightPx) / float64(img.originalSize.WidthPx)
		size.HeightPx = int(float64(size.WidthPx) * ratio)
		size.HeightEMU = size.HeightPx * 9525
	}

	img.size = size
	return nil
}

// Data returns the raw image data.
func (img *docxImage) Data() []byte {
	// Return copy to prevent external modification
	data := make([]byte, len(img.data))
	copy(data, img.data)
	return data
}

// RelationshipID returns the relationship ID for this image.
func (img *docxImage) RelationshipID() string {
	return img.relationshipID
}

// SetRelationshipID sets the relationship ID (called by manager).
func (img *docxImage) SetRelationshipID(id string) {
	img.relationshipID = id
}

// Target returns the target path in the .docx package.
func (img *docxImage) Target() string {
	return img.target
}

// setTarget updates the internal target path for the image within the DOCX package.
func (img *docxImage) setTarget(target string) {
	img.target = target
}

// Description returns the alt text description.
func (img *docxImage) Description() string {
	return img.description
}

// SetDescription sets the alt text description.
func (img *docxImage) SetDescription(desc string) error {
	img.description = desc
	return nil
}

// Position returns the image position settings.
func (img *docxImage) Position() domain.ImagePosition {
	return img.position
}

// SetPosition sets the image position metadata.
func (img *docxImage) SetPosition(pos domain.ImagePosition) error {
	img.position = pos
	return nil
}

// detectImageFormat detects the image format from file extension.
func detectImageFormat(path string) domain.ImageFormat {
	ext := strings.ToLower(filepath.Ext(path))
	ext = strings.TrimPrefix(ext, ".")

	switch ext {
	case "png":
		return domain.ImageFormatPNG
	case "jpg", "jpeg":
		return domain.ImageFormatJPEG
	case "gif":
		return domain.ImageFormatGIF
	case "bmp":
		return domain.ImageFormatBMP
	case "tif", "tiff":
		return domain.ImageFormatTIFF
	case "svg":
		return domain.ImageFormatSVG
	case "webp":
		return domain.ImageFormatWEBP
	default:
		return ""
	}
}

func formatFromContentType(contentType string) domain.ImageFormat {
	switch strings.ToLower(contentType) {
	case constants.ContentTypePNG:
		return domain.ImageFormatPNG
	case constants.ContentTypeJPEG:
		return domain.ImageFormatJPEG
	case constants.ContentTypeGIF:
		return domain.ImageFormatGIF
	case constants.ContentTypeBMP:
		return domain.ImageFormatBMP
	case constants.ContentTypeTIFF:
		return domain.ImageFormatTIFF
	default:
		return ""
	}
}

// getImageDimensions reads image dimensions from image data.
func getImageDimensions(data []byte) (domain.ImageSize, error) {
	// Decode image to get dimensions
	img, format, err := image.DecodeConfig(strings.NewReader(string(data)))
	if err != nil {
		// If decode fails, try reading as binary
		reader := strings.NewReader(string(data))
		img, format, err = image.DecodeConfig(reader)
		if err != nil {
			return domain.ImageSize{}, errors.Wrap(err, "getImageDimensions")
		}
	}

	_ = format // format string is for logging if needed

	return domain.NewImageSize(img.Width, img.Height), nil
}

// ReadImageFromReader creates an image from an io.Reader.
func ReadImageFromReader(id string, reader io.Reader, format domain.ImageFormat) (domain.Image, error) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, "ReadImageFromReader")
	}

	// Get dimensions
	size, err := getImageDimensions(data)
	if err != nil {
		return nil, errors.Wrap(err, "ReadImageFromReader")
	}

	// Generate target
	target := fmt.Sprintf("media/image%s.%s", id, format)

	return &docxImage{
		id:           id,
		format:       format,
		size:         size,
		originalSize: size,
		data:         data,
		target:       target,
		description:  "",
		position:     domain.DefaultImagePosition(),
	}, nil
}
