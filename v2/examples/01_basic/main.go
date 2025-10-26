package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo"
)

func main() {
	// Create a new document using the builder pattern with options
	doc, err := docx.NewDocumentBuilder(
		docx.WithTitle("Builder Pattern Demo"),
		docx.WithAuthor("go-docx v2"),
		docx.WithDefaultFont("Calibri", 11),
		docx.WithPageSize(docx.PageSizeA4),
		docx.WithMargins(docx.MarginsNormal),
	).Build()
	
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	// Add title using builder
	err = doc.AddParagraph().
		Text("Welcome to go-docx v2 Builder Pattern").
		Bold().
		FontSize(16).
		Color(docx.Blue).
		Alignment(docx.AlignmentCenter).
		End()
	
	if err != nil {
		log.Fatalf("Failed to add title: %v", err)
	}

	// Add subtitle
	doc.AddParagraph().
		Text("Creating documents is now easier than ever").
		Italic().
		FontSize(12).
		Color(docx.Gray).
		Alignment(docx.AlignmentCenter).
		End()

	// Add blank line
	doc.AddParagraph().End()

	// Add section heading
	doc.AddParagraph().
		Text("1. Introduction").
		Bold().
		FontSize(14).
		Color(docx.Black).
		End()

	// Add normal paragraph
	doc.AddParagraph().
		Text("The builder pattern provides a fluent, chainable API for creating documents. ").
		Text("It makes code more readable and reduces boilerplate.").
		End()

	doc.AddParagraph().End()

	// Add another section
	doc.AddParagraph().
		Text("2. Features").
		Bold().
		FontSize(14).
		End()

	// Add list items
	doc.AddParagraph().
		Text("• Fluent API - chain multiple formatting calls").
		End()

	doc.AddParagraph().
		Text("• Type-safe colors - use predefined color constants").
		Color(docx.Red).
		End()

	doc.AddParagraph().
		Text("• Easy formatting - bold, italic, underline, and more").
		Bold().
		End()

	doc.AddParagraph().
		Text("• Alignment control - left, center, right, justify").
		Alignment(docx.AlignmentCenter).
		End()

	doc.AddParagraph().End()

	// Add section with mixed formatting
	doc.AddParagraph().
		Text("3. Mixed Formatting Example").
		Bold().
		FontSize(14).
		End()

	// This demonstrates multiple text runs with different formatting
	doc.AddParagraph().
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

	doc.AddParagraph().End()

	// Add table example
	doc.AddParagraph().
		Text("4. Simple Table").
		Bold().
		FontSize(14).
		End()

	// Create a simple 2x3 table
	doc.AddTable().
		Width(5000).
		Row().
			Cell().Text("Header 1").Bold().End().
			Cell().Text("Header 2").Bold().End().
			Cell().Text("Header 3").Bold().End().
		End().
		Row().
			Cell().Text("Row 1, Col 1").End().
			Cell().Text("Row 1, Col 2").End().
			Cell().Text("Row 1, Col 3").End().
		End().
		Row().
			Cell().Text("Row 2, Col 1").End().
			Cell().Text("Row 2, Col 2").Color(docx.Blue).End().
			Cell().Text("Row 2, Col 3").End().
		End().
	End()

	doc.AddParagraph().End()

	// Add conclusion
	doc.AddParagraph().
		Text("Conclusion").
		Bold().
		FontSize(14).
		End()

	doc.AddParagraph().
		Text("The builder pattern makes document creation intuitive and enjoyable. ").
		Text("Try it in your next project!").
		Bold().
		Color(docx.Purple).
		End()

	// Save the document
	if err := doc.SaveToFile("01_basic_builder.docx"); err != nil {
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
