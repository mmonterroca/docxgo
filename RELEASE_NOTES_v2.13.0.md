# Release Notes - v2.13.0

**Release Date:** August 13, 2026

## Summary

The largest batch since v2.10.0: header/footer tables as a first-class part
of the domain model, a full round-trip overhaul for header/footer edits
(previously *any* in-memory edit to a header or footer on an opened document
was silently discarded on save), a complete pass at table property
hydration (borders, widths, alignment, row heights, cell shading), and a
full fix for issue #101 — hyperlinks and bookmarks now round-trip correctly,
including a review-driven follow-up pass that caught three more edge cases
after the first fix landed. Table-level borders and run-level "All Caps"
are new public API. Theme-linked table cell shading now survives a
round-trip instead of freezing into a flat color or, in one case, being
dropped outright.

Thanks to [@djmoch](https://github.com/djmoch) for [#102](https://github.com/mmonterroca/docxgo/issues/102) — a single detailed repro (one manually-authored document, one small edit) that surfaced three independent formatting losses at once: the all-caps title, the section break, and the table borders. All three are fixed here.

It's a minor release: several new interface methods (see Compatibility
below), no removed public API. If you're building documents with headers,
footers, tables, hyperlinks, or bookmarks, this release fixes real data
loss you may already be hitting silently.

## Added

### Header/footer tables

`domain.Header`/`domain.Footer` can now hold tables, not just paragraphs — a
common Word pattern (logo + address + page-number letterheads) that was
previously completely unmodeled.

- `AddTable(rows, cols int) (Table, error)` — same bounds validation as
  `Document.AddTable`.
- `Tables() []Table` — top-level tables only, same contract `Paragraphs()`
  already has.
- `Blocks() []Block` — the authoritative ordering view across paragraphs
  and tables, reusing `domain.Block`.

Opening a document now hydrates header/footer tables instead of silently
dropping them (best-effort: a table docxgo can't represent is skipped, not
fatal). `ReplaceText`/`MergeTemplate`/`FindPlaceholders` now reach
placeholders inside header/footer table cells, and the CLI/RPC
`headers`/`footers` content arrays accept `{"type": "table", ...}` items.

**Known limitation:** the remaining table properties (`w:tblInd`,
`w:tblCellMar`, `w:tblLayout`, `w:tblLook`, table-level `w:shd`,
`w:trPr/w:cantSplit`, `w:trPr/w:tblHeader`) are still not hydrated — the
same gap body tables have. A nested table inside a table cell is still
dropped on read, in both the body and a header/footer.

### `domain.Table.SetBorders` / `Borders`

Sets the borders drawn around and inside a table as a whole (`w:tblBorders`),
as opposed to `TableCell.SetBorders`, which only ever reaches one cell's
four sides. Reuses the existing `TableLevelBorders` shape (four outer sides
plus `InsideH`/`InsideV`) that table *styles* already used. Opening an
existing `.docx` now hydrates this from `w:tblPr/w:tblBorders` — without it,
a table drawing its own borders instead of referencing a style still came
back looking borderless even after the table-style fix below.

### `domain.Run.SetCaps` / `Caps`

Sets whether a run displays in all capitals (`w:caps`) — a display-only
override, the same distinction as Word's own "All Caps" character
formatting versus actually typing in capitals: it does not change
`run.Text()` or the run's stored `<w:t>` content, only how it renders.
Same shape as `SetBold`/`Bold`. Opening an existing `.docx` now hydrates
each run's `Caps()`. Added because #102 reported it as a real loss: a title
formatted all-caps in Word round-tripped back to its literal mixed-case
stored text, since docxgo had no representation for the property at all.

## Fixed

### Headers and footers now produce a valid package, and edits actually reach the saved file

Two defects meant a header or footer holding an image was silently
corrupt, and *any* in-memory edit to a header or footer on an opened
document — `SetText`, `AddParagraph`, `AddTable`, a whole new header, all
of it — was silently dropped on save:

- **`<w:hdr>`/`<w:ftr>` now declare `xmlns:wp`.** Every `<w:drawing>`
  wrapper element carries that prefix; a header holding an image was an
  undeclared-prefix error Word refused to open outright.
- **Each header and footer now owns its relationships**, written to
  `word/_rels/headerN.xml.rels`/`footerN.xml.rels` with their own ID
  sequence. A header cannot resolve an `r:id` declared in
  `word/_rels/document.xml.rels`, which is where every header relationship
  used to be minted — so a hyperlink in a header was rejected outright
  before this release, and an image there produced a broken package.
- **Header/footer relationship parts are now preserved through their own
  dedicated path.** Previously they survived a round-trip only
  incidentally, as part of an opaque unrelated-parts bucket.
- **An edit to a header or footer on an opened document now reaches the
  saved file.** `WriteTo` was all-or-nothing: a single preserved header
  discarded the *entire* generated map. It's now a per-name merge — a part
  the caller edited is written from the model, and every untouched part
  still goes back to the file byte-for-byte. "Edited" is decided by
  comparing each part's serialized form against a snapshot taken at
  hydration, so a deep edit five levels down inside a header table cell is
  caught for free.

  Six further defects were only reachable once headers actually started
  regenerating, and are fixed alongside: a newly added header could be
  handed an already-taken part name; `[Content_Types].xml` got no
  `Override` for it; a header target like `/word/header1.xml` produced a
  malformed entry name (`word//word/header1.xml`); a regenerated part's
  relationships could pair new content with stale `r:id`s pointing at
  nothing; a hydrated hyperlink in a header registered no relationship at
  all in that header's own `.rels`; an untouched header could be
  regenerated (and silently lose whatever the reader couldn't model)
  purely because the document *body* held an image; and a header's own
  `.rels` file stored in a subdirectory could be mistaken for a phantom
  header, leaving a duplicate, possibly-stale entry in the saved package.

Verified end to end with `DocxValidator` (Open XML SDK schema validation)
and by opening in Microsoft Word.

**Known limitation:** a regenerated header/footer is rebuilt from what
docxgo can model, so a content control, unrecognized field, or nested table
inside it is dropped from that part even though the edit itself lands.
`Document.HasPreservedHeadersOrFooters` reports the round-tripped case if
you want to decide differently. When two sections reference the same
header part, only the first section's copy is written — editing the object
belonging to a later section changes nothing (not a regression: before
this release no header edit reached the file at all).

### Hyperlinks and bookmarks (#101), including a review-driven second pass

`Paragraph.AddHyperlink` previously minted a relationship but never
attached it to a field, so the serializer had no signal to wrap it in a
real `<w:hyperlink>` element — Word rendered it as plain blue underlined
text, not an actual link, and nothing caught it. `AddHyperlink("#anchor",
...)` separately minted a broken *external* relationship whose target was
the literal `"#anchor"` string instead of producing a working internal
link. Both are fixed, and several more defects surfaced while fixing them:

- A hyperlink's display text no longer appears twice in the saved
  document (a pre-existing bug, unrelated to the two above, invisible to
  every existing test since none of them checked rendered text).
- A hyperlink built from several differently-formatted runs no longer
  serializes as one `<w:hyperlink>` element per run — something Word
  itself never produces for a single link. Adjacent hyperlink elements
  sharing one target now fold back into one on save.
- `w:history` is now read from and written to the source instead of being
  hardcoded per code path (an explicit `"0"` is a real, distinct value
  from the attribute being absent).
- **Bookmarks are hydrated for the first time.** Previously every `REF`
  field's target, internal hyperlink anchor, and Word's own
  `_Ref…`/`_GoBack` bookmark vanished on the next save. A bookmark whose
  `w:bookmarkStart`/`w:bookmarkEnd` fall in the same paragraph *and* wrap
  that paragraph's entire content now survives a round trip.
  `generateHeadingBookmarks` no longer overwrites a bookmark that came
  from the source, and its own numbering no longer collides with a
  hydrated bookmark's numeric ID.
- Adding a hyperlink (or any relationship) to a document opened with
  `OpenDocument` no longer produces a package Word offers to repair — the
  rels part is now regenerated from the relationship manager instead of
  written back stale whenever something new was added since the document
  was opened.

A code review of the first pass at this fix (before it was released) found
three more edge cases, all fixed in the same release rather than shipped
broken and patched later:

- A bookmark whose start/end fall in the same paragraph but wrap only
  *part* of it (e.g. `"target"` inside `"prefix target suffix"`) was
  hydrated and re-emitted at the paragraph's own boundaries — silently
  widening it to cover text it never bookmarked, which would have
  corrupted any `REF` field pointed at it. It's now dropped instead,
  matching how a bookmark spanning multiple paragraphs was already handled.
- An internal (`#anchor`) hyperlink hydrated from a source that omitted
  `w:history` entirely came back with an invented `w:history="1"`. The
  default now lives at construction (only reachable from the public
  `AddHyperlink`), so it can never leak onto a hydrated link.
- A table cell shaded purely via a theme reference, with no cached
  fallback color in the source, lost its `<w:shd>` entirely on round-trip
  — see theme-linked cell shading, below.

**Known limitation:** a bookmark spanning multiple paragraphs, or hanging
directly off `w:body` (where Word's own `_GoBack` usually sits), still has
nowhere to live in the current single-paragraph model and is dropped —
unchanged from before this release.

### Table property hydration

A table's own layout — widths, alignment, borders, row heights, cell
shading — is no longer dropped when a `.docx` is opened and resaved (#102).
Previously only a table's style reference and merge attributes
(`w:gridSpan`, `w:vMerge`) were hydrated, so a hand-laid-out table came
back as an unstyled auto-width grid on the next save. A table's named
style (`<w:tblStyle>`) is also no longer dropped, and `w:tblPr`/`w:tcPr` no
longer emit their children out of schema order — a defect reachable from
the plain public API with no round-trip involved, which the Open XML SDK
rejects outright even though Word is lenient about it. A table whose first
row is a merged full-width cell no longer gets invented, evenly-split
column widths.

**Known limitation:** table property hydration is lossy in eight specific,
deliberate ways — each a place where the domain model has no way to
express what the source said, and inventing one would be worse than
dropping it. The most visible: a real pattern fill (anything other than
`clear`/`solid`) still has no single color to map onto `SetShading`, a
`w:val="solid"` shading is re-emitted as the visually-equivalent `clear`
fill, an explicit `<w:… w:val="none"/>` border is indistinguishable from
an absent one, and `w:tblLook` is still written as a hardcoded `04A0`. See
`CHANGELOG.md` for the full list.

### Theme-linked cell shading

A table cell shaded via a theme reference (`w:themeFill`/`w:themeColor`) no
longer freezes into a flat color, or — for a cell whose source cached no
concrete fallback color at all — gets dropped outright. Per
[MS-OE376](https://learn.microsoft.com/en-us/openspecs/office_standards/ms-oe376/c7a6a5fd-538c-4a77-8cbb-0f447298dace),
the theme reference is the primary color and a cached fallback is only
consulted when the reference is absent, so a producer writing
`w:themeFill` alone is valid input, not malformed — and that link is now
kept: `<w:shd w:val="clear" w:themeFill="..."/>` with no `w:fill` is a
shape this library has never emitted before, verified schema-valid with
`DocxValidator` and confirmed to open without a repair prompt in Word.
Calling `SetShading` explicitly still clears any cached theme reference,
since a caller-supplied color has nothing to do with the source's theme.

An explicitly white-shaded cell (`<w:shd w:fill="FFFFFF">`) is also no
longer indistinguishable from a cell nobody touched — `TableCell` now
tracks whether `SetShading` was ever called, the same explicit-versus-unset
shape `w:caps` needed.

### Run "All Caps" fidelity

A run's "All Caps" display formatting is no longer dropped on round-trip
(#102) — see `domain.Run.SetCaps`/`Caps` under Added, above.
`ConsolidateRuns` (and `MergeTemplate`/`ReplaceText`, which call it) no
longer merges two adjacent runs with different `Caps()` values, silently
turning the formatting on or off for whichever side lost. An explicit
`<w:caps w:val="false"/>` — a run overriding a style's own All Caps back
off — no longer disappears on resave either.

### Section breaks and other correctness fixes

- A mid-document section break embedded in a paragraph's own `pPr` is no
  longer dropped when its optional `w:sectPr/w:type` is absent — its
  schema default is `nextPage`, not "no break," and treating it as "not a
  section break" silently merged what the source modeled as separate
  sections into one.
- The first image added to a document opened via `OpenDocument` that had
  none is now declared in `[Content_Types].xml` — previously reproduced on
  every released version and produced a package Word offers to repair.
- Several `document.open`-scoped CLI/RPC calls (`section.add`,
  `template.inspect`, header/footer table hydration) no longer leave
  half-applied state behind when they fail partway through, or misreport a
  placeholder's location inside a header/footer table cell.

**Known limitation:** a section-ending paragraph that also carries real
content still splits into two paragraphs on round-trip — this fix only
de-duplicates the *bare* case (a paragraph whose sole content is its
section break). See `CHANGELOG.md` for the detail on why, and the specific
tests that pin current behavior for both shapes.

## Compatibility

- **`domain.Header`/`domain.Footer` each gained three methods**
  (`AddTable`, `Tables`, `Blocks`). A direct implementation needs to add
  all three; embedding either interface is unaffected. No exported docxgo
  API accepts one of these from a caller — only returns one.
- **`domain.Table` gained two methods** (`SetBorders`, `Borders`). Same
  reasoning as above — a direct implementation needs both; embedding is
  unaffected.
- **`domain.Run` gained two methods** (`SetCaps`, `Caps`). Same reasoning.
- **`AddHyperlink` now rejects a url containing a double quote**, where it
  previously accepted it and produced a broken link silently.
- **A run returned by `AddHyperlink` now carries a field.**
  `run.Fields()` returns one entry where it previously returned none;
  `ReplaceText`/`MergeTemplate` now skip that run's display text like any
  other field-bearing run instead of rewriting it.
- **`run.SetText` on the run `AddHyperlink` returns no longer changes the
  saved document.** The field captures display text at construction and
  the serializer always re-emits it; call `AddHyperlink` again, or build
  the run via `AddRun` + `run.AddField(docx.NewHyperlinkField(...))`
  directly, to change what a hyperlink run displays.
- **`pkg/template`'s `ReplaceText`/`MergeTemplate` no longer skip header
  and footer matches on a round-tripped document.** If you relied on the
  old behavior — a strict-mode merge failing rather than touching a
  preserved header — check `Document.HasPreservedHeadersOrFooters` before
  calling.
- **`Document.Paragraphs()` no longer includes a bare paragraph whose sole
  purpose was carrying a mid-document section break**, on a document
  opened via `OpenDocument`. It's now represented purely by the section
  boundary, matching what the writer already produces for its own section
  breaks.

No public Go interface method was removed, and no existing method's
signature changed.
