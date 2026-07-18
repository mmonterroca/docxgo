# Release Notes - v2.7.0

**Release Date:** July 2026

## Summary

v2.7.0 makes docxgo usable from **any language, not just Go**. It adds a JSON-RPC command-line interface (`cmd/docxgo`) that speaks over stdin/stdout, and a fully-typed Node.js wrapper (`@mmonterroca/docxgo`) built on top of it — so you can create and manipulate Word documents from Node.js, AWS Lambda, Python, or a shell script, on any platform (Linux/macOS/Windows, arm64/x64), with no config, ports, or auth.

It also surfaces v2.6.0's proofing-language support through a new `document.setLanguage` method, and ships the CI/CD automation that builds and publishes the multi-platform binaries and npm packages.

This release closes [#19](https://github.com/mmonterroca/docxgo/issues/19) (PR [#24](https://github.com/mmonterroca/docxgo/pull/24)).

## New Features

### Command-line interface — `cmd/docxgo` (closes #19)

A single self-contained binary exposes docxgo as JSON-RPC over stdin/stdout. No server, no ports, no auth — you pipe a request in and read a response out.

- **Two modes:**
  - `docxgo exec --request '<JSON>'` — one-shot execution, ideal for `child_process` calls and CI scripts.
  - `docxgo rpc` — a persistent session that reads newline-delimited JSON requests from stdin and writes responses to stdout, ideal for high-frequency batch work and AWS Lambda warm starts.
- **22 methods** across the `system.*`, `document.*`, `paragraph.*`, `table.*`, `section.*`, and `template.*` namespaces — including `document.applyPatch` for multi-operation mutation and `system.batch` for pipelining several calls in one round-trip.
- **File and base64/buffer I/O in both directions**, so a Node or Lambda caller never has to touch the filesystem.

```bash
# One-shot: create a document and get it back as base64
docxgo exec --request '{
  "id": 1,
  "method": "document.create",
  "params": {
    "content": [{ "type": "paragraph", "runs": [{ "text": "Hello!" }] }],
    "output": "buffer"
  }
}'
```

Full protocol reference: [docs/CLI_GUIDE.md](docs/CLI_GUIDE.md).

### Node.js wrapper — `@mmonterroca/docxgo`

A TypeScript package (CommonJS + ESM) that wraps the CLI binary with three API levels, from lowest to highest:

- `DocxgoExec` — synchronous one-shot client (spawns the binary per call).
- `DocxgoRPC` — low-level persistent client over the NDJSON protocol.
- `DocumentBuilder` — high-level fluent API that mirrors the Go builder.

It ships full TypeScript type definitions and resolves the correct platform binary automatically via `optionalDependencies`.

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

Full guide and API reference: [npm/README.md](npm/README.md).

### `document.setLanguage`

Exposes v2.6.0's `WithLanguage` / `WithLanguageEx` through the CLI and npm wrapper, so you can set a document's default proofing language (BCP 47 tags such as `es-MX`) from Node or the shell. It is also available as a `setLanguage` patch operation, and the current language is reported by `document.inspect`.

It honors the same round-trip guard as `Document.SetLanguage`: it works on documents created via `document.create` (or `DocumentBuilder.create()`), and is rejected on documents opened via `document.open`, whose `styles.xml`/`settings.xml` are preserved verbatim.

### Release & publish automation

- `.github/workflows/release.yml` — builds binaries for darwin/linux/windows × arm64/x64, generates checksums, and publishes a GitHub Release on any `v*` tag.
- `.github/workflows/npm-publish.yml` — publishes the five platform packages and the main `@mmonterroca/docxgo` package with npm OIDC provenance when a GitHub Release is published.

## Bug Fixes

- **`docx.Version`** now reports the correct version. It was stale at `2.5.0` (the v2.6.0 release did not bump it), so `docxgo version` and `system.version` reported the wrong value.
- **`template.ConsolidateRuns`** now returns an error and stops at the first run-setter failure instead of silently leaving a paragraph partially rebuilt; `MergeTemplate` and `FindPlaceholders` propagate it.
- **`document.applyPatch`** error responses now include an `applied` count, so callers can tell how many operations succeeded before a mid-sequence failure. `applyPatch` is documented as **not** atomic — there is no rollback.
- **`template.render`** no longer reports `"error"`-severity findings in an otherwise-successful (`ok: true`) response; in non-strict mode every finding that reaches a successful response is labeled `"warning"`.
- **npm `DocumentBuilder.create()` / `createToFile()`** now track the new document's ID, so `applyPatch`/`inspect`/`saveToBuffer`/etc. can be chained directly after creating a document — no save-and-reopen round-trip, which would otherwise trip `setLanguage`'s round-trip guard.
- **npm `DocxgoRPC.close()` / `kill()`** now mark the client closed synchronously, closing a race window where a call issued during shutdown could hang.

## Installation

**Go library** (unchanged):

```bash
go get github.com/mmonterroca/docxgo/v2@v2.7.0
```

**Node.js / CLI**:

```bash
npm install @mmonterroca/docxgo
```

## Compatibility

- **Purely additive to the Go library.** The CLI (`cmd/docxgo`) and the npm package are brand-new surfaces; no existing Go API changed. There are no `domain` interface changes beyond the `SetLanguage`/`Language` methods already added in v2.6.0.
- **Documents written with earlier versions are unaffected.**
- The Node.js package targets Node 16+ and ships prebuilt binaries for Linux, macOS, and Windows (x64 and arm64).

## Acknowledgements

- PR [#24](https://github.com/mmonterroca/docxgo/pull/24) (`dev`) — CLI JSON-RPC wrapper + Node.js integration.

## Related Issues

- Closes [#19](https://github.com/mmonterroca/docxgo/issues/19) — CLI JSON-RPC wrapper for docxgo and Node.js integration.
