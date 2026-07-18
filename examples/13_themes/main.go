// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Example: Document Themes
//
// This example demonstrates the theme system in go-docx:
//  1. Lists all available preset themes
//  2. Creates a themed document for each preset
//  3. Creates a comparison document showing all themes side by side
//
// Themes control:
//   - Colors (primary, secondary, accent, text, headings)
//   - Fonts (body, heading, monospace)
//   - Spacing (paragraph, line, heading, section)
//   - Heading styles (size, bold, uppercase, color)
package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/themes"
)

func main() {
	fmt.Println("🎨 Document Themes Example")
	fmt.Println("==========================")
	fmt.Println()

	// Get all available themes (Corporate, Startup, Modern, Fintech, Academic)
	allThemes := themes.AllThemes()

	// Also include the tech themes
	allThemes = append(allThemes, themes.TechPresentation, themes.TechDarkMode)

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
		fmt.Printf("  ✅ Created: %s_theme.docx\n", theme.Name())
	}
	fmt.Println()

	// Also create a comparison document showing all themes
	fmt.Println("Creating theme comparison document...")
	if err := createComparisonDocument(allThemes); err != nil {
		log.Fatalf("Error creating comparison document: %v", err)
	}
	fmt.Println("✅ Created: theme_comparison.docx")
	fmt.Println()
	fmt.Println("✨ All documents created successfully!")
}

// createThemedDocument creates a sample document using the specified theme.
func createThemedDocument(theme themes.Theme) error {
	// Create document with theme applied via builder
	builder := docx.NewDocumentBuilder(
		docx.WithTheme(theme),
		docx.WithTitle(fmt.Sprintf("Sample Document - %s Theme", theme.DisplayName())),
		docx.WithAuthor("go-docx Themes"),
	)

	// Add initial content via builder, then Build() to get the document.
	// After building, use the domain API to apply styles and add more content.
	builder.AddParagraph().
		Text("Sample Document").
		End()

	doc, err := builder.Build()
	if err != nil {
		return err
	}

	// Apply Title style to the first paragraph
	paras := doc.Paragraphs()
	if len(paras) > 0 {
		paras[0].SetStyle(domain.StyleIDTitle)
	}

	// Subtitle with theme name
	subtitlePara, _ := doc.AddParagraph()
	subtitlePara.SetStyle(domain.StyleIDSubtitle)
	subtitlePara.SetAlignment(domain.AlignmentCenter)
	subtitleRun, _ := subtitlePara.AddRun()
	subtitleRun.AddText(fmt.Sprintf("Using %s Theme", theme.DisplayName()))

	// Spacer
	doc.AddParagraph()

	// --- Section 1: Introduction ---
	h1, _ := doc.AddParagraph()
	h1.SetStyle(domain.StyleIDHeading1)
	h1Run, _ := h1.AddRun()
	h1Run.AddText("1. Introduction")

	introPara, _ := doc.AddParagraph()
	introPara.SetStyle(domain.StyleIDNormal)
	introRun, _ := introPara.AddRun()
	introRun.AddText(theme.Description() + ". This document demonstrates the visual style and formatting applied by this theme.")

	// --- Section 2: Typography and Hierarchy ---
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

	// --- Section 3: Special Elements ---
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

	// Quotes
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

	// --- Section 4: Theme Details ---
	h1_4, _ := doc.AddParagraph()
	h1_4.SetStyle(domain.StyleIDHeading1)
	h1_4Run, _ := h1_4.AddRun()
	h1_4Run.AddText("4. Theme Details")

	// Theme colors
	colors := theme.Colors()
	colorsPara, _ := doc.AddParagraph()
	colorsPara.SetStyle(domain.StyleIDNormal)
	colorsRun, _ := colorsPara.AddRun()
	colorsRun.AddText(fmt.Sprintf(
		"Primary Color: RGB(%d, %d, %d) • Secondary: RGB(%d, %d, %d) • Accent: RGB(%d, %d, %d)",
		colors.Primary.R, colors.Primary.G, colors.Primary.B,
		colors.Secondary.R, colors.Secondary.G, colors.Secondary.B,
		colors.Accent.R, colors.Accent.G, colors.Accent.B,
	))

	// Theme fonts
	fonts := theme.Fonts()
	fontsPara, _ := doc.AddParagraph()
	fontsPara.SetStyle(domain.StyleIDNormal)
	fontsRun, _ := fontsPara.AddRun()
	fontsRun.AddText(fmt.Sprintf(
		"Body Font: %s (%dpt) • Heading Font: %s • Monospace: %s",
		fonts.Body, fonts.BodySize/2, fonts.Heading, fonts.Monospace,
	))

	// Save the document
	filename := fmt.Sprintf("%s_theme.docx", theme.Name())
	return doc.SaveAs(filename)
}

// createComparisonDocument creates a single document comparing all themes.
func createComparisonDocument(allThemes []themes.Theme) error {
	doc := docx.NewDocument()

	// Title
	titlePara, _ := doc.AddParagraph()
	titlePara.SetStyle(domain.StyleIDTitle)
	titlePara.SetAlignment(domain.AlignmentCenter)
	titleRun, _ := titlePara.AddRun()
	titleRun.AddText("Theme Comparison Guide")

	// Subtitle
	subtitlePara, _ := doc.AddParagraph()
	subtitlePara.SetStyle(domain.StyleIDSubtitle)
	subtitlePara.SetAlignment(domain.AlignmentCenter)
	subtitleRun, _ := subtitlePara.AddRun()
	subtitleRun.AddText("go-docx Document Themes")

	doc.AddParagraph()

	// Introduction
	introPara, _ := doc.AddParagraph()
	introPara.SetStyle(domain.StyleIDNormal)
	introRun, _ := introPara.AddRun()
	introRun.AddText("This document provides an overview of all available themes in go-docx. Each theme defines a complete visual style including colors, fonts, spacing, and formatting rules.")

	doc.AddParagraph()

	// List each theme
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

		// Theme ID
		namePara, _ := doc.AddParagraph()
		namePara.SetStyle(domain.StyleIDNormal)
		nameRun1, _ := namePara.AddRun()
		nameRun1.SetBold(true)
		nameRun1.AddText("Theme ID: ")
		nameRun2, _ := namePara.AddRun()
		nameRun2.AddText(theme.Name())

		// Usage example
		usagePara, _ := doc.AddParagraph()
		usagePara.SetStyle(domain.StyleIDNormal)
		usageRun, _ := usagePara.AddRun()
		usageRun.SetBold(true)
		usageRun.AddText("Usage:")

		codePara, _ := doc.AddParagraph()
		codePara.SetStyle(domain.StyleIDListParagraph)
		codeRun, _ := codePara.AddRun()
		codeRun.SetFont(domain.Font{Name: "Courier New"})
		codeRun.AddText("docx.NewDocumentBuilder(docx.WithTheme(theme))")

		// Colors info
		colors := theme.Colors()
		colorPara, _ := doc.AddParagraph()
		colorPara.SetStyle(domain.StyleIDNormal)
		colorRun1, _ := colorPara.AddRun()
		colorRun1.SetBold(true)
		colorRun1.AddText("Colors: ")
		colorRun2, _ := colorPara.AddRun()
		colorRun2.AddText(fmt.Sprintf("Primary RGB(%d,%d,%d), Accent RGB(%d,%d,%d)",
			colors.Primary.R, colors.Primary.G, colors.Primary.B,
			colors.Accent.R, colors.Accent.G, colors.Accent.B))

		// Fonts info
		fonts := theme.Fonts()
		fontPara, _ := doc.AddParagraph()
		fontPara.SetStyle(domain.StyleIDNormal)
		fontRun1, _ := fontPara.AddRun()
		fontRun1.SetBold(true)
		fontRun1.AddText("Fonts: ")
		fontRun2, _ := fontPara.AddRun()
		fontRun2.AddText(fmt.Sprintf("Body: %s (%dpt), Heading: %s",
			fonts.Body, fonts.BodySize/2, fonts.Heading))

		doc.AddParagraph()
	}

	// Custom themes section
	conclusionPara, _ := doc.AddParagraph()
	conclusionPara.SetStyle(domain.StyleIDHeading1)
	conclusionRun, _ := conclusionPara.AddRun()
	conclusionRun.AddText("Custom Themes")

	customPara, _ := doc.AddParagraph()
	customPara.SetStyle(domain.StyleIDNormal)
	customRun, _ := customPara.AddRun()
	customRun.AddText("You can also create custom themes or modify existing ones using Clone() and With*() methods:")

	codeExamples := []string{
		"customTheme := themes.Corporate.Clone()",
		"colors := customTheme.Colors()",
		"customTheme = customTheme.WithColors(colors)",
		"customTheme = customTheme.WithFonts(newFonts)",
	}
	for _, code := range codeExamples {
		codePara, _ := doc.AddParagraph()
		codePara.SetStyle(domain.StyleIDListParagraph)
		codeRun, _ := codePara.AddRun()
		codeRun.SetFont(domain.Font{Name: "Courier New"})
		codeRun.AddText(code)
	}

	return doc.SaveAs("theme_comparison.docx")
}
