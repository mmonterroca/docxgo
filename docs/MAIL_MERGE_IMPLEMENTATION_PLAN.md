# Mail Merge / Template MVP — Implementation Plan

**Author**: Misael Monterroca
**Date**: February 2026
**Target Version**: v2.3.0
**Estimated Effort**: 29-36 hours (MVP)
**Branch**: `feature/mail-merge`

---

## Objective

Implement a text-based mail merge / template system that allows users to:

1. Open a `.docx` template designed in Word
2. Replace `{{placeholder}}` tokens with data values
3. Save the merged document — preserving all formatting, styles, and layout

**Scope: MVP (text-only)**. Image placeholders, conditional blocks, and table row repetition are deferred to v2.4.0.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────┐
│                  Public API (pkg/template/)           │
│                                                      │
│  MergeTemplate(doc, data) error                      │
│  MergeOptions { Delimiter, StrictMode }              │
│  FindPlaceholders(doc) []Placeholder                 │
│  ValidateTemplate(doc, data) []ValidationError       │
└───────────────────┬──────────────────────────────────┘
                    │ operates on domain interfaces
                    ▼
┌──────────────────────────────────────────────────────┐
│            Domain Mutations (domain/ + core/)         │
│                                                      │
│  Paragraph: ClearRuns(), RemoveRun(i), InsertRunAt(i)│
│  Run: (no new methods needed)                        │
│  Document: (no new methods needed for MVP)           │
└──────────────────────────────────────────────────────┘
```

The template engine operates **entirely through `domain` interfaces** — no type assertions to internal types. It reads paragraphs/runs, consolidates, finds placeholders, replaces, and the existing serializer handles writing.

---

## Implementation Phases

### Phase 1: Paragraph Mutation APIs (3-4 hours)

**Goal**: Add run manipulation methods to `Paragraph` so we can clear, remove, and insert runs.

**Files to modify:**

| File                         | Changes                                                                                             |
| ---------------------------- | --------------------------------------------------------------------------------------------------- |
| `domain/paragraph.go`        | Add `ClearRuns()`, `RemoveRun(index int) error`, `InsertRunAt(index int) (Run, error)` to interface |
| `internal/core/paragraph.go` | Implement the 3 new methods on `paragraph` struct                                                   |
| `internal/core/core_test.go` | Add tests for ClearRuns, RemoveRun, InsertRunAt                                                     |

**Implementation details:**

```go
// domain/paragraph.go — additions to Paragraph interface
ClearRuns()
RemoveRun(index int) error
InsertRunAt(index int) (Run, error)
```

```go
// internal/core/paragraph.go — implementations
func (p *paragraph) ClearRuns() {
    p.runs = p.runs[:0]
}

func (p *paragraph) RemoveRun(index int) error {
    if index < 0 || index >= len(p.runs) {
        return errors.New(errors.ErrValidation, "run index out of range")
    }
    p.runs = append(p.runs[:index], p.runs[index+1:]...)
    return nil
}

func (p *paragraph) InsertRunAt(index int) (domain.Run, error) {
    if index < 0 || index > len(p.runs) {
        return nil, errors.New(errors.ErrValidation, "run index out of range")
    }
    id := p.idGen.NextRunID()
    r := NewRun(id, p.relManager)
    if index == len(p.runs) {
        p.runs = append(p.runs, r)
    } else {
        p.runs = append(p.runs[:index+1], p.runs[index:]...)
        p.runs[index] = r
    }
    return r, nil
}
```

**Tests:**

- `TestParagraph_ClearRuns` — add 3 runs, clear, verify len=0
- `TestParagraph_RemoveRun` — add 3 runs, remove middle, verify order
- `TestParagraph_RemoveRun_OutOfRange` — verify error on invalid index
- `TestParagraph_InsertRunAt` — insert at beginning, middle, end
- `TestParagraph_InsertRunAt_OutOfRange` — verify error on invalid index

**Acceptance criteria:**

- [ ] All 3 methods added to `domain.Paragraph` interface
- [ ] Implementations handle edge cases (empty runs, boundary indices)
- [ ] Tests pass with `go test ./domain/... ./internal/core/...`

---

### Phase 2: Run Consolidation (6-8 hours)

**Goal**: Merge adjacent runs with identical formatting into single runs. This solves the "split placeholder" problem where Word fragments `{{name}}` across multiple `<w:r>` elements.

**New file:** `pkg/template/consolidate.go`

**The problem:**

```xml
<!-- Word splits {{first_name}} into separate runs -->
<w:r><w:rPr><w:b/></w:rPr><w:t>{{</w:t></w:r>
<w:r><w:rPr><w:b/></w:rPr><w:t>first</w:t></w:r>
<w:r><w:rPr><w:b/><w:lang w:val="es-MX"/></w:rPr><w:t>_</w:t></w:r>
<w:r><w:rPr><w:b/></w:rPr><w:t>name}}</w:t></w:r>
```

**Algorithm:**

```
ConsolidateRuns(paragraph):
    1. Get runs list
    2. Walk runs left to right
    3. For each pair of adjacent runs (i, i+1):
       a. Compare formatting (font, color, size, bold, italic, underline, strike, highlight)
       b. Skip runs that have fields, breaks, or images (non-text runs)
       c. If formatting matches AND both are text-only:
          - Append run[i+1].Text() to run[i] via SetText()
          - Mark run[i+1] for removal
    4. Remove marked runs (right to left using RemoveRun())
```

**Formatting equality function:**

```go
func formatsEqual(a, b domain.Run) bool {
    return a.Font() == b.Font() &&
        a.Color() == b.Color() &&
        a.Size() == b.Size() &&
        a.Bold() == b.Bold() &&
        a.Italic() == b.Italic() &&
        a.Underline() == b.Underline() &&
        a.Strike() == b.Strike() &&
        a.Highlight() == b.Highlight()
}
```

**Key decision — proofing attributes:** Word adds `<w:lang>`, `<w:rPrChange>`, and spelling attributes that cause "different" formatting on what the user sees as identical. Our hydrator does NOT hydrate `<w:lang>` though, so this is not a problem in our domain model. In our model, two runs that visually look the same **will** have equal formatting fields.

**Files:**

| File                               | Purpose                                           |
| ---------------------------------- | ------------------------------------------------- |
| `pkg/template/consolidate.go`      | `ConsolidateRuns(para domain.Paragraph)` function |
| `pkg/template/format.go`           | `formatsEqual(a, b domain.Run) bool` helper       |
| `pkg/template/consolidate_test.go` | Tests                                             |

**Tests:**

- `TestConsolidateRuns_IdenticalFormatting` — 3 runs with same formatting → 1 run
- `TestConsolidateRuns_DifferentFormatting` — bold + normal + bold → 3 runs unchanged
- `TestConsolidateRuns_MixedTextAndFields` — runs with fields are not merged
- `TestConsolidateRuns_SingleRun` — no-op
- `TestConsolidateRuns_EmptyParagraph` — no-op
- `TestConsolidateRuns_SplitPlaceholder` — `{{` + `name` + `}}` → `{{name}}`
- `TestConsolidateRuns_PreservesFormatting` — merged run retains bold/color/font from first run

**Acceptance criteria:**

- [ ] `ConsolidateRuns()` correctly merges adjacent text runs with identical formatting
- [ ] Runs with fields, breaks, or images are never merged
- [ ] Text content is preserved exactly
- [ ] Formatting from the first run in a merge group is applied
- [ ] Tests cover edge cases (empty, single, all-different, all-same)

---

### Phase 3: Placeholder Detection (4-5 hours)

**Goal**: Scan a document for `{{placeholder}}` tokens and return structured results.

**New file:** `pkg/template/placeholder.go`

**Data structures:**

```go
// Placeholder represents a found placeholder in the document
type Placeholder struct {
    Key       string  // e.g., "first_name" (without delimiters)
    Location  Location
}

// Location describes where a placeholder was found
type Location struct {
    Type          LocationType  // LocationParagraph, LocationHeader, LocationFooter, LocationTableCell
    ParagraphIdx  int           // index within parent container
    RunIdx        int           // run index within paragraph
    StartOffset   int           // character offset within run text
    EndOffset     int           // character offset (exclusive) within run text
    // For table cells:
    TableIdx      int
    RowIdx        int
    CellIdx       int
    // For headers/footers:
    SectionIdx    int
    HeaderType    domain.HeaderType
    FooterType    domain.FooterType
}

type LocationType int
const (
    LocationParagraph LocationType = iota
    LocationHeader
    LocationFooter
    LocationTableCell
)
```

**Algorithm:**

```
FindPlaceholders(doc, options):
    results = []

    // 1. Scan document paragraphs
    for each paragraph in doc.Paragraphs():
        ConsolidateRuns(paragraph)  // Phase 2
        results += scanParagraph(paragraph, LocationParagraph, ...)

    // 2. Scan table cells
    for each table in doc.Tables():
        for each row in table.Rows():
            for each cell in row.Cells():
                for each paragraph in cell.Paragraphs():
                    ConsolidateRuns(paragraph)
                    results += scanParagraph(paragraph, LocationTableCell, ...)

    // 3. Scan headers/footers
    for each section in doc.Sections():
        for each headerType in [Default, First, Even]:
            header = section.Header(headerType)
            for each paragraph in header.Paragraphs():
                ConsolidateRuns(paragraph)
                results += scanParagraph(paragraph, LocationHeader, ...)
        // same for footers

    return results
```

**Files:**

| File                               | Purpose                                                 |
| ---------------------------------- | ------------------------------------------------------- |
| `pkg/template/placeholder.go`      | `FindPlaceholders(doc, opts)`, types, `scanParagraph()` |
| `pkg/template/placeholder_test.go` | Tests                                                   |

**Tests:**

- `TestFindPlaceholders_Simple` — single `{{name}}` in a paragraph
- `TestFindPlaceholders_Multiple` — `{{first}}` and `{{last}}` in same run
- `TestFindPlaceholders_InTable` — placeholder inside a table cell
- `TestFindPlaceholders_InHeader` — placeholder in section header
- `TestFindPlaceholders_InFooter` — placeholder in section footer
- `TestFindPlaceholders_NoDuplicates` — same key appears multiple times, all returned
- `TestFindPlaceholders_NoPlaceholders` — returns empty slice
- `TestFindPlaceholders_CustomDelimiters` — `${key}` with custom options

**Acceptance criteria:**

- [ ] Finds placeholders in paragraphs, table cells, headers, and footers
- [ ] Runs are consolidated before scanning (Phase 2 dependency)
- [ ] Returns accurate location metadata for each placeholder
- [ ] Custom delimiter support via options

---

### Phase 4: Text Replacement Engine (8-10 hours)

**Goal**: Given a document and a `map[string]string` of key→value pairs, replace all `{{key}}` placeholders with their values while preserving formatting.

**New file:** `pkg/template/merge.go`

**Public API:**

```go
// MergeData is the data to merge into the template
type MergeData map[string]string

// MergeOptions configures the merge behavior
type MergeOptions struct {
    OpenDelimiter  string // default: "{{"
    CloseDelimiter string // default: "}}"
    StrictMode     bool   // if true, error on missing keys; if false, leave unreplaced
}

// MergeTemplate replaces all placeholders in doc with values from data.
// It modifies the document in place.
func MergeTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) error

// ValidateTemplate checks that all placeholders in doc have corresponding keys in data.
// Returns a list of validation errors (missing keys, unused keys).
func ValidateTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) []ValidationError
```

**Replacement algorithm — single-run case (most common after consolidation):**

```
For each paragraph (in body, tables, headers, footers):
    ConsolidateRuns(paragraph)
    for each run in paragraph.Runs():
        text = run.Text()
        if text contains "{{...}}":
            newText = regex.ReplaceAllStringFunc(text, func(match):
                key = extract key from match
                if data has key:
                    return data[key]
                if strictMode:
                    record error
                return match  // leave unreplaced
            )
            run.SetText(newText)
```

**Replacement algorithm — cross-run case (fallback):**

After consolidation, most placeholders will be in a single run. However, if formatting differs mid-placeholder (e.g., `{{na` is bold but `me}}` is not), the placeholder spans multiple runs. This case is handled by:

1. Concatenate text of all runs in the paragraph
2. Find placeholder positions in the concatenated text
3. Map positions back to individual runs
4. For a cross-run placeholder:
   a. Replace text in the first run (everything from `{{` through end of that run → replacement value)
   b. Clear text in middle runs, mark for removal
   c. Replace text in the last run (everything from start through `}}` → empty)
   d. Remove marked runs

**Traversal helper (shared with Phase 3):**

```go
// walkParagraphs calls fn for every paragraph in the document,
// including those in table cells, headers, and footers.
func walkParagraphs(doc domain.Document, fn func(para domain.Paragraph) error) error
```

**Files:**

| File                         | Purpose                                                    |
| ---------------------------- | ---------------------------------------------------------- |
| `pkg/template/merge.go`      | `MergeTemplate()`, `ValidateTemplate()`, replacement logic |
| `pkg/template/walk.go`       | `walkParagraphs()` traversal helper                        |
| `pkg/template/options.go`    | `MergeData`, `MergeOptions`, `ValidationError` types       |
| `pkg/template/merge_test.go` | Tests                                                      |

**Tests:**

- `TestMergeTemplate_SinglePlaceholder` — `{{name}}` → "John"
- `TestMergeTemplate_MultiplePlaceholders` — `{{first}} {{last}}` → "John Doe"
- `TestMergeTemplate_RepeatedPlaceholder` — `{{name}}` appears 3 times
- `TestMergeTemplate_PreservesFormatting` — bold `{{name}}` becomes bold "John"
- `TestMergeTemplate_InTableCell` — placeholder inside table → replaced
- `TestMergeTemplate_InHeader` — placeholder in header → replaced
- `TestMergeTemplate_InFooter` — placeholder in footer → replaced
- `TestMergeTemplate_MissingKey_Lenient` — unreplaced placeholder left as-is
- `TestMergeTemplate_MissingKey_Strict` — returns error
- `TestMergeTemplate_EmptyValue` — `{{name}}` → "" (empty string)
- `TestMergeTemplate_SpecialChars` — value with `<`, `>`, `&` (XML-safe)
- `TestMergeTemplate_CustomDelimiters` — `${key}` → "value"
- `TestMergeTemplate_NoPlaceholders` — no-op, no error
- `TestValidateTemplate_AllKeysPresent` — no errors
- `TestValidateTemplate_MissingKeys` — reports missing keys
- `TestValidateTemplate_UnusedKeys` — reports unused keys (warning level)

**Acceptance criteria:**

- [ ] All placeholders in body, tables, headers, footers are replaced
- [ ] Formatting (bold, italic, color, font, size) is preserved on replaced text
- [ ] Cross-run placeholders are handled correctly
- [ ] StrictMode correctly reports missing keys
- [ ] ValidateTemplate returns actionable errors
- [ ] No panics on edge cases (empty doc, no runs, nil data)

---

### Phase 5: Public API & Example (4-5 hours)

**Goal**: Create the clean public API surface and a working example.

**Public package:** `pkg/template/`

**Public API surface (final):**

```go
package template

// MergeTemplate replaces all {{key}} placeholders in doc with values from data.
// The document is modified in place. Use doc.SaveAs() to write the result.
//
// Example:
//
//   doc, _ := docx.OpenDocument("invoice_template.docx")
//   err := template.MergeTemplate(doc, template.MergeData{
//       "customer_name": "Acme Corp",
//       "invoice_date":  "2026-02-27",
//       "total":         "$1,234.56",
//   })
//   doc.SaveAs("invoice_acme.docx")
func MergeTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) error

// FindPlaceholders returns all {{key}} placeholders found in the document.
func FindPlaceholders(doc domain.Document, opts ...MergeOptions) []Placeholder

// ValidateTemplate checks that all placeholders have matching data keys.
func ValidateTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) []ValidationError

// ConsolidateRuns merges adjacent runs with identical formatting in a paragraph.
// This is also called automatically by MergeTemplate and FindPlaceholders.
func ConsolidateRuns(para domain.Paragraph)
```

**Example program:** `examples/14_mail_merge/main.go`

The example will:

1. **Include** a pre-designed `template.docx` in the repo with placeholders (`{{name}}`, `{{date}}`, `{{total}}`, etc.) and professional formatting (designed in Word)
2. **Open** it via `docx.OpenDocument("template.docx")`
3. **Merge** with sample data (customer info, invoice items)
4. **Save** as `merged_output.docx`
5. Also demonstrate batch merge (3 customers from a slice)

This demonstrates the real-world workflow: designers create templates in Word, developers inject data via code.

**Files:**

| File                             | Purpose               |
| -------------------------------- | --------------------- |
| `pkg/template/doc.go`            | Package documentation |
| `examples/14_mail_merge/main.go` | Working example       |

**Acceptance criteria:**

- [ ] Example runs and produces a valid `.docx` file
- [ ] `go doc ./pkg/template/` shows clean API documentation
- [ ] Package is importable as `github.com/mmonterroca/docxgo/v2/pkg/template`

---

### Phase 6: Integration Tests & Documentation (4-5 hours)

**Goal**: End-to-end integration tests and documentation updates.

**Integration tests:** `pkg/template/integration_test.go`

- `TestIntegration_CreateTemplateAndMerge` — full round-trip: create doc → add placeholders → save → open → merge → save → verify
- `TestIntegration_ComplexTemplate` — multiple sections, tables, headers with placeholders
- `TestIntegration_BatchMerge` — merge same template with 10 different data sets
- `TestIntegration_PreservesDocumentStructure` — verify styles, formatting, layout survive merge

**Documentation updates:**

| File                            | Update                                                     |
| ------------------------------- | ---------------------------------------------------------- |
| `README.md`                     | Add "Template / Mail Merge" to features, add usage example |
| `docs/IMPLEMENTATION_STATUS.md` | Move Mail Merge from "Planned" to "Implemented"            |
| `docs/V2_API_GUIDE.md`          | Add template/merge section                                 |
| `examples/README.md`            | Add example 14                                             |
| `CHANGELOG.md`                  | Add v2.3.0 entry                                           |

**Acceptance criteria:**

- [ ] All integration tests pass
- [ ] `go test ./...` passes (all packages)
- [ ] Documentation accurately reflects the new feature
- [ ] No regressions in existing tests

---

## File Summary

### New Files (12)

| File                               | Lines (est.) | Phase |
| ---------------------------------- | ------------ | ----- |
| `pkg/template/doc.go`              | ~20          | 5     |
| `pkg/template/consolidate.go`      | ~80          | 2     |
| `pkg/template/consolidate_test.go` | ~150         | 2     |
| `pkg/template/format.go`           | ~30          | 2     |
| `pkg/template/placeholder.go`      | ~120         | 3     |
| `pkg/template/placeholder_test.go` | ~200         | 3     |
| `pkg/template/merge.go`            | ~180         | 4     |
| `pkg/template/merge_test.go`       | ~300         | 4     |
| `pkg/template/options.go`          | ~50          | 4     |
| `pkg/template/walk.go`             | ~60          | 4     |
| `pkg/template/integration_test.go` | ~200         | 6     |
| `examples/14_mail_merge/main.go`   | ~120         | 5     |

**Total new code**: ~1,500 lines (est.)

### Modified Files (5)

| File                            | Changes                   | Phase |
| ------------------------------- | ------------------------- | ----- |
| `domain/paragraph.go`           | +3 methods to interface   | 1     |
| `internal/core/paragraph.go`    | +3 method implementations | 1     |
| `internal/core/core_test.go`    | +5 test functions         | 1     |
| `README.md`                     | Features + example        | 6     |
| `docs/IMPLEMENTATION_STATUS.md` | Status update             | 6     |

---

## Execution Order & Dependencies

```
Phase 1 ──────► Phase 2 ──────► Phase 3 ──────► Phase 4 ──────► Phase 5 ──────► Phase 6
Mutations       Consolidation   Detection       Replacement     API + Example   Tests + Docs
(3-4h)          (6-8h)          (4-5h)          (8-10h)         (4-5h)          (4-5h)

No deps         Needs Phase 1   Needs Phase 2   Needs 1+2+3    Needs 1-4       Needs 1-5
```

Each phase produces a working, testable increment. Phases cannot be parallelized (each builds on the previous).

---

## Risk Assessment

| Risk                                       | Impact | Mitigation                                                                                                                                           |
| ------------------------------------------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Split-run problem harder than expected     | HIGH   | Phase 2 is allocated the most time. The hydrator doesn't preserve `<w:lang>`, which eliminates the most common cause of "phantom" formatting splits. |
| Cross-run placeholders after consolidation | MEDIUM | Phase 4 includes a cross-run fallback algorithm. In practice, ~95% of placeholders will be single-run after consolidation.                           |
| Round-trip corruption                      | MEDIUM | We don't create new relationships (text-only MVP), so issue #8 is not a blocker. Existing round-trip tests will catch regressions.                   |
| Performance on large documents             | LOW    | The algorithm is O(paragraphs × runs). Documents rarely exceed 10K paragraphs. No optimization needed for MVP.                                       |
| Custom delimiters conflict with content    | LOW    | Default `{{ }}` delimiters are very unlikely to appear in natural text. Custom delimiter option available.                                           |

---

## Out of Scope (Deferred to v2.4.0)

| Feature                                    | Reason                                       | Dependency               |
| ------------------------------------------ | -------------------------------------------- | ------------------------ |
| Image placeholders (`{{image:logo}}`)      | Requires issue #8 fix (relationship merging) | Issue #8                 |
| Conditional sections (`{{#if}}...{{/if}}`) | Requires `RemoveParagraph()` on Document     | Document mutation APIs   |
| Table row repetition (`{{#each items}}`)   | Requires deep row cloning                    | Row clone helper         |
| Loop blocks (`{{#loop}}...{{/endloop}}`)   | Complex, needs block-level operations        | Conditional + repetition |
| Nested placeholders                        | Edge case, low demand                        | v2.5.0+                  |

---

## Success Metrics

- [ ] `go test ./...` — all tests pass (existing + new)
- [ ] Example 14 generates valid `.docx` files that open in Word
- [ ] Placeholders in body, tables, headers, and footers are all replaced
- [ ] Formatting is preserved on all replaced text
- [ ] Zero external dependencies maintained
- [ ] API is clean: 4 public functions, 3 public types
