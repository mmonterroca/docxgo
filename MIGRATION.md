# Migration Guide: v1 → docxgo v2

**Current target:** v2.7.2
**Minimum Go version:** 1.23

This guide covers migration from the historical `fumiama/go-docx` API to the
current public module, `github.com/mmonterroca/docxgo/v2`. v2 is a separate
architecture with intentionally breaking API changes, so migration is a source
update rather than a module-path-only replacement.

For the complete current API, see the [API guide](docs/V2_API_GUIDE.md) and
[pkg.go.dev](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2).

## What changes

- Every operation that can fail reports an error.
- Document creation is available through a fluent builder or the direct domain
  interfaces.
- Font sizes use points in the builder and half-points in the direct `Run` API.
- Colors, alignments, underlines, page sizes, and other options use typed values.
- The public module carries the required semantic-import suffix `/v2`.
- `internal/*` packages are implementation details and must not be imported by
  consumers.
- A single `Document` is not safe for concurrent mutation; synchronize access
  externally if multiple goroutines share one instance.

## 1. Update the module

```bash
go get github.com/mmonterroca/docxgo/v2@v2.7.2
go mod tidy
```

Replace the old import:

```go
import "github.com/fumiama/go-docx"
```

with the current root package:

```go
import docx "github.com/mmonterroca/docxgo/v2"
```

Advanced code may also import public subpackages such as:

```go
import (
    "github.com/mmonterroca/docxgo/v2/domain"
    "github.com/mmonterroca/docxgo/v2/pkg/template"
    "github.com/mmonterroca/docxgo/v2/themes"
)
```

Do not import `github.com/mmonterroca/docxgo/v2/internal/core`; Go intentionally
prevents external modules from importing it. Use `docx.NewDocument()` instead.

## 2. Choose an API style

### Fluent builder

The builder is the shortest path for new documents. It accumulates operation
errors and returns the first one from `Build`.

```go
package main

import (
    "log"

    docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
    builder := docx.NewDocumentBuilder(
        docx.WithTitle("Migration example"),
        docx.WithAuthor("Example author"),
        docx.WithLanguage("en-US"),
    )

    builder.AddParagraph().
        Text("Hello from docxgo v2").
        Bold().
        Color(docx.Red).
        FontSize(14).
        Alignment(docx.AlignmentCenter).
        End()

    doc, err := builder.Build()
    if err != nil {
        log.Fatal(err)
    }
    if err := doc.SaveAs("output.docx"); err != nil {
        log.Fatal(err)
    }
}
```

### Direct domain API

Use the direct interfaces when you need fine-grained control or want to inspect
the objects you create.

```go
package main

import (
    "log"

    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
    document := docx.NewDocument()

    paragraph, err := document.AddParagraph()
    if err != nil {
        log.Fatal(err)
    }
    if err := paragraph.SetAlignment(domain.AlignmentCenter); err != nil {
        log.Fatal(err)
    }

    run, err := paragraph.AddRun()
    if err != nil {
        log.Fatal(err)
    }
    if err := run.SetText("Hello from the direct API"); err != nil {
        log.Fatal(err)
    }
    if err := run.SetBold(true); err != nil {
        log.Fatal(err)
    }
    if err := run.SetSize(28); err != nil { // 14 pt = 28 half-points
        log.Fatal(err)
    }
    if err := run.SetColor(domain.Color{R: 255, G: 0, B: 0}); err != nil {
        log.Fatal(err)
    }

    if err := document.SaveAs("output.docx"); err != nil {
        log.Fatal(err)
    }
}
```

## 3. Translate common operations

| v1 pattern | v2 builder | v2 direct API |
|---|---|---|
| `docx.New()` | `docx.NewDocumentBuilder()` | `docx.NewDocument()` |
| `para.AddText("Text")` | `.Text("Text")` | `para.AddRun()` then `run.SetText("Text")` |
| `run.Bold()` | `.Bold()` | `run.SetBold(true)` |
| `run.Size("28")` | `.FontSize(28)` | `run.SetSize(56)` |
| `run.Color("FF0000")` | `.Color(docx.Red)` | `run.SetColor(domain.Color{R: 255})` |
| paragraph alignment string | `.Alignment(docx.AlignmentCenter)` | `para.SetAlignment(domain.AlignmentCenter)` |
| chained write with ignored failures | `Build()` then `SaveAs()` | check each returned error |

The builder's `FontSize` accepts points. The direct `Run.SetSize` method accepts
half-points, matching OOXML conventions.

## 4. Migrate document structure

### Paragraphs and runs

```go
paragraph, err := document.AddParagraph()
if err != nil {
    return err
}

run, err := paragraph.AddRun()
if err != nil {
    return err
}
if err := run.SetText("Text"); err != nil {
    return err
}
```

### Tables

```go
table, err := document.AddTable(3, 2)
if err != nil {
    return err
}

row, err := table.Row(0)
if err != nil {
    return err
}
cell, err := row.Cell(0)
if err != nil {
    return err
}
paragraph, err := cell.AddParagraph()
if err != nil {
    return err
}
run, err := paragraph.AddRun()
if err != nil {
    return err
}
return run.SetText("Cell content")
```

### Sections

```go
import "github.com/mmonterroca/docxgo/v2/domain"

section, err := document.DefaultSection()
if err != nil {
    return err
}
if err := section.SetPageSize(domain.PageSizeA4); err != nil {
    return err
}
if err := section.SetOrientation(domain.OrientationLandscape); err != nil {
    return err
}
```

The builder exposes the same settings through `DefaultSection()` and
`AddSection()`.

### Fields

```go
run, err := paragraph.AddRun()
if err != nil {
    return err
}
if err := run.AddField(docx.NewPageNumberField()); err != nil {
    return err
}
```

Factories are available for page numbers, total pages, tables of contents,
hyperlinks, style references, and custom fields.

### Themes

```go
import "github.com/mmonterroca/docxgo/v2/themes"

builder := docx.NewDocumentBuilder(
    docx.WithTheme(themes.Corporate),
)
```

## 5. Read and modify an existing document

```go
document, err := docx.OpenDocument("input.docx")
if err != nil {
    return err
}

paragraph, err := document.AddParagraph()
if err != nil {
    return err
}
run, err := paragraph.AddRun()
if err != nil {
    return err
}
if err := run.SetText("Appended by docxgo v2"); err != nil {
    return err
}
return document.SaveAs("output.docx")
```

Unknown OOXML parts are preserved where supported. Review the
[implementation status](docs/IMPLEMENTATION_STATUS.md) for the current
round-trip coverage and test your own templates before production rollout.

## 6. Migration checklist

- [ ] Require Go 1.23 or later.
- [ ] Change every docxgo module import to include `/v2`.
- [ ] Remove consumer imports of `internal/*` packages.
- [ ] Replace silent method chains with checked errors or the builder.
- [ ] Convert direct font sizes from points to half-points.
- [ ] Replace string colors, alignments, and styles with typed values.
- [ ] Add external synchronization if a `Document` is shared across goroutines.
- [ ] Run `go test ./...` in the consuming project.
- [ ] Generate representative documents and open them in each target editor.

## License note

docxgo v2 is distributed under MIT. The repository history contains historical
AGPL-era upstream snapshots; the definitive scope and current-release
determination are recorded in the
[provenance audit](docs/PROVENANCE_AUDIT.md). Check the license of any separate
v1 version you continue to distribute.

## Support

- [Issues](https://github.com/mmonterroca/docxgo/issues) for reproducible bugs
- [Issues](https://github.com/mmonterroca/docxgo/issues) for migration questions
- [Examples](examples/) for working end-to-end code
- [Troubleshooting guide](docs/TROUBLESHOOTING_DOCX_VALIDATION.md) for OOXML validation

**Last updated:** July 2026
