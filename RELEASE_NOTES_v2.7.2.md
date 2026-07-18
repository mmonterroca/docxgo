# Release Notes - v2.7.2

**Release Date:** July 18, 2026

## Summary

v2.7.2 closes the docxgo v2 license and provenance review with a definitive
MIT determination and makes every new binary package carry the required MIT
notice. It also aligns release metadata and current documentation with the
actual v2 module, CLI, and npm package.

There are no public API or document-generation changes in this patch.

## License and provenance determination

The completed [`docs/PROVENANCE_AUDIT.md`](docs/PROVENANCE_AUDIT.md) records the
reproducible result:

- docxgo v2.7.2 may be distributed under MIT;
- AGPL-3.0 remains applicable to the historical `fumiama/go-docx` snapshots in
  Git history;
- the current v2 tree contains no protectable implementation copied from the
  AGPL-only period; and
- no project-wide AGPL relicense or further licensing consultation remains as
  a release prerequisite.

The root `LICENSE` now preserves only the exact MIT-era predecessor notices.
Later AGPL-era authorship remains credited in `CREDITS.md` without presenting
that work as MIT-licensed.

## Packaging fixes

- Every platform-specific npm package includes `LICENSE`.
- Every GitHub binary ZIP or tar archive includes `LICENSE`.
- The main npm package remains license-complete through its prepack step.
- npm and Go release metadata consistently report 2.7.2.

## Documentation and automation

- Current guides now use the correct `github.com/mmonterroca/docxgo/v2`
  module path and `docxgo` branding.
- CLI version examples, minimum Go version, release status, and documentation
  indexes now match the release candidate.
- GitHub Actions use `actions/setup-go@v7` consistently.
- npm publication documentation now reflects the existing automatic
  `RELEASE_PAT` release-event path, while retaining the idempotent manual
  fallback.

## Installation

**Go library:**

```bash
go get github.com/mmonterroca/docxgo/v2@v2.7.2
```

**Node.js / CLI:**

```bash
npm install @mmonterroca/docxgo@2.7.2
```

## Compatibility

- Backward-compatible with v2.7.1.
- No Go API signature changed.
- No JSON-RPC method or response shape changed.
- No Node.js API or supported platform changed.
