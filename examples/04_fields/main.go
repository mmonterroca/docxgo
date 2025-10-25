/*
   Copyright (c) 2025 SlideLang

   Example: Fields - Table of Contents, Page Numbers, and Hyperlinks

   This example demonstrates how to use fields in go-docx v2:
   - Adding page numbers to headers/footers
   - Creating a Table of Contents
   - Inserting hyperlinks
   - Using STYLEREF for running headers

   Fields are dynamic elements that are calculated by Word when the document
   is opened or updated (F9 key).
*/

package main

import (
	"fmt"
	"log"

	"github.com/SlideLang/go-docx"
)

func main() {
	// Create a new document
	doc, err := docx.New()
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	// Example 1: Add page numbers to footer
	if err := addPageNumbers(doc); err != nil {
		log.Fatalf("Failed to add page numbers: %v", err)
	}

	// Example 2: Add Table of Contents
	if err := addTableOfContents(doc); err != nil {
		log.Fatalf("Failed to add TOC: %v", err)
	}

	// Example 3: Add content with headings
	if err := addContent(doc); err != nil {
		log.Fatalf("Failed to add content: %v", err)
	}

	// Example 4: Add hyperlinks
	if err := addHyperlinks(doc); err != nil {
		log.Fatalf("Failed to add hyperlinks: %v", err)
	}

	// Save the document
	if err := doc.SaveToFile("fields_example.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Document created successfully: fields_example.docx")
	fmt.Println("\nNote: Open the document in Word and press F9 to update all fields.")
}

// addPageNumbers adds page numbers to the default footer
func addPageNumbers(doc *docx.Document) error {
	// Get the default section
	section, err := doc.DefaultSection()
	if err != nil {
		return fmt.Errorf("get default section: %w", err)
	}

	// Get the default footer
	footer, err := section.Footer(docx.FooterDefault)
	if err != nil {
		return fmt.Errorf("get footer: %w", err)
	}

	// Add a paragraph to the footer
	para, err := footer.AddParagraph()
	if err != nil {
		return fmt.Errorf("add paragraph: %w", err)
	}

	// Set center alignment
	para.SetAlignment(docx.AlignmentCenter)

	// Add text before page number
	run1, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	run1.AddText("Page ")

	// Add page number field
	run2, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	pageField := docx.NewPageNumberField()
	run2.AddField(pageField)

	// Add text after page number
	run3, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	run3.AddText(" of ")

	// Add total page count field
	run4, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	pageCountField := docx.NewPageCountField()
	run4.AddField(pageCountField)

	return nil
}

// addTableOfContents adds a TOC at the beginning of the document
func addTableOfContents(doc *docx.Document) error {
	// Add TOC heading
	heading, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add heading: %w", err)
	}
	heading.SetStyle("Heading1")
	run, err := heading.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	run.AddText("Table of Contents")

	// Add TOC paragraph
	tocPara, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add TOC paragraph: %w", err)
	}

	// Create TOC field with custom options
	tocOptions := map[string]string{
		"levels":      "1-3",      // Include heading levels 1-3
		"hyperlinks":  "true",     // Enable hyperlinks
	}
	tocField := docx.NewTOCField(tocOptions)

	// Add TOC field to paragraph
	tocRun, err := tocPara.AddRun()
	if err != nil {
		return fmt.Errorf("add TOC run: %w", err)
	}
	tocRun.AddField(tocField)

	// Add page break after TOC
	doc.AddPageBreak()

	return nil
}

// addContent adds sample content with headings
func addContent(doc *docx.Document) error {
	sections := []struct {
		heading string
		content string
		level   string
	}{
		{
			heading: "Introduction",
			content: "This document demonstrates the use of fields in Word documents. Fields are dynamic elements that can display automatically calculated values such as page numbers, dates, or cross-references.",
			level:   "Heading1",
		},
		{
			heading: "Page Numbering",
			content: "The footer contains page numbers that are automatically calculated. The format used is 'Page X of Y', where X is the current page and Y is the total number of pages.",
			level:   "Heading2",
		},
		{
			heading: "Table of Contents",
			content: "The TOC at the beginning is generated from the heading styles in the document. In Word, you can update it by right-clicking and selecting 'Update Field', or by pressing F9.",
			level:   "Heading2",
		},
		{
			heading: "Hyperlinks",
			content: "Hyperlinks are also implemented as fields. They allow linking to external URLs or internal bookmarks.",
			level:   "Heading2",
		},
		{
			heading: "Advanced Features",
			content: "Additional field types supported include STYLEREF for running headers, SEQ for sequence numbering, and custom fields for specialized needs.",
			level:   "Heading1",
		},
		{
			heading: "Cross-References",
			content: "Cross-references can be created using REF fields, allowing you to reference other parts of the document by bookmark name.",
			level:   "Heading2",
		},
		{
			heading: "Date and Time",
			content: "DATE and TIME fields automatically insert the current date or time, with optional formatting switches.",
			level:   "Heading2",
		},
	}

	for _, section := range sections {
		// Add heading
		heading, err := doc.AddParagraph()
		if err != nil {
			return fmt.Errorf("add heading: %w", err)
		}
		heading.SetStyle(section.level)
		run, err := heading.AddRun()
		if err != nil {
			return fmt.Errorf("add run: %w", err)
		}
		run.AddText(section.heading)

		// Add content
		para, err := doc.AddParagraph()
		if err != nil {
			return fmt.Errorf("add paragraph: %w", err)
		}
		contentRun, err := para.AddRun()
		if err != nil {
			return fmt.Errorf("add content run: %w", err)
		}
		contentRun.AddText(section.content)

		// Add spacing
		para.SetSpacingAfter(200) // 200 twips
	}

	return nil
}

// addHyperlinks adds example hyperlinks to the document
func addHyperlinks(doc *docx.Document) error {
	// Add section heading
	heading, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add heading: %w", err)
	}
	heading.SetStyle("Heading1")
	run, err := heading.AddRun()
	if err != nil {
		return fmt.Errorf("add run: %w", err)
	}
	run.AddText("Useful Links")

	// Add paragraph with hyperlinks
	para, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add paragraph: %w", err)
	}

	// Text before link
	textRun, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add text run: %w", err)
	}
	textRun.AddText("For more information, visit ")

	// Hyperlink
	linkRun, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add link run: %w", err)
	}
	linkRun.SetColor(docx.ColorBlue)
	linkRun.SetUnderline(docx.UnderlineSingle)
	
	hyperlinkField := docx.NewHyperlinkField("https://github.com/SlideLang/go-docx", "go-docx repository")
	linkRun.AddField(hyperlinkField)

	// Text after link
	textRun2, err := para.AddRun()
	if err != nil {
		return fmt.Errorf("add text run: %w", err)
	}
	textRun2.AddText(" on GitHub.")

	return nil
}
