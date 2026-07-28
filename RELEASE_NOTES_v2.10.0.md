# Release Notes - v2.10.0

**Release Date:** July 27, 2026

## Summary

v2.10.0 bundles a batch of work that would otherwise have shipped as several
small patch releases: a critical security fix, in-place editing RPC methods
recovered from a closed pull request, mandatory CI checks, and three
correctness fixes that had been documented as known limitations for a while.
It's a minor release — new public interface methods, no breaking changes.

## Added

### In-place editing RPC methods, recovered from #64

This release recovers and completes #64, originally contributed by
@asaf-ramati1. That pull request was closed unmerged after its source fork
was deleted — but GitHub retained the two commits under `refs/pull/64/head`,
so they're merged here with original authorship intact, not reimplemented or
squashed.

Adds five new surfaces:

- `document.replaceText` (and the underlying `template.ReplaceText(doc, find,
  replace) (ReplaceResult, error)` Go API) — replaces every occurrence of
  `find` across body paragraphs, table cells, headers, and footers. Runs are
  consolidated first so text Word fragmented across runs with identical
  formatting is matched. A match touching a run that carries a field is
  always skipped (such a run's text is never serialized, so "replacing" it
  would report a change Word never makes); a match spanning several runs is
  skipped if any of them carries a break or image, but a match confined to a
  single run carrying either is replaced safely.
- `paragraph.setText` — replaces a paragraph's entire text content in one
  call.
- `table.getCell` / `table.setCell` — read and write a single cell's content
  directly, without walking the whole table.
- `table.list` gains an `includeText` flag to return cell text inline.

Two defects were found and fixed as part of recovering this work (not present
in the original #64 diff): `ReplaceText` called `ConsolidateRuns`
unconditionally on every paragraph before checking for a match, so a call
whose `find` string matched nothing still rewrote run structure across the
entire document; and a match inside a run carrying a field could be replaced
in memory while the serializer silently discarded the change at write time,
with the RPC still reporting it as `replaced`.

### New domain interface methods

- `domain.TableCell.RemoveParagraph(index int) error` — lets `table.setCell`
  actually shrink a cell's paragraph count instead of leaving cleared-but-
  present leftovers behind.
- `domain.Paragraph.SetIndentLeft` / `SetIndentRight` / `SetIndentFirstLine` /
  `SetIndentHanging(twips int) error` — each sets exactly one side of a
  paragraph's indentation, including to 0 to explicitly override a style's
  own value on just that side, without the ambiguity `SetIndent`'s
  whole-struct call has (a zero-valued side is indistinguishable from a side
  that was never touched).

Both follow the precedent set by `domain.Document`'s two additions in v2.6.0
and `Paragraph.RemoveRun` itself in v2.3.0: each has exactly one implementor
in this repo, and nothing in the public API accepts one of these interfaces
from a caller — only returns one. Minor, not breaking.

## Fixed

### Field code injection via unescaped double quotes (critical, CodeQL `go/unsafe-quoting`)

`NewHyperlinkField`'s `url`, `NewStyleRefField`'s `styleName`, and
`NewTOCField`'s `"levels"` switch were interpolated directly inside a quoted
argument of the generated Word field code. A value containing `"` broke out
of that argument, letting arbitrary field-code syntax reach `word/document.xml`
on save. The sink was reachable through the RPC layer via any field-style
parameter read from JSON input on stdin.

Escaping the quote was considered and rejected: the reader's own field-code
parser splits instructions on bare quotes with no escape awareness, so any
`\"`-style scheme would corrupt docxgo's own open→save round-trip. A value
containing a quote is rejected instead, at construction. None of the three
field constructors previously returned an error, so rejection is recorded on
the field and surfaces the next time it's used — `Run.AddField` is the one
place every field must pass through to reach a run's XML, so that's where the
check lives.

The fix combines a `strings.ReplaceAll`-shaped sanitizer with a
compare-and-reject wrapper around it, rather than a simple
`strings.Contains`-and-return-early guard: CodeQL's `go/unsafe-quoting` query
only recognizes a `ReplaceAll`-shaped transformation as a barrier, not a
guard-and-return, so the semantically equivalent early-return form would have
left the alerts open despite being a correct fix.

### `table.setCell` can now shrink a cell's paragraph count

Writing fewer paragraphs than a cell already had used to clear the extra
paragraphs' content but leave them present, so `paragraphCount` in the result
could exceed the number of items actually written. `table.setCell` now
removes the trailing paragraphs a shrinking write doesn't need, using the new
`TableCell.RemoveParagraph`. `paragraphCount` now always equals the number of
items provided.

### Paragraph indentation no longer loses explicit sides on read

A source `<w:ind>` element naming only some sides (say, just `left`) used to
come back out of docxgo re-serialized with the *other* sides — `right`,
`firstLine`, `hanging` — as explicit `0`, because the reader merged whatever
it read into one `SetIndent` call, and that call can't tell "this side was 0
in the source" apart from "this side was never mentioned." A style's own
non-zero indentation on an untouched side was silently overridden as a
result. The reader now calls the four new per-side setters directly, one per
attribute actually present in the source, so a side the source document never
mentioned stays untouched on save.

### Builder options that used to do nothing

`WithDefaultFont`, `WithDefaultFontSize`, `WithPageSize`, and `WithMargins`
were silent no-ops: `NewDocumentBuilder` built a `Config` from every applied
`Option` but only ever read `Metadata` and `Theme` off it. All four compiled,
ran without error, and produced a document completely unaffected by the
value passed in — the existing tests for this were false positives that
happened to pass for unrelated reasons (an empty-document error, or a
coincidental match with the actual default).

`WithPageSize`/`WithMargins` now flow into the document's default section.
`WithDefaultFont`/`WithDefaultFontSize` now write into `styles.xml`'s
`w:docDefaults/w:rPrDefault`, which sits below the Normal style in the OOXML
cascade — a theme applied via `WithTheme` still wins for any style that sets
its own font or size.

`WithStrictValidation` is deprecated rather than given invented semantics:
there has never been a strict-validation subsystem for it to enable (the
`service.NewValidator` from the original v2 design was never built), and
choosing behavior now, as a side effect of a no-op bugfix, would have baked
undiscussed policy into a permanent public contract. See #92 for the
follow-up to design real semantics for it.

## Changed

### CI now enforces what CONTRIBUTING.md always asked for

- `gofmt -l` and `go test -race ./...` now run in CI — both documented
  requirements that were never actually checked.
- The golangci-lint step no longer limits itself to `only-new-issues`, so
  "Lint, Build and Test" checks the whole repository rather than just each
  PR's diff.
- **`Lint, Build and Test`, `Node.js Tests`, and `CodeQL` are now required
  status checks** on the `master` branch ruleset. Previously nothing was
  required to pass before a merge. A `RepositoryRole: Admin` bypass actor was
  added to the ruleset at the same time — there was none before, so a
  required context that stopped reporting could otherwise have blocked every
  merge, including the repository owner's.
- `CONTRIBUTING.md` documents the three required checks, which weren't
  mentioned anywhere before.

## Compatibility

- **No breaking changes.** `domain.TableCell.RemoveParagraph` and
  `domain.Paragraph.SetIndentLeft/Right/FirstLine/Hanging` are additive
  interface methods; see **Added** above for why they don't break existing
  implementors.
- `WithStrictValidation` is deprecated (a no-op, as it always has been);
  `Config.StrictValidation` is kept only so existing callers don't fail to
  compile.
- A builder calling none of `WithDefaultFont`/`WithDefaultFontSize`/
  `WithPageSize`/`WithMargins` produces byte-identical output to before this
  release. A builder that *does* call them gets, for the first time, the
  document it actually asked for — most notably, `WithPageSize(docx.Letter)`
  previously had zero effect (the section always defaulted to A4 regardless
  of this option), so any caller relying on it will see a real page-size
  change in generated output.
- **CLI:** `paragraph.add`/`paragraph.setText`'s `indent` fields (`left`/
  `right`/`firstLine`/`hanging`) become nullable (`*int` server-side).
  Omitting a field is unchanged; `0` now behaves differently (honored, rather
  than silently dropped) on a side the caller explicitly set.
- A source document whose `<w:ind>` names both `firstLine` and `hanging` on
  the same paragraph now loads successfully, with each side applied
  independently, instead of failing the whole document read — which is what
  `SetIndent`'s mutual-exclusivity check used to do when the reader routed
  through it.
