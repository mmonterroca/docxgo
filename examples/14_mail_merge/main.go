// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Example: Mail Merge / Template Engine
//
// This example demonstrates the template / mail merge workflow:
//  1. Create a template document with {{placeholders}}
//  2. Open the template and inspect its placeholders
//  3. Merge with data to produce personalized documents
//  4. Demonstrate batch merge for multiple recipients
//
// This showcases:
//   - Creating reusable document templates
//   - Placeholder detection and validation
//   - Single and batch mail merge
//   - Formatting preservation during merge
//   - Table cell placeholder replacement
package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/template"
)

func main() {
	// Step 1: Create a template document with placeholders
	fmt.Println("Step 1: Creating invoice template with placeholders...")
	templatePath := createTemplate()
	fmt.Printf("  Created: %s\n\n", templatePath)

	// Step 2: Open the template and inspect placeholders
	fmt.Println("Step 2: Inspecting template placeholders...")
	doc, err := docx.OpenDocument(templatePath)
	if err != nil {
		log.Fatalf("Failed to open template: %v", err)
	}

	names := template.PlaceholderNames(doc)
	fmt.Printf("  Found %d unique placeholders:\n", len(names))
	for _, name := range names {
		fmt.Printf("    - {{%s}}\n", name)
	}

	// Step 3: Validate before merging
	fmt.Println("\nStep 3: Validating template against data...")
	data := template.MergeData{
		"company_name":    "Acme Corporation",
		"company_address": "123 Main Street, Springfield, IL 62701",
		"customer_name":   "John Smith",
		"customer_email":  "john.smith@example.com",
		"invoice_number":  "INV-2025-001",
		"invoice_date":    "January 15, 2025",
		"due_date":        "February 15, 2025",
		"item1_desc":      "Web Development Services",
		"item1_qty":       "40 hours",
		"item1_price":     "$150.00",
		"item1_total":     "$6,000.00",
		"item2_desc":      "UI/UX Design",
		"item2_qty":       "20 hours",
		"item2_price":     "$120.00",
		"item2_total":     "$2,400.00",
		"subtotal":        "$8,400.00",
		"tax":             "$672.00",
		"grand_total":     "$9,072.00",
	}

	// Re-open for validation (FindPlaceholders modifies runs via consolidation)
	doc, _ = docx.OpenDocument(templatePath)
	validationErrors := template.ValidateTemplate(doc, data)
	if len(validationErrors) == 0 {
		fmt.Println("  All placeholders have matching data keys!")
	} else {
		for _, ve := range validationErrors {
			fmt.Printf("  %s\n", ve.Error())
		}
	}

	// Step 4: Merge single document
	fmt.Println("\nStep 4: Merging template with data...")
	doc, _ = docx.OpenDocument(templatePath)
	if err := template.MergeTemplate(doc, data); err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	outputPath := "invoice_john_smith.docx"
	if err := doc.SaveAs(outputPath); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	fmt.Printf("  Saved: %s\n", outputPath)

	// Step 5: Batch merge for multiple customers
	fmt.Println("\nStep 5: Batch merge for multiple customers...")
	customers := []template.MergeData{
		{
			"company_name":    "Acme Corporation",
			"company_address": "123 Main Street, Springfield, IL 62701",
			"customer_name":   "Alice Johnson",
			"customer_email":  "alice@example.com",
			"invoice_number":  "INV-2025-002",
			"invoice_date":    "January 20, 2025",
			"due_date":        "February 20, 2025",
			"item1_desc":      "Mobile App Development",
			"item1_qty":       "60 hours",
			"item1_price":     "$160.00",
			"item1_total":     "$9,600.00",
			"item2_desc":      "Quality Assurance",
			"item2_qty":       "15 hours",
			"item2_price":     "$100.00",
			"item2_total":     "$1,500.00",
			"subtotal":        "$11,100.00",
			"tax":             "$888.00",
			"grand_total":     "$11,988.00",
		},
		{
			"company_name":    "Acme Corporation",
			"company_address": "123 Main Street, Springfield, IL 62701",
			"customer_name":   "Bob Williams",
			"customer_email":  "bob@example.com",
			"invoice_number":  "INV-2025-003",
			"invoice_date":    "January 25, 2025",
			"due_date":        "February 25, 2025",
			"item1_desc":      "Database Optimization",
			"item1_qty":       "25 hours",
			"item1_price":     "$175.00",
			"item1_total":     "$4,375.00",
			"item2_desc":      "Performance Testing",
			"item2_qty":       "10 hours",
			"item2_price":     "$130.00",
			"item2_total":     "$1,300.00",
			"subtotal":        "$5,675.00",
			"tax":             "$454.00",
			"grand_total":     "$6,129.00",
		},
	}

	for i, customerData := range customers {
		doc, err := docx.OpenDocument(templatePath)
		if err != nil {
			log.Fatalf("Failed to open template for customer %d: %v", i+1, err)
		}

		if err := template.MergeTemplate(doc, customerData); err != nil {
			log.Fatalf("Merge failed for customer %d: %v", i+1, err)
		}

		filename := fmt.Sprintf("invoice_%s.docx", customerData["customer_name"])
		if err := doc.SaveAs(filename); err != nil {
			log.Fatalf("Failed to save for customer %d: %v", i+1, err)
		}
		fmt.Printf("  Saved: %s\n", filename)
	}

	fmt.Println("\nDone! All invoices generated successfully.")
}

// createTemplate creates a professional invoice template with placeholders.
func createTemplate() string {
	builder := docx.NewDocumentBuilder(
		docx.WithTitle("Invoice Template"),
		docx.WithAuthor("docxgo v2"),
		docx.WithDefaultFont("Calibri"),
		docx.WithDefaultFontSize(22),
		docx.WithPageSize(docx.A4),
		docx.WithMargins(docx.NormalMargins),
	)

	// Company header
	builder.AddParagraph().
		Text("{{company_name}}").Bold().FontSize(14).
		Alignment(domain.AlignmentCenter).End()
	builder.AddParagraph().
		Text("{{company_address}}").FontSize(10).
		Alignment(domain.AlignmentCenter).End()

	// Blank line
	builder.AddParagraph().End()

	// Invoice title
	builder.AddParagraph().
		Text("INVOICE").Bold().FontSize(16).
		Alignment(domain.AlignmentCenter).End()

	// Blank line
	builder.AddParagraph().End()

	// Invoice details
	builder.AddParagraph().Text("Invoice #: {{invoice_number}}").End()
	builder.AddParagraph().Text("Date: {{invoice_date}}").End()
	builder.AddParagraph().Text("Due Date: {{due_date}}").End()

	// Blank line
	builder.AddParagraph().End()

	// Bill To
	builder.AddParagraph().Text("Bill To:").Bold().End()
	builder.AddParagraph().Text("{{customer_name}}").End()
	builder.AddParagraph().Text("{{customer_email}}").End()

	// Blank line
	builder.AddParagraph().End()

	// Build first to get the document, then add table
	doc, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build template document: %v", err)
	}

	// Items table
	table, _ := doc.AddTable(3, 4) // header + 2 items

	// Header row
	setTableCell(table, 0, 0, "Description")
	setTableCell(table, 0, 1, "Quantity")
	setTableCell(table, 0, 2, "Unit Price")
	setTableCell(table, 0, 3, "Total")

	// Item 1
	setTableCell(table, 1, 0, "{{item1_desc}}")
	setTableCell(table, 1, 1, "{{item1_qty}}")
	setTableCell(table, 1, 2, "{{item1_price}}")
	setTableCell(table, 1, 3, "{{item1_total}}")

	// Item 2
	setTableCell(table, 2, 0, "{{item2_desc}}")
	setTableCell(table, 2, 1, "{{item2_qty}}")
	setTableCell(table, 2, 2, "{{item2_price}}")
	setTableCell(table, 2, 3, "{{item2_total}}")

	// Blank line + totals
	doc.AddParagraph()
	addStyledParagraph(doc, "Subtotal: {{subtotal}}", 22, false, domain.AlignmentRight)
	addStyledParagraph(doc, "Tax (8%): {{tax}}", 22, false, domain.AlignmentRight)
	addStyledParagraph(doc, "Grand Total: {{grand_total}}", 24, true, domain.AlignmentRight)

	templatePath := "invoice_template.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		log.Fatalf("Failed to save template: %v", err)
	}

	return templatePath
}

func addStyledParagraph(doc domain.Document, text string, size int, bold bool, align domain.Alignment) {
	para, _ := doc.AddParagraph()
	para.SetAlignment(align)
	r, _ := para.AddRun()
	r.SetText(text)
	r.SetSize(size)
	if bold {
		r.SetBold(true)
	}
}

func setTableCell(table domain.Table, row, col int, text string) {
	rows := table.Rows()
	cell, _ := rows[row].Cell(col)
	para, _ := cell.AddParagraph()
	r, _ := para.AddRun()
	r.SetText(text)
}
