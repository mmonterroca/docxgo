# Release Notes - v2.1.0

## 🎨 go-docx v2.1.0 - Themes & Advanced Styling

**Release Date**: October 31, 2025

We're excited to announce **v2.1.0**, bringing powerful **theme support** and **advanced styling capabilities** to go-docx!

---

## 🆕 What's New

### 🎨 Theme System

Complete theme infrastructure for consistent document styling:

- ✅ **Pre-built Themes** - 6 professional themes ready to use
- ✅ **Theme Colors** - Comprehensive color palettes (Primary, Secondary, Accent, Background, Text, etc.)
- ✅ **Theme Fonts** - Font families for body, headings, and monospace
- ✅ **Theme Spacing** - Configurable spacing for paragraphs, headings, and sections
- ✅ **Custom Themes** - Clone and customize existing themes or create from scratch
- ✅ **Theme Application** - Apply themes globally to documents with one method call

**Available Themes:**

| Theme | Description | Best For |
|-------|-------------|----------|
| `DefaultTheme` | Classic professional look | Business documents, reports |
| `ModernLight` | Clean, contemporary design | Marketing materials, proposals |
| `TechPresentation` | Tech-focused with blue accents | Technical documentation, specs |
| `TechDarkMode` | Dark theme for technical docs | Developer documentation, coding guides |
| `AcademicFormal` | Traditional academic styling | Research papers, academic reports |
| `MinimalistClean` | Minimal, distraction-free | Focus documents, minimalist designs |

**Example: Using Themes**

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/domain"
    "github.com/mmonterroca/docxgo/themes"
)

func main() {
    // Create document with theme
    doc := docx.NewDocument()
    themes.TechPresentation.ApplyTo(doc)
    
    // Get theme colors and fonts
    colors := themes.TechPresentation.Colors()
    fonts := themes.TechPresentation.Fonts()
    
    // Use theme settings
    title, _ := doc.AddParagraph()
    title.SetStyle(domain.StyleIDHeading1)
    titleRun, _ := title.AddRun()
    titleRun.AddText("Technical Architecture Document")
    
    body, _ := doc.AddParagraph()
    bodyRun, _ := body.AddRun()
    bodyRun.AddText("This document uses the Tech Presentation theme")
    bodyRun.SetFont(domain.Font{Name: fonts.Body})
    bodyRun.SetColor(colors.Text)
    
    if err := doc.SaveAs("themed.docx"); err != nil {
        log.Fatal(err)
    }
}
```

**Example: Custom Theme**

```go
// Clone and customize existing theme
customTheme := themes.ModernLight.Clone()
customColors := themes.ThemeColors{
    Primary:    domain.Color{R: 200, G: 50, B: 50},  // Custom red
    Secondary:  domain.Color{R: 50, G: 50, B: 200},  // Custom blue
    Background: domain.Color{R: 255, G: 255, B: 255},
    Text:       domain.Color{R: 33, G: 33, B: 33},
    // ... other colors
}
customTheme = customTheme.WithColors(customColors)
customTheme.ApplyTo(doc)
```

### 📐 Advanced Examples

#### Example 13: Technical Architecture Documents

New comprehensive example demonstrating:
- Theme application (Light & Dark modes)
- PlantUML diagram integration
- Code blocks with syntax highlighting
- Professional tables with tech styling
- Architecture decision records
- Multiple sections with consistent branding

See [`examples/13_themes/04_tech_architecture/`](examples/13_themes/04_tech_architecture/) for complete implementation.

**Features:**
- Generates both light and dark mode versions
- Integrates with PlantUML server for UML diagrams (Class, Sequence, Component)
- Styled code blocks with language indicators
- Technology comparison tables
- Cover page with metadata
- Headers and footers with theme colors

---

## 🔧 Improvements

### Style Management
- Enhanced built-in style library (40+ styles)
- Improved style inheritance and customization
- Better style serialization and persistence

### Error Handling
- More descriptive error messages for theme operations
- Validation of theme color values
- Clear feedback on theme application failures

### Documentation
- Complete theme system documentation
- New example showcasing all theme capabilities
- Updated API documentation with theme methods

---

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.1.0
```

---

## 🚀 Migration from v2.0.0

v2.1.0 is **fully backward compatible** with v2.0.0. No breaking changes.

### New APIs Added

```go
// Theme package
import "github.com/mmonterroca/docxgo/themes"

// Pre-built themes
themes.DefaultTheme
themes.ModernLight
themes.TechPresentation
themes.TechDarkMode
themes.AcademicFormal
themes.MinimalistClean

// Theme application
theme.ApplyTo(doc)

// Theme properties
colors := theme.Colors()
fonts := theme.Fonts()
spacing := theme.Spacing()

// Customization
customTheme := theme.Clone()
customTheme = theme.WithColors(newColors)
customTheme = theme.WithFonts(newFonts)
customTheme = theme.WithSpacing(newSpacing)
```

---

## 🐛 Bug Fixes

- Fixed style preservation when applying themes
- Corrected font inheritance in themed documents
- Improved color serialization in OOXML
- Fixed spacing inconsistencies in themed headings

---

## 📊 Examples

All examples from v2.0.0 continue to work, plus:

| Example | Description | Key Features |
|---------|-------------|--------------|
| `13_themes/04_tech_architecture` | Technical documentation | Themes, PlantUML, code blocks, tables |

### Running the New Example

```bash
cd examples/13_themes/04_tech_architecture
go run main.go
```

Generates:
- `tech_architecture_light.docx` - Light mode technical document
- `tech_architecture_dark.docx` - Dark mode technical document

---

## 🔗 Compatibility

- **Go Version**: 1.23+
- **OOXML**: Office Open XML (ISO/IEC 29500)
- **Microsoft Word**: 2007+ (Windows/Mac)
- **LibreOffice**: 6.0+ (all platforms)
- **Google Docs**: Full compatibility
- **Operating Systems**: Linux, macOS, Windows

---

## 📚 Documentation

### New Documentation
- **[themes/README.md](themes/README.md)** - Complete theme system guide
- **[examples/13_themes/README.md](examples/13_themes/README.md)** - Theme example documentation

### Updated Documentation
- **[README.md](README.md)** - Added theme examples
- **[docs/V2_API_GUIDE.md](docs/V2_API_GUIDE.md)** - Theme API reference

---

## 🗺️ Roadmap

### v2.2.0 (Q1 2026) - Enhanced Reading
- Complete Phase 10 (Document Reading to 100%)
- Read headers/footers from existing documents
- Read images and complex tables
- Comments and change tracking

### v2.3.0 (Q2 2026) - Advanced Content
- Custom XML parts
- Advanced drawing shapes
- Enhanced image manipulation
- Content controls

See [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for detailed roadmap.

---

## 🙏 Credits

**v2.1.0 Theme System**: [@mmonterroca](https://github.com/mmonterroca)

**Contributors**: See [CONTRIBUTORS](CONTRIBUTORS) file

---

## 📄 License

**MIT License** - See [LICENSE](LICENSE) file for details.

---

## 🎉 Thank You!

Thank you for using go-docx v2.1.0! We're excited to see the beautiful themed documents you create.

If you find this library useful, please:
- ⭐ Star the repository on GitHub
- 📣 Share with your colleagues
- 🐛 Report issues you encounter
- 💡 Suggest features you'd like to see

Happy theming! 🎨
