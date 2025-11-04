package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
	// Create a new document
	doc := docx.NewDocument()

	// Configure page layout
	section, _ := doc.DefaultSection()
	section.SetPageSize(domain.PageSizeA4)
	section.SetOrientation(domain.OrientationPortrait)
	section.SetMargins(domain.Margins{
		Top:    1440,
		Right:  1440,
		Bottom: 1440,
		Left:   1440,
		Header: 720,
		Footer: 720,
	})

	// Setup header with document info
	header, _ := section.Header(domain.HeaderDefault)
	setupHeader(header)

	// Setup footer with page numbers
	footer, _ := section.Footer(domain.FooterDefault)
	setupFooter(footer)

	// Add cover page
	addCoverPage(doc)

	// Add table of contents
	addTableOfContents(doc)

	// Add main content sections
	addIntroduction(doc)
	addFeatures(doc)
	addExamples(doc)
	addConclusion(doc)

	// Save the document
	if err := doc.SaveAs("07_advanced_demo.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Advanced document created successfully: 07_advanced_demo.docx")
	fmt.Println("\nThis document demonstrates:")
	fmt.Println("✓ Professional page layout (A4, portrait, 1-inch margins)")
	fmt.Println("✓ Custom header and footer with fields")
	fmt.Println("✓ Cover page with Title style")
	fmt.Println("✓ Table of Contents with hyperlinks")
	fmt.Println("✓ Multiple heading levels (H1, H2, H3)")
	fmt.Println("✓ Various paragraph styles (Normal, Quote, List)")
	fmt.Println("✓ Character formatting (bold, italic, color, hyperlinks)")
	fmt.Println("✓ Dynamic fields (page numbers, TOC, hyperlinks)")
	fmt.Println("\nTo update the TOC, open the document and press F9 or right-click > Update Field")
}

func setupHeader(header domain.Header) {
	para, _ := header.AddParagraph()
	para.SetAlignment(domain.AlignmentRight)

	run, _ := para.AddRun()
	run.AddText("go-docx v2 • Advanced Features Demo")
	run.SetSize(20)                                       // 10pt in half-points
	run.SetColor(domain.Color{R: 0x44, G: 0x72, B: 0xC4}) // Blue
}

func setupFooter(footer domain.Footer) {
	para, _ := footer.AddParagraph()
	para.SetAlignment(domain.AlignmentCenter)

	// Left part: Document info
	r1, _ := para.AddRun()
	r1.AddText("Page ")
	r1.SetSize(20) // 10pt in half-points

	// Page number field
	r2, _ := para.AddRun()
	pageField := docx.NewPageNumberField()
	r2.AddField(pageField)
	r2.SetSize(20)

	r3, _ := para.AddRun()
	r3.AddText(" of ")
	r3.SetSize(20)

	// Total pages field
	r4, _ := para.AddRun()
	totalField := docx.NewPageCountField()
	r4.AddField(totalField)
	r4.SetSize(20)
}

func addCoverPage(doc domain.Document) {
	// Title
	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)
	title.SetAlignment(domain.AlignmentCenter)
	titleRun, _ := title.AddRun()
	titleRun.AddText("Advanced Document Creation")

	doc.AddParagraph()

	// Subtitle
	subtitle, _ := doc.AddParagraph()
	subtitle.SetStyle(domain.StyleIDSubtitle)
	subtitle.SetAlignment(domain.AlignmentCenter)
	subtitleRun, _ := subtitle.AddRun()
	subtitleRun.AddText("Demonstrating All Phase 6 Features")

	doc.AddParagraph()
	doc.AddParagraph()

	// Author info
	author, _ := doc.AddParagraph()
	author.SetAlignment(domain.AlignmentCenter)
	authorRun, _ := author.AddRun()
	authorRun.AddText("Created with go-docx v2")
	authorRun.SetItalic(true)

	doc.AddParagraph()
	doc.AddParagraph()
}

func addTableOfContents(doc domain.Document) {
	// TOC heading
	tocHeading, _ := doc.AddParagraph()
	tocHeading.SetStyle(domain.StyleIDHeading1)
	tocRun, _ := tocHeading.AddRun()
	tocRun.AddText("Table of Contents")

	doc.AddParagraph()

	// Add TOC field
	tocPara, _ := doc.AddParagraph()
	tocFieldRun, _ := tocPara.AddRun()

	tocOptions := map[string]string{
		"levels":          "1-3",
		"hyperlinks":      "true",
		"hidePageNumbers": "false",
	}
	tocField := docx.NewTOCField(tocOptions)
	tocFieldRun.AddField(tocField)

	// Instructions
	doc.AddParagraph()
	instruction, _ := doc.AddParagraph()
	instruction.SetStyle(domain.StyleIDIntenseQuote)
	instRun, _ := instruction.AddRun()
	instRun.AddText("Note: To update this table of contents, open the document in Word, right-click the TOC, and select \"Update Field\", or press F9.")
	instRun.SetItalic(true)
	instRun.SetSize(20) // 10pt in half-points

	doc.AddParagraph()
	doc.AddParagraph()
}

func addIntroduction(doc domain.Document) {
	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Introduction")

	// Paragraph
	intro, _ := doc.AddParagraph()
	intro.SetStyle(domain.StyleIDNormal)
	introRun, _ := intro.AddRun()
	introRun.AddText("This document demonstrates the advanced features available in go-docx v2. It showcases sections, headers, footers, fields, and comprehensive style management.")

	doc.AddParagraph()

	// Heading 2
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("1.1 Purpose")

	purpose, _ := doc.AddParagraph()
	purpose.SetStyle(domain.StyleIDNormal)
	purposeRun, _ := purpose.AddRun()
	purposeRun.AddText("The goal is to provide a complete example that developers can reference when building professional documents programmatically.")

	doc.AddParagraph()
}

func addFeatures(doc domain.Document) {
	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("2. Features Overview")

	// Intro
	intro, _ := doc.AddParagraph()
	intro.SetStyle(domain.StyleIDNormal)
	introRun, _ := intro.AddRun()
	introRun.AddText("go-docx v2 Phase 6 introduces several powerful capabilities:")

	doc.AddParagraph()

	// Heading 2 - Sections
	h2_1, _ := doc.AddParagraph()
	h2_1.SetStyle(domain.StyleIDHeading2)
	h2_1Run, _ := h2_1.AddRun()
	h2_1Run.AddText("2.1 Section Management")

	sectDesc, _ := doc.AddParagraph()
	sectDesc.SetStyle(domain.StyleIDNormal)
	sectDescRun, _ := sectDesc.AddRun()
	sectDescRun.AddText("Complete control over page layout:")

	features := []string{
		"Page sizes (A3, A4, A5, Letter, Legal, Tabloid)",
		"Orientation (Portrait, Landscape)",
		"Margins (all sides configurable)",
		"Multi-column layouts",
	}

	for _, feature := range features {
		p, _ := doc.AddParagraph()
		p.SetStyle(domain.StyleIDListParagraph)
		r, _ := p.AddRun()
		r.AddText("• " + feature)
	}

	doc.AddParagraph()

	// Heading 2 - Headers/Footers
	h2_2, _ := doc.AddParagraph()
	h2_2.SetStyle(domain.StyleIDHeading2)
	h2_2Run, _ := h2_2.AddRun()
	h2_2Run.AddText("2.2 Headers and Footers")

	hfDesc, _ := doc.AddParagraph()
	hfDesc.SetStyle(domain.StyleIDNormal)
	hfDescRun, _ := hfDesc.AddRun()
	hfDescRun.AddText("Professional header and footer support:")

	hfFeatures := []string{
		"Three types: Default, First, Even",
		"Dynamic fields (page numbers, document properties)",
		"Full formatting capabilities",
	}

	for _, feature := range hfFeatures {
		p, _ := doc.AddParagraph()
		p.SetStyle(domain.StyleIDListParagraph)
		r, _ := p.AddRun()
		r.AddText("• " + feature)
	}

	doc.AddParagraph()

	// Heading 2 - Fields
	h2_3, _ := doc.AddParagraph()
	h2_3.SetStyle(domain.StyleIDHeading2)
	h2_3Run, _ := h2_3.AddRun()
	h2_3Run.AddText("2.3 Field System")

	fieldDesc, _ := doc.AddParagraph()
	fieldDesc.SetStyle(domain.StyleIDNormal)
	fieldDescRun, _ := fieldDesc.AddRun()
	fieldDescRun.AddText("Nine field types for dynamic content:")

	fieldTypes := []string{
		"Page numbers and page count",
		"Table of Contents (see above)",
		"Hyperlinks (see examples below)",
		"Date and time",
		"Document properties",
	}

	for _, field := range fieldTypes {
		p, _ := doc.AddParagraph()
		p.SetStyle(domain.StyleIDListParagraph)
		r, _ := p.AddRun()
		r.AddText("• " + field)
	}

	doc.AddParagraph()

	// Heading 2 - Styles
	h2_4, _ := doc.AddParagraph()
	h2_4.SetStyle(domain.StyleIDHeading2)
	h2_4Run, _ := h2_4.AddRun()
	h2_4Run.AddText("2.4 Style Management")

	styleDesc, _ := doc.AddParagraph()
	styleDesc.SetStyle(domain.StyleIDNormal)
	styleDescRun, _ := styleDesc.AddRun()
	styleDescRun.AddText("Comprehensive style system:")

	styleFeatures := []string{
		"40+ built-in styles with type-safe constants",
		"StyleManager for querying and custom styles",
		"Paragraph and character styles",
	}

	for _, feature := range styleFeatures {
		p, _ := doc.AddParagraph()
		p.SetStyle(domain.StyleIDListParagraph)
		r, _ := p.AddRun()
		r.AddText("• " + feature)
	}

	doc.AddParagraph()
}

func addExamples(doc domain.Document) {
	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("3. Code Examples")

	// Heading 2
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("3.1 Hyperlink Example")

	linkDesc, _ := doc.AddParagraph()
	linkDesc.SetStyle(domain.StyleIDNormal)

	r1, _ := linkDesc.AddRun()
	r1.AddText("Visit the ")

	// Hyperlink
	r2, _ := linkDesc.AddRun()
	linkField := docx.NewHyperlinkField(
		"https://github.com/mmonterroca/docxgo/v2",
		"go-docx GitHub repository",
	)
	r2.SetColor(domain.Color{R: 0x00, G: 0x00, B: 0xFF}) // Blue
	r2.SetUnderline(domain.UnderlineSingle)
	r2.AddField(linkField)

	r3, _ := linkDesc.AddRun()
	r3.AddText(" for more information.")

	doc.AddParagraph()

	// Heading 2
	h2_2, _ := doc.AddParagraph()
	h2_2.SetStyle(domain.StyleIDHeading2)
	h2_2Run, _ := h2_2.AddRun()
	h2_2Run.AddText("3.2 Quote Example")

	quote, _ := doc.AddParagraph()
	quote.SetStyle(domain.StyleIDIntenseQuote)
	quoteRun, _ := quote.AddRun()
	quoteRun.AddText("\"The best way to predict the future is to create it.\" - This demonstrates the IntenseQuote style.")

	doc.AddParagraph()

	// Heading 2
	h2_3, _ := doc.AddParagraph()
	h2_3.SetStyle(domain.StyleIDHeading2)
	h2_3Run, _ := h2_3.AddRun()
	h2_3Run.AddText("3.3 Mixed Formatting")

	mixed, _ := doc.AddParagraph()
	mixed.SetStyle(domain.StyleIDNormal)

	mr1, _ := mixed.AddRun()
	mr1.AddText("This paragraph demonstrates ")

	mr2, _ := mixed.AddRun()
	mr2.SetBold(true)
	mr2.AddText("bold text")

	mr3, _ := mixed.AddRun()
	mr3.AddText(", ")

	mr4, _ := mixed.AddRun()
	mr4.SetItalic(true)
	mr4.AddText("italic text")

	mr5, _ := mixed.AddRun()
	mr5.AddText(", ")

	mr6, _ := mixed.AddRun()
	mr6.SetColor(domain.Color{R: 0xFF, G: 0x00, B: 0x00}) // Red
	mr6.AddText("colored text")

	mr7, _ := mixed.AddRun()
	mr7.AddText(", and ")

	mr8, _ := mixed.AddRun()
	mr8.SetBold(true)
	mr8.SetItalic(true)
	mr8.SetColor(domain.Color{R: 0x00, G: 0xAA, B: 0x00}) // Green
	mr8.AddText("combined formatting")

	mr9, _ := mixed.AddRun()
	mr9.AddText(".")

	doc.AddParagraph()
}

func addConclusion(doc domain.Document) {
	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("4. Conclusion")

	// Summary
	summary, _ := doc.AddParagraph()
	summary.SetStyle(domain.StyleIDNormal)
	summaryRun, _ := summary.AddRun()
	summaryRun.AddText("This document demonstrates the complete feature set of go-docx v2 Phase 6. All advanced capabilities work together seamlessly to create professional documents programmatically.")

	doc.AddParagraph()

	// Next steps heading
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("4.1 Next Steps")

	// Next steps list
	nextSteps := []string{
		"Explore individual examples (05_styles, 06_sections, 04_fields)",
		"Read the API documentation for detailed reference",
		"Check the migration guide for v1 to v2 transition",
		"Contribute to the project on GitHub",
	}

	for _, step := range nextSteps {
		p, _ := doc.AddParagraph()
		p.SetStyle(domain.StyleIDListParagraph)
		r, _ := p.AddRun()
		r.AddText("• " + step)
	}

	doc.AddParagraph()
	doc.AddParagraph()

	// Final note
	final, _ := doc.AddParagraph()
	final.SetStyle(domain.StyleIDIntenseQuote)
	final.SetAlignment(domain.AlignmentCenter)
	finalRun, _ := final.AddRun()
	finalRun.AddText("Thank you for using go-docx v2!")
	finalRun.SetBold(true)
}
