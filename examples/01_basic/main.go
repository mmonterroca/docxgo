package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo"
)

func main() {
	// Create a new document builder with options
	builder := docx.NewDocumentBuilder(
		docx.WithTitle("Builder Pattern Demo"),
		docx.WithAuthor("go-docx v2"),
		docx.WithDefaultFont("Calibri"),
		docx.WithDefaultFontSize(22), // 11pt = 22 half-points
		docx.WithPageSize(docx.A4),
		docx.WithMargins(docx.NormalMargins),
	)

	// Add title using builder fluent API
	builder.AddParagraph().
		Text("Welcome to go-docx v2 Builder Pattern").
		Bold().
		FontSize(16).
		Color(docx.Blue).
		Alignment(docx.AlignmentCenter).
		End()

	// Add subtitle
	builder.AddParagraph().
		Text("Creating documents is now easier than ever").
		Italic().
		FontSize(12).
		Color(docx.Gray).
		Alignment(docx.AlignmentCenter).
		End()

	// Add blank line
	builder.AddParagraph().End()

	// Add section heading
	builder.AddParagraph().
		Text("1. Introduction").
		Bold().
		FontSize(14).
		Color(docx.Black).
		End()

	// Add normal paragraph
	builder.AddParagraph().
		Text("The builder pattern provides a fluent, chainable API for creating documents. ").
		Text("It makes code more readable and reduces boilerplate.").
		End()

	builder.AddParagraph().End()

	// Add another section
	builder.AddParagraph().
		Text("2. Features").
		Bold().
		FontSize(14).
		End()

	// Add list items
	builder.AddParagraph().
		Text("• Fluent API - chain multiple formatting calls").
		End()

	builder.AddParagraph().
		Text("• Type-safe colors - use predefined color constants").
		Color(docx.Red).
		End()

	builder.AddParagraph().
		Text("• Easy formatting - bold, italic, underline, and more").
		Bold().
		End()

	builder.AddParagraph().
		Text("• Alignment control - left, center, right, justify").
		Alignment(docx.AlignmentCenter).
		End()

	builder.AddParagraph().End()

	// Add section with mixed formatting
	builder.AddParagraph().
		Text("3. Mixed Formatting Example").
		Bold().
		FontSize(14).
		End()

	// This demonstrates multiple text runs with different formatting
	builder.AddParagraph().
		Text("This paragraph has ").
		Text("bold text").Bold().
		Text(", ").
		Text("italic text").Italic().
		Text(", ").
		Text("colored text").Color(docx.Green).
		Text(", and ").
		Text("underlined text").Underline(docx.UnderlineSingle).
		Text(".").
		End()

	builder.AddParagraph().End()

	// Add table example
	builder.AddParagraph().
		Text("4. Simple Table").
		Bold().
		FontSize(14).
		End()

	// Create a simple 3x3 table
	builder.AddTable(3, 3).
		Row(0).
		Cell(0).Text("Header 1").Bold().End().
		Cell(1).Text("Header 2").Bold().End().
		Cell(2).Text("Header 3").Bold().End().
		End().
		Row(1).
		Cell(0).Text("Row 1, Col 1").End().
		Cell(1).Text("Row 1, Col 2").End().
		Cell(2).Text("Row 1, Col 3").End().
		End().
		Row(2).
		Cell(0).Text("Row 2, Col 1").End().
		Cell(1).Text("Row 2, Col 2").Shading(docx.Blue).End().
		Cell(2).Text("Row 2, Col 3").End().
		End().
		End()

	builder.AddParagraph().End()

	// Add conclusion
	builder.AddParagraph().
		Text("Conclusion").
		Bold().
		FontSize(14).
		End()

	builder.AddParagraph().
		Text("The builder pattern makes document creation intuitive and enjoyable. ").
		Text("Try it in your next project!").
		Bold().
		Color(docx.Purple).
		End()

	// Build the document (validates and finalizes)
	doc, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build document: %v", err)
	}

	// Save the document
	if err := doc.SaveAs("01_basic_builder.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Document created successfully: 01_basic_builder.docx")
	fmt.Println("\nThis example demonstrates:")
	fmt.Println("✓ Builder pattern with fluent API")
	fmt.Println("✓ Document options (title, author, font, margins)")
	fmt.Println("✓ Predefined color constants")
	fmt.Println("✓ Text formatting (bold, italic, color, size)")
	fmt.Println("✓ Alignment control")
	fmt.Println("✓ Simple table creation")
	fmt.Println("✓ Mixed formatting in paragraphs")
}
