# docxgo

A Go library for creating and editing Microsoft Word (`.docx` / OOXML) documents, with a JSON-RPC CLI and Node.js wrapper for use from other languages.

**🌐 Official Website:** [docxgo.dev](https://docxgo.dev/)

[![Go Reference](https://pkg.go.dev/badge/github.com/mmonterroca/docxgo/v2.svg)](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/mmonterroca/docxgo)](https://goreportcard.com/report/github.com/mmonterroca/docxgo)

## Overview

**docxgo** builds and edits Word documents in Go through a fluent, type-safe API. It covers document creation, a template and mail-merge engine, themes, images, tables, and full read-modify-save round-tripping, with structured error handling on every operation.

Since v2.7.0 it also ships a JSON-RPC command-line interface and a Node.js wrapper (`@mmonterroca/docxgo`), so the same engine can be driven from Node, Python, a shell script, or AWS Lambda without a running server.

- **Fluent builder** — chainable API for documents, paragraphs, runs, tables, sections
- **Read & modify** existing `.docx` files with round-trip style preservation
- **Templates / mail merge** — placeholder detection, data merge, batch generation, external Word templates (MERGEFIELD)
- **Themes** — 7 preset themes (Corporate, Startup, Modern, Fintech, Academic, TechPresentation, TechDarkMode) for coordinated colors, fonts, and spacing
- **Rich content** — tables (merging, nesting, 8 styles), in-memory images (PNG, JPEG, GIF), fields (TOC, page numbers, hyperlinks), headers/footers, 40+ built-in styles
- **Proofing language** — `WithLanguage` / `WithLanguageEx` for spell-check, grammar, hyphenation
- **Production quality** — clean architecture, explicit errors, thread-safe internals

**Status:** v2.13.0 · Production-ready · Requires Go 1.23+ (no external C dependencies; runs on Linux, macOS, Windows). See the [CHANGELOG](CHANGELOG.md) for version history.

---

## Installation

**Go library:**

```bash
go get github.com/mmonterroca/docxgo/v2
```

**Node.js / CLI:**

```bash
npm install @mmonterroca/docxgo
```

---

## Quick Start (Go)

### Builder API (fluent, recommended)

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
    builder := docx.NewDocumentBuilder(
        docx.WithTitle("My Report"),
        docx.WithAuthor("John Doe"),
        docx.WithDefaultFont("Calibri"),
        docx.WithPageSize(docx.A4),
        docx.WithLanguage("en-US"), // spell-check, grammar, hyphenation
    )

    builder.AddParagraph().
        Text("Project Report").Bold().FontSize(16).
        Color(docx.Blue).Alignment(domain.AlignmentCenter).
        End()

    builder.AddParagraph().
        Text("This is bold text").Bold().
        Text(" and this is ").
        Text("colored text").Color(docx.Red).
        End()

    doc, err := builder.Build()
    if err != nil {
        log.Fatal(err)
    }
    if err := doc.SaveAs("report.docx"); err != nil {
        log.Fatal(err)
    }
}
```

The lower-level domain API is also available directly (`doc := docx.NewDocument()`, `doc.AddParagraph()`, …) when you want fine-grained control.

### Read & modify an existing document

```go
doc, err := docx.OpenDocument("template.docx")
if err != nil {
    log.Fatal(err)
}

for _, para := range doc.Paragraphs() {
    for _, run := range para.Runs() {
        if run.Text() == "PLACEHOLDER" {
            run.SetText("Updated Value")
            run.SetBold(true)
        }
    }
}

if err := doc.SaveAs("modified.docx"); err != nil {
    log.Fatal(err)
}
```

---

## Node.js & CLI

docxgo ships a JSON-RPC binary (`cmd/docxgo`) that speaks over stdin/stdout, plus a fully-typed Node.js wrapper built on top of it. Documents can be created and manipulated from any language — with file **or** base64/buffer I/O, so binaries never have to touch the filesystem.

```ts
import { DocumentBuilder } from '@mmonterroca/docxgo';

const doc = new DocumentBuilder();
await doc
  .setOptions({ title: 'Report', author: 'docxgo' })
  .addHeading('Quarterly Report', 1)
  .addParagraph('Generated from Node.js.')
  .createToFile('report.docx');
doc.dispose();
```

The wrapper offers three API levels — `DocxgoExec` (one-shot), `DocxgoRPC` (persistent), and `DocumentBuilder` (fluent). The binary itself has two modes: `docxgo exec` (single request) and `docxgo rpc` (persistent NDJSON session), ideal for batch work and AWS Lambda.

- **[docs/CLI_GUIDE.md](docs/CLI_GUIDE.md)** — full JSON-RPC protocol and method reference
- **[npm/README.md](npm/README.md)** — Node.js API reference and examples

---

## Examples

The [`examples/`](examples/) directory has 15 runnable Go programs:

| | | |
|---|---|---|
| [01_basic](examples/01_basic/) — builder basics | [06_sections](examples/06_sections/) — page layout | [11_multi_section](examples/11_multi_section/) — multi-section |
| [02_intermediate](examples/02_intermediate/) — product catalog | [07_advanced](examples/07_advanced/) — everything combined | [12_read_and_modify](examples/12_read_and_modify/) — edit existing |
| [03_toc](examples/03_toc/) — table of contents | [08_images](examples/08_images/) — images | [13_themes](examples/13_themes/) — theme system |
| [04_fields](examples/04_fields/) — TOC, page numbers, links | [09_advanced_tables](examples/09_advanced_tables/) — merging, nesting | [14_mail_merge](examples/14_mail_merge/) — template engine |
| [05_styles](examples/05_styles/) — style system | [10_paragraph_spacing](examples/10_paragraph_spacing/) — spacing | [15_external_template](examples/15_external_template/) — MERGEFIELD |

Node.js examples live in [`npm/examples/`](npm/examples/).

---

## Architecture

Clean architecture with clear layer boundaries — `domain/` (public interfaces), `internal/` (implementations), `pkg/` (utilities), `themes/`, plus the `cmd/docxgo/` CLI binary and `npm/` Node.js wrapper. Interface segregation, dependency injection, explicit errors, a strongly typed public API, and thread-safe internal managers (a single `Document` is not thread-safe — guard concurrent access; see the package godoc's Thread Safety section). See **[docs/V2_DESIGN.md](docs/V2_DESIGN.md)** for the full rationale.

---

## Error Handling

Every operation returns explicit, structured errors carrying operation context and an error code — no silent failures. The fluent builder accumulates errors and surfaces the first at `Build()`. See **[docs/ERROR_HANDLING.md](docs/ERROR_HANDLING.md)** for the error system and patterns.

---

## Testing

```bash
go test ./...              # run all tests
go test -cover ./...       # with coverage
go test -bench=. ./...     # benchmarks
```

Coverage varies by package — `domain` and `pkg/errors` are at 100%, `internal/xml` ~96%, and the core packages sit around 50–70%.

---

## Documentation

- **[V2 API Guide](docs/V2_API_GUIDE.md)** — complete API reference with examples (start here)
- **[CLI Guide](docs/CLI_GUIDE.md)** — JSON-RPC protocol for the CLI / Node.js wrapper
- **[Node.js wrapper](npm/README.md)** — `@mmonterroca/docxgo` API reference
- **[Implementation Status](docs/IMPLEMENTATION_STATUS.md)** — what's implemented and planned
- **[Migration Guide](MIGRATION.md)** — upgrading from v1
- **[V2 Design](docs/V2_DESIGN.md)** · **[Error Handling](docs/ERROR_HANDLING.md)** · **[Contributing](CONTRIBUTING.md)**
- **[API Reference (pkg.go.dev)](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2)**

---

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow (feature branches → `master`), testing requirements, and PR process. In short: fork, branch from `master`, add tests, run `go test ./...`, and open a PR against `master`.

---

## License & Credits

docxgo v2 is distributed under the MIT license ([LICENSE](LICENSE)). Its Git history descends from [`fumiama/go-docx`](https://github.com/fumiama/go-docx) and includes AGPL-era upstream snapshots; the completed [provenance audit](docs/PROVENANCE_AUDIT.md) determines that the current rewritten release contains no protectable AGPL implementation. Historical commits retain their original licenses. See [CREDITS.md](CREDITS.md) for the full genealogy.

## Support

- **Support:** [GitHub Issues](https://github.com/mmonterroca/docxgo/issues) for bugs, questions, and feature requests
- **Email:** misael@monterroca.com
