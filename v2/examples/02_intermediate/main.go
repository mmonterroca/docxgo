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
		docx.WithDefaultFont("Arial", 11),
		docx.WithPageSize(docx.PageSizeLetter),
		docx.WithMargins(docx.MarginsNormal),
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

	// Product table
	builder.AddTable().
		Width(8000).
		Row().
			Cell().Text("Product").Bold().Width(2000).End().
			Cell().Text("Description").Bold().Width(4000).End().
			Cell().Text("Price").Bold().Width(2000).End().
		End().
		Row().
			Cell().Text("Laptop Pro X1").End().
			Cell().Text("15-inch display, 16GB RAM, 512GB SSD").End().
			Cell().Text("$1,299").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Wireless Earbuds").End().
			Cell().Text("Active noise cancellation, 24h battery").End().
			Cell().Text("$199").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Smart Watch").End().
			Cell().Text("Fitness tracking, heart rate monitor").End().
			Cell().Text("$349").Color(docx.Green).Bold().End().
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

	builder.AddTable().
		Width(8000).
		Row().
			Cell().Text("Product").Bold().Width(2000).End().
			Cell().Text("Description").Bold().Width(4000).End().
			Cell().Text("Price").Bold().Width(2000).End().
		End().
		Row().
			Cell().Text("LED Desk Lamp").End().
			Cell().Text("Adjustable brightness, USB charging port").End().
			Cell().Text("$49").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Garden Tool Set").End().
			Cell().Text("10-piece set with carrying case").End().
			Cell().Text("$89").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Indoor Plant Kit").End().
			Cell().Text("Includes 3 plants, pots, and soil").End().
			Cell().Text("$39").Color(docx.Green).Bold().End().
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

	builder.AddTable().
		Width(8000).
		Row().
			Cell().Text("Product").Bold().Width(2000).End().
			Cell().Text("Description").Bold().Width(4000).End().
			Cell().Text("Price").Bold().Width(2000).End().
		End().
		Row().
			Cell().Text("Hiking Backpack").End().
			Cell().Text("40L capacity, waterproof, ergonomic").End().
			Cell().Text("$129").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Camping Tent").End().
			Cell().Text("4-person, weather-resistant, easy setup").End().
			Cell().Text("$249").Color(docx.Green).Bold().End().
		End().
		Row().
			Cell().Text("Running Shoes").End().
			Cell().Text("Lightweight, cushioned, breathable").End().
			Cell().Text("$119").Color(docx.Green).Bold().End().
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
