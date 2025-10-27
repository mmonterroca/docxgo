/*
MIT License

Copyright (c) 2025 Misael Monterroca
Copyright (c) 2020-2023 fumiama (original go-docx)

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

// Example demonstrating basic usage of docxgo.
package main

import (
	"fmt"
	"log"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/core"
)

func main() {
	fmt.Println("go-docx v2 - Basic Example")
	fmt.Println("==========================")
	fmt.Println()

	// Create a new document
	doc := core.NewDocument()
	fmt.Println("✓ Created new document")

	// Set metadata
	metadata := &domain.Metadata{
		Title:       "Sample Document",
		Creator:     "go-docx v2",
		Subject:     "Demonstration of clean architecture",
		Description: "This document showcases the new v2 API",
	}
	if err := doc.SetMetadata(metadata); err != nil {
		log.Fatalf("Failed to set metadata: %v", err)
	}
	fmt.Println("✓ Set document metadata")

	// Add a title paragraph
	title, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph: %v", err)
	}

	titleRun, err := title.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}

	if err := titleRun.SetText("Welcome to go-docx v2"); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}
	if err := titleRun.SetBold(true); err != nil {
		log.Fatalf("Failed to set bold: %v", err)
	}
	if err := titleRun.SetSize(32); err != nil { // 16pt
		log.Fatalf("Failed to set size: %v", err)
	}
	if err := title.SetAlignment(domain.AlignmentCenter); err != nil {
		log.Fatalf("Failed to set alignment: %v", err)
	}

	fmt.Println("✓ Added title paragraph")

	// Add spacing
	if err := title.SetSpacingAfter(720); err != nil { // 0.5 inch
		log.Fatalf("Failed to set spacing: %v", err)
	}

	// Add a regular paragraph with mixed formatting
	para, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph: %v", err)
	}

	// First run - normal text
	run1, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := run1.SetText("This is "); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}

	// Second run - bold text
	run2, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := run2.SetText("bold"); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}
	if err := run2.SetBold(true); err != nil {
		log.Fatalf("Failed to set bold: %v", err)
	}

	// Third run - normal text
	run3, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := run3.SetText(" and this is "); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}

	// Fourth run - italic colored text
	run4, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := run4.SetText("italic colored"); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}
	if err := run4.SetItalic(true); err != nil {
		log.Fatalf("Failed to set italic: %v", err)
	}
	if err := run4.SetColor(domain.ColorBlue); err != nil {
		log.Fatalf("Failed to set color: %v", err)
	}

	// Fifth run - normal text
	run5, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := run5.SetText(" text."); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}

	fmt.Println("✓ Added formatted paragraph")

	// Add an indented paragraph
	indentedPara, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph: %v", err)
	}

	indentRun, err := indentedPara.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := indentRun.SetText("This paragraph is indented with a first line indent."); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}

	indent := domain.Indentation{
		Left:      720,  // 0.5 inch left
		FirstLine: 360,  // 0.25 inch first line
	}
	if err := indentedPara.SetIndent(indent); err != nil {
		log.Fatalf("Failed to set indent: %v", err)
	}

	fmt.Println("✓ Added indented paragraph")

	// Add a hyperlink
	linkPara, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph: %v", err)
	}

	linkText, err := linkPara.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}
	if err := linkText.SetText("Visit: "); err != nil {
		log.Fatalf("Failed to set text: %v", err)
	}

	_, err = linkPara.AddHyperlink("https://github.com/mmonterroca/docxgo", "GitHub Repository")
	if err != nil {
		log.Fatalf("Failed to add hyperlink: %v", err)
	}

	fmt.Println("✓ Added hyperlink")

	// Add a table
	table, err := doc.AddTable(3, 4)
	if err != nil {
		log.Fatalf("Failed to add table: %v", err)
	}

	// Fill first row (header)
	row0, _ := table.Row(0)
	for col := 0; col < 4; col++ {
		cell, _ := row0.Cell(col)
		cellPara, _ := cell.AddParagraph()
		cellRun, _ := cellPara.AddRun()
		cellRun.SetText(fmt.Sprintf("Header %d", col+1))
		cellRun.SetBold(true)
		cellPara.SetAlignment(domain.AlignmentCenter)
	}

	// Fill data rows
	for row := 1; row < 3; row++ {
		tableRow, _ := table.Row(row)
		for col := 0; col < 4; col++ {
			cell, _ := tableRow.Cell(col)
			cellPara, _ := cell.AddParagraph()
			cellRun, _ := cellPara.AddRun()
			cellRun.SetText(fmt.Sprintf("R%d C%d", row, col+1))
		}
	}

	fmt.Println("✓ Added 3x4 table")

	// Validate document
	if err := doc.Validate(); err != nil {
		log.Fatalf("Document validation failed: %v", err)
	}
	fmt.Println("✓ Document validation passed")

	// Print summary
	fmt.Println("\nDocument Summary:")
	fmt.Printf("  Paragraphs: %d\n", len(doc.Paragraphs()))
	fmt.Printf("  Tables:     %d\n", len(doc.Tables()))
	fmt.Printf("  Title:      %s\n", doc.Metadata().Title)

	fmt.Println("\nFirst paragraph text:", doc.Paragraphs()[0].Text())

	// Note: Saving not yet implemented
	fmt.Println("\nNote: File saving will be implemented in the next phase.")
	// doc.SaveAs("example.docx")
}
