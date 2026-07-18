# docxgo — Code Provenance Audit (v2 vs. AGPL upstream)

**Scope:** Record the provenance *facts* needed to assess docxgo v2's license
position relative to `fumiama/go-docx`, which is licensed **AGPL-3.0**
(fumiama relicensed it from MIT to AGPL-3.0 on 2023-02-24). docxgo v2 is
currently distributed under the **MIT** license.

> **This document does not, and cannot, resolve docxgo's license position.**
> Whether docxgo v2 may be distributed under MIT is a *derivative-work*
> determination under copyright law — a legal question for an IP attorney,
> not a conclusion this technical analysis reaches. The purpose here is to
> give that review an accurate, reproducible factual record. See **Status &
> open question** at the end. **Not legal advice.**

**Date:** 2026-07-18

---

## Status & open question (read this first)

Two facts, both verified, point in different directions and must be weighed
together by counsel:

1. **docxgo v2's shipped repository is a git descendant of the AGPL-era
   `fumiama/go-docx`.** It is not an independent, from-scratch project: its
   `master` contains fumiama's entire commit history through 2025-05-06
   (including the 2023-02-24 AGPL relicense), with docxgo's own development
   layered on top starting 2025-10-21. See Finding 1. On its face this makes
   docxgo a **fork of** — and a strong candidate for a **derivative work
   of** — AGPL-licensed code.

2. **The current docxgo source tree retains almost no verbatim overlap with
   fumiama** (0.4% of distinctive lines, all OOXML schema-dictated struct
   tags or generic Go idioms; the files with real logic are at an exact 0%).
   See Findings 2–6. This reflects a substantial rewrite.

Neither fact settles the matter on its own. Derivative-work status is **not**
determined by how much verbatim text remains today — a work adapted from an
original can remain a derivative work even after heavy rewriting. Fact 1 is
the stronger signal and cannot be waved away by Fact 2. **The prior version
of this document, and the README, claimed docxgo is an "independent" rewrite
that "AGPL only ever entered fumiama's fork, never this codebase." That claim
is contradicted by Fact 1 and has been removed.**

**Required next step:** IP counsel review before docxgo makes or relies on any
MIT-licensing representation (including the current `LICENSE` file and the
`@mmonterroca/docxgo` npm package's `MIT` license field), and before any paid
license warranty or indemnification. This document is designed to make that
review fast and cheap, not to substitute for it.

---

## Method

- **Line-level comparison** (Findings 2–6): distinctive lines (≥25 chars,
  comments excluded, whitespace-normalized) from every docxgo `.go` file,
  checked for verbatim presence in the `fumiama/go-docx` corpus pinned at
  commit [`0c30fd0`](https://github.com/fumiama/go-docx/commit/0c30fd09304b17fdb42b0dcea142962b2f4883a3)
  (2025-05-06, its HEAD as of this writing). Fully reproducible: clone that
  commit and run [`docs/provenance/compare_line_overlap.py`](provenance/compare_line_overlap.py)
  against this repo — see that script's docstring for the exact commands.
- **Git-history comparison** (Finding 1): commit-ancestry checks between this
  repo and the same pinned upstream, reproducible with the commands shown
  inline.

---

## Findings

### 1. Git-history provenance — docxgo descends from the AGPL-era upstream

This is the dominant fact. docxgo v2's shipped repository shares git history
with `fumiama/go-docx` and descends from it *through the AGPL period*, rather
than being an independent project or a fork taken during the earlier MIT
window.

Reproducible (run in a clone of this repo, with `0c30fd0` = fumiama HEAD):

```
# All 140 fumiama commits (through 2025-05-06) are ancestors of docxgo master:
git merge-base --is-ancestor 0c30fd09304b17fdb42b0dcea142962b2f4883a3 master   # exit 0 = yes
git rev-list --count 0c30fd09304b17fdb42b0dcea142962b2f4883a3                  # 140
git rev-list --count master                                                    # 385

# docxgo's first *own* commit sits on top of that AGPL-era history:
git rev-list --reverse master --not 0c30fd09304b17fdb42b0dcea142962b2f4883a3 | head -1
#  -> f36dd59  2025-10-21  "feat: Implement Phase 1-2 - Bookmarks, Fields, and TOC Builder"
```

Timeline (all dates from the git history in this repo):

| Date | Event | Upstream license |
|---|---|---|
| 2020–2023-02 | gingfrederik/docx → gonfva/docxlib → fumiama/go-docx (early) | **MIT** |
| **2023-02-24** | fumiama relicenses (`70ec491d` "change LICENSE to AGPLv3") | **→ AGPL-3.0** |
| 2023-02 → 2025-05 | fumiama continues development | AGPL-3.0 |
| **2025-10-21** | docxgo's first own commit, on top of fumiama's 2025-05 AGPL HEAD | (distributed as MIT) |

The upstream *was* MIT in 2021–early-2023, but docxgo did **not** take the
code during that window — its own work begins in October 2025, on a base that
had been AGPL-licensed for ~2.5 years. So the "the overlap predates AGPL"
argument (Finding 6) applies only to the small residue of schema boilerplate,
**not** to the codebase's provenance as a whole.

### 2. Verbatim line overlap of the current tree (102 files)

Distinctive lines from every docxgo `.go` file were checked against the
fumiama corpus (its 36 non-test `.go` files, 1,247 distinctive lines, at the
pinned commit above). Reproduce with `docs/provenance/compare_line_overlap.py`.

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
| `internal/serializer/serializer.go` | 0.0% (0/550) |
| `internal/serializer/latent_styles.go` | 0.0% (0/214) |
| `internal/serializer/serializer_test.go` | 0.0% (0/318) |
| `internal/writer/zip.go` | 0.0% (0/284) |

Across the whole tree (102 `.go` files, 11,564 distinctive lines total),
47 lines match anything in the fumiama corpus — **0.4% of docxgo's
distinctive lines.** Overlap is concentrated in `internal/xml/*.go`, which
hold OOXML struct *definitions*; the files carrying real implementation logic
(`serializer.go`, `latent_styles.go`, `zip.go`) are at an exact 0%.

*This measures present-day textual similarity. It is relevant to, but does not
by itself decide, the derivative-work question — see Status above.*

### 3. What the overlapping lines actually are

Of the 47 matched lines, 43 are struct field or struct type declarations whose
form is dictated by the ECMA-376 (OOXML) schema, e.g.:

```
type RunProperties struct {
Val string `xml:"w:val,attr"`
ASCII string `xml:"w:ascii,attr,omitempty"`
EastAsia string `xml:"w:eastAsia,attr,omitempty"`
```

There is essentially only one correct way to map an OOXML element such as
`w:color val="…"` onto a Go struct tag. Under the copyright **merger
doctrine**, expression dictated by an external constraint (here, the schema)
is generally not protectable, and its coincidence across implementations is
expected — but whether merger applies to any given line is itself a legal
judgment, noted here for counsel rather than asserted as settled.

The remaining 4 matched lines are **not** struct declarations — they're
generic Go idioms unrelated to the OOXML schema: `writer_test.go` and
`io_test.go` each match the standard `for _, f := range zipReader.File {`
zip-iteration loop (`writer_test.go` also matches the literal path check
`if f.Name == "word/document.xml" {`), and `handlers.go` matches the generic
`items = append(items, item)` idiom.

### 4. The `Run` type is structurally different

The `Run` type — the file with the highest surface overlap — is built on
different principles in the two projects:

| | fumiama/go-docx `Run` (AGPL) | docxgo v2 `Run` (MIT-distributed) |
|---|---|---|
| Child content | `Children []interface{}` (polymorphic) | typed optional fields (`Text *Text`, `Break *Break`, `Drawing *Drawing`, …) |
| Parsing | custom `UnmarshalXML` token walk | standard `encoding/xml` struct mapping |
| Doc coupling | holds `file *Docx` back-pointer | none |

Note this is scoped to the `Run` type specifically, **not** a project-wide
"docxgo never uses `interface{}`" claim: docxgo's public `domain` API has no
`interface{}`, but several internal serialization types
(`internal/xml/document.go`, `paragraph.go`, `table.go`) do use
`[]interface{}` to model OOXML's polymorphic "any child" content — the same
general technique fumiama uses.

### 5. Shared functions, types, and assets

- **Function names:** none of fumiama's function names appear in docxgo's
  seven flagged files (verified: 0 shared, whether or not fumiama's own test
  files are counted).
- **Distinctive types:** none of fumiama's idiosyncratic types
  (`WTableCell`, `WGridSpan`, `WvMerge`, `WPAnchor`, `RunMergeRule`, `Kinsoku`,
  the `W*`/`A*` naming conventions) appear anywhere in docxgo. Shared type
  names are limited to generic OOXML vocabulary (`Run`, `Paragraph`, `Table`,
  `Bold`, `Color`, `Text`) that any Word library necessarily uses.
- **Embedded default XML** (docxgo's built-in `settings.xml` / `fontTable.xml`
  content in `internal/writer/zip.go`): its distinctive markers —
  `<w:panose1 w:val="020F0502020204030204"/>`,
  `<w:characterSpacingControl w:val="doNotCompress"/>`, and the
  `compatibilityMode` compat setting — do not appear in fumiama. (Both
  projects' default `theme1.xml` share the literal `name="Office Theme"`, but
  that's Microsoft's own generic default theme name and is not a distinctive
  marker either way.)

### 6. The residual schema overlap predates AGPL

The schema-dictated struct-tag lines that do coincide (Finding 3) also exist
in the **MIT-era ancestors** (`gingfrederik/docx`, `gonfva/docxlib`), which
predate fumiama's 2023-02-24 AGPL relicense. So that *specific residue* traces
to MIT-era code. This does **not** extend to the codebase as a whole, whose
git provenance runs through the AGPL period (Finding 1).

---

## What can and cannot be claimed

**Supportable as fact (verified above):**
- docxgo's current tree shares only 0.4% of its distinctive lines with the
  AGPL upstream, and that residue is schema boilerplate / generic Go idioms.
- No fumiama function names, distinctive types, or embedded assets carry over.
- docxgo has been substantially rewritten relative to the upstream.

**NOT supportable, and removed from README/CREDITS:**
- "Independent" / "clean-room" / "from-scratch" framing.
- "AGPL only ever entered fumiama's fork, never this codebase" — the codebase
  descends from AGPL-era fumiama (Finding 1).

**Open — for IP counsel, not resolved here:**
- Whether docxgo v2 is a derivative work of the AGPL `fumiama/go-docx`.
- Whether docxgo may be distributed under MIT, or whether the AGPL obligations
  attach (in whole or part) to the shipped module and npm package.
- What, if anything, must change in `LICENSE`, `package.json`, attribution, or
  the public claims as a result.

## Recommended follow-ups

1. **Obtain IP counsel review** on the derivative-work question before making
   or relying on any MIT representation. Provide them this document, the pinned
   upstream commit, and `docs/provenance/compare_line_overlap.py`.
2. Do **not** change `LICENSE` or `package.json`'s `license` field on the
   basis of this technical analysis alone — that decision is counsel's.
3. Keep this document and the comparison script in-repo as the factual record.
4. Re-run the comparison on major refactors touching `internal/xml/*` to keep
   the textual-overlap figures current.
