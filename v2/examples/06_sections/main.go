package main

import (
	"fmt"
	"log"

	docx "github.com/SlideLang/go-docx/v2"
	"github.com/SlideLang/go-docx/v2/domain"
)

func main() {
	// Create a new document
	doc := docx.NewDocument()

	// Get the default section
	section, err := doc.DefaultSection()
	if err != nil {
		log.Fatalf("Failed to get default section: %v", err)
	}

	// Configure page layout
	section.SetPageSize(domain.PageSizeA4)
	section.SetOrientation(domain.OrientationPortrait)

	// Set margins (in twips: 1440 = 1 inch)
	margins := domain.Margins{
		Top:    1440, // 1 inch
		Right:  1440, // 1 inch
		Bottom: 1440, // 1 inch
		Left:   1440, // 1 inch
		Header: 720,  // 0.5 inch
		Footer: 720,  // 0.5 inch
	}
	section.SetMargins(margins)

	// Add header to default section
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		log.Fatalf("Failed to get header: %v", err)
	}

	headerPara, _ := header.AddParagraph()
	headerPara.SetAlignment(domain.AlignmentRight)
	headerRun, _ := headerPara.AddRun()
	headerRun.AddText("Section & Layout Demo")
	headerRun.SetBold(true)
	headerRun.SetColor(0x4472C4) // Blue

	// Add footer with page numbers
	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		log.Fatalf("Failed to get footer: %v", err)
	}

	footerPara, _ := footer.AddParagraph()
	footerPara.SetAlignment(domain.AlignmentCenter)

	footerRun1, _ := footerPara.AddRun()
	footerRun1.AddText("Page ")

	footerRun2, _ := footerPara.AddRun()
	pageField := docx.NewPageNumberField()
	footerRun2.AddField(pageField)

	footerRun3, _ := footerPara.AddRun()
	footerRun3.AddText(" of ")

	footerRun4, _ := footerPara.AddRun()
	totalField := docx.NewPageCountField()
	footerRun4.AddField(totalField)

	// Add title
	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)
	titleRun, _ := title.AddRun()
	titleRun.AddText("Section Management Demo")

	doc.AddParagraph()

	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Page Layout")

	// Description
	desc1, _ := doc.AddParagraph()
	desc1.SetStyle(domain.StyleIDNormal)
	desc1Run, _ := desc1.AddRun()
	desc1Run.AddText("This document demonstrates section management and page layout control:")

	// List items
	item1, _ := doc.AddParagraph()
	item1.SetStyle(domain.StyleIDListParagraph)
	item1Run, _ := item1.AddRun()
	item1Run.AddText("• Page size: A4 (210mm × 297mm)")

	item2, _ := doc.AddParagraph()
	item2.SetStyle(domain.StyleIDListParagraph)
	item2Run, _ := item2.AddRun()
	item2Run.AddText("• Orientation: Portrait")

	item3, _ := doc.AddParagraph()
	item3.SetStyle(domain.StyleIDListParagraph)
	item3Run, _ := item3.AddRun()
	item3Run.AddText("• Margins: 1 inch on all sides")

	item4, _ := doc.AddParagraph()
	item4.SetStyle(domain.StyleIDListParagraph)
	item4Run, _ := item4.AddRun()
	item4Run.AddText("• Header margin: 0.5 inch")

	item5, _ := doc.AddParagraph()
	item5.SetStyle(domain.StyleIDListParagraph)
	item5Run, _ := item5.AddRun()
	item5Run.AddText("• Footer margin: 0.5 inch")

	doc.AddParagraph()

	// Heading 1
	h1_2, _ := doc.AddParagraph()
	h1_2.SetStyle(domain.StyleIDHeading1)
	h1_2Run, _ := h1_2.AddRun()
	h1_2Run.AddText("2. Headers and Footers")

	// Description
	desc2, _ := doc.AddParagraph()
	desc2.SetStyle(domain.StyleIDNormal)
	desc2Run, _ := desc2.AddRun()
	desc2Run.AddText("This document has:")

	// Header info
	hItem1, _ := doc.AddParagraph()
	hItem1.SetStyle(domain.StyleIDListParagraph)
	hItem1Run, _ := hItem1.AddRun()
	hItem1Run.AddText("• Header: Right-aligned title (see top of page)")

	// Footer info
	hItem2, _ := doc.AddParagraph()
	hItem2.SetStyle(domain.StyleIDListParagraph)
	hItem2Run, _ := hItem2.AddRun()
	hItem2Run.AddText("• Footer: Center-aligned page numbers (see bottom)")

	doc.AddParagraph()

	// Heading 1
	h1_3, _ := doc.AddParagraph()
	h1_3.SetStyle(domain.StyleIDHeading1)
	h1_3Run, _ := h1_3.AddRun()
	h1_3Run.AddText("3. Available Page Sizes")

	// Page size info
	psDesc, _ := doc.AddParagraph()
	psDesc.SetStyle(domain.StyleIDNormal)
	psDescRun, _ := psDesc.AddRun()
	psDescRun.AddText("go-docx v2 supports these predefined page sizes:")

	sizes := []struct {
		name string
		desc string
	}{
		{"A3", "297mm × 420mm"},
		{"A4", "210mm × 297mm (used in this document)"},
		{"A5", "148mm × 210mm"},
		{"Letter", "8.5in × 11in"},
		{"Legal", "8.5in × 14in"},
		{"Tabloid", "11in × 17in"},
	}

	for _, size := range sizes {
		sizePara, _ := doc.AddParagraph()
		sizePara.SetStyle(domain.StyleIDListParagraph)
		sizeRun, _ := sizePara.AddRun()
		sizeRun.AddText(fmt.Sprintf("• %s: %s", size.name, size.desc))
	}

	doc.AddParagraph()

	// Heading 1
	h1_4, _ := doc.AddParagraph()
	h1_4.SetStyle(domain.StyleIDHeading1)
	h1_4Run, _ := h1_4.AddRun()
	h1_4Run.AddText("4. Orientation Options")

	// Orientation info
	oriDesc, _ := doc.AddParagraph()
	oriDesc.SetStyle(domain.StyleIDNormal)
	oriDescRun, _ := oriDesc.AddRun()
	oriDescRun.AddText("You can set page orientation:")

	ori1, _ := doc.AddParagraph()
	ori1.SetStyle(domain.StyleIDListParagraph)
	ori1Run, _ := ori1.AddRun()
	ori1Run.AddText("• Portrait (default, used here)")

	ori2, _ := doc.AddParagraph()
	ori2.SetStyle(domain.StyleIDListParagraph)
	ori2Run, _ := ori2.AddRun()
	ori2Run.AddText("• Landscape (for wide content)")

	doc.AddParagraph()

	// Heading 1
	h1_5, _ := doc.AddParagraph()
	h1_5.SetStyle(domain.StyleIDHeading1)
	h1_5Run, _ := h1_5.AddRun()
	h1_5Run.AddText("5. Column Layouts")

	// Column info
	colDesc, _ := doc.AddParagraph()
	colDesc.SetStyle(domain.StyleIDNormal)
	colDescRun, _ := colDesc.AddRun()
	colDescRun.AddText("Sections support multiple column layouts:")

	col1, _ := doc.AddParagraph()
	col1.SetStyle(domain.StyleIDListParagraph)
	col1Run, _ := col1.AddRun()
	col1Run.AddText("• Single column (default, used here)")

	col2, _ := doc.AddParagraph()
	col2.SetStyle(domain.StyleIDListParagraph)
	col2Run, _ := col2.AddRun()
	col2Run.AddText("• Two columns (newspaper style)")

	col3, _ := doc.AddParagraph()
	col3.SetStyle(domain.StyleIDListParagraph)
	col3Run, _ := col3.AddRun()
	col3Run.AddText("• Three or more columns")

	// Add filler content to demonstrate multi-page
	doc.AddParagraph()
	h1_6, _ := doc.AddParagraph()
	h1_6.SetStyle(domain.StyleIDHeading1)
	h1_6Run, _ := h1_6.AddRun()
	h1_6Run.AddText("6. Sample Content")

	for i := 1; i <= 15; i++ {
		para, _ := doc.AddParagraph()
		para.SetStyle(domain.StyleIDNormal)
		run, _ := para.AddRun()
		run.AddText(fmt.Sprintf("This is paragraph %d. It provides sample content to demonstrate how the page layout, margins, and headers/footers work across multiple pages. Notice the consistent header at the top and page numbers at the bottom.", i))
	}

	// Save the document
	if err := doc.SaveToFile("06_sections_demo.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Document created successfully: 06_sections_demo.docx")
	fmt.Println("\nThe document demonstrates:")
	fmt.Println("- A4 page size with portrait orientation")
	fmt.Println("- 1-inch margins on all sides")
	fmt.Println("- Custom header with document title")
	fmt.Println("- Footer with dynamic page numbers (Page X of Y)")
	fmt.Println("- Multiple pages to show consistent layout")
}
