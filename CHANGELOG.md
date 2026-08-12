## Unreleased

### Added

- **`domain.Header`/`domain.Footer` can now hold tables, not just paragraphs.** Tables in headers/footers are a common Word pattern (logo + address + page-number letterheads) that was previously completely unmodeled: `header.Paragraphs()` returned 0 for a table-only header, and a table there survived a resave only incidentally, via preserved bytes. Three new methods on both interfaces:
  - `AddTable(rows, cols int) (Table, error)` — same bounds validation as `Document.AddTable`.
  - `Tables() []Table` — top-level tables only, same "top-level only" contract `Paragraphs()` already has (a paragraph inside a header table cell is not in `Paragraphs()`).
  - `Blocks() []Block` — the authoritative ordering view across paragraphs and tables, reusing the same `domain.Block` type `Document.Blocks()` already returns. `Block.SectionBreak` is never set for a header/footer, since `w:hdr`/`w:ftr` carry no section properties of their own.

  `OpenDocument`/`OpenDocumentFromBytes`/`OpenDocumentFromReader` now hydrate header/footer tables into the domain model instead of silently dropping them (best-effort: a table docxgo can't represent, e.g. a grid wider than 63 columns, is skipped rather than failing the whole document — the same tolerance header/footer parts already have for other malformed content). `AddHyperlink` on a paragraph inside a header/footer table cell is rejected for the same reason it's rejected on a top-level header/footer paragraph (no per-part relationships file yet). `pkg/template`'s `ReplaceText`/`MergeTemplate`/`FindPlaceholders` now reach placeholders inside header/footer table cells. The CLI/RPC `headers`/`footers` content arrays accept `{"type": "table", ...}` items (previously silently decoded as an empty paragraph).

  **Known limitations:** table properties (`w:tblPr` beyond style, `w:tblGrid`, `w:trPr`, most of `w:tcPr`) are not yet hydrated from an existing document — the same gap body tables already have. A nested table inside a table cell is still dropped on read, in both the body and a header/footer. A header/footer table in a document opened via `OpenDocument` is still written back from preserved bytes on an untouched resave — the domain model isn't consulted yet for an *opened* document's header/footer (it is for one built fresh via `AddTable`).

  **Pre-existing, unrelated to this change, but worth calling out here:** an image added to *any* header/footer paragraph — inside a table cell or not — produces a package Word will flag for repair. `internal/xml.Header`/`Footer` never declare the `xmlns:wp`/`xmlns:a`/`xmlns:pic` namespaces a `<w:drawing>` needs (unlike `internal/xml.Document`, which does), so `<w:hdr>`/`<w:ftr>` end up with an undeclared-prefix error; separately, `attachImage` always mints its relationship into `document.xml.rels`, which a header/footer part can't resolve at all. Confirmed on `master` before this PR (plain paragraph, no table involved) via `DocxValidator`, so it is not new here and this PR does not attempt to fix it — flagging it since PR 2b's own header-table serializer reuses the image-capable `tableSerializer`, which makes the path newly reachable through a table cell too, not just a plain paragraph.
- **Header/footer relationship parts (`word/_rels/headerN.xml.rels`, `word/_rels/footerN.xml.rels`) are now preserved through their own dedicated path instead of the opaque, unrelated-parts bucket.** Previously these files survived a round-trip only incidentally, as part of `PreservedParts.Additional` — indistinguishable from any other unrecognized part in the package. They're now read into `Package.HeaderRels`/`FooterRels`, threaded through `core.PreservedParts.HeaderRels`/`FooterRels`, and written back verbatim on an untouched resave. No behavior change yet (headers/footers still don't mint their own relationships); this is the plumbing PR 2c (per-part relationships) builds on.
- **`domain.Run.SetCaps` / `Caps`** — set whether a run displays in all capitals (`w:caps`). This is a display-only override, the same distinction as Word's own "All Caps" character formatting versus actually typing in capitals: it does not change `run.Text()` or the run's stored `<w:t>` content, only how it renders.
  - `SetCaps(caps bool) error` / `Caps() bool` — same shape as the existing `SetBold`/`Bold`, `SetItalic`/`Italic`, `SetStrike`/`Strike`.
  - Written as `w:caps` in the run's own `w:rPr`.
  - Opening an existing `.docx` via `OpenDocument`/`OpenDocumentFromBytes`/`OpenDocumentFromReader` now hydrates each run's `Caps()` from its `w:rPr/w:caps`, if present — same round-trip fidelity as every other run property (`Bold`, `Highlight`, etc). Added because #102 reported it as a real loss: a title formatted all-caps in Word round-tripped back to its literal (mixed-case) stored text, since docxgo had no representation for the property at all.

### Fixed

- **A header/footer table that fails to hydrate partway through is no longer left half-added.** `hydrateTable`'s own `AddTable` call attaches the table to the header/footer before its rows and cells are populated, so a later cell failure (e.g. an unparseable `gridSpan`) was returning an error that the best-effort header/footer path correctly tolerated — but the partially-populated table stayed behind in `Tables()`/`Blocks()` anyway, instead of being skipped entirely as the "best-effort" tolerance is supposed to mean. `docxHeader`/`docxFooter` gained an unexported `RemoveLastTable`, called on this error path to undo the attach.
- **`section.add` (CLI/RPC) no longer leaves a half-configured section behind when it fails partway through.** `AddSectionWithBreak` attaches the new section to the document immediately, before this call's own validation (page size, margins, headers, footers, ...) runs — and that document is the same instance the server keeps in its store across calls, keyed by `documentId`. An error partway through (e.g. an unrecognized `"type"` in a header's content array) used to leave the new section permanently in `Sections()`, visible to every later operation on that `documentId` despite the call itself having failed. `document` (internal/core) gained `RemoveLastSection`, called on this error path.
- **`template.inspect` (CLI/RPC) no longer omits `table`/`row`/`cell` coordinates for a placeholder found inside a header or footer table cell.** It gated those keys on `Location.Type == LocationTableCell`, but a header/footer table cell match deliberately keeps `Type == LocationHeader`/`LocationFooter` (see `walkHeaderFooterTables`'s doc comment, above) — so every such placeholder reported only `"location": "header"`/`"footer"`, indistinguishable from a plain header/footer paragraph match. `template.Location` gained a new `InTableCell bool` field, the real discriminator (`Type` alone can't tell a header table cell match from a plain header paragraph match, since both keep the same `Type` and zero-valued indices); `template.inspect` now gates on it instead.
- **`Paragraph.AddHyperlink` now emits a real `<w:hyperlink>` element instead of an orphaned relationship (#101).** It minted a hyperlink relationship in `word/_rels/document.xml.rels` but never attached it to anything — the run it created carried no field, so the serializer had no signal to wrap it in `<w:hyperlink r:id="...">`. Word opened the file and rendered the run as plain blue, underlined text: not a link, and not visibly wrong either, so nothing caught it. `AddHyperlink` now builds the run through `NewHyperlinkField` + `run.AddField` — the same path `docx.NewHyperlinkField` already used correctly — which is what makes the serializer emit the real element. This also fixes the CLI/npm `{"hyperlink": {"url": ...}}` run shortcut (`paragraph.add`'s `runs`, `table.setCell`'s cell paragraphs), which was `AddHyperlink`'s only production caller.
- **`AddHyperlink("#anchor", text)` now produces a working internal link.** Previously it unconditionally minted an `External` relationship whose target was the literal `"#anchor"` string. The serializer already resolves an internal link straight from the url, so the fix above surfaced a pre-existing bug on this path (and on the reader's own hydration of a real Word document's internal link) as a visible orphaned relationship; `run.AddField` no longer mints one when the hyperlink field's url starts with `#`. The target of an internal link still doesn't survive being read back from an existing `.docx` — the reader has no `w:bookmarkStart` hydration — so round-tripping a bookmark target remains a separate gap.
- **A hyperlink's display text no longer appears twice in the saved document.** `expandRunWithFields` (the serializer's field-expansion path) always serialized a run's leftover text into a trailing plain `<w:r>` after handling its fields, without checking whether a hyperlink branch had already serialized that same text into the `<w:hyperlink>`'s own `<w:r>`. Word rendered both — the link text, immediately followed by an identical plain-text run with the same formatting — for every hyperlink docxgo ever wrote or resaved, on both the external-URL and internal-anchor paths. This predates #101/#103 entirely (the serializer's hyperlink branch was untouched by that fix) and was invisible to every existing test, since they all counted `<w:hyperlink>` elements or checked the in-memory model rather than the rendered text — found while planning #101's follow-ups, by actually running the writer instead of reading the code. `expandRunWithFields` now tracks whether a hyperlink branch consumed the run's text and skips the trailing block when it did.
- **A run's "All Caps" display formatting is no longer dropped when a `.docx` is opened and resaved (#102).** See `domain.Run.SetCaps`/`Caps` under Added, above — this was a straight-up missing feature (no reader hydration, no writer serialization, no domain representation at all), not a narrower bug.
- **`ConsolidateRuns` (and `MergeTemplate`/`ReplaceText`, which call it) no longer merges two adjacent runs with different `Caps()` values.** `formatsEqual`, the check `ConsolidateRuns` uses to decide whether two runs are visually identical enough to combine, was never updated when `Caps` was added — it compared the other 7 visual attributes plus `Language` but not `Caps`, so a run with `Caps=true` sitting next to one with `Caps=false` would be merged into one run carrying only the leader's formatting, silently turning off (or on) "All Caps" for whichever side lost.
- **An explicit `<w:caps w:val="false"/>` — a run overriding a run/paragraph style's own All Caps back off — no longer disappears on resave.** `Caps() bool` has no way to represent "explicitly set to false" versus "never mentioned"; the serializer only emitted `w:caps` when `Caps()` was `true`, so both cases wrote nothing, indistinguishably. The internal run type now tracks whether `SetCaps` was ever called (`CapsSet()`, unexported, mirroring the existing `spacingSetter`/`indentSetter` pattern), and the serializer emits `w:val="false"` when it was explicitly set false. `formatsEqual` also now compares this explicit-vs-unset state, not just `Caps()` itself, so `ConsolidateRuns` no longer merges an explicit-false run into a same-valued-but-never-set neighbor and silently drops the override.
- **A table's named style is no longer dropped when a `.docx` is opened and resaved (#102).** A table commonly gets its visible borders, shading, and banding from a style referenced by `<w:tblStyle w:val="...">` — a built-in like Word's "TableGrid", or a custom one — rather than from an explicit `<w:tblBorders>` on the table itself. The reader never hydrated `tblStyle` at all, so it silently vanished on the next save even though the style's own definition survived untouched in `styles.xml`: the borders looked "removed" with nothing in the diff to explain why. `hydrateTable` now reads a table's `<w:tblPr>/<w:tblStyle>` and applies it via the existing `domain.Table.SetStyle`, which the writer already serializes correctly (`TableSerializer`) — this was purely a reader gap.
- **A mid-document section break is no longer dropped when a `.docx` is opened and resaved (#102).** A `<w:sectPr>` embedded in a paragraph's own `pPr` (rather than as the body's last child) marks that paragraph as the end of a section — per the OOXML schema, this holds even when the optional `w:sectPr/w:type` child is absent, since its schema default is `nextPage`, not "no break." The reader's `applySectionProperties` treated a missing `w:type` as "not a section break" and skipped starting a new section, silently merging what the source file modeled as two (or more) sections into one — collapsing distinct page setup, margins, and header/footer references along with it. It now distinguishes a `pPr`-embedded `sectPr`, which always starts a new section, from the body's own final `sectPr`, which never does (nothing follows it). Fixed alongside: a paragraph that exists in the source purely to carry a mid-body break (no runs, no other `pPr` content) is no longer also hydrated as its own empty `domain.Paragraph` — the writer already synthesizes a paragraph like it for every section break, and keeping both doubled it on the next save.
- **Adding a hyperlink (or any other relationship) to a document opened with `OpenDocument` no longer produces a package Word offers to repair.** `word/_rels/document.xml.rels` is normally written back verbatim on a round-tripped document, for fidelity. That's correct as long as nothing new was added to the relationship manager since the document was opened; once something is (e.g. a hyperlink attached via `AddHyperlink` or `run.AddField(docx.NewHyperlinkField(...))` after `OpenDocument`), writing the preserved bytes unchanged leaves a dangling `r:id` in the resaved `document.xml`. `Document.WriteTo` now detects that case and regenerates the rels part from the relationship manager instead of the preserved bytes — which is exact, because opening a document now defers assigning docxgo's own default relationship IDs (styles, fontTable, theme, settings, webSettings) until after the source file's own relationships are registered under their real IDs, so a source relationship numbered e.g. `rId1` — for anything, in whatever order the authoring application originally assigned it — is never silently displaced by one of docxgo's own defaults. A resave keeps writing the preserved bytes untouched only when the relationship manager still holds exactly the source file's own relationships; a source `.docx` that was already missing one of docxgo's five defaults (styles/fontTable/theme/settings/webSettings — Word tolerates that, docxgo does not for a document it writes) gets that default relationship added on save even without the caller adding anything, same as it always has for a freshly created document.

### Compatibility

- **`domain.Header` and `domain.Footer` each gained three methods** (`AddTable`, `Tables`, `Blocks`, above). If you implement either interface directly with your own type, you'll need to add all three; embedding `domain.Header`/`domain.Footer` in your type is unaffected, since the new methods are promoted automatically. No exported docxgo API accepts a `domain.Header`/`domain.Footer` from a caller (only returns one) — same reasoning as `domain.Run.SetLanguage` in v2.12.0.
- **`AddHyperlink` now rejects a url containing a double quote**, where it previously accepted it and produced a broken link silently. It routes through the same `NewHyperlinkField` constructor the field API already used, which rejects such a url because the field's `HYPERLINK "url"` code has no escape a reader can round-trip.
- **A run returned by `AddHyperlink` now carries a field.** `run.Fields()` returns one entry where it previously returned none; `ReplaceText`/`MergeTemplate` now skip that run's display text the same way they already skip any other field-bearing run, rather than rewriting it.
- **`run.SetText` on the run `AddHyperlink` returns no longer changes the saved document.** The field captures the display text at construction and the serializer always re-emits that captured text; call `AddHyperlink` again, or build the run via `AddRun` + `run.AddField(docx.NewHyperlinkField(...))` directly, to change what a hyperlink run displays. Formatting setters (`SetBold`, `SetColor`, ...) are unaffected.
- **`AddHyperlink` now returns an error on a header or footer paragraph**, where it previously produced a relationship that referenced nothing (see Fixed, above). docxgo does not yet write a per-part relationships file (`word/_rels/headerN.xml.rels`/`footerN.xml.rels`), so a hyperlink relationship minted there would reference a part that doesn't exist; real support is a follow-up.
- **`domain.Run` interface gained two methods** (`SetCaps`, `Caps`, above). If you implement `domain.Run` directly with your own type — most commonly a hand-written test double — you'll need to add both methods; embedding `domain.Run` in your type is unaffected, since the new methods are promoted automatically. No exported docxgo API accepts a `domain.Run` from a caller (only returns one), same reasoning as `Document.SetLanguage`/`Language` in v2.6.0 and `Run.SetLanguage`/`Language` in v2.12.0.
- **`Document.Paragraphs()` no longer includes a bare paragraph whose sole purpose was carrying a mid-document section break**, when the source document is opened via `OpenDocument`. Previously that paragraph was hydrated like any other (see Fixed, above); it is now represented purely by the section boundary, matching what the writer already produces for section breaks it creates itself.

### Known limitations

- **A section-ending paragraph that also carries real content still splits into two paragraphs on round-trip.** This fix (see above) only de-duplicates the *bare* case — a paragraph whose sole content is its `pPr`/`sectPr` — which is the shape the real `.docx` behind #102 uses and the shape Word produces when a section break sits on its own dedicated paragraph. It's also legal OOXML (ECMA-376 17.6.17) for the section-ending paragraph to carry text directly, which Word does whenever the boundary falls on an already-written paragraph rather than a fresh blank one; docxgo still hydrates that paragraph's text onto one `domain.Paragraph` and starts a new `domain.Section` as a separate block, and the writer always synthesizes its own empty carrier paragraph for a section break, so the source's one paragraph becomes two on resave. When the embedded `sectPr` carries an explicit `w:type`, this splitting is pre-existing, unrelated to this PR (confirmed against `master` before this change). When `w:type` is *absent* — the shape this PR's own fix targets — the splitting is a new, unavoidable side effect of this fix, not pre-existing: `master` never recognized that embedded `sectPr` as a section break at all in this content-bearing case, so it silently dropped the section boundary entirely (one merged section, no tripling) instead of splitting the paragraph. This fix trades that silent data loss for a correctly preserved section boundary at the cost of one extra paragraph on resave — a strict improvement, but the paragraph-count regression itself is new here, not carried over. Fixing it needs the writer to fold a section break into the *preceding* content paragraph instead of always minting a new one, a behavior change affecting every docxgo-authored document, not just round-tripped ones — out of scope here; see `TestReconstructSectionBreakWithContent_AgainstHandAuthoredPackage` and `TestReconstructSectionBreakWithContent_NoExplicitType_AgainstHandAuthoredPackage` in `internal/reader/reader_test.go`, which pin the current behavior for both shapes.

## v2.12.0 — 2026-08-02

### Added

- **`domain.Run.SetLanguage` / `Language`** — set a per-run proofing language override, used by Word for spell-checking, grammar-checking, and hyphenation of just that run (e.g. a foreign-language phrase inside an otherwise single-language paragraph). Complements `Document.SetLanguage`/`WithLanguage`, which only set the document-wide default; a run without an override inherits that default.
  - `SetLanguage(lang *domain.Language) error` — same `Language{Val, EastAsia, Bidi}` shape as `Document.SetLanguage`; at least one field must be non-empty. Pass `nil` to clear the override and fall back to the document default.
  - `Language() *domain.Language` — returns a defensive copy, or `nil` if unset. Mutating it has no effect on the run.
  - Written as `w:lang` in the run's own `w:rPr`, into the field `internal/xml.RunProperties.Lang` that already existed but nothing wrote to.
  - Opening an existing `.docx` via `OpenDocument`/`OpenDocumentFromBytes`/`OpenDocumentFromReader` now hydrates each run's `Language()` from its `w:rPr/w:lang`, if present — same round-trip fidelity as every other run property (`Bold`, `Highlight`, etc).

### Fixed

- **`ConsolidateRuns` (and therefore `MergeTemplate`/`ReplaceText`) could merge across a language boundary.** `formatsEqual`, which decides whether two adjacent runs are mergeable, compared 8 visual formatting attributes but not the new `Language`. Two visually-identical runs — one carrying a `SetLanguage` override (e.g. a `bonjour` run tagged `fr` inside otherwise-untagged prose), one not — were merged into a single run, silently discarding whichever run's language the merge didn't keep. Found in review of this same release before it was tagged. `formatsEqual` now compares `Language()` too (nil-safe: both unset, or both set to the same `Val`/`EastAsia`/`Bidi`), so a language boundary blocks the merge like any other formatting difference.

### Changed

- **`domain.Run` interface gained two methods** (`SetLanguage`, `Language`, above). If you implement `domain.Run` directly with your own type — most commonly a hand-written test double — you'll need to add both methods; embedding `domain.Run` in your type is unaffected, since the new methods are promoted automatically. No exported docxgo API accepts a `domain.Run` from a caller (only returns one), same reasoning as `Document.SetLanguage`/`Language` in v2.6.0.

---

## v2.11.0 — 2026-07-29

v2.10.0 was merged and published without a code review — CI alone. A
post-hoc review of that batch found ten confirmed defects that gofmt, `go
vet`, `go test -race`, and golangci-lint cannot see by construction: they're
about what the code *means*, not whether it compiles or races. This release
fixes all ten in one batch rather than spreading them across several
patches, since several touch the same request paths.

### Fixed

- **`document.replaceText`/`template.ReplaceText` no longer report a header or footer match as replaced when the write is discarded on save.** On a document opened via `document.open`, `WriteTo` writes preserved header/footer XML verbatim — an in-memory replacement there never reached the saved file, but the RPC reported it as `replaced` anyway. Those matches are now counted in `skipped` instead. `template.MergeTemplate` had the identical gap for placeholders inside a preserved header/footer; its `StrictMode` now surfaces them as missing keys rather than silently discarding the merge.
- **`document.replaceText`, `table.getCell`, and `table.setCell` now decode params strictly.** They used plain `json.Unmarshal` for top-level params while nested paragraph items were already decoded strictly — a misspelled field (`replacement` instead of `replace`) silently decoded to a zero value indistinguishable from a deliberately minimal request, so a typo in a `replaceText` call deleted every occurrence of `find` and reported success.
- **`paragraph.add`, `table.add`, and `document.addContent` no longer leave orphaned content behind when a request is rejected.** All three mutated the real document before content that could still fail (e.g. a field constructor rejecting a quote) had a chance to, so a rejected request could leave a stray, half-populated paragraph or table appended anyway. Each now validates against a scratch document first, the same pattern `paragraph.setText`/`table.setCell` already used.
- **`SetIndent` no longer leaves a stale per-side flag from an earlier `SetIndentLeft`/`Right`/`FirstLine`/`Hanging` call.** A paragraph hydrated by the reader via `SetIndentLeft` and then given a `SetIndent` call for a different side kept the earlier flag, so the serializer emitted an explicit `w:left="0"` the `SetIndent` call never asked for — clobbering the paragraph's style indentation on that side. This was a round-trip regression introduced by v2.10.0's own indentation fix (#76). `SetIndent` now clears all four flags along with the struct it replaces.
- **The serializer never emits both `w:ind`'s `firstLine` and `hanging` together.** `SetIndent` rejects setting both, but the per-side setters `SetIndentFirstLine`/`SetIndentHanging` don't share that check — reachable via the reader hydrating a source `<w:ind>` that already carries both attributes. Word treats the two as mutually exclusive even though the schema allows both; the serializer now emits only `hanging` when both are set, matching Word's own behavior.
- **TOC field switches (`\o`, `\h`, `\n`, `\p`, `\z`, `\u`) are now written with a single backslash.** They were built as Go raw string literals containing two literal backslashes each — XML chardata escaping never touches a bare backslash, so the doubled backslash reached `word/document.xml` verbatim, and Word treats `\\o` as an unrecognized switch. A TOC field built by docxgo never actually applied any of its switches. Rebuilding an existing document doesn't change this on its own — only a document containing a TOC field constructed by `NewTOCField` (directly, or via `field: {type: "toc"}` in the RPC) is affected.
- **`Field.SetCode` now rejects control characters and line breaks.** It previously accepted any non-blank string verbatim and `serializer.go` emits `Code()` straight into `<w:instrText>` — a way to reach the same class of corruption v2.10.0's field-code sanitizer (#85) exists to prevent, bypassing it entirely since `SetCode` sits outside the `New*Field` constructors' quote check. `SetCode` still accepts balanced quotes (e.g. `STYLEREF "Heading 1"`), since a well-formed instruction legitimately contains them. The reader hydrates a field's code from a real file's own `<w:instrText>` through a new, unexported `SetCodeRaw` path that bypasses this guard — the source file is the ground truth for a valid instruction, not this package's own validation, so `OpenDocument` doesn't fail on a normal `.docx` because of this new check.
- **Release workflow no longer hard-fails when `RELEASE_PAT` is unset.** Passing an empty `secrets.RELEASE_PAT` as `softprops/action-gh-release`'s `token` input overrides the action's own fallback to `github.token` — an explicitly-passed empty `with:` input wins over the action's default, unlike an unset env var. `release.yml` now falls back explicitly (`secrets.RELEASE_PAT || github.token`), matching the degraded-but-working behavior the emptiness check next to it already documented.

### Changed

- **Docs corrected to match actual `paragraph.setText`/`table.setCell` behavior.** Both replace paragraph *content* (text and runs) only — any paragraph property (style, alignment, indent, spacing) omitted from the request keeps its existing value rather than being reset. `CLI_GUIDE.md` and the npm `ParagraphSetTextParams` JSDoc previously described these as full replacements without drawing that distinction.
- **`TableSetCellResult`'s npm JSDoc corrected.** It said the cell "can grow but never shrink" — the pre-v2.10.0 behavior, which #84 (also v2.10.0) made false in the same release it shipped in.

### Compatibility

- **`document.replaceText`/`template.ReplaceText` response semantics change for documents with preserved headers/footers.** A call that used to report a header/footer match as `replaced` now reports it as `skipped`; the total match count is unchanged. Every other document shape is unaffected.
- **`document.replaceText`, `table.getCell`, and `table.setCell` now reject a request containing an unrecognized top-level field**, where it used to be silently ignored. A well-formed request (no typos, no extra fields) is unaffected.
- **TOC field code bytes change.** A document created with `NewTOCField`/`field: {type: "toc"}` under this release has single-backslash switches (`\o "1-3" \h \z \u`) instead of the doubled-backslash bytes v2.10.0 and earlier produced. This is a fix, not a behavior a caller could have depended on: the doubled-backslash form was never a working TOC field in Word.
- **`Field.SetCode` (Go API only; not reachable from the RPC layer) now rejects a control character or line break** where it previously accepted anything non-blank. No CLI protocol or npm surface exposes `SetCode` directly.
- No public Go interface changed. No new public Go, CLI protocol, or Node.js API surface.
- **Retroactive note on v2.10.0's `Config` pointerization** (`### Compatibility` there described the change but not its concrete consequence): `docx.Config`'s `DefaultFont`/`DefaultFontSize`/`PageSize`/`Margins` fields becoming pointers is source-breaking for code that writes to a `*docx.Config` directly (e.g. a custom `docx.Option`), not for code that only calls `docx.With*` functions. `c.DefaultFont = "Arial"` no longer compiles. See `MIGRATION.md`.

---

## v2.10.0 — 2026-07-27

### Added

- **In-place editing RPC methods, recovered from #64** (PR #83; originally contributed by @asaf-ramati1 in #64). That PR was closed unmerged after its source fork was deleted, but GitHub retained the two commits under `refs/pull/64/head` — they're merged here with original authorship intact, not reimplemented or squashed. Adds `document.replaceText`, `paragraph.setText`, `table.getCell`, `table.setCell`, and a `includeText` flag on `table.list`, plus the underlying Go API: `template.ReplaceText(doc, find, replace) (ReplaceResult, error)`, which replaces every occurrence of `find` across body paragraphs, table cells, headers, and footers, consolidating fragmented runs first and skipping (not corrupting) matches that touch a field, break, or image in a way that can't be replaced safely.
- `domain.TableCell.RemoveParagraph(index int) error` (#84), so `table.setCell` can actually shrink a cell's paragraph count instead of leaving cleared-but-present leftovers. `paragraphCount` in the result now always equals the number of items provided.
- `domain.Paragraph.SetIndentLeft/SetIndentRight/SetIndentFirstLine/SetIndentHanging(twips int) error` (#76), each setting exactly one side of a paragraph's indentation — including to 0, to explicitly override a style's own value on just that side — without touching the other three the way `SetIndent`'s whole-struct call must. The CLI's `paragraph.add`/`paragraph.setText` `indent` fields become nullable (`*int` server-side) for the same reason v2.9.0 did this for spacing.

### Fixed

- **Rejected double quotes in generated field-code arguments** (#85, critical CodeQL `go/unsafe-quoting`). `NewHyperlinkField`'s `url`, `NewStyleRefField`'s `styleName`, and `NewTOCField`'s `"levels"` switch were interpolated unescaped inside a quoted argument of the generated Word field code — a value containing `"` broke out of the quoted argument and injected arbitrary field-code syntax. Escaping was ruled out: the reader's field-code parser splits on bare quotes with no escape awareness, so any `\"` scheme would corrupt docxgo's own open→save round-trip. A value containing a quote is rejected instead, surfaced through `run.AddField` — the one chokepoint every field must pass through to reach a run's XML. The fix uses `strings.ReplaceAll` as the barrier (not an early-return `strings.Contains` guard): CodeQL's `go/unsafe-quoting` only recognizes a `ReplaceAll`-shaped sanitizer, not guard-and-return, so a semantically-equivalent early-return form would have left the alerts open.
- **`ReplaceText` no longer mutates the whole document on a miss.** It called `ConsolidateRuns` unconditionally on every paragraph before checking for a match, so a `document.replaceText` whose `find` string appears nowhere still rewrote run structure across the entire body, every table cell, and every header/footer. Fixed by checking the concatenated paragraph text first.
- **A match inside a run carrying a field is now skipped, not silently discarded.** The single-run skip guard was gated on `len(spanned) > 1`, so e.g. a PAGE field's cached text could be replaced in memory while the serializer discarded the change at write time — the RPC still reported it as `replaced`.
- **`table.setCell` can now shrink a cell's paragraph count**, using the new `TableCell.RemoveParagraph` above instead of clearing-but-keeping leftover paragraphs.
- **A source `<w:ind>` naming only some sides no longer clobbers the others on save.** The reader merged every attribute it read into one `SetIndent` call, which can't distinguish "this side was 0 in the source" from "this side was never mentioned" — so a document with only `left` set in its source XML came back out with `right`/`firstLine`/`hanging` as explicit `0`, silently overriding a style's own indentation on sides the source never touched. The reader now calls the new per-side setters above directly, one per attribute actually present.
- **`WithDefaultFont`, `WithDefaultFontSize`, `WithPageSize`, and `WithMargins` are no longer silent no-ops** (#45). `NewDocumentBuilder` built a `Config` from every applied `Option` but only ever read `Metadata` and `Theme` off it — the other four compiled, ran without error, and produced a document completely unaffected by the value passed in. `WithPageSize`/`WithMargins` now flow into the default section; `WithDefaultFont`/`WithDefaultFontSize` now write into `styles.xml`'s `w:docDefaults/w:rPrDefault`, below the Normal style in the OOXML cascade so a theme's own font/size still wins. `WithStrictValidation` is deprecated rather than given invented semantics — there was never a validation subsystem for it to enable (see `### Compatibility`).

### Changed

- **CI now runs `gofmt -l` and `go test -race ./...`**, both required by `CONTRIBUTING.md` but never actually checked, and the golangci-lint step no longer limits itself to `only-new-issues`, so "Lint, Build and Test" checks the whole repository.
- **`Lint, Build and Test`, `Node.js Tests`, and `CodeQL` are now required status checks** on the `master` branch ruleset — previously nothing was required to pass before merge. A `RepositoryRole: Admin` bypass actor was added at the same time (there was none before), so a context that stops reporting can't block every merge including the owner's.
- `CONTRIBUTING.md` documents the three required checks.

### Compatibility

- **New public Go interface methods, not breaking.** `domain.TableCell.RemoveParagraph` and `domain.Paragraph.SetIndentLeft/Right/FirstLine/Hanging` follow the precedent of `domain.Document`'s two additions in v2.6.0 and `Paragraph.RemoveRun` itself in v2.3.0: each has exactly one implementor in this repo, and no exported API accepts one of these interfaces from a caller — only returns one.
- `WithStrictValidation` is deprecated (no-op, as it always has been); `Config.StrictValidation` is kept only so existing callers don't fail to compile. See #92 for the follow-up to give it real semantics.
- A builder that calls none of `WithDefaultFont`/`WithDefaultFontSize`/`WithPageSize`/`WithMargins` produces byte-identical output to before this release. A builder that does call them now gets the document it asked for, for the first time — most notably, `WithPageSize(docx.Letter)` previously had zero effect (the section always defaulted to A4 regardless), so a caller relying on that option now sees a real page-size change.
- CLI: `paragraph.add`/`paragraph.setText`'s `indent` fields (`left`/`right`/`firstLine`/`hanging`) become nullable (`*int` server-side); omitting a field is unchanged, but `0` now behaves differently (honored, rather than silently dropped) on a side the caller explicitly set.
- A source document whose `<w:ind>` names both `firstLine` and `hanging` on the same paragraph now loads successfully (each side applies independently) instead of failing the whole document read, which `SetIndent`'s mutual-exclusivity check used to do.

---

## v2.9.1 — 2026-07-27

### Fixed

- **An explicit spacing override no longer clobbers a style's line spacing.** v2.9.0's fix for explicit-zero spacing widened the emit gate to fire whenever `spacingBefore`/`spacingAfter` was explicitly set, but `Line`/`LineRule` were filled into the same `<w:spacing>` element unconditionally — a paragraph that only called `SetSpacingAfter(0)` gained an unintended direct `w:line="240" w:lineRule="auto"`, silently overriding a style's real line spacing (e.g. 1.5 lines rendering as single-spaced). `Line`/`LineRule` are now gated by their own departure-from-default check, independent of before/after.
- **`FindPlaceholders`'s reported `Location` is now correct for the common, non-split case.** v2.9.0's offset translation used an inclusive comparison for both the start and end of a match; applied to the start, it misattributed a match beginning exactly at a run boundary — the ordinary Word `"Hello "` + `"{{Name}}"` split — to the end of the *previous* run instead of the start of the one actually holding it, so `RunIndex`/`StartOffset` pointed at text that didn't contain the placeholder and slicing it could panic. It also misattributed matches following a leading empty run. `Location`'s doc comment now states the contract explicitly: when `RunIndex == EndRunIndex`, single-run slicing works exactly as it did before v2.9.0.

### Changed

- **CI is now resilient to the npm lock's platform-package gap between releases.** `npm/package-lock.json` is regenerated with `npm install --package-lock-only` during release prep, which necessarily drops the platform-package entries (they don't exist on the registry yet at that point) — but `npm ci` requires every declared `optionalDependency` to have a resolved lock entry, so CI's Node.js Tests job failed on any commit between a release-prep commit and the next lock regeneration. This reproduced identically after both the v2.8.0 and v2.9.0 release commits. CI now uses `npm install`, which reconciles instead of requiring exact pre-sync.
- **Hardened the release workflows.** `inputs.version` is now bound via `env:` and validated against a semver pattern before use, instead of being interpolated directly into `run:` shell scripts in a job holding `NPM_TOKEN` and `id-token: write`. Dropped the now-redundant `release: types: [published]` trigger from `npm-publish.yml` — `release.yml`'s `publish-npm` job already calls it unconditionally (see v2.9.0), so keeping both would race two `npm publish` calls for the same tag once `RELEASE_PAT` is configured. The `RELEASE_PAT` emptiness check in `release.yml` is also now bound via `env:` rather than expanded directly into a shell condition.

### Compatibility

- No public Go, CLI protocol, or Node.js API changed.
- Generated output for a paragraph that both carries a style with non-default line spacing *and* explicitly sets only `spacingBefore`/`spacingAfter` (not line spacing) changes back to what v2.9.0 intended: the style's line spacing is inherited again instead of being overridden by an unintended direct `0/240/auto`. Any other paragraph is unaffected.
- `FindPlaceholders`/`FindPlaceholdersCustom`/`PlaceholderNames`/`ValidateTemplate` report different (corrected) `RunIndex`/`StartOffset` values for matches that sit exactly at a run boundary or follow a leading empty run — a narrow input shape most callers won't have hit, since v2.9.0 shipped hours before this fix.

---

## v2.9.0 — 2026-07-27

### Fixed

- **npm publish no longer depends on the GitHub release event.** `release.yml` authors the GitHub Release with a `RELEASE_PAT` so the `release: published` event fires `npm-publish.yml` automatically — but when that secret is empty or expired, the release silently falls back to being authored by `github-actions[bot]`, and GitHub's anti-recursion rule means a `GITHUB_TOKEN`-authored event can never trigger another workflow. Every release since at least v2.7.2 had to be published to npm by hand via `workflow_dispatch`, with `release.yml` reporting success regardless. `release.yml` now invokes `npm-publish.yml` directly as a `workflow_call` job, passing the version it already computed, so publishing no longer depends on that event firing at all. The `RELEASE_PAT` check is now a build warning instead of the only thing standing between "released" and "published to npm".
- **`FindPlaceholders`, `FindPlaceholdersCustom`, `PlaceholderNames`, and `ValidateTemplate` no longer mutate the document they scan.** All four read as pure queries but called `ConsolidateRuns` internally to heal placeholders Word split across runs — v2.8.0 removed the worst consequence of that (silently dropped images), but the mutation itself remained: a caller who scanned a template to preview its placeholders got their in-memory document restructured as a side effect. The run-grouping logic `ConsolidateRuns` uses to decide what's mergeable is now shared with a new scan path that matches against each group's concatenated text and reports offsets back against the paragraph's real, unmodified runs — nothing merges unless `MergeTemplate` actually needs to write replacement text. `Location` gains `EndRunIndex` so a match spanning multiple runs can report where it ends, not just where it starts.
- **An explicit zero spacing (or line spacing) is now honored on a paragraph carrying a style.** v2.8.0 fixed the *unstyled* case: a paragraph with no style and no explicit spacing correctly inherits `0` from the new `w:pPrDefault`. It did not fix the styled case, because `Paragraph.SpacingBefore()`/`SpacingAfter() int` can't distinguish "never set" from "explicitly set to 0" — so `SetSpacingAfter(0)` on a paragraph with a style supplying non-zero spacing was silently dropped, and the style's value won instead of the caller's explicit zero. The same gap applied to line spacing: an explicit `SetLineSpacing(Auto, 240)` meant to override a style's non-auto rule back to single-spacing was indistinguishable from never calling it, since `Auto`/`240` are also the defaults. The concrete paragraph type now tracks whether each setter was ever called and the serializer honors an explicit value even at zero. **Indentation has the identical gap** but needs a different fix shape — tracked in [#76](https://github.com/mmonterroca/docxgo/issues/76).

### Changed

- **Removed the unreachable no-style-manager `styles.xml` fallback.** `ZipWriter.writeDefaultStyles` wrote a hand-maintained raw XML string, reachable only when `writeStyles` received a `nil` `*Styles` — which never happens from the one production caller. It was a second, independent source of truth for default paragraph properties that had to be kept in sync by hand with the serializer; `writeStyles` now returns an error on `nil` instead of degrading to it.
- **CLI:** `paragraph.add`'s `spacingBefore`/`spacingAfter` fields become nullable (`*int` server-side) so a JSON `"spacingAfter": 0` can be told apart from omitting the field — otherwise the spacing fix above would only reach the Go API, not the CLI or Node.js wrapper.
- Refreshed `npm/package-lock.json` for the version bump; this incidentally restores `npm ci` in CI, which had been broken since the v2.8.0 release commit landed with an unrelated lockfile desync (unrelated to any fix in this release).

### Compatibility

- **No public Go interface changed.** `domain.Paragraph` is unchanged; the new `SpacingBeforeSet()`/`SpacingAfterSet()`/`LineSpacingSet()` methods are exposed only on the concrete internal type, read by the serializer via a type assertion that degrades gracefully for any other `domain.Paragraph` implementation (third-party, wrapped) — see `TestParagraphSerializer_WrappedParagraphDegradesGracefully`.
- **`template.Location`** gains a new field, `EndRunIndex` (additive). `RunIndex`/`StartOffset` now point at the paragraph's own run and offset rather than a post-consolidation run that no longer exists for scans — no non-test consumer of these fields exists in this repo, and no `RunIndex`/`StartOffset`/`EndOffset` semantics changed for the common case of a match inside a single run.
- **CLI JSON-RPC:** `spacingBefore`/`spacingAfter` remain optional integer fields; omitting them is unchanged, but `0` now behaves differently (honored, rather than ignored) when the paragraph carries a style.
- Generated output changes as described under **Fixed** for paragraphs that explicitly set spacing to 0 on a styled paragraph.

---

## v2.8.0 — 2026-07-27

### Fixed

- **Inline images could be silently deleted by template operations.** `template.ConsolidateRuns` merges adjacent identically-formatted runs to heal placeholders that Word split across several `<w:r>` elements. It did so by rebuilding the paragraph's run list and copying each run's properties across through the `domain.Run` interface — which cannot carry a run's image. Any run holding a picture was therefore dropped the moment consolidation touched its paragraph: the `<w:drawing>` vanished from `word/document.xml` and the `word/media/` part was orphaned. This reached further than mail merge: `FindPlaceholders` reads as a pure query but consolidates internally, so **scanning a document for placeholders could delete an image from it**, and because `walkParagraphs` covers headers and footers, a merge touching only the body could remove a header logo. Consolidation no longer rebuilds runs at all — each merge group's own first run receives the combined text and the runs it absorbed are removed, so nothing is copied and nothing can be lost. Documents already damaged cannot be recovered from the saved file and must be regenerated.
- **Generated documents now state their own default paragraph spacing.** `word/styles.xml` carried an empty `<w:pPrDefault/>`, so a paragraph that set no spacing inherited nothing and Word applied its own defaults (8pt after, 1.15 line) instead of the 0pt/single the domain model specifies. The same gap made an explicit `SpacingAfter(0)` indistinguishable from never setting it. `w:docDefaults` now carries a real `w:pPrDefault` of `before=0, after=0, line=240, lineRule=auto`, in both the style-manager path and the no-style-manager fallback. Reported and diagnosed by **@b52es** in [#63](https://github.com/mmonterroca/docxgo/pull/63); the fix sets document defaults rather than forcing direct formatting onto every paragraph, which would have overridden paragraph styles in the OOXML cascade and stripped the spacing from every `Heading`.
- **Exact and at-least line spacing at 240 twips is now emitted.** `SetLineSpacing({Rule: Exact, Value: 240})` and `{Rule: AtLeast, Value: 240}` produced no `<w:spacing>` element, because the emit decision looked only at the value and 240 is the default. With document defaults now in place such a paragraph would have inherited `lineRule="auto"`, silently converting a caller's exact 12pt line height to automatic spacing. A non-auto rule now counts as a departure from the defaults regardless of its value.

### Changed

- **Generated output renders more compactly.** Documents produced without explicit spacing previously inherited Word's own defaults (8pt after each paragraph, 1.15 line spacing) because docxgo stated none; they now render at the 0pt / single spacing the API has always claimed. This is the point of the spacing fix, but it is visible in every such document, and it is why this is a minor release rather than a patch. To keep the previous look, set spacing explicitly — `para.SetSpacingAfter(160)` and `para.SetLineSpacing(domain.LineSpacing{Rule: domain.LineSpacingAuto, Value: 276})` — or apply it once through a style. Documents opened from disk are unaffected: `OpenDocument` preserves the source `word/styles.xml` verbatim for round-trip fidelity, so a `.docx` produced by v2.7.2 or earlier must be regenerated rather than round-tripped to pick up the new defaults.
- **The `dev` integration branch has been retired; contributions now branch from and target `master`.** `dev` was introduced as a Git Flow integration branch, but since ~v2.4 its only traffic had been `master → dev` sync merges — it had stopped being a real integration point while still being the branch `CONTRIBUTING.md` told contributors to use. That cost real contributions: [#35](https://github.com/mmonterroca/docxgo/pull/35) targeted `dev`, went stale there, and had to be re-shipped through `master` as [#39](https://github.com/mmonterroca/docxgo/pull/39); [#57](https://github.com/mmonterroca/docxgo/pull/57) was closed; and the author of [#64](https://github.com/mmonterroca/docxgo/pull/64) declined to target `dev` at all because it was behind `master` and doing so would have pulled unrelated history into the diff. Releases are cut by tagging `master`, so the branch was never the release gate either. Short-lived `integration/<topic>` branches remain available for the rare case where two in-flight changes must be co-staged before either lands.

### Known limitations

- An explicit `SetSpacingBefore(0)`/`SetSpacingAfter(0)` is still discarded on a paragraph carrying a style with non-zero spacing: the domain model cannot distinguish "never set" from "explicitly set to zero", so the attribute is omitted and the style's value wins. This release fixes the unstyled case; the styled case needs unset-tracking in the domain model and is tracked in [#69](https://github.com/mmonterroca/docxgo/issues/69). `Indentation` has the same limitation.
- `FindPlaceholders` still mutates the document it scans by consolidating runs. This release removes the damage that mutation could do, but not the mutation itself — tracked in [#68](https://github.com/mmonterroca/docxgo/issues/68).

### Compatibility

- No public Go, CLI protocol, or Node.js API changed — no signature, interface, or method was added, removed, or altered. `ConsolidateRuns`, `MergeTemplate`, and `FindPlaceholders` keep their behavior for every `domain.Paragraph` implementation, including third-party and wrapped ones. Generated document output changes as described under **Changed**.

---

## v2.7.2 — 2026-07-18

### Compliance

- **Definitive MIT provenance determination.** `docs/PROVENANCE_AUDIT.md` now records the reproducible conclusion that the current v2 implementation may remain MIT: AGPL applies to the historical upstream snapshots in Git history, not to the current release tree. No project-wide AGPL relicense or further licensing consultation is an open release task.
- **Corrected copyright notices.** The root `LICENSE` preserves the exact MIT-era predecessor notices without implying that fumiama's later AGPL-era work was licensed as MIT. `CREDITS.md`, the README, package godoc, and design documentation now use the same determination.
- **License-complete artifacts.** Platform npm packages and GitHub binary archives now include the root MIT license text. This repairs the notice omission identified in the published v2.7.1 artifacts.

### Fixed

- **Release metadata and documentation** now consistently report v2.7.2 across the Go version constant, CLI examples, API/design/status guides, and npm package metadata.
- **Documentation import paths and branding** now use the public `/v2` module path and the `docxgo` project name.
- **Release automation comments and actions** now match the actual PAT-triggered npm publication path, and Go setup actions are aligned on `actions/setup-go@v7`.

### Compatibility

- No public Go, CLI protocol, or Node.js API changed. This is a backward-compatible compliance, packaging, and documentation patch.

---

## v2.7.1 — 2026-07-18

### Added

- **The two Tech themes are now discoverable through the public API.** `TechPresentation` and `TechDarkMode` existed but were not resolvable by name; they are now returned by `themes.AllThemes()` (which now yields **7** themes), `themes.ThemeNames()`, and `themes.GetTheme()`, and are accepted by the CLI's theme option and the Node.js `ThemeName` type (PR #56). Consumers that relied on `AllThemes()` returning exactly five themes should be aware of the new count.
- **`docs/PROVENANCE_AUDIT.md`** — a reproducible code-provenance record for docxgo v2 relative to the AGPL-licensed `fumiama/go-docx` upstream, with an accompanying comparison script (`docs/provenance/compare_line_overlap.py`).

### Fixed

- **Output branding.** Generated documents reported `go-docx/v2` (and a frozen `go-docx v2.0.0`) as the authoring application in `docProps/app.xml`/`core.xml`; they now report `docxgo`.

### Changed

- **Documentation aligned with actual behavior** (no API or generation changes): image support is stated as PNG/JPEG/GIF — the formats that decode end-to-end (see #55 for the previously over-stated set); thread-safety wording clarified (a single `Document` is not thread-safe, the internal managers are); the package godoc version string corrected; and the public README prose tightened. The license/provenance narrative across README, CREDITS, and the godoc now defers to `docs/PROVENANCE_AUDIT.md` as the single source of truth.

---

## v2.7.0 — 2026-07-17

### Added

- **Command-line interface (`cmd/docxgo`)** — a JSON-RPC binary that exposes docxgo over stdin/stdout, so documents can be created and manipulated from any language (Node.js, Python, shell, AWS Lambda), on any platform (Linux/macOS/Windows, arm64/x64), with zero config, ports, or auth (closes #19, PR #24)
  - Two modes: `exec` (one-shot, ideal for `child_process`) and `rpc` (persistent newline-delimited JSON session, ideal for batch/Lambda usage)
  - 22 methods across the `system.*`, `document.*`, `paragraph.*`, `table.*`, `section.*`, and `template.*` namespaces, including `document.applyPatch` for multi-operation mutation and `system.batch` for pipelining
  - File **and** base64/buffer I/O in both directions, so binaries never need to touch the filesystem
  - Full protocol reference in [docs/CLI_GUIDE.md](docs/CLI_GUIDE.md)
- **Node.js wrapper — `@mmonterroca/docxgo`** — a TypeScript package (CommonJS + ESM) wrapping the CLI binary with three API levels: `DocxgoExec` (synchronous one-shot), `DocxgoRPC` (low-level persistent client), and `DocumentBuilder` (high-level fluent API). Ships full type definitions and resolves a platform-specific binary via `optionalDependencies`. See [npm/README.md](npm/README.md)
- **`document.setLanguage` RPC method** — exposes v2.6.0's `WithLanguage`/`WithLanguageEx` through the CLI and npm wrapper. Also available as a `setLanguage` patch operation, and the current language is reported by `document.inspect`. Honors the same round-trip guard as `Document.SetLanguage` (works on documents created via `document.create`, not ones opened via `document.open`)
- **Release automation** — `.github/workflows/release.yml` builds multi-platform binaries and publishes a GitHub Release on `v*` tags; `.github/workflows/npm-publish.yml` publishes the platform packages and the main npm package with OIDC provenance when a Release is published

### Fixed

- **`docx.Version`** now reports the correct version. It was stale at `2.5.0` (the v2.6.0 release did not bump it), so `docxgo version` and `system.version` reported the wrong value
- **`template.ConsolidateRuns`** now returns an error and stops at the first run-setter failure instead of silently leaving a paragraph partially rebuilt; `MergeTemplate` and `FindPlaceholders` propagate it
- **`document.applyPatch`** error responses now include an `applied` count, so callers can tell how many operations succeeded before a mid-sequence failure. `applyPatch` is documented as **not** atomic — there is no rollback
- **`template.render`** no longer reports `"error"`-severity findings in an otherwise-successful (`ok: true`) response; in non-strict mode all findings that reach a successful response are labeled `"warning"`
- **npm `DocumentBuilder.create()`/`createToFile()`** now track the new document's ID, so `applyPatch`/`inspect`/`saveToBuffer`/etc. can be chained directly after creating a document — no save-and-reopen round-trip, which would otherwise trip `setLanguage`'s round-trip guard
- **npm `DocxgoRPC.close()`/`kill()`** now mark the client closed synchronously, closing a race window where a call issued during shutdown could hang

---

## v2.6.0 — 2026-07-17

### Added

- **`WithLanguage` / `WithLanguageEx`** — set the document's default proofing language, used by Word for spell-checking, grammar-checking, and hyphenation (closes #44)
  - `WithLanguage(lang string)` sets the primary language (BCP 47 tag, e.g. `"es-MX"`)
  - `WithLanguageEx(docx.Language{Val, EastAsia, Bidi})` additionally sets East Asian (CJK) and right-to-left (bidi) script languages; at least one of the three must be non-empty
  - Written as `w:lang` in `word/styles.xml`'s `docDefaults/rPrDefault/rPr` and as `w:themeFontLang` in `word/settings.xml`
  - `NewDocumentBuilder` always starts from a new document, so `WithLanguage`/`WithLanguageEx` always take effect
  - Opening an existing `.docx` via `OpenDocument`/`OpenDocumentFromBytes`/`OpenDocumentFromReader` now hydrates `Language()` from its `styles.xml`, if it declares one
  - `domain.Document` gained `SetLanguage(*Language) error` and `Language() *Language`; `domain.Language` is the new public type (aliased as `docx.Language`). `Language()` returns a defensive copy — mutating it has no effect on the document
  - `SetLanguage` returns an error (rather than silently no-op) on a document opened via `OpenDocument` whose `styles.xml`/`settings.xml` were preserved verbatim for round-trip fidelity, since a language set there could never actually reach the saved file. Use `WithLanguage`/`WithLanguageEx` when building a new document instead

### Changed

- **`domain.Document` interface gained two methods** (`SetLanguage`, `Language`, above). If you implement `domain.Document` directly with your own type — most commonly a hand-written test double — you'll need to add both methods; embedding `domain.Document` in your type is unaffected, since the new methods are promoted automatically. No exported docxgo API accepts a `domain.Document` from a caller (only returns one), so this is consistent with `SetBackgroundColor`/`BackgroundColor`, which were added to this same interface in the `v2.0.1` patch release.

### Fixed

- **`word/styles.xml` element order** — `internal/xml.Styles` now marshals `w:docDefaults` before `w:latentStyles`, matching the `CT_Styles` schema. This was previously backwards but unobservable, since `DocDefaults` was never populated before this change.
- **CI never ran against `master`** — `.github/workflows/ci.yml` triggered on `branches: [ main, dev ]`, but this repo's default/release branch is `master`, not `main`. The Lint/Build/Test job had silently never fired for any push or PR targeting `master` since the workflow was added. Fixed to trigger on `master` and `dev`.

---

## v2.5.0 — 2026-07-04

### Added

- **Run-level formatting on `CellBuilder`** — table cells now support the full fluent formatting set, matching `ParagraphBuilder` (PR #39; `Italic`/`Color`/`FontSize` contributed by @SlashLight in #35)
  - `CellBuilder.Italic()` — italicize the last run in the last paragraph of the cell
  - `CellBuilder.Color(color)` — set the last run's text color
  - `CellBuilder.FontSize(points)` — set the last run's font size (points, converted to half-points internally)
  - `CellBuilder.Underline(style)` — set the last run's underline style
  - `lastRun()` helper extracted and reused by `Bold()`; existing `Bold()` behavior and error messages are unchanged (backward-compatible)

### Fixed

- **Per-part relationship ID resolution for headers/footers** (PR #40, closes #37)
  - Relationship IDs in OOXML are scoped per-part: a header's own `word/_rels/header1.xml.rels` may reuse an ID (e.g. `rId1`) that also exists in `word/_rels/document.xml.rels` for something unrelated. Drawings inside headers/footers now resolve their `r:embed` against that part's own `.rels`, falling back to the document-wide map — fixing wrong or missing media in headers and footers
  - `internal/reader/parser.go` parses each header/footer part's own `.rels` into a `PartRelationships` map on `ParsedPackage`; an unparseable per-part `.rels` is skipped rather than aborting the whole document open
  - `internal/reader/reconstruct.go` adds an `activeRelationships` scope to `reconstructContext`; header/footer hydration shares a single `hydratePartParagraphs` helper

### Tests

- `TestCellBuilder_RunFormatting` / `TestCellBuilder_RunFormattingErrors` — happy-path and error coverage for `Italic`/`Color`/`FontSize`/`Underline` on cells
- `TestOpenDocument_CollidingRelationshipIDsAcrossParts` — a header image whose `rId1` collides with an unrelated document-level `rId1` resolves to the header's own media (plus a save/reopen round-trip guard)
- `TestOpenDocument_MalformedHeaderRelsIsTolerated` — a corrupt per-part `.rels` no longer fails the open

---

## v2.4.0 — 2026-04-30

### Added

- **In-memory image API** (`pkg/builder` + `domain.Paragraph`) — insert images from byte slices without touching the file system (PR #30, closes #29)
  - `ParagraphBuilder.AddImageFromBytes(data, format)` — inline image from bytes
  - `ParagraphBuilder.AddImageFromBytesWithSize(data, format, size)` — with custom dimensions
  - `ParagraphBuilder.AddImageFromBytesWithPosition(data, format, size, pos)` — floating with positioning
  - Matching methods on `domain.Paragraph` (`AddImageFromBytes`, `AddImageFromBytesWithSize`, `AddImageFromBytesWithPosition`)
  - New `internal/core` constructors: `NewImageFromBytes`, `NewImageFromBytesWithSize`, `NewImageFromBytesWithPosition`
  - Format normalization (`JPG` → `jpeg`, leading `.` trimmed) and validation against the supported set
  - Defensive copy of the input byte slice so callers can safely reuse buffers
- `examples/08_images` updated to demonstrate the new in-memory image flow

### Fixed

- **Round-trip preservation of `w:gridSpan` and `w:vMerge`** for merged table cells (PR #26, closes #25)
  - `hydrateTableCell` now parses `<w:tcPr>` and applies horizontal merges through `cell.Merge(span, 1)` so spanned-over cells are correctly marked as `IsHorizontallyMergedContinuation()`
  - Restores `w:vMerge` (`restart` / `continue`) onto reconstructed `domain.TableCell`s
  - `hydrateTable` tracks `colOffset` and recomputes `maxCols` from gridSpan sums so XML cells map to the correct grid columns
  - Numeric parse errors wrapped with `errors.WrapWithContext` (attribute + raw value) for clearer diagnostics

### Changed

- CLI handler (`cmd/docxgo/handlers.go`) `applyImage()` now uses `AddImageFromBytes*` directly for base64 images, eliminating the temp-file write/read round-trip

### Tests

- `TestGridSpanPreservedAfterRoundTrip` — verifies horizontal merge survives save + reopen and that continuation cells are flagged correctly
- `TestVMergePreservedAfterRoundTrip` — verifies vertical merge `restart` / `continue` survives save + reopen
- Unit tests for all three `NewImageFromBytes*` constructors (valid data, empty data, empty/invalid format)
- Builder tests for all three `AddImageFromBytes*` methods (error path validation)

---

## v2.3.0 — 2026-02-27

### Added

- **Template / Mail Merge** (`pkg/template/`) — new package for document template processing
  - `MergeTemplate()` — replace `{{placeholder}}` tokens with data values, preserving formatting
  - `FindPlaceholders()` / `PlaceholderNames()` — detect all placeholders in body, tables, headers, footers
  - `ValidateTemplate()` — check for missing/unused data keys before merging
  - `ConsolidateRuns()` — merge adjacent runs with identical formatting to heal split placeholders
  - Custom delimiter support (e.g., `${key}`, `«key»` instead of `{{key}}`)
  - Strict mode for missing key error reporting
  - Batch merge support (reopen template for each record)
  - Full MERGEFIELD compatibility with real Microsoft Word templates
- `Paragraph.ClearRuns()`, `Paragraph.RemoveRun(index)`, `Paragraph.InsertRunAt(index)` — paragraph mutation APIs
- `Run.Fields()`, `Run.Breaks()`, `Run.Image()` — run content inspection methods
- `Run.ClearFields()` — clear field instructions from a run
- Example 14: Mail merge invoice template with batch generation
- Example 15: External Word template with MERGEFIELDs and «» delimiters

### Fixed

- `walkParagraphs()` auto-creating phantom headers/footers causing "Word found unreadable content" errors
- MERGEFIELD text duplication after merge due to field instructions persisting in preceding runs

---

## v2.2.2 — 2026-02-26

### Fixed

- Table style borders not rendering in Word: styles like `TableGrid` now emit proper `w:tblPr > w:tblBorders` in `styles.xml` (#15)
- Grid column width calculation: derive widths from first row cells instead of emitting `w:w="0"`

### Added

- `TableStyleDef` interface for table-specific style properties (borders, cell margins)
- `TableLevelBorders` struct with `InsideH`/`InsideV` for inner grid borders
- Example 13_themes restored with full v2 API (7 preset themes)

### Changed

- Examples 01_basic and 02_intermediate now use `Style(domain.TableStyleGrid)` for visible table borders
- Comprehensive documentation overhaul across all docs files (#14)
- CHANGELOG.md now contains complete version history (v2.0.0-beta through v2.2.2)

---

## v2.2.1 — 2026-01-22

### Fixed

- Hyperlink RelationshipID preservation during round-trip (read -> modify -> save)
- Drawing position serialization: fixed `wp:align` empty value error for floating images; added `UseOffsetX`/`UseOffsetY` flags
- Internal hyperlink anchors: support for anchor-based bookmarks and TOC links
- Custom styles preservation: all 55+ custom styles from templates maintained during round-trip

### Added

- Troubleshooting documentation (`docs/TROUBLESHOOTING_DOCX_VALIDATION.md`)

---

## v2.2.0 — 2026-01-06

### Fixed

- Table cell border serialization: ensure each side serializes with correct style ("single"), width (Sz=4), and color ("FF0000" hex)

### Tests

- Added `TestTableSerializer_CellBorders` validating per-side style, width, and color

### Acknowledgements

- Original fix contributed by @g-mero; the historical PR is no longer available
- Validation & extension: [PR #4](https://github.com/mmonterroca/docxgo/pull/4) by @Copilot

---

## v2.1.1 — 2025-11-04

### Fixed

- Go module path: re-applied `/v2` suffix to `go.mod` (regression from v2.1.0)

---

## v2.1.0 — 2025-10-31

### Added

- Theme system with 7 preset themes (Corporate, Startup, Modern, Fintech, Academic, TechPresentation, TechDarkMode)
- Theme colors, fonts, and spacing configuration
- Custom theme support (clone, customize, create from scratch)
- `WithTheme()` builder option for one-call theme application
- Example 13: themes showcase

### Fixed

- Style preservation when applying themes
- Font inheritance in themed documents
- Color serialization in OOXML

---

## v2.0.1 — 2025-11-04

### Fixed

- Go module path: added required `/v2` suffix to `go.mod` (`module github.com/mmonterroca/docxgo` -> `module github.com/mmonterroca/docxgo/v2`)

---

## v2.0.0 — 2025-10-29

### Added

- Document reading: `docx.OpenDocument()` to open existing .docx files
- Read document structure (paragraphs, runs, tables, styles)
- Modify existing content (edit text, change formatting, update table cells)
- Round-trip capability (Create -> Save -> Open -> Modify -> Save)
- Example 12: read and modify documents

### Fixed

- Style preservation (Title, Subtitle, Heading1-9, Quote, Normal, ListParagraph) during round-trip
- README API examples corrected for accurate signatures

---

## v2.0.0-beta — 2025-10-28

### Added

- Complete rewrite of go-docx with domain-driven clean architecture
- Builder pattern API with fluent chaining and error accumulation
- Direct domain API for programmatic control
- Section management with independent page settings
- Headers and footers (per-section, first-page, odd/even variants)
- Page layout (sizes, orientations, margins, columns)
- Fields support (TOC, page numbers, hyperlinks, dates, 9 field types)
- Advanced tables (cell merging, vertical alignment, shading, borders, 8 styles)
- Inline and floating images with precise positioning (9 formats)
- 40+ built-in Word styles
- Rich text formatting (bold, italic, underline, colors, fonts, sizes, highlights)
- Comprehensive error types with context and validation
- 11 working examples covering all major features
- Complete API documentation and migration guide from v1.x

### Fixed

- Character encoding issues in XML serialization
- Relationship ID management for images and hyperlinks
- Section break serialization order
- Table cell border rendering
