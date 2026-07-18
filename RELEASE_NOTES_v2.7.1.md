# Release Notes - v2.7.1

**Release Date:** July 2026

## Summary

v2.7.1 is a small maintenance release. It makes the two Tech themes usable by
name across every API surface, fixes the authoring-application name stamped into
generated documents, and aligns the documentation with the library's actual
behavior.

## Improvements

### Tech themes are now discoverable through the public API

`TechPresentation` and `TechDarkMode` were implemented but could only be used by
referencing their package variables directly — they were not returned by the
theme-discovery functions, so looking them up by name failed. They are now fully
wired through:

- `themes.AllThemes()` — now returns **7** themes (the 5 general-purpose presets
  plus the 2 Tech themes).
- `themes.ThemeNames()` and `themes.GetTheme(name)` — `GetTheme("tech-presentation")`
  and `GetTheme("tech-darkmode")` now resolve instead of returning `nil`.
- The CLI / JSON-RPC `theme` option — `"theme": "TechPresentation"` (or
  `"tech-presentation"`) now works.
- The Node.js `ThemeName` type — the two Tech themes are now valid values.

```bash
# Now works end-to-end through the CLI / Node wrapper
docxgo exec --request '{
  "id": 1,
  "method": "document.create",
  "params": {
    "options": { "theme": "TechPresentation" },
    "content": [{ "type": "paragraph", "runs": [{ "text": "Hello!" }] }],
    "output": "buffer"
  }
}'
```

## Bug Fixes

- **Output branding.** Generated documents stamped `go-docx/v2` (and a frozen
  `go-docx v2.0.0`) as the authoring application in `docProps/app.xml` /
  `docProps/core.xml`. They now report `docxgo`.

## Documentation

- **Image formats.** The docs previously listed up to nine image formats; in
  practice only **PNG, JPEG, and GIF** decode end-to-end (the only decoders the
  library registers, with no external dependencies). The documentation now
  states PNG/JPEG/GIF. Broader format support is tracked in
  [#55](https://github.com/mmonterroca/docxgo/issues/55).
- **Thread safety.** Wording clarified: a single `Document` is not thread-safe
  and must be guarded for concurrent access; the internal managers are
  thread-safe (`sync.RWMutex`).
- **Version strings** across the godoc and docs corrected to 2.7.1.
- **README** rewritten for a tighter, professional read, with duplicated
  sections removed. Quick-start snippets and examples are unchanged.

## Installation

**Go library:**

```bash
go get github.com/mmonterroca/docxgo/v2@v2.7.1
```

**Node.js / CLI:**

```bash
npm install @mmonterroca/docxgo
```

## Compatibility

- **Backward-compatible.** No existing Go API signatures changed and no
  generation behavior changed for existing documents.
- The one observable API change is additive: `themes.AllThemes()` now returns
  seven themes instead of five, and the two Tech themes are resolvable by name.
  Callers that assumed `AllThemes()` returns exactly five should be aware.
- The Node.js package continues to target Node 16+ with prebuilt binaries for
  Linux, macOS, and Windows (x64 and arm64).

## Related Issues

- [#55](https://github.com/mmonterroca/docxgo/issues/55) — image format
  validation names more formats than actually decode (tracked for a future
  release).
