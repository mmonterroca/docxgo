// Package main demonstrates the v0.3.0 features: headers, footers, hyperlinks, and modern fonts
package main

import (
	"fmt"
	"os"

	"github.com/fumiama/go-docx"
)

const modernFont = "Calibri"

func main() {
	fmt.Println("🎓 Go-Docx v0.3.0 Professional Demo")
	fmt.Println("====================================")

	// Create document with theme
	doc := docx.New().WithDefaultTheme()

	// ADD FOOTER WITH PAGE NUMBERS (v0.3.0)
	fmt.Println("🔢 Adding page number footer...")
	doc.AddPageNumberFooter()

	// COVER PAGE
	fmt.Println("🎨 Creating cover page...")
	
	// Add spacing
	doc.AddParagraph()
	doc.AddParagraph()
	
	// Logo/Company name
	logo := doc.AddParagraph()
	logo.AddText("SLIDELANG").Size("28").Bold().Color("2E75B5")
	logo.Justification("center")
	
	doc.AddParagraph()
	
	// Document title
	title := doc.AddParagraph()
	title.AddText("Go-Docx Professional Document").Size("36").Bold().Color("1F4E78")
	title.Justification("center")
	
	// Subtitle
	subtitle := doc.AddParagraph()
	subtitle.AddText("Enhanced Document Generation Library").Size("18").Color("5B9BD5")
	subtitle.Justification("center")
	
	doc.AddParagraph()
	doc.AddParagraph()
	
	// Version
	version := doc.AddParagraph()
	version.AddText("Version v0.3.0").Size("16")
	version.Justification("center")
	
	// Date
	date := doc.AddParagraph()
	date.AddText("October 22, 2025").Size("14").Color("7F7F7F")
	date.Justification("center")
	
	// Page break after cover
	doc.AddParagraph().AddPageBreaks()

	// TABLE OF CONTENTS
	fmt.Println("📖 Adding Table of Contents...")
	opts := docx.DefaultTOCOptions()
	opts.Title = "Table of Contents"
	opts.Depth = 3
	opts.PageNumbers = true
	opts.Hyperlinks = true
	doc.AddTOC(opts)

	doc.AddParagraph().AddPageBreaks()

	// CHAPTER 1: INTRODUCTION
	fmt.Println("📝 Adding Chapter 1: Introduction...")
	doc.AddSmartHeading("Introduction", 1)

	p1 := doc.AddParagraph()
	p1.AddText("The go-docx library is a powerful tool for generating Microsoft Word documents in Go. This version includes:")

	features := []string{
		"Table of Contents with automatic bookmarks",
		"Native heading styles (Heading 1-4)",
		"Field codes (PAGE, PAGEREF, SEQ, HYPERLINK, STYLEREF)",
		"Paragraph indentation API",
		"Modern typography with Calibri font",
		"Headers and footers with page numbers",
		"Hyperlinks for external and internal references",
	}

	for _, feature := range features {
		bullet := doc.AddParagraph()
		bullet.AddText("• " + feature)
		bullet.Indent(720, 0, 0)
	}

	doc.AddParagraph().AddPageBreaks()

	// CHAPTER 2: FEATURES
	fmt.Println("✨ Adding Chapter 2: Features...")
	doc.AddSmartHeading("Features Overview", 1)

	doc.AddSmartHeading("Implemented Features", 2)

	p2 := doc.AddParagraph()
	p2.AddText("The following features are fully implemented:")

	implementedFeatures := []struct {
		name string
		desc string
	}{
		{"Bookmarks", "Named locations for cross-referencing"},
		{"Field Codes", "Dynamic content with various field types"},
		{"TOC Builder", "Automatic Table of Contents generation"},
		{"Native Styles", "Built-in Heading 1-4 styles"},
		{"Indentation", "Precise paragraph indent control"},
		{"Modern Fonts", "Calibri font support"},
		{"Headers/Footers", "Page headers and footers"},
		{"Hyperlinks", "External and internal links"},
	}

	for _, feat := range implementedFeatures {
		bullet := doc.AddParagraph()
		bullet.AddText("• ")
		bullet.AddText(feat.name).Bold()
		bullet.AddText(": " + feat.desc)
		bullet.Indent(720, 0, 0)
	}

	doc.AddParagraph().AddPageBreaks()

	// CHAPTER 3: CODE EXAMPLES
	fmt.Println("💻 Adding Chapter 3: Code Examples...")
	doc.AddSmartHeading("Code Examples", 1)

	doc.AddSmartHeading("Basic Document Creation", 2)

	p3 := doc.AddParagraph()
	p3.AddText("Create a document with modern features:")

	doc.AddParagraph()

	code1 := doc.AddParagraph()
	code1.AddText(`doc := docx.New().WithDefaultTheme()

// Add footer with page numbers
doc.AddPageNumberFooter()

// Add content
para := doc.AddParagraph()
para.AddText("Hello, World!")`).Size("18")
	code1.Indent(720, 0, 0)

	doc.AddParagraph()

	doc.AddSmartHeading("Adding Hyperlinks", 2)

	p4 := doc.AddParagraph()
	p4.AddText("Create external and internal links:")

	doc.AddParagraph()

	code2 := doc.AddParagraph()
	code2.AddText(`para := doc.AddParagraph()
para.AddText("Visit ")
para.AddHyperlinkField("https://github.com/SlideLang/go-docx",
    "GitHub", "Repository link")`).Size("18")
	code2.Indent(720, 0, 0)

	doc.AddParagraph().AddPageBreaks()

	// CHAPTER 4: TABLES
	fmt.Println("📊 Adding Chapter 4: Tables...")
	doc.AddSmartHeading("Tables and Data", 1)

	p5 := doc.AddParagraph()
	p5.AddText("Tables for structured data presentation.")

	doc.AddParagraph()

	doc.AddSmartHeading("Version History", 2)

	doc.AddParagraph()
	caption := doc.AddParagraph()
	caption.AddText("Table ")
	caption.AddSeqField("Table", "ARABIC")
	caption.AddText(": Version History").Italic()

	doc.AddParagraph()

	table := doc.AddTable(4, 3, 0, nil)

	// Header row
	table.TableRows[0].TableCells[0].AddParagraph().AddText("Version").Bold()
	table.TableRows[0].TableCells[1].AddParagraph().AddText("Date").Bold()
	table.TableRows[0].TableCells[2].AddParagraph().AddText("Features").Bold()

	// Data rows
	table.TableRows[1].TableCells[0].AddParagraph().AddText("v0.1.0")
	table.TableRows[1].TableCells[1].AddParagraph().AddText("Oct 21, 2025")
	table.TableRows[1].TableCells[2].AddParagraph().AddText("TOC, Bookmarks, Fields")

	table.TableRows[2].TableCells[0].AddParagraph().AddText("v0.2.0")
	table.TableRows[2].TableCells[1].AddParagraph().AddText("Oct 22, 2025")
	table.TableRows[2].TableCells[2].AddParagraph().AddText("Indentation, Spacing Fix")

	table.TableRows[3].TableCells[0].AddParagraph().AddText("v0.3.0")
	table.TableRows[3].TableCells[1].AddParagraph().AddText("Oct 22, 2025")
	table.TableRows[3].TableCells[2].AddParagraph().AddText("Headers/Footers, Hyperlinks")

	doc.AddParagraph().AddPageBreaks()

	// CHAPTER 5: REFERENCES WITH HYPERLINKS (v0.3.0)
	fmt.Println("🔗 Adding Chapter 5: References...")
	doc.AddSmartHeading("References and Links", 1)

	p6 := doc.AddParagraph()
	p6.AddText("External resources demonstrating hyperlink functionality:")

	doc.AddParagraph()

	doc.AddSmartHeading("External Resources", 2)

	ref1 := doc.AddParagraph()
	ref1.AddText("[1] ")
	ref1.AddText("SlideLang Repository: ").Bold()
	ref1.AddHyperlinkField("https://github.com/SlideLang/go-docx", "GitHub", "Visit repository")

	ref2 := doc.AddParagraph()
	ref2.AddText("[2] ")
	ref2.AddText("Office Open XML Spec: ").Bold()
	ref2.AddHyperlinkField("https://docs.microsoft.com/office/open-xml", "Microsoft Docs", "Learn OOXML")

	ref3 := doc.AddParagraph()
	ref3.AddText("[3] ")
	ref3.AddText("Go Language: ").Bold()
	ref3.AddHyperlinkField("https://go.dev", "go.dev", "Official Go website")

	doc.AddParagraph()
	doc.AddParagraph()

	// Conclusion
	doc.AddSmartHeading("Conclusion", 2)

	conclusion := doc.AddParagraph()
	conclusion.AddText("This document demonstrates the v0.3.0 capabilities including modern fonts, page numbering, and hyperlinks. For more information, visit the ")
	conclusion.AddHyperlinkField("https://github.com/SlideLang/go-docx", "GitHub repository", "")
	conclusion.AddText(".")

	// Save document
	outputPath := "v030_demo.docx"
	fmt.Printf("💾 Saving document to %s...\n", outputPath)

	w, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()

	if _, err := doc.WriteTo(w); err != nil {
		fmt.Printf("❌ Error writing: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Document generated successfully!")
	fmt.Printf("📄 Open %s in Microsoft Word\n", outputPath)
}
