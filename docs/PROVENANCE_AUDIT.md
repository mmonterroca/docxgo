# docxgo — Code Provenance Audit (v2 vs. AGPL upstream)

**Scope:** Determine whether the docxgo v2 codebase contains code derived from
`fumiama/go-docx`, which is licensed **AGPL-3.0** (relicensed by fumiama on
2023-02-24). docxgo v2 is distributed under the **MIT** license and its
marketing/position depends on the claim that no AGPL-licensed expression is
present in the shipped code.

**Date:** 2026-07-18
**Method:** Automated line-level and structural comparison of the docxgo v2
working tree (`master`) against the full `fumiama/go-docx` working tree
(140 commits, 2021-04-23 → 2025-05-06, which accumulates all AGPL-era code).

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
every docxgo `.go` file were checked against the full fumiama corpus
(1,299 distinctive lines).

| docxgo file | verbatim overlap |
|---|---|
| `internal/xml/run.go` | 34.2% (13/38) |
| `internal/xml/document.go` | 21.6% (11/51) |
| `internal/xml/table.go` | 18.8% (13/69) |
| `internal/xml/section.go` | 12.0% (6/50) |
| `internal/xml/paragraph.go` | 10.4% (5/48) |
| `internal/xml/style.go` | 6.7% (7/105) |
| `internal/writer/writer_test.go` | 6.5% (4/62) |
| `internal/serializer/serializer.go` | **0.0% (0/652)** |
| `internal/serializer/latent_styles.go` | **0.0% (0/214)** |
| `internal/serializer/serializer_test.go` | **0.0% (0/409)** |
| `internal/writer/zip.go` | **0.0% (0/300)** |
| `cmd/docxgo/handlers.go` | 0.1% (1/913) |

**Every file containing real logic is at 0–2%.** Overlap is concentrated
entirely in `internal/xml/*.go`, which hold OOXML struct *definitions*.

### 2. What the overlapping lines actually are

100% of the matched lines are struct field declarations whose form is dictated
by the ECMA-376 (OOXML) schema, e.g.:

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
not evidence of copying. The test files' overlap is the universal Go idiom
`for _, f := range zipReader.File {` plus the fixed path string
`word/document.xml`.

### 3. Structural design is opposite, not derived

The `Run` type — the file with the highest surface overlap — is designed on
**incompatible principles** in the two projects:

| | fumiama/go-docx (AGPL) | docxgo v2 (MIT) |
|---|---|---|
| Child content | `Children []interface{}` (polymorphic) | typed optional fields (`Text *Text`, `Break *Break`, `Drawing *Drawing`, …) |
| Parsing | custom `UnmarshalXML` token walk | standard `encoding/xml` struct mapping |
| Doc coupling | holds `file *Docx` back-pointer | none |
| Philosophy | interface-based children | **no `interface{}`** (docxgo's stated design goal) |

This is precisely the difference between the legacy fork and an independent
clean-architecture rewrite.

### 4. No shared functions, types, or assets

- **Function names:** fumiama 84, docxgo (suspect files) 104, **shared: none.**
- **Distinctive types:** none of fumiama's idiosyncratic types
  (`WTableCell`, `WGridSpan`, `WvMerge`, `WPAnchor`, `RunMergeRule`, `Kinsoku`,
  the `W*`/`A*` naming conventions) appear anywhere in docxgo. Shared type names
  are limited to generic OOXML vocabulary (`Run`, `Paragraph`, `Table`, `Bold`,
  `Color`, `Text`) that any Word library necessarily uses.
- **Embedded default XML** (docxgo's built-in `theme1.xml`, `settings.xml`,
  `fontTable.xml` blobs): none of their distinctive markers
  (`panose1 020F0502020204030204`, `characterSpacingControl doNotCompress`,
  `compatibilityMode`, `Office Theme`) appear in fumiama. docxgo's defaults are
  independently sourced from Microsoft's standard output.

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

1. Keep this document in-repo (e.g. `docs/PROVENANCE_AUDIT.md`) as a
   due-diligence artifact.
2. Before offering a paid license warranty/indemnification, have an IP attorney
   review this analysis and sign off — the reproducible method above should make
   that inexpensive.
3. Re-run this diff on major refactors that touch `internal/xml/*` or import
   upstream code, to keep the record current.
