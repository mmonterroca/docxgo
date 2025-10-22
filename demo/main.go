// Package main demonstrates the enhanced go-docx features for SlideLang
package main

import (
	"fmt"
	"os"

	"github.com/fumiama/go-docx"
)

func main() {
	fmt.Println("🚀 SlideLang/go-docx Enhanced Fork Demo")
	fmt.Println("======================================")

	// Create a new document with enhanced features
	doc := docx.New().WithDefaultTheme()

	// Add document title
	titlePara := doc.AddParagraph()
	titlePara.AddText("SlideLang Enhanced Document").Bold().Size("32").Color("2E75B6")

	// Add some space
	doc.AddParagraph()

	// Add Table of Contents with all options
	fmt.Println("📖 Adding Table of Contents...")
	opts := docx.DefaultTOCOptions()
	opts.Title = "Table of Contents"
	opts.Depth = 3
	opts.PageNumbers = true
	opts.Hyperlinks = true
	err := doc.AddTOC(opts)
	if err != nil {
		panic(err)
	}

	// Add page break after TOC
	pageBreakPara := doc.AddParagraph()
	pageBreakPara.AddPageBreaks()

	// Add main content with headings and bookmarks
	fmt.Println("📝 Adding content with headings and bookmarks...")

	// Chapter 1
	h1 := doc.AddHeadingWithTOC("1. Introduction", 1, 1)
	h1.Style("Heading1")

	introPara := doc.AddParagraph()
	introPara.AddText("Welcome to the SlideLang enhanced fork of go-docx! This document demonstrates the new features including:")

	// Bullet list with proper indentation (720 twips = 0.5 inch)
	bullet1 := doc.AddParagraph()
	bullet1.AddText("•  Dynamic Table of Contents with page numbers")
	bullet1.Indent(720, 0, 0)

	bullet2 := doc.AddParagraph()
	bullet2.AddText("•  Bookmarks for cross-references")
	bullet2.Indent(720, 0, 0)

	bullet3 := doc.AddParagraph()
	bullet3.AddText("•  Field codes for auto-updating content")
	bullet3.Indent(720, 0, 0)

	bullet4 := doc.AddParagraph()
	bullet4.AddText("•  Professional heading styles")
	bullet4.Indent(720, 0, 0)

	// Chapter 1.1
	h11 := doc.AddHeadingWithTOC("1.1 Key Features", 2, 2)
	h11.Style("Heading2")

	featuresPara := doc.AddParagraph()
	featuresPara.AddText("The enhanced library provides professional document generation capabilities needed for ")
	featuresPara.AddText("DocLang").Bold()
	featuresPara.AddText(" and ")
	featuresPara.AddText("SlideLang").Bold()
	featuresPara.AddText(" exporters.")

	// Chapter 2
	h2 := doc.AddHeadingWithTOC("2. Technical Implementation", 1, 3)
	h2.Style("Heading1")

	techPara := doc.AddParagraph()
	techPara.AddText("This section covers the technical details. See section ")
	// Add cross-reference to introduction
	techPara.AddRefField("_Toc000000001", true)
	techPara.AddText(" for background information.")

	// Chapter 2.1 - demonstrate fields
	h21 := doc.AddHeadingWithTOC("2.1 Field Codes", 2, 4)
	h21.Style("Heading2")

	fieldPara := doc.AddParagraph()
	fieldPara.AddText("This document was generated on page ")
	fieldPara.AddPageField()
	fieldPara.AddText(" of ")
	fieldPara.AddNumPagesField()
	fieldPara.AddText(" total pages.")

	// Add figure with caption
	figPara := doc.AddParagraph()
	figPara.AddText("Figure ")
	figPara.AddSeqField("Figure", "ARABIC")
	figPara.AddText(": Document structure diagram")

	// Chapter 3
	h3 := doc.AddHeadingWithTOC("3. Future Enhancements", 1, 5)
	h3.Style("Heading1")

	futurePara := doc.AddParagraph()
	futurePara.AddText("Future versions will include:")

	// Enhancement bullet list with proper indentation
	enh1 := doc.AddParagraph()
	enh1.AddText("•  Native Heading1-4 style definitions")
	enh1.Indent(720, 0, 0)

	enh2 := doc.AddParagraph()
	enh2.AddText("•  Advanced headers and footers")
	enh2.Indent(720, 0, 0)

	enh3 := doc.AddParagraph()
	enh3.AddText("•  Style customization API")
	enh3.Indent(720, 0, 0)

	enh4 := doc.AddParagraph()
	enh4.AddText("•  Bibliography support")
	enh4.Indent(720, 0, 0)

	enh5 := doc.AddParagraph()
	enh5.AddText("•  Track changes integration")
	enh5.Indent(720, 0, 0)

	// Add page number indicator at end of document
	doc.AddParagraph() // spacing
	footerPara := doc.AddParagraph()
	footerPara.AddText("Page ")
	footerPara.AddPageField()

	// Add page size at the END (important!)
	doc.WithA4Page()

	// Save the document
	filename := "slidelang_enhanced_demo.docx"
	fmt.Printf("💾 Saving document as %s...\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = doc.WriteTo(file)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Demo document created successfully!")
	fmt.Printf("📁 Open %s in Microsoft Word to see:\n", filename)
	fmt.Println("   • Dynamic Table of Contents (press F9 to update)")
	fmt.Println("   • Clickable hyperlinks in TOC")
	fmt.Println("   • Auto-updating page numbers")
	fmt.Println("   • Cross-references between sections")
	fmt.Println("   • Professional document structure")
	fmt.Println()
	fmt.Println("🎯 This demonstrates the core features needed for DocLang/SlideLang!")
}
