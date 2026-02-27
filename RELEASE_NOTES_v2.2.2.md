# Release Notes - v2.2.2

**Release Date:** February 2026

## Summary

This release fixes a critical issue where tables using built-in styles (e.g., `TableGrid`, `PlainTable1`) rendered without visible borders in Microsoft Word. It also includes a comprehensive documentation overhaul to bring all docs up to date, and restores the themes example.

## Bug Fixes

### Table Style Borders Not Rendering ([#15](https://github.com/mmonterroca/docxgo/issues/15))

Tables with applied styles like `TableGrid` appeared without any borders in Word because the style definitions in `styles.xml` were missing `w:tblPr` / `w:tblBorders` elements.

**Root Cause**: Table styles were registered in the style manager but had no border data attached. The serializer had no logic to emit `w:tblPr` for table-type styles.

**Fix**:
- Added `TableStyleDef` interface to `domain/style.go` with `TableBorders()`, `SetTableBorders()`, `HasTableBorders()`, and `CellMargins()` methods
- Added `TableLevelBorders` struct to `domain/table.go` with `InsideH` and `InsideV` fields for inner grid borders
- Updated `tableStyle` in `internal/manager/table_style.go` to implement `TableStyleDef` with border and cell margin storage
- Built-in styles `TableGrid` and `PlainTable1` now ship with all-around + inside single borders (0.5pt)
- Added `serializeTableStyleProperties()` to the document serializer, emitting proper `w:tblPr > w:tblBorders` XML
- Added `TableStyleProperties`, `TableCellMargins`, `TableCellMargin`, and `TableLevelBorders` XML structs

### Grid Column Width Calculation

- `serializeGrid()` now derives column widths from the first row's cell widths instead of emitting `w:w="0"` for every column
- Handles cell span (`gridSpan`) by distributing width evenly across spanned columns
- Falls back to omitting widths (Word auto-calculates) when no explicit widths are set

## Improvements

### Examples Updated
- **01_basic** and **02_intermediate**: Added `.Style(domain.TableStyleGrid)` to table creation so tables render with visible borders out of the box
- **13_themes**: Restored with updated v2 API — generates 7 themed documents (Corporate, Startup, Modern, Fintech, Academic, TechPresentation, TechDarkMode)

### Documentation Overhaul ([#14](https://github.com/mmonterroca/docxgo/issues/14))

Comprehensive update across all documentation files to reflect the current state of the project:

- **README.md**: Updated version info, example count (13), architecture tree, features section, roadmap with actual release history, copyright year
- **CHANGELOG.md**: Added all missing version entries (v2.0.0-beta through v2.2.1 — previously only had v2.2.0)
- **docs/IMPLEMENTATION_STATUS.md**: Complete rewrite — accurate release history table, removed outdated roadmap/recommendations, fixed known limitations (document reading works since v2.0.0), added mail merge as top planned feature
- **docs/README.md**: Updated example count to 13, added missing examples to index
- **examples/README.md**: Added `13_themes` section, removed reference to nonexistent `basic/` directory
- **docs/V2_API_GUIDE.md**: Updated version from v2.0.0-beta to v2.2.1
- **docs/V2_DESIGN.md**: Updated status from "v2.0.0-beta Ready" to "v2.2.1 Stable"

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.2.2
```

## Usage Example

```go
package main

import (
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
    builder := docx.NewDocumentBuilder(
        docx.WithTitle("Invoice"),
    )

    builder.AddParagraph().Text("Product List").Style(domain.StyleHeading1).End()

    // Tables now render with visible borders when using TableStyleGrid
    builder.AddTable(2, 3).
        Style(domain.TableStyleGrid).
        Row(0).Cell(0).Text("Product").Bold().End().
        Row(0).Cell(1).Text("Qty").Bold().End().
        Row(0).Cell(2).Text("Price").Bold().End().
        Row(1).Cell(0).Text("Widget").End().
        Row(1).Cell(1).Text("10").End().
        Row(1).Cell(2).Text("$9.99").End()

    doc, _ := builder.Build()
    doc.SaveAs("invoice.docx")
}
```

## Files Changed

### Code Changes
- `domain/style.go` — Added `TableStyleDef` interface
- `domain/table.go` — Added `TableLevelBorders` struct, clarified `TableBorders` doc comment
- `internal/manager/style.go` — Built-in table styles now carry border definitions
- `internal/manager/table_style.go` — Implements `TableStyleDef` (borders + cell margins)
- `internal/serializer/serializer.go` — New `serializeTableStyleProperties()`, `serializeStyleBorder()`, `borderLineStyleToString()` functions; improved `serializeGrid()` with width derivation
- `internal/xml/style.go` — Added `TableStyleProperties`, `TableCellMargins`, `TableCellMargin` structs
- `internal/xml/table.go` — Added `TableLevelBorders` struct, `Space` attr on `Border`, `Borders` field on `TableProperties`

### Example Changes
- `examples/01_basic/main.go` — Added `Style(domain.TableStyleGrid)` to table
- `examples/02_intermediate/main.go` — Added `Style(domain.TableStyleGrid)` to tables
- `examples/13_themes/main.go` — Restored with full v2 API (355 lines)

### Documentation Changes
- `CHANGELOG.md` — Complete version history (v2.0.0-beta through v2.2.2)
- `README.md` — Version, features, roadmap, examples updated
- `docs/IMPLEMENTATION_STATUS.md` — Full rewrite (722 → 376 lines)
- `docs/README.md` — Example count and index updated
- `docs/V2_API_GUIDE.md` — Version header updated
- `docs/V2_DESIGN.md` — Status header updated
- `examples/README.md` — Added 13_themes, removed stale references

## Acknowledgements

- Issue reported by [@g-mero](https://github.com/g-mero) in [#15](https://github.com/mmonterroca/docxgo/issues/15)
- Roadmap feedback from [@jmls](https://github.com/jmls) in [#14](https://github.com/mmonterroca/docxgo/issues/14)
