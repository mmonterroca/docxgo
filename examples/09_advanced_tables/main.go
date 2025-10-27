/*
MIT License

Copyright (c) 2025 Misael Monterroca

Example: Advanced table features - cell merging, nested tables, and styling

This example demonstrates:
- Horizontal cell merging (colspan)
- Vertical cell merging (rowspan)
- Combined horizontal and vertical merging
- Nested tables within cells
- Table styling
- Row height control
- Complex table layouts (calendar, invoice, etc.)
*/

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
)

func main() {
	builder := docx.NewDocumentBuilder()

	// Title
	builder.AddParagraph().
		Text("Advanced Table Features").
		Bold().
		FontSize(24).
		Alignment(domain.AlignmentCenter).
		End()

	builder.AddParagraph().Text("").End()

	// Example 1: Horizontal cell merging
	builder.AddParagraph().
		Text("1. Horizontal Cell Merging (Colspan)").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Merging cells horizontally to create headers or spans:").
		End()

	builder.AddTable(3, 4).
		Style(domain.TableStyleGrid).
		Row(0).
		Cell(0).Text("Header (spans 4 columns)").Bold().Merge(4, 1).End().
		End().
		Row(1).
		Cell(0).Text("A1").End().
		Cell(1).Text("B1").End().
		Cell(2).Text("C1").End().
		Cell(3).Text("D1").End().
		End().
		Row(2).
		Cell(0).Text("A2 (spans 2)").Merge(2, 1).End().
		Cell(2).Text("C2 (spans 2)").Merge(2, 1).End().
		End().
		End()

	builder.AddParagraph().Text("").End()

	// Example 2: Vertical cell merging
	builder.AddParagraph().
		Text("2. Vertical Cell Merging (Rowspan)").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Merging cells vertically for labels or categories:").
		End()

	builder.AddTable(4, 3).
		Style(domain.TableStyleGrid).
		Row(0).
		Cell(0).Text("Header 1").Bold().End().
		Cell(1).Text("Header 2").Bold().End().
		Cell(2).Text("Header 3").Bold().End().
		End().
		Row(1).
		Cell(0).Text("Category A\n(3 rows)").VerticalAlignment(domain.VerticalAlignCenter).Merge(1, 3).End().
		Cell(1).Text("Data 1").End().
		Cell(2).Text("Value 1").End().
		End().
		Row(2).
		Cell(1).Text("Data 2").End().
		Cell(2).Text("Value 2").End().
		End().
		Row(3).
		Cell(1).Text("Data 3").End().
		Cell(2).Text("Value 3").End().
		End().
		End()

	builder.AddParagraph().Text("").End()

	// Example 3: Combined merging (2x2 region)
	builder.AddParagraph().
		Text("3. Combined Horizontal and Vertical Merging").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Creating complex layouts with both colspan and rowspan:").
		End()

	builder.AddTable(4, 4).
		Style(domain.TableStyleGrid).
		Row(0).
		Cell(0).Text("2x2 Merged\nCell").
		VerticalAlignment(domain.VerticalAlignCenter).
		Merge(2, 2).End().
		Cell(2).Text("C1").End().
		Cell(3).Text("D1").End().
		End().
		Row(1).
		Cell(2).Text("C2").End().
		Cell(3).Text("D2").End().
		End().
		Row(2).
		Cell(0).Text("A3").End().
		Cell(1).Text("B3").End().
		Cell(2).Text("C3").End().
		Cell(3).Text("D3").End().
		End().
		Row(3).
		Cell(0).Text("Footer (spans all 4 columns)").Merge(4, 1).End().
		End().
		End()

	builder.AddParagraph().Text("").End()

	// Example 4: Calendar-style layout
	builder.AddParagraph().
		Text("4. Calendar Layout").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Using merging to create a monthly calendar:").
		End()

	now := time.Now()
	builder.AddTable(7, 7).
		Style(domain.TableStyleMediumShading).
		Row(0).
		Cell(0).Text(now.Format("January 2006")).Bold().Merge(7, 1).End().
		End().
		Row(1).
		Cell(0).Text("Sun").Bold().End().
		Cell(1).Text("Mon").Bold().End().
		Cell(2).Text("Tue").Bold().End().
		Cell(3).Text("Wed").Bold().End().
		Cell(4).Text("Thu").Bold().End().
		Cell(5).Text("Fri").Bold().End().
		Cell(6).Text("Sat").Bold().End().
		End().
		Row(2).
		Cell(0).Text("").End().
		Cell(1).Text("").End().
		Cell(2).Text("1").End().
		Cell(3).Text("2").End().
		Cell(4).Text("3").End().
		Cell(5).Text("4").End().
		Cell(6).Text("5").End().
		End().
		Row(3).
		Cell(0).Text("6").End().
		Cell(1).Text("7").End().
		Cell(2).Text("8").End().
		Cell(3).Text("9").End().
		Cell(4).Text("10").End().
		Cell(5).Text("11").End().
		Cell(6).Text("12").End().
		End().
		Row(4).
		Cell(0).Text("13").End().
		Cell(1).Text("14").End().
		Cell(2).Text("15").End().
		Cell(3).Text("16").End().
		Cell(4).Text("17").End().
		Cell(5).Text("18").End().
		Cell(6).Text("19").End().
		End().
		Row(5).
		Cell(0).Text("20").End().
		Cell(1).Text("21").End().
		Cell(2).Text("22").End().
		Cell(3).Text("23").End().
		Cell(4).Text("24").End().
		Cell(5).Text("25").End().
		Cell(6).Text("26").End().
		End().
		Row(6).
		Cell(0).Text("27").End().
		Cell(1).Text("28").End().
		Cell(2).Text("29").End().
		Cell(3).Text("30").End().
		Cell(4).Text("31").End().
		Cell(5).Text("").End().
		Cell(6).Text("").End().
		End().
		End()

	builder.AddParagraph().Text("").End()

	// Example 5: Nested tables
	builder.AddParagraph().
		Text("5. Nested Tables").
		Bold().
		FontSize(16).
		End()

	builder.AddParagraph().
		Text("Tables within table cells for complex layouts:").
		End()

	// Note: Nested tables need to be added via direct API, not builder yet
	doc, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build document: %v", err)
	}

	// Add nested table example using direct API
	outerTable, err := doc.AddTable(2, 2)
	if err != nil {
		log.Fatalf("Failed to add outer table: %v", err)
	}

	outerTable.SetStyle(domain.TableStyleGrid)

	row0, _ := outerTable.Row(0)
	cell00, _ := row0.Cell(0)
	para00, _ := cell00.AddParagraph()
	run00, _ := para00.AddRun()
	run00.SetText("Outer Cell A1")
	run00.SetBold(true)

	// Add nested table in cell (0,1)
	cell01, _ := row0.Cell(1)
	nested, err := cell01.AddTable(2, 2)
	if err != nil {
		log.Fatalf("Failed to add nested table: %v", err)
	}

	// Populate nested table
	nRow0, _ := nested.Row(0)
	nCell00, _ := nRow0.Cell(0)
	nPara00, _ := nCell00.AddParagraph()
	nRun00, _ := nPara00.AddRun()
	nRun00.SetText("Nested A1")

	nCell01, _ := nRow0.Cell(1)
	nPara01, _ := nCell01.AddParagraph()
	nRun01, _ := nPara01.AddRun()
	nRun01.SetText("Nested B1")

	nRow1, _ := nested.Row(1)
	nCell10, _ := nRow1.Cell(0)
	nPara10, _ := nCell10.AddParagraph()
	nRun10, _ := nPara10.AddRun()
	nRun10.SetText("Nested A2")

	nCell11, _ := nRow1.Cell(1)
	nPara11, _ := nCell11.AddParagraph()
	nRun11, _ := nPara11.AddRun()
	nRun11.SetText("Nested B2")

	// Continue with outer table
	row1, _ := outerTable.Row(1)
	cell10, _ := row1.Cell(0)
	para10, _ := cell10.AddParagraph()
	run10, _ := para10.AddRun()
	run10.SetText("Outer Cell A2")

	cell11, _ := row1.Cell(1)
	para11, _ := cell11.AddParagraph()
	run11, _ := para11.AddRun()
	run11.SetText("Outer Cell B2")

	// Example 6: Invoice-style table
	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("")

	para, _ = doc.AddParagraph()
	run, _ = para.AddRun()
	run.SetText("6. Invoice Layout")
	run.SetBold(true)
	run.SetSize(16)

	para, _ = doc.AddParagraph()
	run, _ = para.AddRun()
	run.SetText("Professional invoice with merged header and footer:")

	invoice, err := doc.AddTable(6, 4)
	if err != nil {
		log.Fatalf("Failed to add invoice table: %v", err)
	}

	invoice.SetStyle(domain.TableStyleGrid)

	// Header
	iRow0, _ := invoice.Row(0)
	iCell0, _ := iRow0.Cell(0)
	iCell0.SetShading(domain.Color{R: 220, G: 220, B: 220})
	iPara0, _ := iCell0.AddParagraph()
	iRun0, _ := iPara0.AddRun()
	iRun0.SetText("INVOICE #12345")
	iRun0.SetBold(true)
	iCell0.Merge(4, 1)

	// Column headers
	iRow1, _ := invoice.Row(1)
	cells1 := []string{"Item", "Quantity", "Price", "Total"}
	for i, text := range cells1 {
		cell, _ := iRow1.Cell(i)
		cell.SetShading(domain.Color{R: 240, G: 240, B: 240})
		para, _ := cell.AddParagraph()
		run, _ := para.AddRun()
		run.SetText(text)
		run.SetBold(true)
	}

	// Items
	items := [][]string{
		{"Product A", "2", "$10.00", "$20.00"},
		{"Product B", "1", "$25.00", "$25.00"},
		{"Product C", "3", "$5.00", "$15.00"},
	}

	for rowIdx, item := range items {
		iRow, _ := invoice.Row(rowIdx + 2)
		for colIdx, val := range item {
			cell, _ := iRow.Cell(colIdx)
			para, _ := cell.AddParagraph()
			run, _ := para.AddRun()
			run.SetText(val)
		}
	}

	// Total
	iRow5, _ := invoice.Row(5)
	iCell50, _ := iRow5.Cell(0)
	iCell50.Merge(2, 1)
	iCell51, _ := iRow5.Cell(2)
	iCell51.SetShading(domain.Color{R: 240, G: 240, B: 240})
	iPara51, _ := iCell51.AddParagraph()
	iRun51, _ := iPara51.AddRun()
	iRun51.SetText("TOTAL:")
	iRun51.SetBold(true)
	iCell52, _ := iRow5.Cell(3)
	iCell52.SetShading(domain.Color{R: 240, G: 240, B: 240})
	iPara52, _ := iCell52.AddParagraph()
	iRun52, _ := iPara52.AddRun()
	iRun52.SetText("$60.00")
	iRun52.SetBold(true)

	// Save document
	outputPath := "09_advanced_tables_output.docx"
	if err := doc.SaveAs(outputPath); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Printf("✅ Document created successfully: %s\n", outputPath)
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("  - Horizontal cell merging (colspan)")
	fmt.Println("  - Vertical cell merging (rowspan)")
	fmt.Println("  - Combined 2D merging")
	fmt.Println("  - Calendar layout")
	fmt.Println("  - Nested tables")
	fmt.Println("  - Invoice-style layout")
	fmt.Println("  - Table styles")
	fmt.Println("  - Cell shading and formatting")
}
