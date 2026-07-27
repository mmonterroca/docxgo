# Documentation Index

**docxgo v2 - Complete Documentation Guide**

This index helps you find the right documentation for your needs.

---

## 🎯 I Want To...

### Get Started Quickly
→ **[README.md](../README.md)** - Quick start guide and installation  
→ **[V2_API_GUIDE.md](./V2_API_GUIDE.md)** - Complete API reference with examples

### Learn the Architecture
→ **[V2_DESIGN.md](./V2_DESIGN.md)** - Design decisions, patterns, and architecture  
→ **[IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md)** - What's implemented and what's planned

### Migrate from v1
→ **[MIGRATION.md](../MIGRATION.md)** - Step-by-step migration from v1 to v2

### See Working Examples
→ **[examples/README.md](../examples/README.md)** - 15 working examples (including read/modify, mail merge, and themes)

### Understand Error Handling
→ **[ERROR_HANDLING.md](./ERROR_HANDLING.md)** - Error handling patterns

### Use the CLI
→ **[CLI_GUIDE.md](./CLI_GUIDE.md)** - JSON-RPC CLI binary usage

### Contribute
→ **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Contribution guidelines  
→ **[CREDITS.md](../CREDITS.md)** - Project history and contributors

---

## 📚 Documentation by Type

### Primary Documentation (Current v2)

#### For Users
| Document | Purpose | Audience |
|----------|---------|----------|
| [V2_API_GUIDE.md](./V2_API_GUIDE.md) | Complete API reference | All users |
| [README.md](../README.md) | Quick start guide | New users |
| [examples/](../examples/) | Working code examples | All users |
| [MIGRATION.md](../MIGRATION.md) | v1 to v2 migration | Upgrading users |

#### For Developers
| Document | Purpose | Audience |
|----------|---------|----------|
| [V2_DESIGN.md](./V2_DESIGN.md) | Architecture & design | Contributors |
| [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) | Current status | Contributors |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | How to contribute | Contributors |
| [ERROR_HANDLING.md](./ERROR_HANDLING.md) | Error patterns | Contributors |
| [CLI_GUIDE.md](./CLI_GUIDE.md) | JSON-RPC CLI binary usage | CLI / integration users |
| [PARAGRAPH_BORDERS.md](./PARAGRAPH_BORDERS.md) | Paragraph border API | Contributors |
| [TROUBLESHOOTING_DOCX_VALIDATION.md](./TROUBLESHOOTING_DOCX_VALIDATION.md) | Debugging docx validation | Contributors |

### Legacy Documentation (Archived - v1 Pre-rewrite)

> **Note**: All legacy v1 documentation has been removed. This project is now v2 only.
> For migration guidance, see [MIGRATION.md](../MIGRATION.md).

### Project Information

| Document | Purpose |
|----------|---------|
| [CREDITS.md](../CREDITS.md) | Project history & attribution |
| [LICENSE](../LICENSE) | MIT License |
| [CHANGELOG.md](../CHANGELOG.md) | Version history |

---

## 🗺️ Documentation by Topic

### API and Usage
1. **[V2_API_GUIDE.md](./V2_API_GUIDE.md)** - Complete API guide
   - Builder pattern API
   - Direct domain API
   - All features with examples
   - Migration from v1

2. **[examples/README.md](../examples/README.md)** - Working examples
   - 01_basic - Builder pattern basics
   - 02_intermediate - Multi-section documents
   - 03_toc - Table of Contents
   - 04_fields - Field system
   - 05_styles - Style management
   - 06_sections - Page layout
   - 07_advanced - Advanced integration
   - 08_images - Image handling
   - 09_advanced_tables - Table features
   - 10_paragraph_spacing - Line and paragraph spacing
   - 11_multi_section - Multi-section layouts
   - 12_read_and_modify - Read/modify existing documents
   - 13_themes - Theme system (7 preset themes)
   - 14_mail_merge - Template engine with mail merge
   - 15_external_template - Mail merge with external Word template

### Architecture and Design
1. **[V2_DESIGN.md](./V2_DESIGN.md)** - Architecture overview
   - Design goals
   - Package structure
   - Design patterns
   - Implementation phases

2. **[IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md)** - Current status
   - ✅ What's implemented (production ready)
   - 🚧 What's partial
   - ⏳ What's planned
   - Known limitations

3. **[ERROR_HANDLING.md](./ERROR_HANDLING.md)** - Error handling
   - Error types
   - Error patterns
   - Best practices

### Migration and History
1. **[MIGRATION.md](../MIGRATION.md)** - v1 to v2 migration
   - API changes
   - Code examples
   - Breaking changes

2. **[CREDITS.md](../CREDITS.md)** - Project history
   - Original fork attribution
   - Complete rewrite history
   - Contributors

3. **[CHANGELOG.md](../CHANGELOG.md)** - Version history
   - Release notes
   - Breaking changes
   - Bug fixes

---

## 📖 Reading Order by Role

### New User
1. [README.md](../README.md) - Installation and quick start
2. [V2_API_GUIDE.md](./V2_API_GUIDE.md) - Learn the API
3. [examples/01_basic/](../examples/01_basic/) - Try basic example
4. [examples/](../examples/) - Explore more examples

### Migrating from v1
1. [MIGRATION.md](../MIGRATION.md) - Migration guide
2. [V2_API_GUIDE.md](./V2_API_GUIDE.md) - New API reference
3. [examples/](../examples/) - See v2 patterns

### Contributor
1. [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
2. [V2_DESIGN.md](./V2_DESIGN.md) - Understand architecture
3. [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) - See what needs work
4. [ERROR_HANDLING.md](./ERROR_HANDLING.md) - Error patterns
5. Code in `internal/` and `domain/` - Implementation

### Maintainer
1. [V2_DESIGN.md](./V2_DESIGN.md) - Architecture overview
2. [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) - Track progress
3. [CHANGELOG.md](../CHANGELOG.md) - Maintain version history

---

## 🔍 Quick Reference

### Common Tasks

**Create a simple document:**
```go
// See: V2_API_GUIDE.md - Quick Start
builder := docx.NewDocumentBuilder()
builder.AddParagraph().Text("Hello, World!").End()
doc, _ := builder.Build()
doc.SaveAs("hello.docx")
```

**Add a table:**
```go
// See: V2_API_GUIDE.md - Tables
builder.AddTable(3, 3).
    Row(0).Cell(0).Text("Header").Bold().End()
```

**Add an image:**
```go
// See: V2_API_GUIDE.md - Images
builder.AddParagraph().
    AddImage("logo.png").
    End()
```

**Add page numbers:**
```go
// See: V2_API_GUIDE.md - Fields
section, _ := doc.DefaultSection()
footer, _ := section.Footer(domain.FooterDefault)
para, _ := footer.AddParagraph()
run, _ := para.AddRun()
run.AddField(docx.NewPageNumberField())
```

**Check implementation status:**
```
See: IMPLEMENTATION_STATUS.md - Features list
```

---

## 🆘 Need Help?

**Can't find what you're looking for?**

1. Check [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) - Feature might not be implemented yet
2. Review [examples/](../examples/) - Working code often explains best
3. Read [V2_API_GUIDE.md](./V2_API_GUIDE.md) - Comprehensive API reference
4. Open a GitHub Issue - Community can help

**Found an issue in the docs?**

1. Open a GitHub Issue
2. Submit a Pull Request (see [CONTRIBUTING.md](../CONTRIBUTING.md))
3. The docs are in Markdown - easy to fix!

---

**Last Updated**: July 2026  
**Documentation Version**: v2.9.0
