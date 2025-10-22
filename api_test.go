/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package docx

import (
	"testing"
)

func TestAddLink(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()

	link := para.AddLink("Google", "https://www.google.com")

	if link == nil {
		t.Fatal("AddLink returned nil")
	}
	if link.ID == "" {
		t.Fatal("Link ID should not be empty")
	}
	if link.Run.InstrText != "Google" {
		t.Errorf("Expected InstrText 'Google', got '%s'", link.Run.InstrText)
	}
}

func TestParagraphStyle(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph().Style("Heading1")

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}
	if para.Properties.Style == nil {
		t.Fatal("Style should not be nil")
	}
	if para.Properties.Style.Val != "Heading1" {
		t.Errorf("Expected Style 'Heading1', got '%s'", para.Properties.Style.Val)
	}
}

func TestParagraphAddPageBreaks(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddPageBreaks()

	if run == nil {
		t.Fatal("AddPageBreaks returned nil")
	}
	if len(run.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(run.Children))
	}
	br, ok := run.Children[0].(*BarterRabbet)
	if !ok {
		t.Fatal("Expected child to be BarterRabbet")
	}
	if br.Type != "page" {
		t.Errorf("Expected Type 'page', got '%s'", br.Type)
	}
}

func TestParagraphNumPr(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph().NumPr("1", "0")

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}
	if para.Properties.NumProperties == nil {
		t.Fatal("NumProperties should not be nil")
	}
	if para.Properties.NumProperties.NumID == nil || para.Properties.NumProperties.NumID.Val != "1" {
		t.Error("NumID not set correctly")
	}
	if para.Properties.NumProperties.Ilvl == nil || para.Properties.NumProperties.Ilvl.Val != "0" {
		t.Error("Ilvl not set correctly")
	}
}

func TestParagraphNumFont(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph().NumFont("Arial", "SimSun", "Arial", "eastAsia")

	if para.Properties == nil || para.Properties.RunProperties == nil {
		t.Fatal("Properties should not be nil")
	}
	if para.Properties.RunProperties.Fonts == nil {
		t.Fatal("Fonts should not be nil")
	}
	if para.Properties.RunProperties.Fonts.ASCII != "Arial" {
		t.Errorf("Expected ASCII 'Arial', got '%s'", para.Properties.RunProperties.Fonts.ASCII)
	}
}

func TestParagraphNumSize(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph().NumSize("24")

	if para.Properties == nil || para.Properties.RunProperties == nil {
		t.Fatal("Properties should not be nil")
	}
	if para.Properties.RunProperties.Size == nil {
		t.Fatal("Size should not be nil")
	}
	if para.Properties.RunProperties.Size.Val != "24" {
		t.Errorf("Expected Size '24', got '%s'", para.Properties.RunProperties.Size.Val)
	}
}

func TestRunColor(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("test")
	run.Color("FF0000")

	if run.RunProperties.Color == nil {
		t.Fatal("Color should not be nil")
	}
	if run.RunProperties.Color.Val != "FF0000" {
		t.Errorf("Expected Color 'FF0000', got '%s'", run.RunProperties.Color.Val)
	}
}

func TestRunSize(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("test")
	run.Size("24")

	if run.RunProperties.Size == nil {
		t.Fatal("Size should not be nil")
	}
	if run.RunProperties.Size.Val != "24" {
		t.Errorf("Expected Size '24', got '%s'", run.RunProperties.Size.Val)
	}
}

func TestRunSizeCs(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("test")
	run.SizeCs("24")

	if run.RunProperties.SizeCs == nil {
		t.Fatal("SizeCs should not be nil")
	}
	if run.RunProperties.SizeCs.Val != "24" {
		t.Errorf("Expected SizeCs '24', got '%s'", run.RunProperties.SizeCs.Val)
	}
}

func TestRunSpacing(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("test")
	run.Spacing(100)

	if run.RunProperties.Spacing == nil {
		t.Fatal("Spacing should not be nil")
	}
	if run.RunProperties.Spacing.Line != 100 {
		t.Errorf("Expected Spacing 100, got %d", run.RunProperties.Spacing.Line)
	}
}

func TestRunStrike(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run := para.AddText("test")

	run.Strike(true)
	if run.RunProperties.Strike == nil {
		t.Fatal("Strike should not be nil")
	}
	if run.RunProperties.Strike.Val != "true" {
		t.Errorf("Expected Strike 'true', got '%s'", run.RunProperties.Strike.Val)
	}

	run.Strike(false)
	if run.RunProperties.Strike.Val != "false" {
		t.Errorf("Expected Strike 'false', got '%s'", run.RunProperties.Strike.Val)
	}
}

func TestTableJustification(t *testing.T) {
	doc := New().WithDefaultTheme()
	table := doc.AddTable(2, 2, 5000, nil)
	table.Justification("center")

	if table.TableProperties.Justification == nil {
		t.Fatal("Justification should not be nil")
	}
	if table.TableProperties.Justification.Val != "center" {
		t.Errorf("Expected Justification 'center', got '%s'", table.TableProperties.Justification.Val)
	}
}

func TestTableRowJustification(t *testing.T) {
	doc := New().WithDefaultTheme()
	table := doc.AddTable(2, 2, 5000, nil)
	row := table.TableRows[0].Justification("right")

	if row.TableRowProperties.Justification == nil {
		t.Fatal("Justification should not be nil")
	}
	if row.TableRowProperties.Justification.Val != "right" {
		t.Errorf("Expected Justification 'right', got '%s'", row.TableRowProperties.Justification.Val)
	}
}

func TestRangeRelationships(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddLink("Test", "http://example.com")

	count := 0
	err := doc.RangeRelationships(func(rel *Relationship) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("RangeRelationships failed: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least one relationship")
	}
}

func TestWithA3Page(t *testing.T) {
	doc := New().WithDefaultTheme().WithA3Page()

	if len(doc.Document.Body.Items) == 0 {
		t.Fatal("Expected at least one body item")
	}

	sectpr, ok := doc.Document.Body.Items[len(doc.Document.Body.Items)-1].(*SectPr)
	if !ok {
		t.Fatal("Expected last item to be SectPr")
	}
	if sectpr.PgSz == nil {
		t.Fatal("PgSz should not be nil")
	}
	if sectpr.PgSz.W != 16838 || sectpr.PgSz.H != 23811 {
		t.Errorf("Expected A3 dimensions (16838x23811), got (%dx%d)", sectpr.PgSz.W, sectpr.PgSz.H)
	}
}

func TestWithA4Page(t *testing.T) {
	doc := New().WithDefaultTheme().WithA4Page()

	if len(doc.Document.Body.Items) == 0 {
		t.Fatal("Expected at least one body item")
	}

	sectpr, ok := doc.Document.Body.Items[len(doc.Document.Body.Items)-1].(*SectPr)
	if !ok {
		t.Fatal("Expected last item to be SectPr")
	}
	if sectpr.PgSz == nil {
		t.Fatal("PgSz should not be nil")
	}
	if sectpr.PgSz.W != 11906 || sectpr.PgSz.H != 16838 {
		t.Errorf("Expected A4 dimensions (11906x16838), got (%dx%d)", sectpr.PgSz.W, sectpr.PgSz.H)
	}
}

func TestAddTableTwips(t *testing.T) {
	doc := New().WithDefaultTheme()
	rowHeights := []int64{500, 600, 700}
	colWidths := []int64{1000, 2000, 3000}
	table := doc.AddTableTwips(rowHeights, colWidths, 6000, nil)

	if table == nil {
		t.Fatal("AddTableTwips returned nil")
	}
	if len(table.TableRows) != len(rowHeights) {
		t.Errorf("Expected %d rows, got %d", len(rowHeights), len(table.TableRows))
	}
	if len(table.TableRows[0].TableCells) != len(colWidths) {
		t.Errorf("Expected %d columns, got %d", len(colWidths), len(table.TableRows[0].TableCells))
	}
	if len(table.TableGrid.GridCols) != len(colWidths) {
		t.Errorf("Expected %d grid cols, got %d", len(colWidths), len(table.TableGrid.GridCols))
	}
}

func TestTableString(t *testing.T) {
	doc := New().WithDefaultTheme()
	table := doc.AddTable(2, 2, 5000, nil)

	// Add some text to cells
	table.TableRows[0].TableCells[0].AddParagraph().AddText("Cell 1")
	table.TableRows[0].TableCells[1].AddParagraph().AddText("Cell 2")
	table.TableRows[1].TableCells[0].AddParagraph().AddText("Cell 3")
	table.TableRows[1].TableCells[1].AddParagraph().AddText("Cell 4")

	str := table.String()
	if str == "" {
		t.Error("Table.String() returned empty string")
	}
}

func TestParagraphString(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Hello")
	para.AddText(" ")
	para.AddText("World")

	str := para.String()
	if str != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", str)
	}
}

func TestParagraphStringWithLink(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	para.AddText("Visit ")
	para.AddLink("Google", "https://www.google.com")

	str := para.String()
	if str == "" {
		t.Error("Paragraph.String() returned empty string")
	}
}
