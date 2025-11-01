# Document Themes

The `themes` package provides a comprehensive theme system for go-docx, allowing you to apply consistent visual styling to your documents with a single configuration.

## Overview

Themes define a complete visual style for documents, including:
- **Colors**: Primary, secondary, accent, and semantic colors
- **Fonts**: Body, heading, and monospace font families and sizes
- **Spacing**: Paragraph, line, heading, and section spacing
- **Headings**: Sizes, weights, and color usage for H1-H3

## Quick Start

```go
import (
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/themes"
)

// Apply a preset theme
builder := docx.NewDocumentBuilder(
    docx.WithTheme(themes.Corporate),
)

builder.AddParagraph().
    Text("Professional document with Corporate theme").
    End()

doc, _ := builder.Build()
doc.SaveAs("output.docx")
```

## Preset Themes

### Corporate
Professional business theme with navy blue and red accents.
- **Colors**: Navy Blue (#2F5496), Red (#C00000)
- **Best For**: Business reports, proposals, corporate documentation
- **Font**: Calibri, 11pt body text

### Startup
Energetic, modern theme with vibrant colors.
- **Colors**: Slate Blue (#6A5ACD), Turquoise (#48D1CC), Tomato Red (#FF6347)
- **Best For**: Pitch decks, business plans, innovation reports
- **Font**: Calibri Light for headings, 11pt body text

### Modern
Clean, minimalist theme with contemporary styling.
- **Colors**: Wet Asphalt (#34495E), Peter River Blue (#2980B9)
- **Best For**: Technical documentation, white papers, product specs
- **Font**: Segoe UI, 11pt body text

### Fintech
Professional financial theme with trustworthy blues and greens.
- **Colors**: Dark Cerulean (#005288), Teal (#009688)
- **Best For**: Financial reports, investment documents, banking materials
- **Font**: Arial, 11pt body text

### Academic
Traditional scholarly theme for academic documents.
- **Colors**: Black, Maroon (#800000)
- **Best For**: Research papers, academic reports, thesis documents
- **Font**: Times New Roman, 12pt body text, double-spaced

## Theme API

### Theme Interface

```go
type Theme interface {
    Name() string
    DisplayName() string
    Description() string
    Colors() ThemeColors
    Fonts() ThemeFonts
    Spacing() ThemeSpacing
    Headings() ThemeHeadings
    ApplyTo(doc domain.Document) error
    Clone() Theme
    WithColors(colors ThemeColors) Theme
    WithFonts(fonts ThemeFonts) Theme
    WithSpacing(spacing ThemeSpacing) Theme
}
```

### Customizing Themes

#### Modify Colors

```go
// Clone and customize
customTheme := themes.Corporate.Clone()

colors := customTheme.Colors()
colors.Primary = domain.Color{R: 100, G: 50, B: 150}  // Purple
colors.Accent = domain.Color{R: 255, G: 140, B: 0}    // Orange

customTheme = customTheme.WithColors(colors)

builder := docx.NewDocumentBuilder(docx.WithTheme(customTheme))
```

#### Modify Fonts

```go
customTheme := themes.Modern.Clone()

fonts := customTheme.Fonts()
fonts.Body = "Georgia"
fonts.Heading = "Helvetica Neue"
fonts.BodySize = 24  // 12pt (sizes are in half-points)

customTheme = customTheme.WithFonts(fonts)
```

#### Modify Spacing

```go
customTheme := themes.Academic.Clone()

spacing := customTheme.Spacing()
spacing.ParagraphAfter = 240  // 12pt after paragraphs
spacing.LineSpacing = 360     // 1.5 line spacing

customTheme = customTheme.WithSpacing(spacing)
```

### Apply Theme to Existing Document

```go
// Open existing document
doc, _ := docx.OpenDocument("existing.docx")

// Apply theme
themes.Corporate.ApplyTo(doc)

// Save
doc.SaveAs("restyled.docx")
```

## Creating Custom Themes

Create a completely custom theme from scratch:

```go
customTheme := themes.NewTheme(
    "my-brand",
    "My Brand Theme",
    "Custom theme for company branding",
)

// Define colors
colors := themes.ThemeColors{
    Primary:    domain.Color{R: 0, G: 102, B: 204},    // Brand blue
    Secondary:  domain.Color{R: 51, G: 153, B: 255},   // Light blue
    Accent:     domain.Color{R: 255, G: 102, B: 0},    // Brand orange
    Background: domain.Color{R: 255, G: 255, B: 255},
    Text:       domain.Color{R: 33, G: 33, B: 33},
    TextLight:  domain.Color{R: 117, G: 117, B: 117},
    Heading:    domain.Color{R: 0, G: 102, B: 204},
    Muted:      domain.Color{R: 230, G: 230, B: 230},
    Success:    domain.Color{R: 0, G: 180, B: 0},
    Warning:    domain.Color{R: 255, G: 180, B: 0},
    Error:      domain.Color{R: 220, G: 0, B: 0},
}
customTheme = customTheme.WithColors(colors)

// Define fonts
fonts := themes.ThemeFonts{
    Body:      "Open Sans",
    Heading:   "Montserrat",
    Monospace: "Fira Code",
    BodySize:  22,  // 11pt
    SmallSize: 18,  // 9pt
}
customTheme = customTheme.WithFonts(fonts)

// Define spacing
spacing := themes.ThemeSpacing{
    ParagraphBefore: 0,
    ParagraphAfter:  200,    // 10pt
    LineSpacing:     240,    // Single
    HeadingBefore:   320,    // 16pt
    HeadingAfter:    160,    // 8pt
    SectionSpacing:  480,    // 24pt
}
customTheme = customTheme.WithSpacing(spacing)

// Define heading styles
headings := themes.ThemeHeadings{
    H1Size:      36,  // 18pt
    H2Size:      30,  // 15pt
    H3Size:      26,  // 13pt
    H1Bold:      true,
    H2Bold:      true,
    H3Bold:      false,
    H1Uppercase: false,
    UseColor:    true,
}
customTheme = customTheme.WithHeadings(headings)  // Note: WithHeadings not yet implemented
```

## Theme Structure

### ThemeColors

```go
type ThemeColors struct {
    Primary    domain.Color  // Main brand color
    Secondary  domain.Color  // Supporting brand color
    Accent     domain.Color  // Highlight color
    Background domain.Color  // Page background
    Text       domain.Color  // Body text
    TextLight  domain.Color  // Secondary text
    Heading    domain.Color  // Heading color
    Muted      domain.Color  // Borders, disabled elements
    Success    domain.Color  // Success messages
    Warning    domain.Color  // Warning messages
    Error      domain.Color  // Error messages
}
```

### ThemeFonts

```go
type ThemeFonts struct {
    Body      string  // Body text font family
    Heading   string  // Heading font family
    Monospace string  // Code/monospace font
    BodySize  int     // Body font size in half-points
    SmallSize int     // Small text size in half-points
}
```

### ThemeSpacing

All measurements in twips (1/1440 inch), where 20 twips = 1pt.

```go
type ThemeSpacing struct {
    ParagraphBefore int  // Space before paragraphs
    ParagraphAfter  int  // Space after paragraphs
    LineSpacing     int  // Line height (240=single, 480=double)
    HeadingBefore   int  // Space before headings
    HeadingAfter    int  // Space after headings
    SectionSpacing  int  // Major section breaks
}
```

### ThemeHeadings

```go
type ThemeHeadings struct {
    H1Size      int   // H1 font size in half-points
    H2Size      int   // H2 font size in half-points
    H3Size      int   // H3 font size in half-points
    H1Bold      bool  // H1 bold setting
    H2Bold      bool  // H2 bold setting
    H3Bold      bool  // H3 bold setting
    H1Uppercase bool  // H1 uppercase transform
    UseColor    bool  // Use theme color for headings
}
```

## Helper Functions

### AllThemes()
Returns all available preset themes.

```go
themes := themes.AllThemes()
for _, theme := range themes {
    fmt.Printf("%s: %s\n", theme.DisplayName(), theme.Description())
}
```

### GetTheme(name)
Get a theme by its name.

```go
theme := themes.GetTheme("corporate")
if theme != nil {
    // Use theme
}
```

### ThemeNames()
Get names of all available themes.

```go
names := themes.ThemeNames()
// ["corporate", "startup", "modern", "fintech", "academic"]
```

## Best Practices

1. **Choose Appropriately**: Select themes that match your document's purpose
2. **Brand Consistency**: Customize themes to match your brand colors
3. **Test Thoroughly**: Always test documents in target applications (Word, LibreOffice)
4. **Accessibility**: Ensure sufficient color contrast for readability
5. **Font Availability**: Use commonly available fonts or embed custom fonts

## Compatibility

Themes use standard OOXML properties and work with:
- Microsoft Word 2010+
- LibreOffice Writer
- Google Docs (with limitations)
- Other OOXML applications

## Examples

See the [examples/13_themes](../examples/13_themes) directory for comprehensive examples of all themes and customization techniques.

## Future Enhancements

Planned features:
- Table styles per theme
- Custom header/footer styling
- Theme presets for specific industries
- Theme export/import (JSON/YAML)
- Dark mode themes
- Accessibility-focused themes
