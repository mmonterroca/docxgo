package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/themes"
)

func main() {
	fmt.Println("🎨 Document Themes Example")
	fmt.Println("==========================")
	fmt.Println()

	// Get all available themes
	allThemes := themes.AllThemes()
	
	fmt.Printf("Available themes: %d\n", len(allThemes))
	for _, theme := range allThemes {
		fmt.Printf("  - %s: %s\n", theme.DisplayName(), theme.Description())
	}
	fmt.Println()

	// Create a document for each theme
	for _, theme := range allThemes {
		fmt.Printf("Creating document with %s theme...\n", theme.DisplayName())
		if err := createThemedDocument(theme); err != nil {
			log.Printf("Error creating %s document: %v\n", theme.DisplayName(), err)
			continue
		}
		fmt.Printf("✅ Created: %s_theme.docx\n\n", theme.Name())
	}

	// Also create a comparison document showing all themes
	fmt.Println("Creating theme comparison document...")
	if err := createComparisonDocument(); err != nil {
		log.Fatalf("Error creating comparison document: %v", err)
	}
	fmt.Println("✅ Created: theme_comparison.docx")
	fmt.Println()
	fmt.Println("✨ All documents created successfully!")
}

// createThemedDocument creates a sample document using the specified theme.
func createThemedDocument(theme themes.Theme) error {
	// Create document with theme
	builder := docx.NewDocumentBuilder(
		docx.WithTheme(theme),
		docx.WithTitle(fmt.Sprintf("Sample Document - %s Theme", theme.DisplayName())),
		docx.WithAuthor("go-docx Themes"),
	)

	// Add title
	builder.AddParagraph().
		Text(fmt.Sprintf("Sample Document")).
		End()

	// Apply Title style manually to the last paragraph
	doc, err := builder.Build()
	if err != nil {
		return err
	}
	
	paras := doc.Paragraphs()
	if len(paras) > 0 {
		paras[0].SetStyle(domain.StyleIDTitle)
	}

	// Add subtitle with theme name
	subtitlePara, _ := doc.AddParagraph()
	subtitlePara.SetStyle(domain.StyleIDSubtitle)
	subtitleRun, _ := subtitlePara.AddRun()
	subtitleRun.AddText(fmt.Sprintf("Using %s Theme", theme.DisplayName()))

	// Add spacing
	doc.AddParagraph()

	// Section 1: Introduction
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Introduction")

	introPara, _ := doc.AddParagraph()
	introPara.SetStyle(domain.StyleIDNormal)
	introRun, _ := introPara.AddRun()
	introRun.AddText(theme.Description() + ". This document demonstrates the visual style and formatting applied by this theme.")

	// Section 2: Colors and Formatting
	h1_2, _ := doc.AddParagraph()
	h1_2.SetStyle(domain.StyleIDHeading1)
	h1_2Run, _ := h1_2.AddRun()
	h1_2Run.AddText("2. Typography and Hierarchy")

	// H2 example
	h2, _ := doc.AddParagraph()
	h2.SetStyle(domain.StyleIDHeading2)
	h2Run, _ := h2.AddRun()
	h2Run.AddText("2.1 Heading Level 2")

	para1, _ := doc.AddParagraph()
	para1.SetStyle(domain.StyleIDNormal)
	run1, _ := para1.AddRun()
	run1.AddText("This is a paragraph using the Normal style. The theme controls the font family, size, color, and spacing. ")
	run2, _ := para1.AddRun()
	run2.SetBold(true)
	run2.AddText("Bold text")
	run3, _ := para1.AddRun()
	run3.AddText(" and ")
	run4, _ := para1.AddRun()
	run4.SetItalic(true)
	run4.AddText("italic text")
	run5, _ := para1.AddRun()
	run5.AddText(" are also styled according to the theme.")

	// H3 example
	h3, _ := doc.AddParagraph()
	h3.SetStyle(domain.StyleIDHeading3)
	h3Run, _ := h3.AddRun()
	h3Run.AddText("2.1.1 Heading Level 3")

	para2, _ := doc.AddParagraph()
	para2.SetStyle(domain.StyleIDNormal)
	run6, _ := para2.AddRun()
	run6.AddText("Headings provide a clear visual hierarchy. Each theme defines specific sizes, colors, and weights for H1, H2, and H3 styles.")

	// Section 3: Lists and Quotes
	h1_3, _ := doc.AddParagraph()
	h1_3.SetStyle(domain.StyleIDHeading1)
	h1_3Run, _ := h1_3.AddRun()
	h1_3Run.AddText("3. Special Elements")

	h2_2, _ := doc.AddParagraph()
	h2_2.SetStyle(domain.StyleIDHeading2)
	h2_2Run, _ := h2_2.AddRun()
	h2_2Run.AddText("3.1 List Paragraphs")

	// List items
	listItems := []string{
		"First item with List Paragraph style",
		"Second item showing consistent formatting",
		"Third item demonstrating spacing and indentation",
	}

	for _, item := range listItems {
		listPara, _ := doc.AddParagraph()
		listPara.SetStyle(domain.StyleIDListParagraph)
		listRun, _ := listPara.AddRun()
		listRun.AddText("• " + item)
	}

	// Quote section
	h2_3, _ := doc.AddParagraph()
	h2_3.SetStyle(domain.StyleIDHeading2)
	h2_3Run, _ := h2_3.AddRun()
	h2_3Run.AddText("3.2 Quotations")

	quotePara, _ := doc.AddParagraph()
	quotePara.SetStyle(domain.StyleIDQuote)
	quoteRun, _ := quotePara.AddRun()
	quoteRun.AddText("This is a quote using the Quote style. Quotes are typically styled with subtle differences to distinguish them from body text.")

	intenseQuotePara, _ := doc.AddParagraph()
	intenseQuotePara.SetStyle(domain.StyleIDIntenseQuote)
	intenseQuoteRun, _ := intenseQuotePara.AddRun()
	intenseQuoteRun.AddText("This is an Intense Quote, designed to stand out more prominently.")

	// Section 4: Theme Information
	h1_4, _ := doc.AddParagraph()
	h1_4.SetStyle(domain.StyleIDHeading1)
	h1_4Run, _ := h1_4.AddRun()
	h1_4Run.AddText("4. Theme Details")

	// Theme colors
	colors := theme.Colors()
	colorsPara, _ := doc.AddParagraph()
	colorsPara.SetStyle(domain.StyleIDNormal)
	colorsPara.AddRun()
	colorsPara.Runs()[0].AddText(fmt.Sprintf(
		"Primary Color: RGB(%d, %d, %d) • Secondary: RGB(%d, %d, %d) • Accent: RGB(%d, %d, %d)",
		colors.Primary.R, colors.Primary.G, colors.Primary.B,
		colors.Secondary.R, colors.Secondary.G, colors.Secondary.B,
		colors.Accent.R, colors.Accent.G, colors.Accent.B,
	))

	// Theme fonts
	fonts := theme.Fonts()
	fontsPara, _ := doc.AddParagraph()
	fontsPara.SetStyle(domain.StyleIDNormal)
	fontsPara.AddRun()
	fontsPara.Runs()[0].AddText(fmt.Sprintf(
		"Body Font: %s (%dpt) • Heading Font: %s • Monospace: %s",
		fonts.Body, fonts.BodySize/2, fonts.Heading, fonts.Monospace,
	))

	// Save the document
	filename := fmt.Sprintf("%s_theme.docx", theme.Name())
	if err := doc.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

// createComparisonDocument creates a single document comparing all themes.
func createComparisonDocument() error {
	// Create document without a theme (will use defaults)
	doc := docx.NewDocument()

	// Add title
	titlePara, _ := doc.AddParagraph()
	titlePara.SetStyle(domain.StyleIDTitle)
	titleRun, _ := titlePara.AddRun()
	titleRun.AddText("Theme Comparison Guide")

	// Add subtitle
	subtitlePara, _ := doc.AddParagraph()
	subtitlePara.SetStyle(domain.StyleIDSubtitle)
	subtitleRun, _ := subtitlePara.AddRun()
	subtitleRun.AddText("go-docx Document Themes")

	doc.AddParagraph()

	// Introduction
	introPara, _ := doc.AddParagraph()
	introPara.SetStyle(domain.StyleIDNormal)
	introRun, _ := introPara.AddRun()
	introRun.AddText("This document provides an overview of all available themes in go-docx. Each theme defines a complete visual style including colors, fonts, spacing, and formatting rules.")

	doc.AddParagraph()

	// List all themes
	allThemes := themes.AllThemes()
	for i, theme := range allThemes {
		// Theme heading
		themePara, _ := doc.AddParagraph()
		themePara.SetStyle(domain.StyleIDHeading1)
		themeRun, _ := themePara.AddRun()
		themeRun.AddText(fmt.Sprintf("%d. %s", i+1, theme.DisplayName()))

		// Description
		descPara, _ := doc.AddParagraph()
		descPara.SetStyle(domain.StyleIDNormal)
		descRun, _ := descPara.AddRun()
		descRun.AddText(theme.Description())

		// Theme name
		namePara, _ := doc.AddParagraph()
		namePara.SetStyle(domain.StyleIDNormal)
		namePara.AddRun()
		namePara.Runs()[0].SetBold(true)
		namePara.Runs()[0].AddText("Theme ID: ")
		nameRun2, _ := namePara.AddRun()
		nameRun2.AddText(theme.Name())

		// Usage example
		usagePara, _ := doc.AddParagraph()
		usagePara.SetStyle(domain.StyleIDNormal)
		usagePara.AddRun()
		usagePara.Runs()[0].SetBold(true)
		usagePara.Runs()[0].AddText("Usage:")
		
		codePara, _ := doc.AddParagraph()
		codePara.SetStyle(domain.StyleIDListParagraph)
		codePara.AddRun()
		codePara.Runs()[0].SetFont(domain.Font{Name: "Courier New"})
		codePara.Runs()[0].AddText(fmt.Sprintf("docx.NewDocumentBuilder(docx.WithTheme(themes.%s))", 
			capitalizeFirst(theme.Name())))

		// Colors info
		colors := theme.Colors()
		colorPara, _ := doc.AddParagraph()
		colorPara.SetStyle(domain.StyleIDNormal)
		colorPara.AddRun()
		colorPara.Runs()[0].SetBold(true)
		colorPara.Runs()[0].AddText("Colors: ")
		colorRun2, _ := colorPara.AddRun()
		colorRun2.AddText(fmt.Sprintf("Primary RGB(%d,%d,%d), Accent RGB(%d,%d,%d)",
			colors.Primary.R, colors.Primary.G, colors.Primary.B,
			colors.Accent.R, colors.Accent.G, colors.Accent.B))

		// Add spacing between themes
		doc.AddParagraph()
	}

	// Conclusion
	conclusionPara, _ := doc.AddParagraph()
	conclusionPara.SetStyle(domain.StyleIDHeading1)
	conclusionRun, _ := conclusionPara.AddRun()
	conclusionRun.AddText("Custom Themes")

	customPara, _ := doc.AddParagraph()
	customPara.SetStyle(domain.StyleIDNormal)
	customRun, _ := customPara.AddRun()
	customRun.AddText("You can also create custom themes or modify existing ones:")

	examplePara, _ := doc.AddParagraph()
	examplePara.SetStyle(domain.StyleIDListParagraph)
	examplePara.AddRun()
	examplePara.Runs()[0].SetFont(domain.Font{Name: "Courier New"})
	examplePara.Runs()[0].AddText("customTheme := themes.Corporate.Clone()")

	examplePara2, _ := doc.AddParagraph()
	examplePara2.SetStyle(domain.StyleIDListParagraph)
	examplePara2.AddRun()
	examplePara2.Runs()[0].SetFont(domain.Font{Name: "Courier New"})
	examplePara2.Runs()[0].AddText("colors := customTheme.Colors()")

	examplePara3, _ := doc.AddParagraph()
	examplePara3.SetStyle(domain.StyleIDListParagraph)
	examplePara3.AddRun()
	examplePara3.Runs()[0].SetFont(domain.Font{Name: "Courier New"})
	examplePara3.Runs()[0].AddText("colors.Primary = docx.Color{R: 255, G: 0, B: 0}")

	examplePara4, _ := doc.AddParagraph()
	examplePara4.SetStyle(domain.StyleIDListParagraph)
	examplePara4.AddRun()
	examplePara4.Runs()[0].SetFont(domain.Font{Name: "Courier New"})
	examplePara4.Runs()[0].AddText("customTheme = customTheme.WithColors(colors)")

	// Save the document
	if err := doc.SaveAs("theme_comparison.docx"); err != nil {
		return fmt.Errorf("failed to save comparison document: %w", err)
	}

	return nil
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
