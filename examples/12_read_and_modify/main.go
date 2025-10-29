/*
MIT License

Copyright (c) 2025 Misael Monterroca

Example: Read and Modify Documents

This example demonstrates the complete read/modify/write workflow:
1. Create a comprehensive showcase document with all features
2. Read it back using OpenDocument()
3. Modify its content (add paragraphs, modify text, add tables)
4. Save the modified version with a new name

This showcases:
- Document reading and parsing
- Content inspection and traversal
- In-place modification
- Preserving existing formatting
- Adding new content to existing documents
*/

package main

import (
	"fmt"
	"log"

	"github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
)

func main() {
	fmt.Println("📝 Step 1: Creating comprehensive showcase document...")
	originalPath := createShowcaseDocument()
	fmt.Printf("   ✅ Created: %s\n\n", originalPath)

	fmt.Println("📖 Step 2: Reading the document back...")
	doc, err := docx.OpenDocument(originalPath)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	fmt.Println("   ✅ Document loaded successfully")

	// Inspect the document
	fmt.Println("\n🔍 Step 3: Inspecting document content...")
	inspectDocument(doc)

	fmt.Println("\n✏️  Step 4: Modifying the document...")
	modifyDocument(doc)
	fmt.Println("   ✅ Modifications applied")

	// Save modified version
	modifiedPath := "12_modified_document.docx"
	fmt.Printf("\n💾 Step 5: Saving modified document as '%s'...\n", modifiedPath)
	if err := doc.SaveAs(modifiedPath); err != nil {
		log.Fatalf("Failed to save modified document: %v", err)
	}
	fmt.Printf("   ✅ Saved: %s\n", modifiedPath)

	fmt.Println("\n🎉 Complete! Compare the original and modified documents to see the changes.")
	fmt.Println("\nGenerated files:")
	fmt.Printf("  📄 %s (original showcase)\n", originalPath)
	fmt.Printf("  📄 %s (modified version)\n", modifiedPath)
}

// createShowcaseDocument creates a comprehensive document with all features
func createShowcaseDocument() string {
	doc := docx.NewDocument()

	// Title
	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)
	title.SetAlignment(domain.AlignmentCenter)
	titleRun, _ := title.AddRun()
	titleRun.AddText("Document Showcase - All Features")

	// Subtitle
	subtitle, _ := doc.AddParagraph()
	subtitle.SetStyle(domain.StyleIDSubtitle)
	subtitle.SetAlignment(domain.AlignmentCenter)
	subtitleRun, _ := subtitle.AddRun()
	subtitleRun.AddText("This document demonstrates all capabilities of go-docx v2")

	doc.AddParagraph() // Empty line

	// Section 1: Text Formatting
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Text Formatting Capabilities")

	para1, _ := doc.AddParagraph()
	r1, _ := para1.AddRun()
	r1.AddText("This paragraph demonstrates ")
	
	r2, _ := para1.AddRun()
	r2.AddText("bold text")
	r2.SetBold(true)
	
	r3, _ := para1.AddRun()
	r3.AddText(", ")
	
	r4, _ := para1.AddRun()
	r4.AddText("italic text")
	r4.SetItalic(true)
	
	r5, _ := para1.AddRun()
	r5.AddText(", ")
	
	r6, _ := para1.AddRun()
	r6.AddText("underlined text")
	r6.SetUnderline(domain.UnderlineSingle)
	
	r7, _ := para1.AddRun()
	r7.AddText(", ")
	
	r8, _ := para1.AddRun()
	r8.AddText("colored text")
	r8.SetColor(docx.Red)
	
	r9, _ := para1.AddRun()
	r9.AddText(", and ")
	
	r10a, _ := para1.AddRun()
	r10a.AddText("large text")
	r10a.SetSize(32) // 16pt in half-points
	
	r11, _ := para1.AddRun()
	r11.AddText(".")

	// Section 2: Styles
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading1)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("2. Paragraph Styles")

	h2_1, _ := doc.AddParagraph()
	h2_1.SetStyle(domain.StyleIDHeading2)
	h2_1Run, _ := h2_1.AddRun()
	h2_1Run.AddText("This is a Heading 2 style")

	h3_1, _ := doc.AddParagraph()
	h3_1.SetStyle(domain.StyleIDHeading3)
	h3_1Run, _ := h3_1.AddRun()
	h3_1Run.AddText("This is a Heading 3 style")

	quote, _ := doc.AddParagraph()
	quote.SetStyle(domain.StyleIDQuote)
	quoteRun, _ := quote.AddRun()
	quoteRun.AddText("This is a quote style paragraph demonstrating how quoted text appears in documents.")

	normal, _ := doc.AddParagraph()
	normal.SetStyle(domain.StyleIDNormal)
	normalRun, _ := normal.AddRun()
	normalRun.AddText("This is normal body text with standard formatting.")

	// Section 3: Tables
	h3, _ := doc.AddParagraph()
	h3.SetStyle(domain.StyleIDHeading1)
	h3Run, _ := h3.AddRun()
	h3Run.AddText("3. Table Features")

	table, _ := doc.AddTable(3, 3)
	table.SetStyle(domain.TableStyleGrid)

	// Header row
	row0, _ := table.Row(0)
	cell00, _ := row0.Cell(0)
	p00, _ := cell00.AddParagraph()
	r00, _ := p00.AddRun()
	r00.AddText("Product")
	r00.SetBold(true)

	cell01, _ := row0.Cell(1)
	p01, _ := cell01.AddParagraph()
	r01, _ := p01.AddRun()
	r01.AddText("Quantity")
	r01.SetBold(true)

	cell02, _ := row0.Cell(2)
	p02, _ := cell02.AddParagraph()
	r02, _ := p02.AddRun()
	r02.AddText("Price")
	r02.SetBold(true)

	// Data rows
	row1, _ := table.Row(1)
	cell10, _ := row1.Cell(0)
	p10, _ := cell10.AddParagraph()
	r10, _ := p10.AddRun()
	r10.AddText("Item A")

	cell11, _ := row1.Cell(1)
	p11, _ := cell11.AddParagraph()
	r11b, _ := p11.AddRun()
	r11b.AddText("5")

	cell12, _ := row1.Cell(2)
	p12, _ := cell12.AddParagraph()
	r12, _ := p12.AddRun()
	r12.AddText("$50.00")

	row2, _ := table.Row(2)
	cell20, _ := row2.Cell(0)
	p20, _ := cell20.AddParagraph()
	r20, _ := p20.AddRun()
	r20.AddText("Item B")

	cell21, _ := row2.Cell(1)
	p21, _ := cell21.AddParagraph()
	r21, _ := p21.AddRun()
	r21.AddText("3")

	cell22, _ := row2.Cell(2)
	p22, _ := cell22.AddParagraph()
	r22, _ := p22.AddRun()
	r22.AddText("$30.00")

	doc.AddParagraph() // Empty line

	// Section 4: Lists and Spacing
	h4, _ := doc.AddParagraph()
	h4.SetStyle(domain.StyleIDHeading1)
	h4Run, _ := h4.AddRun()
	h4Run.AddText("4. Lists and Spacing")

	list1, _ := doc.AddParagraph()
	list1.SetStyle(domain.StyleIDListParagraph)
	list1Run, _ := list1.AddRun()
	list1Run.AddText("• First item in the list")

	list2, _ := doc.AddParagraph()
	list2.SetStyle(domain.StyleIDListParagraph)
	list2Run, _ := list2.AddRun()
	list2Run.AddText("• Second item in the list")

	list3, _ := doc.AddParagraph()
	list3.SetStyle(domain.StyleIDListParagraph)
	list3Run, _ := list3.AddRun()
	list3Run.AddText("• Third item in the list")

	// Save document
	outputPath := "12_showcase_original.docx"
	if err := doc.SaveAs(outputPath); err != nil {
		log.Fatalf("Failed to save showcase document: %v", err)
	}

	return outputPath
}

// inspectDocument demonstrates how to read and inspect document content
func inspectDocument(doc domain.Document) {
	paragraphs := doc.Paragraphs()
	tables := doc.Tables()

	fmt.Printf("   📊 Document statistics:\n")
	fmt.Printf("      • Paragraphs: %d\n", len(paragraphs))
	fmt.Printf("      • Tables: %d\n", len(tables))

	// Show first few paragraphs
	fmt.Printf("\n   📝 First 3 paragraphs:\n")
	for i, para := range paragraphs {
		if i >= 3 {
			break
		}
		text := para.Text()
		if len(text) > 60 {
			text = text[:60] + "..."
		}
		fmt.Printf("      %d. %q\n", i+1, text)
	}

	// Show table info
	if len(tables) > 0 {
		fmt.Printf("\n   📋 First table:\n")
		table := tables[0]
		fmt.Printf("      • Rows: %d\n", table.RowCount())
		fmt.Printf("      • Columns: %d\n", table.ColumnCount())
	}
}

// modifyDocument demonstrates how to modify existing content
func modifyDocument(doc domain.Document) {
	// PART 1: Modify existing content
	fmt.Println("   → Modifying existing paragraphs...")
	paragraphs := doc.Paragraphs()
	
	// Change the title color
	if len(paragraphs) > 0 {
		title := paragraphs[0]
		runs := title.Runs()
		if len(runs) > 0 {
			runs[0].SetColor(docx.Blue) // Change title to blue
		}
	}

	// Add emphasis to subtitle
	if len(paragraphs) > 1 {
		subtitle := paragraphs[1]
		runs := subtitle.Runs()
		if len(runs) > 0 {
			runs[0].SetItalic(true) // Make subtitle italic
		}
	}

	// Modify a heading - find "1. Text Formatting Capabilities"
	for _, para := range paragraphs {
		if para.Text() == "1. Text Formatting Capabilities" {
			para.SetStyle(domain.StyleIDHeading2) // Change from H1 to H2
			runs := para.Runs()
			if len(runs) > 0 {
				runs[0].SetText("1. Text Formatting (MODIFIED)") // Update text
				runs[0].SetColor(docx.Red)                       // Highlight in red
			}
			break
		}
	}

	// Modify table content
	fmt.Println("   → Modifying existing table...")
	tables := doc.Tables()
	if len(tables) > 0 {
		table := tables[0]
		// Update a cell in the table (row 2, column 2 - "Item B" price)
		if row, err := table.Row(2); err == nil {
			if cell, err := row.Cell(2); err == nil {
				paragraphs := cell.Paragraphs()
				if len(paragraphs) > 0 {
					runs := paragraphs[0].Runs()
					if len(runs) > 0 {
						runs[0].SetText("$35.00 (UPDATED)") // Update price
						runs[0].SetBold(true)
						runs[0].SetColor(docx.Green)
					}
				}
			}
		}
	}

	// PART 2: Add new content
	fmt.Println("   → Adding new section...")
	newPara, err := doc.AddParagraph()
	if err != nil {
		log.Printf("Warning: Could not add paragraph: %v", err)
		return
	}

	// Add heading for modifications section
	newPara.SetStyle(domain.StyleIDHeading1)
	run, _ := newPara.AddRun()
	run.SetText("5. Modifications (Added by Reader)")
	run.SetColor(docx.Purple)

	// Add description paragraph
	descPara, _ := doc.AddParagraph()
	descRun, _ := descPara.AddRun()
	descRun.SetText("This section was added after reading the document. The modifications demonstrate:")

	// Add bullet points
	bullet1, _ := doc.AddParagraph()
	bullet1.SetStyle(domain.StyleIDListParagraph)
	b1run, _ := bullet1.AddRun()
	b1run.SetText("✓ Read existing documents")

	bullet2, _ := doc.AddParagraph()
	bullet2.SetStyle(domain.StyleIDListParagraph)
	b2run, _ := bullet2.AddRun()
	b2run.SetText("✓ Inspect and traverse content")

	bullet3, _ := doc.AddParagraph()
	bullet3.SetStyle(domain.StyleIDListParagraph)
	b3run, _ := bullet3.AddRun()
	b3run.SetText("✓ Modify existing text and formatting")

	bullet4, _ := doc.AddParagraph()
	bullet4.SetStyle(domain.StyleIDListParagraph)
	b4run, _ := bullet4.AddRun()
	b4run.SetText("✓ Update table cell values")

	bullet5, _ := doc.AddParagraph()
	bullet5.SetStyle(domain.StyleIDListParagraph)
	b5run, _ := bullet5.AddRun()
	b5run.SetText("✓ Add new content (paragraphs, tables)")

	bullet6, _ := doc.AddParagraph()
	bullet6.SetStyle(domain.StyleIDListParagraph)
	b6run, _ := bullet6.AddRun()
	b6run.SetText("✓ Save modified documents")

	// Add a new table showing the changes
	newTable, err := doc.AddTable(3, 2)
	if err != nil {
		log.Printf("Warning: Could not add table: %v", err)
		return
	}

	newTable.SetStyle(domain.TableStyleMediumShading)

	// Header row
	row0, _ := newTable.Row(0)
	cell00, _ := row0.Cell(0)
	p00, _ := cell00.AddParagraph()
	r00, _ := p00.AddRun()
	r00.SetText("Feature")
	r00.SetBold(true)

	cell01, _ := row0.Cell(1)
	p01, _ := cell01.AddParagraph()
	r01, _ := p01.AddRun()
	r01.SetText("Status")
	r01.SetBold(true)

	// Data rows
	row1, _ := newTable.Row(1)
	cell10, _ := row1.Cell(0)
	p10, _ := cell10.AddParagraph()
	r10, _ := p10.AddRun()
	r10.SetText("Document Reading")

	cell11, _ := row1.Cell(1)
	p11, _ := cell11.AddParagraph()
	r11, _ := p11.AddRun()
	r11.SetText("✅ Working")
	r11.SetColor(docx.Green)

	row2, _ := newTable.Row(2)
	cell20, _ := row2.Cell(0)
	p20, _ := cell20.AddParagraph()
	r20, _ := p20.AddRun()
	r20.SetText("Editing Existing Content")

	cell21, _ := row2.Cell(1)
	p21, _ := cell21.AddParagraph()
	r21, _ := p21.AddRun()
	r21.SetText("✅ Working")
	r21.SetColor(docx.Green)

	// Add summary of changes
	doc.AddParagraph() // Empty line

	summaryPara, _ := doc.AddParagraph()
	summaryPara.SetStyle(domain.StyleIDIntenseQuote)
	summaryRun, _ := summaryPara.AddRun()
	summaryRun.SetText("📝 Changes made to original document:")

	change1, _ := doc.AddParagraph()
	change1.SetStyle(domain.StyleIDListParagraph)
	c1run, _ := change1.AddRun()
	c1run.SetText("• Title color changed to blue")

	change2, _ := doc.AddParagraph()
	change2.SetStyle(domain.StyleIDListParagraph)
	c2run, _ := change2.AddRun()
	c2run.SetText("• Subtitle made italic")

	change3, _ := doc.AddParagraph()
	change3.SetStyle(domain.StyleIDListParagraph)
	c3run, _ := change3.AddRun()
	c3run.SetText("• Section 1 heading text updated and colored red")

	change4, _ := doc.AddParagraph()
	change4.SetStyle(domain.StyleIDListParagraph)
	c4run, _ := change4.AddRun()
	c4run.SetText("• Table price updated from $30.00 to $35.00")

	change5, _ := doc.AddParagraph()
	change5.SetStyle(domain.StyleIDListParagraph)
	c5run, _ := change5.AddRun()
	c5run.SetText("• New section 5 added with content")

	// Add final note
	finalPara, _ := doc.AddParagraph()
	finalPara.SetAlignment(domain.AlignmentCenter)
	finalRun, _ := finalPara.AddRun()
	finalRun.SetText("--- End of Modified Document ---")
	finalRun.SetItalic(true)
	finalRun.SetColor(docx.Gray)
}
