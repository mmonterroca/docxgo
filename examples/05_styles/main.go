package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
)

func main() {
	// Create a new document
	doc := docx.NewDocument()

	// Create title with Title style
	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)
	titleRun, _ := title.AddRun()
	titleRun.AddText("Style Management Demo")

	// Add some spacing
	doc.AddParagraph()

	// Heading 1
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Built-in Styles")

	// Normal paragraph
	normalPara, _ := doc.AddParagraph()
	normalPara.SetStyle(domain.StyleIDNormal)
	normalRun, _ := normalPara.AddRun()
	normalRun.AddText("This paragraph uses the Normal style. It's the default style for body text.")

	// Heading 2
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("1.1 Text Emphasis")

	// Intense Quote
	quote, _ := doc.AddParagraph()
	quote.SetStyle(domain.StyleIDIntenseQuote)
	quoteRun, _ := quote.AddRun()
	quoteRun.AddText("This is an intense quote. It stands out from the normal text with special formatting.")

	// Heading 2
	h2_2, _ := doc.AddParagraph()
	h2_2.SetStyle(domain.StyleIDHeading2)
	h2_2Run, _ := h2_2.AddRun()
	h2_2Run.AddText("1.2 Lists and References")

	// List paragraph
	list, _ := doc.AddParagraph()
	list.SetStyle(domain.StyleIDListParagraph)
	listRun, _ := list.AddRun()
	listRun.AddText("• First item in the list")

	list2, _ := doc.AddParagraph()
	list2.SetStyle(domain.StyleIDListParagraph)
	list2Run, _ := list2.AddRun()
	list2Run.AddText("• Second item in the list")

	doc.AddParagraph()

	// Heading 1
	h1_2, _ := doc.AddParagraph()
	h1_2.SetStyle(domain.StyleIDHeading1)
	h1_2Run, _ := h1_2.AddRun()
	h1_2Run.AddText("2. Available Built-in Styles")

	// Description
	infoPara, _ := doc.AddParagraph()
	infoPara.SetStyle(domain.StyleIDNormal)
	infoRun, _ := infoPara.AddRun()
	infoRun.AddText("This document demonstrates several built-in styles:")

	// Get style manager to query available styles
	styleMgr := doc.StyleManager()

	// List common built-in styles dynamically
	commonStyleIDs := []string{
		domain.StyleIDNormal,
		domain.StyleIDHeading1,
		domain.StyleIDHeading2,
		domain.StyleIDHeading3,
		domain.StyleIDTitle,
		domain.StyleIDSubtitle,
		domain.StyleIDQuote,
		domain.StyleIDIntenseQuote,
		domain.StyleIDListParagraph,
		domain.StyleIDCaption,
	}

	for _, styleID := range commonStyleIDs {
		if styleMgr.HasStyle(styleID) {
			stylePara, _ := doc.AddParagraph()
			stylePara.SetStyle(domain.StyleIDListParagraph)
			styleRun, _ := stylePara.AddRun()
			
			// Get style to display its name
			if style, err := styleMgr.GetStyle(styleID); err == nil {
				styleRun.AddText(fmt.Sprintf("• %s (%s)", style.Name(), styleID))
			}
		}
	}

	doc.AddParagraph()

	// Heading 1
	h1_3, _ := doc.AddParagraph()
	h1_3.SetStyle(domain.StyleIDHeading1)
	h1_3Run, _ := h1_3.AddRun()
	h1_3Run.AddText("3. Character Styles")

	// Paragraph with mixed character styles
	mixedPara, _ := doc.AddParagraph()
	mixedPara.SetStyle(domain.StyleIDNormal)

	r1, _ := mixedPara.AddRun()
	r1.AddText("This paragraph has ")

	r2, _ := mixedPara.AddRun()
	r2.SetBold(true)
	r2.AddText("bold text")

	r3, _ := mixedPara.AddRun()
	r3.AddText(", ")

	r4, _ := mixedPara.AddRun()
	r4.SetItalic(true)
	r4.AddText("italic text")

	r5, _ := mixedPara.AddRun()
	r5.AddText(", and ")

	r6, _ := mixedPara.AddRun()
	r6.SetColor(domain.Color{R: 255, G: 0, B: 0}) // Red
	r6.AddText("colored text")

	r7, _ := mixedPara.AddRun()
	r7.AddText(".")

	doc.AddParagraph()

	// Heading 1
	h1_4, _ := doc.AddParagraph()
	h1_4.SetStyle(domain.StyleIDHeading1)
	h1_4Run, _ := h1_4.AddRun()
	h1_4Run.AddText("4. Subtitle Example")

	// Subtitle
	subtitle, _ := doc.AddParagraph()
	subtitle.SetStyle(domain.StyleIDSubtitle)
	subtitleRun, _ := subtitle.AddRun()
	subtitleRun.AddText("This is a subtitle with special formatting")

	// Save the document
	if err := doc.SaveAs("05_styles_demo.docx"); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Println("Document created successfully: 05_styles_demo.docx")
	fmt.Println("\nThe document demonstrates:")
	fmt.Println("- Title and Subtitle styles")
	fmt.Println("- Heading styles (1-3)")
	fmt.Println("- Normal paragraph style")
	fmt.Println("- Quote and Intense Quote styles")
	fmt.Println("- List paragraph style")
	fmt.Println("- Footnote reference style")
	fmt.Println("- Character-level formatting (bold, italic, color)")
}
