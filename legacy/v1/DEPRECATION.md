# ⚠️ DEPRECATION NOTICE

## This is Legacy Code (v1.x)

**Status**: Deprecated  
**Replacement**: v2.0 (now in project root)  
**Maintenance**: Bug fixes only, no new features

---

## Why is v1 Deprecated?

Version 2.0 represents a **complete architectural rewrite** with:

- ✅ **Clean Architecture** - Interface-based design, dependency injection
- ✅ **Type Safety** - No `interface{}` usage, proper error handling
- ✅ **Testability** - 95%+ test coverage, mockable interfaces
- ✅ **Performance** - Optimized memory allocation, thread-safe managers
- ✅ **Better API** - Builder pattern with fluent interface

v1 has fundamental limitations that cannot be fixed without breaking changes:

- ❌ Silent error handling (no errors returned in fluent API)
- ❌ God objects with too many responsibilities
- ❌ Heavy use of `interface{}` reducing type safety
- ❌ Difficult to test (no interfaces, tight coupling)
- ❌ Global state in some areas

---

## Should You Use v1?

### ✅ Use v1 Legacy If:

- You have **existing code** using v1 and can't migrate yet
- You need **bug fixes** for production code running v1
- You're waiting for v2 to reach stable (currently pre-alpha)

### ❌ Don't Use v1 For:

- **New projects** - Start with v2
- **Long-term maintenance** - v1 will not receive new features
- **Complex documents** - v2 has better architecture for complexity

---

## Migration Path

### Option 1: Wait for v2 Stable (Recommended for Most)

**Timeline**: v2.0.0 stable expected Q1 2026

1. Continue using v1 with current namespace
2. Monitor v2 releases (beta available soon)
3. Migrate when v2.0.0 stable is released
4. Use [MIGRATION.md](../../MIGRATION.md) guide

### Option 2: Migrate to v2 Now (Early Adopters)

**Status**: v2.0.0-alpha (pre-alpha, API may change)

1. Read [v2 README](../../README.md)
2. Follow [MIGRATION.md](../../MIGRATION.md) guide
3. Start with simple documents
4. Test thoroughly (v2 is not production-ready yet)
5. Report issues at https://github.com/SlideLang/go-docx/issues

### Option 3: Stay on v1 Indefinitely (Not Recommended)

If you absolutely cannot migrate:

1. Fork this `legacy/v1/` code to your own repository
2. Maintain your own fork
3. Be aware of security/bug implications

We will provide **critical bug fixes** for v1 through 2025, but no new features.

---

## v1 Documentation

### Installation

```bash
# Use legacy v1 code (not recommended for new projects)
go get github.com/SlideLang/go-docx/legacy/v1
```

### Quick Start

```go
package main

import (
    "os"
    docx "github.com/SlideLang/go-docx/legacy/v1"
)

func main() {
    // v1 API (deprecated)
    w := docx.New().WithDefaultTheme()
    para := w.AddParagraph()
    para.AddText("Hello from v1").Size("24")
    
    f, _ := os.Create("output.docx")
    w.WriteTo(f)
    f.Close()
}
```

### Full v1 Documentation

- [v1 README](README.md) - Original v1 documentation
- [API Documentation](../../docs/API_DOCUMENTATION.md) - v1 API reference (legacy sections)
- [Examples](examples/) - v1 code examples

---

## Key Differences: v1 vs v2

### Error Handling

```go
// v1 - No errors returned (silent failures)
para.AddText("test").Bold().Color("INVALID") // No error!

// v2 - Errors propagate through builder
doc.AddParagraph().
    Text("test").
    Bold().
    Color("INVALID"). // Error recorded
    End()
    
finalDoc, err := doc.Build() // Errors surface here
if err != nil {
    log.Fatal(err)
}
```

### Type Safety

```go
// v1 - interface{} everywhere
type Docx struct {
    Document struct {
        Body struct {
            Items []interface{} // Could be anything!
        }
    }
}

// v2 - Concrete types
type Document interface {
    Paragraphs() []Paragraph
    Tables() []Table
    Sections() []Section
}
```

### Architecture

```go
// v1 - God object
type Docx struct {
    Document, Template, Rels, Media, IDs, Num, ... // 15+ fields
}

// v2 - Dependency injection
type docxDocument struct {
    relationMgr  RelationshipManager
    mediaMgr     MediaManager
    idGen        IDGenerator
    styleMgr     StyleManager
}
```

### Testability

```go
// v1 - Hard to test (concrete structs, no interfaces)
func TestSomething(t *testing.T) {
    doc := docx.New() // Can't mock
}

// v2 - Easy to mock (interface-based)
func TestSomething(t *testing.T) {
    mockDoc := &MockDocument{}
    mockDoc.On("AddParagraph").Return(...)
}
```

---

## Support & Questions

### For v1 Issues

- **Critical bugs**: Report at https://github.com/SlideLang/go-docx/issues
  - Tag with `legacy-v1` label
  - Include "v1" in title

### For Migration Questions

- **Migration help**: Create discussion at https://github.com/SlideLang/go-docx/discussions
- **Email**: misael@monterroca.com (for complex migration scenarios)

### For v2 Questions

- **General questions**: GitHub Discussions
- **Bugs/Features**: GitHub Issues (tag with `v2`)
- **Documentation**: See main [README](../../README.md)

---

## Timeline

| Date | Event |
|------|-------|
| Oct 2025 | v2.0.0-alpha released |
| Nov 2025 | v1 moved to legacy/ (now) |
| Dec 2025 | v2.0.0-beta expected |
| Jan 2026 | v2 release candidate |
| Mar 2026 | v2.0.0 stable release |
| Dec 2025 | v1 critical bug fixes end |
| Mar 2026 | v1 fully deprecated (no support) |

---

## Credits

v1 was built on the shoulders of:

- **Gonzalo Fernández-Victorio** - Original `gonfva/docxlib`
- **fumiama** - Enhanced fork with images, tables, shapes
- **Misael Monterroca** - Professional features (headers, TOC, links)

See [CREDITS.md](../../CREDITS.md) for complete history.

---

## License

v1 code remains under **AGPL-3.0** license.

See [LICENSE](LICENSE) for full text.

---

*This deprecation notice was created: October 25, 2025*  
*For the latest information, see: https://github.com/SlideLang/go-docx*
