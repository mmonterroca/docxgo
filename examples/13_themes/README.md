# Document Themes Example

This example demonstrates the theme system in go-docx, which allows you to apply consistent visual styling to your documents.

## Available Themes

The library includes 5 preset themes:

### 1. Corporate
Professional business theme with navy blue and red accents. Ideal for business reports, proposals, and corporate documentation.

**Color Palette:**
- Primary: Navy Blue (#2F5496)
- Secondary: Light Blue (#4F81BD)
- Accent: Red (#C00000)

**Best For:** Business reports, corporate presentations, formal proposals

### 2. Startup
Energetic modern theme with vibrant colors. Perfect for pitch decks, business plans, and innovative proposals.

**Color Palette:**
- Primary: Slate Blue (#6A5ACD)
- Secondary: Turquoise (#48D1CC)
- Accent: Tomato Red (#FF6347)

**Best For:** Startup documents, pitch decks, innovation reports

### 3. Modern
Clean minimalist theme with contemporary styling. Great for technical documentation, white papers, and modern reports.

**Color Palette:**
- Primary: Wet Asphalt (#34495E)
- Secondary: Concrete (#95A5A6)
- Accent: Peter River Blue (#2980B9)

**Best For:** Technical documentation, white papers, product specs

### 4. Fintech
Professional financial theme with trustworthy blues and greens. Ideal for financial reports, investment documents, and banking materials.

**Color Palette:**
- Primary: Dark Cerulean (#005288)
- Secondary: Blue NCS (#007BA7)
- Accent: Teal (#009688)

**Best For:** Financial reports, investment documents, banking materials

### 5. Academic
Traditional scholarly theme for academic and research documents. Perfect for research papers, academic reports, and thesis documents.

**Color Palette:**
- Primary: Black
- Secondary: Dark Gray
- Accent: Maroon (#800000)

**Best For:** Research papers, academic reports, thesis documents

## Usage

### Basic Usage

```go
import (
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/themes"
)

// Create a document with a theme
builder := docx.NewDocumentBuilder(
    docx.WithTheme(themes.Corporate),
)

builder.AddParagraph().
    Text("Hello, World!").
    End()

doc, _ := builder.Build()
doc.SaveAs("output.docx")
```

### Customizing a Theme

```go
// Clone a theme and customize it
customTheme := themes.Corporate.Clone()

// Modify colors
colors := customTheme.Colors()
colors.Primary = docx.Color{R: 255, G: 0, B: 0} // Change to red
customTheme = customTheme.WithColors(colors)

// Modify fonts
fonts := customTheme.Fonts()
fonts.Body = "Arial"
fonts.Heading = "Arial Black"
customTheme = customTheme.WithFonts(fonts)

// Use the custom theme
builder := docx.NewDocumentBuilder(
    docx.WithTheme(customTheme),
)
```

### Applying Theme to Existing Document

```go
// Open an existing document
doc, _ := docx.OpenDocument("existing.docx")

// Apply a theme
themes.Modern.ApplyTo(doc)

// Save with new styling
doc.SaveAs("restyled.docx")
```

## Running the Examples

The examples are organized in subdirectories:

### 1. Main Themes Example (`01_main/`)
Demonstrates all 5 preset themes with full document content:
```bash
cd examples/13_themes/01_main
go run main.go
```
Generates: `corporate_theme.docx`, `startup_theme.docx`, `modern_theme.docx`, `fintech_theme.docx`, `academic_theme.docx`, `theme_comparison.docx`

### 2. Custom Theme Example (`02_custom_theme/`)
Shows how to clone and customize themes or build from scratch:
```bash
cd examples/13_themes/02_custom_theme
go run main.go
```
Generates: `custom_cloned_theme.docx`, `custom_brand_theme.docx`, `custom_builder_theme.docx`

### 3. External Themes Example (`03_external_themes/`)
Demonstrates using themes from an external package:
```bash
cd examples/13_themes/03_external_themes
go run main.go
```
Generates: `gaming_theme_example.docx`, `medical_theme_example.docx`, `legal_theme_example.docx`

### 4. Technical Architecture Example (`04_tech_architecture/`) 🌟
**Our showcase example!** Modern technical documentation with PlantUML diagrams:
```bash
cd examples/13_themes/04_tech_architecture
go run main.go
```
Generates: `tech_architecture_light.docx`, `tech_architecture_dark.docx`

**Features:**
- Tech Presentation theme (Light & Dark Mode)
- PlantUML diagrams (Class, Component, Sequence)
- Code blocks with syntax highlighting
- Professional tables
- Architecture Decision Records (ADRs)
- Complete microservices architecture example

## Theme Components

Each theme defines:

1. **Colors** - Complete color palette including:
   - Primary, Secondary, Accent
   - Background, Text, TextLight
   - Success, Warning, Error

2. **Fonts** - Font families and sizes:
   - Body font, Heading font, Monospace
   - Body size, Small size

3. **Spacing** - Layout spacing:
   - Paragraph spacing (before/after)
   - Line spacing
   - Heading spacing
   - Section spacing

4. **Headings** - Heading styles:
   - H1, H2, H3 sizes
   - Bold settings
   - Color usage

## Creating Custom Themes

You can create your own theme from scratch:

```go
import "github.com/mmonterroca/docxgo/themes"

customTheme := themes.NewTheme(
    "my-theme",
    "My Custom Theme",
    "A theme tailored for my organization",
)

// Configure colors
colors := themes.ThemeColors{
    Primary:    domain.Color{R: 100, G: 100, B: 200},
    Secondary:  domain.Color{R: 150, G: 150, B: 250},
    // ... configure all colors
}
customTheme = customTheme.WithColors(colors)

// Configure fonts
fonts := themes.ThemeFonts{
    Body:      "Georgia",
    Heading:   "Helvetica",
    BodySize:  24, // 12pt
    // ...
}
customTheme = customTheme.WithFonts(fonts)

// Use your custom theme
builder := docx.NewDocumentBuilder(
    docx.WithTheme(customTheme),
)
```

## Best Practices

1. **Choose the Right Theme**: Select a theme that matches your document's purpose and audience
2. **Consistency**: Use themes to maintain consistency across multiple documents
3. **Customization**: Don't hesitate to customize themes to match your brand
4. **Testing**: Always test themed documents in Microsoft Word to ensure compatibility
5. **Accessibility**: Consider color contrast and readability when customizing colors

## Technical Details

### How Themes Work

Themes work by configuring the document's `StyleManager`. When a theme is applied:

1. The Normal (default) paragraph style is configured with the theme's font and spacing
2. Heading styles (H1-H3) are configured with appropriate sizes, colors, and formatting
3. Special styles (Title, Subtitle, Quote, etc.) are styled according to the theme
4. The configuration is applied to the document's built-in styles

### Style Inheritance

Styles in DOCX follow an inheritance model:
- Most paragraph styles inherit from "Normal"
- Heading styles inherit from "Normal" but override specific properties
- Themes leverage this inheritance to create consistent styling

### Compatibility

Themes use only standard OOXML properties and are compatible with:
- Microsoft Word (2010 and later)
- LibreOffice Writer
- Google Docs (with some limitations)
- Other OOXML-compatible applications

## Creating Custom Themes for External Distribution

External developers can create their own theme packages and distribute them. There are three approaches:

### 1. Clone and Customize (Easiest)

See `custom_theme_example.go` for a complete example of cloning and customizing existing themes.

### 2. Create a Theme Package

Create a separate Go package with your themes:

```go
// mythemes/mythemes.go
package mythemes

import "github.com/mmonterroca/docxgo/themes"

var Gaming = themes.NewTheme("gaming", "Gaming", "Vibrant gaming theme")
var Medical = themes.NewTheme("medical", "Medical", "Clean medical theme")
// ... configure themes
```

Use in your code:

```go
import "your-module/mythemes"

doc := docx.NewDocument()
mythemes.Gaming.ApplyTo(doc)
```

See `external_themes/` directory for a complete example package with Gaming, Medical, and Legal themes.

### 3. Publish Your Themes

You can publish your themes as a Go module for others to use:

```bash
# Create module
go mod init github.com/yourname/my-docx-themes

# Publish to GitHub
git tag v1.0.0
git push --tags

# Others can use it
go get github.com/yourname/my-docx-themes
```

## Running All Examples

```bash
# Main theme demonstration
cd 01_main && go run main.go && cd ..

# Custom theme cloning example
cd 02_custom_theme && go run main.go && cd ..

# External themes package example
cd 03_external_themes && go run main.go && cd ..

# Technical architecture showcase
cd 04_tech_architecture && go run main.go && cd ..
```

Or run all at once with the provided script:
```bash
cd examples/13_themes
./run_all.sh
```

This will execute all examples in sequence and generate **14 documents total**.

## Documentation

For complete documentation on creating custom themes, see:
- [CUSTOM_THEMES_GUIDE.md](./CUSTOM_THEMES_GUIDE.md) - Comprehensive guide for external developers
- `external_themes/` - Example theme package with Gaming, Medical, and Legal themes
- `custom_theme_example.go` - Example of cloning and customizing themes
- `external_example.go` - Example of using external theme packages

## See Also

- [Main README](../../README.md) - Library documentation
- [Styles Example](../05_styles/main.go) - Working with styles
- [V2 API Guide](../../docs/V2_API_GUIDE.md) - Complete API reference
- [Theme System Design](../../themes/README.md) - Theme architecture documentation

````
