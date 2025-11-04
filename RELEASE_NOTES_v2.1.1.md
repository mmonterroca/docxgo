# Release Notes - v2.1.1

**Release Date:** January 2025

## Critical Bug Fix

This is a patch release that fixes a critical Go module compatibility issue in v2.1.0.

### Fixed

- **Go Module Path**: Added required `/v2` suffix to module declaration in `go.mod`
  - Previous: `module github.com/mmonterroca/docxgo`
  - Fixed: `module github.com/mmonterroca/docxgo/v2`
  - This is required by Go's semantic versioning for v2+ major versions

### Impact

Version v2.1.0 could not be imported correctly due to the missing `/v2` suffix. Users attempting to run:

```bash
go get github.com/mmonterroca/docxgo/v2@v2.1.0
```

Would encounter module path mismatch errors. This release resolves that issue.

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.1.1
```

## Usage

All imports now use the `/v2` path:

```go
import (
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/domain"
    "github.com/mmonterroca/docxgo/v2/themes"
)
```

## What's in v2.1.x

This release includes all features from v2.1.0:

### Theme System

A comprehensive theming system with 6 pre-built themes:

- **DefaultTheme**: Professional balanced design
- **ModernLight**: Clean, spacious contemporary look
- **TechPresentation**: Tech-focused with strong visual hierarchy
- **TechDarkMode**: Dark background with bright accents
- **AcademicFormal**: Traditional academic style
- **MinimalistClean**: Ultra-minimal design

See [RELEASE_NOTES_v2.1.0.md](RELEASE_NOTES_v2.1.0.md) for complete feature details.

## Notes

- This release contains the same features as v2.1.0
- Only the module path declaration has been fixed
- All tests passing
- No API changes

## Upgrading from v2.1.0

If you were using v2.1.0 (which was broken), simply update your dependencies:

```bash
go get github.com/mmonterroca/docxgo/v2@v2.1.1
```

No code changes required - just update the version in your `go.mod`.

## Upgrading from v2.0.x

The theme system is fully backward compatible. Existing code continues to work without changes. To use themes:

```go
import "github.com/mmonterroca/docxgo/v2/themes"

theme := themes.ModernLight()
doc := docx.NewDocument(docx.WithTheme(theme))
```

---

**Full Changelog**: See [CHANGELOG.md](CHANGELOG.md) for complete version history.
