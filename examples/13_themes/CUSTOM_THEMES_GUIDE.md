# Creating Custom Themes - Developer Guide

This guide shows external developers how to create their own themes for go-docx.

## Three Ways to Create Custom Themes

### 1. Clone and Modify (Easiest)

Clone an existing theme and customize specific properties:

```go
package main

import (
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/domain"
    "github.com/mmonterroca/docxgo/themes"
)

func main() {
    // Start with an existing theme
    myTheme := themes.Corporate.Clone()
    
    // Customize colors
    colors := myTheme.Colors()
    colors.Primary = domain.Color{R: 255, G: 0, B: 0} // Red
    colors.Accent = domain.Color{R: 0, G: 0, B: 255}  // Blue
    myTheme = myTheme.WithColors(colors)
    
    // Customize fonts
    fonts := myTheme.Fonts()
    fonts.Body = "Georgia"
    fonts.Heading = "Impact"
    myTheme = myTheme.WithFonts(fonts)
    
    // Use it
    doc := docx.NewDocument()
    myTheme.ApplyTo(doc)
    // ... add content
}
```

### 2. Create from Scratch

Build a completely new theme:

```go
package main

import (
    "github.com/mmonterroca/docxgo/domain"
    "github.com/mmonterroca/docxgo/themes"
)

func main() {
    // Create new theme
    brandTheme := themes.NewTheme(
        "my-brand",           // Internal ID
        "My Brand Theme",     // Display name
        "Custom brand colors" // Description
    )
    
    // Define all colors
    colors := themes.ThemeColors{
        Primary:    domain.Color{R: 0, G: 102, B: 204},
        Secondary:  domain.Color{R: 51, G: 153, B: 255},
        Accent:     domain.Color{R: 255, G: 102, B: 0},
        Background: domain.Color{R: 255, G: 255, B: 255},
        Text:       domain.Color{R: 33, G: 33, B: 33},
        TextLight:  domain.Color{R: 117, G: 117, B: 117},
        Heading:    domain.Color{R: 0, G: 102, B: 204},
        Muted:      domain.Color{R: 230, G: 230, B: 230},
        Success:    domain.Color{R: 0, G: 180, B: 0},
        Warning:    domain.Color{R: 255, G: 180, B: 0},
        Error:      domain.Color{R: 220, G: 0, B: 0},
    }
    brandTheme = brandTheme.WithColors(colors)
    
    // Define fonts
    fonts := themes.ThemeFonts{
        Body:      "Open Sans",
        Heading:   "Montserrat",
        Monospace: "Fira Code",
        BodySize:  22, // 11pt (sizes are in half-points)
        SmallSize: 18, // 9pt
    }
    brandTheme = brandTheme.WithFonts(fonts)
    
    // Define spacing (in twips: 1/1440 inch, 20 twips = 1pt)
    spacing := themes.ThemeSpacing{
        ParagraphBefore: 0,
        ParagraphAfter:  200,  // 10pt
        LineSpacing:     260,  // 1.3x line height
        HeadingBefore:   320,  // 16pt
        HeadingAfter:    160,  // 8pt
        SectionSpacing:  480,  // 24pt
    }
    brandTheme = brandTheme.WithSpacing(spacing)
    
    // Use it
    doc := docx.NewDocument()
    brandTheme.ApplyTo(doc)
}
```

### 3. Create a Theme Package (Advanced)

Create a reusable package with multiple themes:

**File: `mythemes/mythemes.go`**

```go
package mythemes

import (
    "github.com/mmonterroca/docxgo/domain"
    "github.com/mmonterroca/docxgo/themes"
)

// Export your themes
var (
    Gaming  = newGamingTheme()
    Medical = newMedicalTheme()
    Legal   = newLegalTheme()
)

func newGamingTheme() themes.Theme {
    theme := themes.NewTheme(
        "gaming",
        "Gaming",
        "Vibrant theme for gaming industry",
    )
    
    colors := themes.ThemeColors{
        Primary:    domain.Color{R: 138, G: 43, B: 226},  // BlueViolet
        Secondary:  domain.Color{R: 0, G: 255, B: 255},   // Cyan
        Accent:     domain.Color{R: 255, G: 20, B: 147},  // DeepPink
        Background: domain.Color{R: 18, G: 18, B: 18},    // Dark
        Text:       domain.Color{R: 240, G: 240, B: 240}, // Light
        TextLight:  domain.Color{R: 180, G: 180, B: 180},
        Heading:    domain.Color{R: 138, G: 43, B: 226},
        Muted:      domain.Color{R: 60, G: 60, B: 60},
        Success:    domain.Color{R: 0, G: 255, B: 127},
        Warning:    domain.Color{R: 255, G: 215, B: 0},
        Error:      domain.Color{R: 255, G: 0, B: 0},
    }
    theme = theme.WithColors(colors)
    
    fonts := themes.ThemeFonts{
        Body:      "Segoe UI",
        Heading:   "Impact",
        Monospace: "Consolas",
        BodySize:  22,
        SmallSize: 18,
    }
    theme = theme.WithFonts(fonts)
    
    spacing := themes.ThemeSpacing{
        ParagraphBefore: 0,
        ParagraphAfter:  240,
        LineSpacing:     280,
        HeadingBefore:   400,
        HeadingAfter:    200,
        SectionSpacing:  640,
    }
    theme = theme.WithSpacing(spacing)
    
    return theme
}

// ... more theme functions
```

**Using your package:**

```go
package main

import (
    "your-module/mythemes"
    docx "github.com/mmonterroca/docxgo"
)

func main() {
    doc := docx.NewDocument()
    mythemes.Gaming.ApplyTo(doc)
    // ... add content
}
```

## Theme Structure Reference

### ThemeColors

```go
type ThemeColors struct {
    Primary    domain.Color // Main brand color (headings, key elements)
    Secondary  domain.Color // Supporting color (subheadings, accents)
    Accent     domain.Color // Highlights, links, call-to-action
    Background domain.Color // Page background (usually white)
    Text       domain.Color // Default body text color
    TextLight  domain.Color // Secondary text (captions, notes)
    Heading    domain.Color // Heading color (often same as Primary)
    Muted      domain.Color // Borders, dividers, disabled elements
    Success    domain.Color // Success messages
    Warning    domain.Color // Warning messages
    Error      domain.Color // Error messages
}
```

### ThemeFonts

```go
type ThemeFonts struct {
    Body      string // Body text font family (e.g., "Calibri", "Arial")
    Heading   string // Heading font family
    Monospace string // Code/monospace font (e.g., "Courier New")
    BodySize  int    // Body font size in half-points (22 = 11pt)
    SmallSize int    // Small text size in half-points (18 = 9pt)
}
```

### ThemeSpacing

All measurements in twips (1/1440 inch). **20 twips = 1pt**.

```go
type ThemeSpacing struct {
    ParagraphBefore int // Space before paragraphs
    ParagraphAfter  int // Space after paragraphs (typically 160-240 twips)
    LineSpacing     int // Line height (240=single, 360=1.5x, 480=double)
    HeadingBefore   int // Space before headings (typically 240-400 twips)
    HeadingAfter    int // Space after headings (typically 120-200 twips)
    SectionSpacing  int // Major section breaks (typically 400-640 twips)
}
```

**Common spacing values:**
- 80 twips = 4pt
- 120 twips = 6pt
- 160 twips = 8pt
- 200 twips = 10pt
- 240 twips = 12pt
- 280 twips = 14pt
- 320 twips = 16pt
- 400 twips = 20pt
- 480 twips = 24pt

## Examples

See the example files in this directory:
- `custom_theme_example.go` - Shows all three approaches
- `external_themes/` - Example of a theme package
- `external_example.go` - How to use external themes

## Running Examples

```bash
# Clone and modify example
go run custom_theme_example.go

# External themes example
go run external_example.go
```

## Publishing Your Themes

To share your themes with others:

1. **Create a Go module:**
   ```bash
   mkdir my-docx-themes
   cd my-docx-themes
   go mod init github.com/yourname/my-docx-themes
   ```

2. **Create your theme file:**
   ```go
   // themes.go
   package mydocxthemes
   
   import (
       "github.com/mmonterroca/docxgo/domain"
       "github.com/mmonterroca/docxgo/themes"
   )
   
   var MyAwesomeTheme = themes.NewTheme(...)
   // ... configure theme
   ```

3. **Publish to GitHub:**
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/yourname/my-docx-themes
   git push -u origin main
   git tag v1.0.0
   git push --tags
   ```

4. **Others can use it:**
   ```bash
   go get github.com/yourname/my-docx-themes
   ```
   
   ```go
   import "github.com/yourname/my-docx-themes"
   
   doc := docx.NewDocument()
   mydocxthemes.MyAwesomeTheme.ApplyTo(doc)
   ```

## Best Practices

1. **Naming:** Use descriptive, lowercase names for theme IDs (e.g., "corporate-blue", "medical-clean")
2. **Colors:** Ensure sufficient contrast for readability (especially Text vs Background)
3. **Fonts:** Use commonly available fonts or provide fallbacks
4. **Testing:** Test themes in Microsoft Word, LibreOffice, and Google Docs
5. **Documentation:** Include screenshots and usage examples
6. **Versioning:** Use semantic versioning for your theme packages

## Color Inspiration

Use these resources for color schemes:
- [Coolors.co](https://coolors.co/) - Color palette generator
- [Adobe Color](https://color.adobe.com/) - Color wheel and themes
- [Material Design Colors](https://materialui.co/colors) - Google's color system
- [Flat UI Colors](https://flatuicolors.com/) - Modern flat color palettes

## Need Help?

- Check the [main themes README](../../../themes/README.md)
- See [examples](../) for more usage patterns
- Open an issue on GitHub for questions

## Contributing

If you create an awesome theme, consider contributing it to the main repository:
1. Fork the repo
2. Add your theme to `themes/presets.go`
3. Add documentation
4. Submit a pull request

Happy theming! 🎨
