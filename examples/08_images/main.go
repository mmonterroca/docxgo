/*
MIT License

Copyright (c) 2025 Misael Monterroca

Example: Image insertion with various positioning options

This example demonstrates:
- Adding inline images to paragraphs
- Custom image sizes
- Floating images with positioning
- Builder pattern for fluent API
*/

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
)

func createSampleImage(path string, width, height int, fillColor color.RGBA) error {
	// Create a simple colored rectangle
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with gradient or solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(int(fillColor.R) * x / width)
			g := uint8(int(fillColor.G) * y / height)
			b := fillColor.B
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Save to file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func main() {
	// Create temporary directory for sample images
	tmpDir := "temp_images"
	defer os.RemoveAll(tmpDir)

	// Create sample images
	redImage := filepath.Join(tmpDir, "red_box.png")
	if err := createSampleImage(redImage, 400, 300, color.RGBA{R: 255, G: 0, B: 0, A: 255}); err != nil {
		log.Fatalf("Failed to create red image: %v", err)
	}

	blueImage := filepath.Join(tmpDir, "blue_box.png")
	if err := createSampleImage(blueImage, 600, 400, color.RGBA{R: 0, G: 0, B: 255, A: 255}); err != nil {
		log.Fatalf("Failed to create blue image: %v", err)
	}

	greenImage := filepath.Join(tmpDir, "green_box.png")
	if err := createSampleImage(greenImage, 300, 300, color.RGBA{R: 0, G: 255, B: 0, A: 255}); err != nil {
		log.Fatalf("Failed to create green image: %v", err)
	}

	// Create document with builder
	builder := docx.NewDocumentBuilder()

	// Title
	builder.AddParagraph().
		Text("Image Insertion Examples").
		Bold().
		FontSize(24).
		Alignment(domain.AlignmentCenter).
		End()

	builder.AddParagraph().Text("").End() // Blank line

	// Example 1: Inline image (default size)
	builder.AddParagraph().
		Text("1. Inline Image (Default Size)").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("This paragraph contains an inline image: ").
		AddImage(redImage).
		Text(" <- Red box image").
		End()

	builder.AddParagraph().Text("").End()

	// Example 2: Custom sized image
	builder.AddParagraph().
		Text("2. Custom Sized Image").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Blue box resized to 300x200 pixels: ").
		End()

	customSize := domain.NewImageSize(300, 200)
	builder.AddParagraph().
		AddImageWithSize(blueImage, customSize).
		End()

	builder.AddParagraph().Text("").End()

	// Example 3: Image sized in inches
	builder.AddParagraph().
		Text("3. Image Sized in Inches").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Green box at 2x2 inches: ").
		End()

	inchSize := domain.NewImageSizeInches(2.0, 2.0)
	builder.AddParagraph().
		AddImageWithSize(greenImage, inchSize).
		End()

	builder.AddParagraph().Text("").End()

	// Example 4: Floating image (centered)
	builder.AddParagraph().
		Text("4. Floating Image (Center Aligned)").
		Bold().
		FontSize(16).
		End()

	floatingSize := domain.NewImageSize(250, 187)
	centerPos := domain.ImagePosition{
		Type:       domain.ImagePositionFloating,
		HAlign:     domain.HAlignCenter,
		VAlign:     domain.VAlignTop,
		WrapText:   domain.WrapSquare,
		BehindText: false,
	}

	builder.AddParagraph().
		Text("This text wraps around the centered floating image below.").
		End()

	builder.AddParagraph().
		AddImageWithPosition(redImage, floatingSize, centerPos).
		End()

	builder.AddParagraph().
		Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ").
		Text("Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ").
		Text("Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.").
		End()

	builder.AddParagraph().Text("").End()

	// Example 5: Floating image (left aligned with text wrap)
	builder.AddParagraph().
		Text("5. Floating Image (Left Aligned with Text Wrap)").
		Bold().
		FontSize(16).
		End()

	leftPos := domain.ImagePosition{
		Type:       domain.ImagePositionFloating,
		HAlign:     domain.HAlignLeft,
		VAlign:     domain.VAlignTop,
		WrapText:   domain.WrapSquare,
		OffsetX:    0,
		OffsetY:    0,
		BehindText: false,
	}

	builder.AddParagraph().
		AddImageWithPosition(greenImage, domain.NewImageSize(200, 200), leftPos).
		Text("This text wraps around a left-aligned image. ").
		Text("The image is positioned to the left and the text flows around it. ").
		Text("This creates a professional magazine-style layout. ").
		Text("You can control the text wrapping behavior using different wrap types. ").
		Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit.").
		End()

	builder.AddParagraph().Text("").End()

	// Example 6: Right-aligned floating image
	builder.AddParagraph().
		Text("6. Floating Image (Right Aligned)").
		Bold().
		FontSize(16).
		End()

	rightPos := domain.ImagePosition{
		Type:       domain.ImagePositionFloating,
		HAlign:     domain.HAlignRight,
		VAlign:     domain.VAlignTop,
		WrapText:   domain.WrapTight,
		BehindText: false,
	}

	builder.AddParagraph().
		AddImageWithPosition(blueImage, domain.NewImageSize(180, 120), rightPos).
		Text("This paragraph has a right-aligned floating image. ").
		Text("The tight wrap setting allows text to flow very close to the image edge. ").
		Text("This is useful for creating compact layouts with minimal white space. ").
		Text("Notice how the text wraps tightly around the image boundary.").
		End()

	builder.AddParagraph().Text("").End()

	// Example 7: Multiple images in a paragraph
	builder.AddParagraph().
		Text("7. Multiple Images in One Paragraph").
		Bold().
		FontSize(16).
		End()

	smallSize := domain.NewImageSize(80, 60)
	builder.AddParagraph().
		Text("Small images inline: ").
		AddImageWithSize(redImage, smallSize).
		Text(" ").
		AddImageWithSize(blueImage, smallSize).
		Text(" ").
		AddImageWithSize(greenImage, smallSize).
		Text(" <- Three inline images").
		End()

	// Save document
	doc, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build document: %v", err)
	}

	outputPath := "08_images_output.docx"
	if err := doc.SaveAs(outputPath); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Printf("✅ Document created successfully: %s\n", outputPath)
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("  - Inline images (default size)")
	fmt.Println("  - Custom image sizes (pixels)")
	fmt.Println("  - Image sizes in inches")
	fmt.Println("  - Floating images (center, left, right)")
	fmt.Println("  - Text wrapping (square, tight)")
	fmt.Println("  - Multiple images in one paragraph")
}
