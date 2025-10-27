# Work in Progress Examples

These examples are currently being updated to work with the v2 API.

## Status

All examples in this directory use APIs that are either:
1. From the legacy v1 codebase
2. Not yet implemented in v2
3. Need refactoring to match current v2 architecture

## Examples

### 04_fields - Field System (TOC, Page Numbers, Hyperlinks)
**Issues:**
- Uses `docx.New()` (v1 API, should be `docx.NewDocument()`)
- Uses `docx.Document` type (should be `domain.Document`)
- Uses `docx.FooterDefault` (should be `domain.FooterDefault`)
- Uses `docx.ColorBlue` (should be `docx.Blue` or `domain.ColorBlue`)

**Required to fix:**
- Update all API calls to v2
- Test with current field implementation

### 05_styles - Style Management
**Issues:**
- Uses `doc.StyleManager()` method (not implemented in v2)
- Uses `domain.StyleIDFootnoteReference` (doesn't exist)
- Uses `Color` as `int` (should be `domain.Color` struct)

**Required to fix:**
- Implement StyleManager if needed, OR
- Rewrite to use direct style setting via `SetStyle()`
- Convert all int colors to `domain.Color{R, G, B}` structs

### 06_sections - Sections and Page Layout
**Issues:**
- Uses `Color` as hex `int` (should be `domain.Color` struct)

**Required to fix:**
- Convert: `0x4472C4` → `domain.Color{R: 0x44, G: 0x72, B: 0xC4}`
- Verify section API is complete

### 07_advanced - Advanced Document Creation
**Issues:**
- Uses `run.SetFontSize()` (method doesn't exist, should be `SetSize()`)
- Uses `docx.Document` type (should be `domain.Document`)
- Uses `Color` as hex `int` (should be `domain.Color` struct)

**Required to fix:**
- Update all method names to v2 API
- Fix color usage
- Update type references

## How to Update

1. **Check current v2 API**: See `domain/` interfaces
2. **Update imports**: Ensure using `github.com/mmonterroca/docxgo` and `github.com/mmonterroca/docxgo/domain`
3. **Fix API calls**: Match signatures in `domain/document.go`, `domain/paragraph.go`, `domain/run.go`
4. **Test**: Run `go build` in example directory
5. **Move back**: Once working, move from `_wip/` to parent `examples/` directory

## Testing

These examples will NOT be compiled during CI/CD runs because they're in a `_wip/` subdirectory.

To test manually:
```bash
cd 04_fields
go build  # Will show errors that need fixing
```

## Contributing

Want to help update these examples? See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.
