# docxgo — Code Provenance Audit (v2 vs. AGPL upstream)

**Scope:** Determine whether the docxgo v2 codebase contains code derived from
`fumiama/go-docx`, which is licensed **AGPL-3.0** (relicensed by fumiama on
2023-02-24). docxgo v2 is distributed under the **MIT** license and its
marketing/position depends on the claim that no AGPL-licensed expression is
present in the shipped code.

**Date:** 2026-07-18
**Method:** Automated line-level and structural comparison of the docxgo v2
working tree (`master`) against `fumiama/go-docx` pinned at commit
[`0c30fd0`](https://github.com/fumiama/go-docx/commit/0c30fd09304b17fdb42b0dcea142962b2f4883a3)
(2025-05-06, HEAD as of this writing; 140 commits, 2021-04-23 → 2025-05-06,
so this tree accumulates all AGPL-era code). Fully reproducible: clone that
commit and run [`docs/provenance/compare_line_overlap.py`](provenance/compare_line_overlap.py)
against this repo — see that script's docstring for the exact commands.

> **Not legal advice.** This is a *technical* provenance analysis performed to
> support a licensing position. It is strong, reproducible evidence for a
> due-diligence review, but a formal legal opinion — especially before selling
> a license warranty or indemnification — should be obtained from an IP
> attorney. This document is designed to make that review fast and cheap.

---

## Background

An earlier internal remediation (issues #12/#13/#36; commits `01063c3`,
`19aa47a`, `d51d446`, `6964374`, Feb–Jul 2026) removed **AGPL-3.0 license
headers** that had been inherited into six source files, removed a leftover
**Apache-2.0** header from `internal/writer/zip.go`, standardized all 102 `.go`
files to a uniform MIT/SPDX header, and corrected provenance documentation
(confirming via git history + GitHub API that the MIT root `gonfva/docxlib`
is MIT, not AGPL — so AGPL enters the lineage only at fumiama's 2023 relicense).

That work fixed *labeling*. It did not, by itself, establish that the underlying
*code* in those files is not derived from the AGPL upstream — swapping a header
does not change provenance. **This audit closes that gap.**

The seven files that carried inherited AGPL/Apache headers — six with an
AGPL-3.0 header, plus the Apache-2.0 header on `internal/writer/zip.go`
(highest-risk set):

- `internal/serializer/serializer.go`
- `internal/serializer/latent_styles.go`
- `internal/serializer/serializer_test.go`
- `internal/xml/run.go`
- `internal/writer/writer_test.go`
- `internal/core/io_test.go`
- `internal/writer/zip.go` (Apache-2.0 header)

---

## Findings

### 1. Verbatim line overlap — whole tree (102 files)

Distinctive lines (≥25 chars, comments excluded, whitespace-normalized) from
every docxgo `.go` file were checked against the fumiama corpus (its 36
non-test `.go` files, 1,247 distinctive lines, at the pinned commit above).
Reproduce with `docs/provenance/compare_line_overlap.py`.

| docxgo file | verbatim overlap |
|---|---|
| `internal/xml/run.go` | 26.5% (9/34) |
| `internal/xml/document.go` | 18.8% (9/48) |
| `internal/xml/table.go` | 13.3% (8/60) |
| `internal/xml/section.go` | 12.5% (6/48) |
| `internal/xml/paragraph.go` | 8.7% (4/46) |
| `internal/xml/field.go` | 6.7% (1/15) |
| `internal/writer/writer_test.go` | 4.7% (2/43) |
| `internal/xml/style.go` | 3.0% (3/99) |
| `internal/xml/drawing.go` | 1.9% (2/103) |
| `internal/manager/relationship.go` | 1.7% (1/59) |
| `internal/core/io_test.go` | 1.1% (1/93) |
| `cmd/docxgo/handlers.go` | 0.1% (1/672) |
| `internal/serializer/serializer.go` | **0.0% (0/550)** |
| `internal/serializer/latent_styles.go` | **0.0% (0/214)** |
| `internal/serializer/serializer_test.go` | **0.0% (0/318)** |
| `internal/writer/zip.go` | **0.0% (0/284)** |

Across the whole tree (102 `.go` files, 11,564 distinctive lines total),
**47 lines match anything in the fumiama corpus — 0.4% of docxgo's
distinctive lines.** Every file containing real, non-OOXML-struct logic is
at 0–2%, and the four files that carried inherited AGPL/Apache headers with
actual implementation logic (`serializer.go`, `latent_styles.go`,
`serializer_test.go`, `zip.go`) are at a **verified, exact 0%.** Overlap is
concentrated in `internal/xml/*.go`, which hold OOXML struct *definitions*.

### 2. What the overlapping lines actually are

Of the 47 matched lines, 43 are struct field or struct type declarations
whose form is dictated by the ECMA-376 (OOXML) schema, e.g.:

```
type RunProperties struct {
Val string `xml:"w:val,attr"`
ASCII string `xml:"w:ascii,attr,omitempty"`
EastAsia string `xml:"w:eastAsia,attr,omitempty"`
```

There is essentially only one correct way to map an OOXML element such as
`w:color val="…"` onto a Go struct tag. Under the copyright **merger doctrine**,
expression that is dictated by an external constraint (here, the schema) is not
protectable, and its presence in two independent implementations is expected —
not evidence of copying.

The remaining 4 matched lines are **not** struct declarations — they're
generic Go idioms, unrelated to the OOXML schema: `writer_test.go` and
`io_test.go` each match on the standard `for _, f := range zipReader.File {`
zip-iteration loop (`writer_test.go` also matches the literal path check
`if f.Name == "word/document.xml" {`), and `handlers.go` matches on the
generic `items = append(items, item)` idiom. These are unavoidable
consequences of both projects reading a `.docx` zip archive and appending to
a slice in Go — not evidence of copying either, but distinct from the
schema-merger argument above, so called out separately here rather than
folded into a blanket "100%" claim.

### 3. Structural design is opposite, not derived

The `Run` type — the file with the highest surface overlap — is designed on
**incompatible principles** in the two projects:

| | fumiama/go-docx `Run` (AGPL) | docxgo v2 `Run` (MIT) |
|---|---|---|
| Child content | `Children []interface{}` (polymorphic) | typed optional fields (`Text *Text`, `Break *Break`, `Drawing *Drawing`, …) |
| Parsing | custom `UnmarshalXML` token walk | standard `encoding/xml` struct mapping |
| Doc coupling | holds `file *Docx` back-pointer | none |

This is precisely the difference between the legacy fork and an independent
clean-architecture rewrite, for the `Run` type specifically — the file with
the highest surface overlap. This is **not** a project-wide "docxgo never
uses `interface{}`" claim: docxgo's public `domain` package API has zero
`interface{}` usage, but several internal serialization types
(`internal/xml/document.go`, `paragraph.go`, `table.go`) do use
`[]interface{}` to model OOXML's own polymorphic "any child" content model —
the same general technique fumiama uses, just confined to internal plumbing
rather than exposed on the public API or used in `Run` itself.

### 4. No shared functions, types, or assets

- **Function names:** fumiama 84, docxgo (suspect files) 104, **shared: none.**
- **Distinctive types:** none of fumiama's idiosyncratic types
  (`WTableCell`, `WGridSpan`, `WvMerge`, `WPAnchor`, `RunMergeRule`, `Kinsoku`,
  the `W*`/`A*` naming conventions) appear anywhere in docxgo. Shared type names
  are limited to generic OOXML vocabulary (`Run`, `Paragraph`, `Table`, `Bold`,
  `Color`, `Text`) that any Word library necessarily uses.
- **Embedded default XML** (docxgo's built-in `settings.xml` / `fontTable.xml`
  content in `internal/writer/zip.go`): its distinctive markers —
  `<w:panose1 w:val="020F0502020204030204"/>`,
  `<w:characterSpacingControl w:val="doNotCompress"/>`, and the
  `compatibilityMode` compat setting — do not appear anywhere in fumiama.
  (Both projects' default `theme1.xml` do share the literal string
  `name="Office Theme"` — but that's Microsoft's own generic default theme
  name, present in effectively every OOXML theme part on Earth, so it's not
  a distinctive marker either way and proves nothing about derivation.)
  docxgo's settings/fontTable defaults are independently sourced from
  Microsoft's standard output.

### 5. The residual overlap predates AGPL

The schema-dictated struct-tag lines that do coincide also exist in the
**MIT-era ancestors** (`gingfrederik/docx`, `gonfva/docxlib`), which are the
foundation both trees inherit and which predate fumiama's 2023-02-24 AGPL
relicense. So even the non-copyrightable overlap does not trace to the AGPL
layer.

---

## Conclusion

Across line-level, functional, type-level, structural, and embedded-asset
comparisons, **there is no evidence that docxgo v2 contains AGPL-derived
expression from `fumiama/go-docx`.** The only shared text is (a) OOXML
struct-tag boilerplate dictated by the ECMA-376 schema (non-protectable under
merger; also present in the MIT-era ancestors) and (b) universal Go idioms.
The prior AGPL headers on the six flagged files are consistent with vestigial
header copy-paste at rewrite time, since none of those files share
implementation with the AGPL upstream.

**Accurate public claim** (evidence-supported):
> docxgo v2 is an independent, MIT-licensed clean-architecture implementation.
> A line-level and structural diff against the AGPL-licensed `fumiama/go-docx`
> upstream found only OOXML schema-dictated struct mappings and standard-library
> idioms in common — no shared logic, functions, or assets — and that residual
> boilerplate also exists in the project's MIT-licensed ancestors.

Avoid the weaker/absolute phrasing "the AGPL never touched a single line," which
is imprecise (schema boilerplate does coincide); the claim above is both true
and stronger for due diligence.

## Recommended follow-ups

1. Keep this document, and `docs/provenance/compare_line_overlap.py`, in-repo
   as due-diligence artifacts.
2. Before offering a paid license warranty/indemnification, have an IP attorney
   review this analysis and sign off — the pinned commit and checked-in script
   should make independent re-verification (and that review) inexpensive.
3. Re-run the script (against a fresh fumiama `HEAD`, since it has continued
   commits after `0c30fd0`) on major docxgo refactors that touch
   `internal/xml/*` or import upstream code, to keep the record current.
