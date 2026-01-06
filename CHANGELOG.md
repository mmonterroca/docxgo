## v2.2.0 — 2026-01-06

### Fixed
- Table cell border serialization: ensure each side serializes with correct style ("single"), width (Sz=4), and color ("FF0000" hex).

### Tests
- Add TestTableSerializer_CellBorders validating per-side style, width, and color.

### Acknowledgements
- Original fix: [PR #3](https://github.com/mmonterroca/docxgo/pull/3) by @g-mero
- Validation & extension: [PR #4](https://github.com/mmonterroca/docxgo/pull/4) by @Copilot