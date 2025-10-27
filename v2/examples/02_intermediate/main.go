package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo"
)

func main() {
	// Create document builder with custom options
	builder := docx.NewDocumentBuilder(
		docx.WithTitle("Product Catalog"),
		docx.WithAuthor("ACME Corporation"),
		docx.WithSubject("Q1 2024 Product Lineup"),
		docx.WithDefaultFont("Arial"),
		docx.WithDefaultFontSize(22), // 11pt = 22 half-points
		docx.WithPageSize(docx.Letter),
		docx.WithMargins(docx.NormalMargins),
	)

	// Cover page
	builder.AddParagraph().
		Text("ACME Corporation").
		Bold().
		FontSize(20).
		Color(docx.Blue).
		Alignment(docx.AlignmentCenter).
		End()

	builder.AddParagraph().
		Text("Product Catalog 2024").
		FontSize(16).
		Color(docx.Gray).
		Alignment(docx.AlignmentCenter).
		End()

	builder.AddParagraph().End()
	builder.AddParagraph().End()

	// Table of contents placeholder
	builder.AddParagraph().
		Text("Table of Contents").
		Bold().
		FontSize(14).
		Underline(docx.UnderlineSingle).
		End()

	builder.AddParagraph().
		Text("1. Electronics.....................................5").
		End()

	builder.AddParagraph().
		Text("2. Home & Garden...............................12").
		End()

	builder.AddParagraph().
		Text("3. Sports & Outdoors..........................18").
		End()

	builder.AddParagraph().End()
	builder.AddParagraph().End()

	// Section 1: Electronics
	builder.AddParagraph().
		Text("1. Electronics").
		Bold().
		FontSize(16).
		Color(docx.Blue).
		End()

	builder.AddParagraph().
		Text("Our electronics lineup features cutting-edge technology and innovative designs.").
		End()

	builder.AddParagraph().End()

	// Product table (4 rows x 3 cols)
	builder.AddTable(4, 3).
		Row(0).
		Cell(0).Text("Product").Bold().End().
		Cell(1).Text("Description").Bold().End().
		Cell(2).Text("Price").Bold().End().
		End().
		Row(1).
		Cell(0).Text("Laptop Pro X1").End().
		Cell(1).Text("15-inch display, 16GB RAM, 512GB SSD").End().
		Cell(2).Text("$1,299").Bold().End().
		End().
		Row(2).
		Cell(0).Text("Wireless Earbuds").End().
		Cell(1).Text("Active noise cancellation, 24h battery").End().
		Cell(2).Text("$199").Bold().End().
		End().
		Row(3).
		Cell(0).Text("Smart Watch").End().
		Cell(1).Text("Fitness tracking, heart rate monitor").End().
		Cell(2).Text("$349").Bold().End().
		End().
		End()

	builder.AddParagraph().End()

	// Features highlight
	builder.AddParagraph().
		Text("Key Features:").
		Bold().
		FontSize(12).
		End()

	builder.AddParagraph().
		Text("✓ Industry-leading performance").
		Color(docx.Green).
		End()

	builder.AddParagraph().
		Text("✓ 2-year warranty on all electronics").
		Color(docx.Green).
		End()

	builder.AddParagraph().
		Text("✓ Free shipping on orders over $500").
		Color(docx.Green).
		End()

	builder.AddParagraph().End()
	builder.AddParagraph().End()

	// Section 2: Home & Garden
	builder.AddParagraph().
		Text("2. Home & Garden").
		Bold().
		FontSize(16).
		Color(docx.Blue).
		End()

	builder.AddParagraph().
		Text("Transform your living space with our curated collection.").
		End()

	builder.AddParagraph().End()

	builder.AddTable(4, 3).
		Row(0).
		Cell(0).Text("Product").Bold().End().
		Cell(1).Text("Description").Bold().End().
		Cell(2).Text("Price").Bold().End().
		End().
		Row(1).
		Cell(0).Text("LED Desk Lamp").End().
		Cell(1).Text("Adjustable brightness, USB charging port").End().
		Cell(2).Text("$49").Bold().End().
		End().
		Row(2).
		Cell(0).Text("Garden Tool Set").End().
		Cell(1).Text("10-piece set with carrying case").End().
		Cell(2).Text("$89").Bold().End().
		End().
		Row(3).
		Cell(0).Text("Indoor Plant Kit").End().
		Cell(1).Text("Includes 3 plants, pots, and soil").End().
		Cell(2).Text("$39").Bold().End().
		End().
		End()

	builder.AddParagraph().End()

	// Section 3: Sports & Outdoors
	builder.AddParagraph().
		Text("3. Sports & Outdoors").
		Bold().
		FontSize(16).
		Color(docx.Blue).
		End()

	builder.AddParagraph().
		Text("Gear up for adventure with our premium outdoor equipment.").
		End()

	builder.AddParagraph().End()

	builder.AddTable(4, 3).
		Row(0).
		Cell(0).Text("Product").Bold().End().
		Cell(1).Text("Description").Bold().End().
		Cell(2).Text("Price").Bold().End().
		End().
		Row(1).
		Cell(0).Text("Hiking Backpack").End().
		Cell(1).Text("40L capacity, waterproof, ergonomic").End().
		Cell(2).Text("$129").Bold().End().
		End().
		Row(2).
		Cell(0).Text("Camping Tent").End().
		Cell(1).Text("4-person, weather-resistant, easy setup").End().
		Cell(2).Text("$249").Bold().End().
		End().
		Row(3).
		Cell(0).Text("Running Shoes").End().
		Cell(1).Text("Lightweight, cushioned, breathable").End().
		Cell(2).Text("$119").Bold().End().
		End().
		End()

	builder.AddParagraph().End()
	builder.AddParagraph().End()

	// Contact info
	builder.AddParagraph().
		Text("Contact Information").
		Bold().
		FontSize(14).
		Underline(docx.UnderlineSingle).
		End()

	builder.AddParagraph().
		Text("ACME Corporation").
		Bold().
		End()

	builder.AddParagraph().
		Text("📧 sales@acme.example.com").
		End()

	builder.AddParagraph().
		Text("📞 1-800-ACME-123").
		End()

	builder.AddParagraph().
		Text("🌐 www.acme.example.com").
		Color(docx.Blue).
		End()

	builder.AddParagraph().End()

	// Footer note
	builder.AddParagraph().
		Text("Prices subject to change. All products include standard warranty. ").
		Text("Visit our website for current pricing and availability.").
		FontSize(9).
		Color(docx.Gray).
		Italic().
		Alignment(docx.AlignmentCenter).
		End()

	// Build the document
	doc, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build document: %v", err)
	}

	// Save the document
	if err := doc.SaveAs("02_intermediate_builder.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Product catalog created successfully: 02_intermediate_builder.docx")
	fmt.Println("\nThis example demonstrates:")
	fmt.Println("✓ Professional document layout")
	fmt.Println("✓ Multiple sections with headings")
	fmt.Println("✓ Product tables with pricing")
	fmt.Println("✓ Mixed text formatting")
	fmt.Println("✓ Color-coded information")
	fmt.Println("✓ Contact information")
	fmt.Println("✓ Document metadata (title, author, subject)")
}
