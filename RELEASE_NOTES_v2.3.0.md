# Release Notes - v2.3.0

**Release Date:** February 2026

## Summary

v2.3.0 introduces the **Template / Mail Merge** engine — a new `pkg/template/` package for document template processing. This release adds placeholder detection, data merging, batch document generation, custom delimiter support, and full MERGEFIELD compatibility with real Microsoft Word templates.

## New Features

### Template / Mail Merge Engine (`pkg/template/`)

A complete template processing package with zero external dependencies:

- **`MergeTemplate()`** — Replace `{{placeholder}}` tokens with data values while preserving all formatting
- **`FindPlaceholders()` / `PlaceholderNames()`** — Detect all placeholders across body, tables, headers, and footers
- **`ValidateTemplate()`** — Check for missing or unused data keys before merging
- **`ConsolidateRuns()`** — Merge adjacent runs with identical formatting to heal split placeholders
- **Custom Delimiters** — Support for `${key}`, `«key»`, or any custom open/close delimiter pair
- **Strict Mode** — Optional error reporting for missing keys during merge
- **Batch Merge** — Generate multiple documents from a single template with different data sets

### MERGEFIELD Support

Full compatibility with Microsoft Word's native mail merge fields:

- Reads Word MERGEFIELD instructions from `.docx` templates
- Correctly handles Word's two-run MERGEFIELD pattern (field instruction + display text)
- Supports `«guillemet»` delimiters used by Word's mail merge UI
- Tested with real production Word templates

### Paragraph Mutation APIs

- `Paragraph.ClearRuns()` — Remove all runs from a paragraph
- `Paragraph.RemoveRun(index)` — Remove a specific run by index
- `Paragraph.InsertRunAt(index)` — Insert a new run at a specific position

### Run Content Inspection

- `Run.Fields()` — Access field instructions in a run
- `Run.Breaks()` — Access break elements in a run
- `Run.Image()` — Access embedded image in a run
- `Run.ClearFields()` — Clear field instructions from a run

### New Examples

- **Example 14**: Mail merge invoice template with batch generation (programmatic)
- **Example 15**: External Word template with real MERGEFIELDs and `«»` delimiters

## Bug Fixes

- **Phantom Header/Footer References** — Fixed `walkParagraphs()` calling `section.Header()` which auto-creates headers/footers on sections that don't have them, causing "Word found unreadable content" errors. Now uses read-only `HeadersAll()`/`FootersAll()` methods.
- **MERGEFIELD Text Duplication** — Fixed issue where MERGEFIELD display text was duplicated after merge because the field instruction in the preceding run continued to emit the original placeholder. Now clears fields from the preceding run when replacing MERGEFIELD text.

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.3.0
```

## Usage Examples

### Basic Template Merge

```go
import (
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/pkg/template"
)

doc := docx.NewDocument()
para, _ := doc.AddParagraph()
para.AddRun().SetText("Hello {{name}}, welcome to {{company}}!")

data := map[string]string{
    "name":    "John",
    "company": "Acme Corp",
}

result, _ := template.MergeTemplate(doc, data, nil)
result.SaveAs("output.docx")
```

### External Word Template with MERGEFIELDs

```go
doc, _ := docx.OpenDocument("template.docx")

opts := &template.MergeOptions{
    OpenDelimiter:  "«",
    CloseDelimiter: "»",
}

data := map[string]string{
    "Contact_FullName": "Jane Doe",
    "Account_Name":     "Acme Corp",
}

result, _ := template.MergeTemplate(doc, data, opts)
result.SaveAs("merged_output.docx")
```

### Batch Merge

```go
doc, _ := docx.OpenDocument("invoice_template.docx")

records := []map[string]string{
    {"name": "John Smith", "amount": "$1,500.00"},
    {"name": "Alice Johnson", "amount": "$2,300.00"},
}

results, _ := template.BatchMerge("invoice_template.docx", records, nil)
for i, result := range results {
    result.SaveAs(fmt.Sprintf("invoice_%d.docx", i+1))
}
```

## Files Added

- `pkg/template/engine.go` — Main template engine
- `pkg/template/merge.go` — Core merge logic
- `pkg/template/merge_test.go` — Merge tests
- `pkg/template/placeholder.go` — Placeholder detection
- `pkg/template/placeholder_test.go` — Placeholder tests
- `pkg/template/consolidate.go` — Run consolidation
- `pkg/template/consolidate_test.go` — Consolidation tests
- `pkg/template/validate.go` — Template validation
- `pkg/template/validate_test.go` — Validation tests
- `pkg/template/walk.go` — Document traversal
- `pkg/template/batch.go` — Batch merge support
- `examples/14_mail_merge/main.go` — Programmatic template example
- `examples/15_external_template/main.go` — External Word template example

## Files Modified

- `domain/paragraph.go` — Added `ClearRuns()`, `RemoveRun()`, `InsertRunAt()`
- `domain/run.go` — Added `Fields()`, `Breaks()`, `Image()`, `ClearFields()`
- `internal/core/paragraph.go` — Implemented paragraph mutation methods
- `internal/core/run.go` — Implemented run content inspection and `ClearFields()`

## Test Coverage

- 41 new tests across the `pkg/template/` package
- All existing tests continue to pass
- Tested with real Microsoft Word templates containing MERGEFIELDs

## Acknowledgements

This release represents the biggest feature addition since the v2.0.0 rewrite, adding a complete template engine while maintaining the library's zero-dependency philosophy.
