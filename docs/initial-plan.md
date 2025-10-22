# SlideLang/go-docx - Enhanced Fork

**Fork de**: https://github.com/fumiama/go-docx  
**Nuestro fork**: https://github.com/SlideLang/go-docx  
**Versión base**: v0.0.0-20250506085032-0c30fd09304b (commit: 0c30fd09304b, 6 Mayo 2025)

## 🎯 Objetivo del Fork

Extender go-docx con funcionalidades profesionales necesarias para generar documentos Word de alta calidad desde **DocLang** y **SlideLang**, especialmente para exportación DOCX.

### ¿Por qué un Fork?

1. **Repositorio inactivo**: Último commit hace 5+ meses, no es desarrollo activo
2. **Funcionalidades críticas faltantes**: Bookmarks, campos de Word, TOC dinámico, estilos nativos
3. **Necesidad de control**: Requerimos features específicas para documentos profesionales
4. **Proyecto independiente**: Biblioteca standalone útil para otros proyectos Go

## 🚀 Setup Inicial

### 1. Clonar el Fork

```bash
# Clonar el repositorio
git clone git@github.com:SlideLang/go-docx.git ~/go-docx-slidelang
cd ~/go-docx-slidelang

# Agregar upstream para sincronizar con original
git remote add upstream https://github.com/fumiama/go-docx.git

# Verificar remotes
git remote -v
# origin    git@github.com:SlideLang/go-docx.git (fetch)
# origin    git@github.com:SlideLang/go-docx.git (push)
# upstream  https://github.com/fumiama/go-docx.git (fetch)
# upstream  https://github.com/fumiama/go-docx.git (push)
```

### 2. Crear Rama de Desarrollo

```bash
# Crear rama permanente para nuestras mejoras
git checkout -b slidelang-enhanced

# Push de la rama
git push -u origin slidelang-enhanced
```

### 3. Configurar Go Module

```bash
# Verificar que el módulo funciona
go mod tidy
go test ./...

# Ver estructura del proyecto
ls -la
```

### 4. Entender la Estructura

```
go-docx/
├── api*.go              # API pública (AddParagraph, AddTable, etc.)
├── struct*.go           # Estructuras XML (Paragraph, Run, Table, etc.)
├── docx.go              # Tipo principal Docx
├── document.go          # Manejo del document.xml
├── styles.go            # Estilos del documento
├── numbering.go         # Numeración de listas
├── rels.go              # Relaciones (imágenes, links)
└── examples/            # Ejemplos de uso
```

## 📋 Funcionalidades a Implementar

### Análisis de Necesidades de DocLang

Basándonos en la especificación de DocLang (`docs/doclang/`), necesitamos soportar:

#### 1. **Tabla de Contenidos (TOC) Dinámica** 🔴 CRÍTICO

**Requisito de DocLang:**
```yaml
toc:
  enabled: true
  depth: 3                    # Niveles H1-H3
  title: "Table of Contents"
  page_numbers: true          # ← REQUIERE BOOKMARKS + PAGEREF
  hyperlinks: true            # ← REQUIERE BOOKMARKS + HYPERLINKS INTERNOS
```

**Features necesarias:**
- ✅ Bookmarks en headings (para referencias)
- ✅ Campo `{ TOC \o "1-3" \h \z \u }` (tabla automática)
- ✅ Campo `{ PAGEREF _Toc123 }` (números de página)
- ✅ Hyperlinks internos a bookmarks (`w:anchor="_Toc123"`)

**Impacto:** SIN esto, el TOC es solo texto estático sin funcionalidad

---

#### 2. **Estilos de Word Nativos (Heading1-4)** 🔴 CRÍTICO

**Problema actual:**
```go
p.Style("Heading1")  // ❌ No funciona - estilo no existe en styles.xml
```

**Requisito de DocLang:**
- Headings H1-H4 deben usar estilos nativos de Word
- Permite actualizar TOC automáticamente
- Navigation Pane funcional en Word
- Formato consistente entre temas

**Features necesarias:**
- ✅ Definiciones de estilos en styles.xml
- ✅ API para crear/modificar estilos
- ✅ Aplicar OutlineLvl a párrafos (jerarquía)
- ✅ Vincular estilos a formato visual

**Impacto:** SIN esto, documentos parecen "amateur" y no tienen estructura Word nativa

---

#### 3. **Numeración Automática** 🟡 IMPORTANTE

**Requisito de DocLang:**
```yaml
numbering:
  enabled: true
  style: "hierarchical"     # 1, 1.1, 1.1.1, 1.2, 2, 2.1...
  sections: true
  figures: true             # Figure 1, Figure 2...
  tables: true              # Table 1, Table 2...
  charts: true
```

**Features necesarias:**
- ✅ Campos `{ SEQ figure \* ARABIC }` para auto-numeración
- ✅ Campos `{ STYLEREF 1 }` para referencias a secciones
- ✅ Numeración multi-nivel configurable
- ✅ Caption automático para figuras/tablas

**Impacto:** Numeración manual propensa a errores, no profesional

---

#### 4. **Headers y Footers con Variables** 🟡 IMPORTANTE

**Requisito de DocLang:**
```yaml
header:
  enabled: true
  odd_pages: "{{title}}"
  even_pages: "{{section_title}}"
  
footer:
  enabled: true
  page_numbers:
    enabled: true
    format: "Page {{current}} of {{total}}"
    alignment: "center"
```

**Features necesarias:**
- ✅ Headers/Footers por sección
- ✅ Campo `{ PAGE }` para número de página actual
- ✅ Campo `{ NUMPAGES }` para total de páginas
- ✅ Campo `{ STYLEREF "Heading1" }` para título de sección
- ✅ Headers/Footers diferentes para pares/impares

**Impacto:** Headers/footers estáticos vs. dinámicos profesionales

---

#### 5. **Referencias Cruzadas** 🟢 DESEABLE

**Requisito de DocLang:**
```markdown
See section {{ref:introduction}} for details.
As shown in {{ref:fig:architecture}}.
```

**Features necesarias:**
- ✅ Bookmarks nombrados
- ✅ Campo `{ REF bookmark_name }` para referencias
- ✅ Campo `{ REF bookmark_name \h }` para hyperlinks
- ✅ Actualización automática de referencias

**Impacto:** Referencias "ver sección X" quedan obsoletas al reorganizar

---

#### 6. **Metadatos del Documento** 🟢 DESEABLE

**Requisito de DocLang:**
```yaml
title: "Technical Specification"
author: "Engineering Team"
date: "2025-10-21"
version: "2.0"
keywords: ["API", "REST", "Microservices"]
```

**Features necesarias:**
- ✅ Core Properties (docProps/core.xml)
- ✅ Custom Properties (docProps/custom.xml)
- ✅ Visible en "Propiedades del documento" en Word

**Impacto:** Metadatos facilitan búsqueda y organización

---

## 🛠️ Plan de Implementación

### Fase 1: Fundamentos (Semana 1) 🔴

#### 1.1 Bookmarks
**Archivos a modificar:**
- `structpara.go` - Agregar `Bookmark` como child de `Paragraph`
- `apibookmark.go` (nuevo) - API pública `AddBookmark(name string)`

**Estructuras necesarias:**
```go
type BookmarkStart struct {
    XMLName xml.Name `xml:"w:bookmarkStart"`
    ID      string   `xml:"w:id,attr"`
    Name    string   `xml:"w:name,attr"`
}

type BookmarkEnd struct {
    XMLName xml.Name `xml:"w:bookmarkEnd"`
    ID      string   `xml:"w:id,attr"`
}
```

**API pública:**
```go
func (p *Paragraph) AddBookmark(name string) *Bookmark {
    id := generateBookmarkID()
    bookmark := &Bookmark{
        ID:   id,
        Name: name,
    }
    p.Children = append(p.Children, bookmark)
    return bookmark
}
```

**Tests:**
```go
func TestBookmark(t *testing.T) {
    w := New().WithDefaultTheme().WithA4Page()
    p := w.AddParagraph()
    p.AddText("Introduction")
    p.AddBookmark("_Toc123456789")
    
    // Verificar XML contiene bookmarkStart y bookmarkEnd
}
```

---

#### 1.2 Field Characters (fldChar)
**Archivos a modificar:**
- `structrun.go` - Agregar `FldChar` como child de `Run`
- `apifield.go` (nuevo) - API para campos de Word

**Estructuras necesarias:**
```go
type FldChar struct {
    XMLName     xml.Name `xml:"w:fldChar"`
    FldCharType string   `xml:"w:fldCharType,attr"` // "begin", "separate", "end"
}

type Field struct {
    Begin     *Run  // w:fldChar type="begin"
    InstrText *Run  // w:instrText
    Separate  *Run  // w:fldChar type="separate"
    Result    *Run  // Resultado del campo
    End       *Run  // w:fldChar type="end"
}
```

**API pública:**
```go
func (p *Paragraph) AddField(instrText string, result string) *Field {
    // Begin
    begin := p.AddRun()
    begin.AddFldChar("begin")
    
    // InstrText
    instr := p.AddRun()
    instr.InstrText = instrText
    
    // Separate
    sep := p.AddRun()
    sep.AddFldChar("separate")
    
    // Result (valor por defecto)
    res := p.AddRun()
    res.AddText(result)
    
    // End
    end := p.AddRun()
    end.AddFldChar("end")
    
    return &Field{begin, instr, sep, res, end}
}
```

**Tests:**
```go
func TestFieldTOC(t *testing.T) {
    w := New().WithDefaultTheme().WithA4Page()
    p := w.AddParagraph()
    p.AddField("TOC \\o \"1-3\" \\h \\z \\u", "Table of Contents")
    
    // Verificar estructura completa del campo
}

func TestFieldPAGEREF(t *testing.T) {
    w := New().WithDefaultTheme().WithA4Page()
    p := w.AddParagraph()
    p.AddField("PAGEREF _Toc123 \\h", "3")
    
    // Verificar campo PAGEREF
}
```

---

### Fase 2: Tabla de Contenidos (Semana 2) 🔴

#### 2.1 TOC Builder
**Archivos a crear:**
- `apitoc.go` - API de alto nivel para TOC

**API pública:**
```go
type TOCOptions struct {
    Title         string   // "Table of Contents"
    Depth         int      // 1-4 (niveles H1-H4)
    PageNumbers   bool     // Mostrar números de página
    Hyperlinks    bool     // Hyperlinks clicables
    RightAlign    bool     // Alinear números a la derecha
    TabLeader     string   // "dot", "hyphen", "underscore", "none"
}

func (d *Docx) AddTOC(opts TOCOptions) error {
    // 1. Insertar campo TOC
    tocPara := d.AddParagraph()
    instrText := fmt.Sprintf("TOC \\o \"1-%d\"", opts.Depth)
    if opts.Hyperlinks {
        instrText += " \\h"
    }
    tocPara.AddField(instrText, opts.Title)
    
    // 2. Generar entradas de ejemplo
    // (Word las regenerará al abrir el documento)
    
    return nil
}

func (p *Paragraph) AddTOCEntry(level int, text string, bookmark string, pageNum string) {
    // Crear entrada TOC con formato correcto
    // - Indentación según level
    // - Hyperlink a bookmark
    // - Campo PAGEREF para número de página
    // - Tab con dotted leader
}
```

**Tests:**
```go
func TestTOC(t *testing.T) {
    w := New().WithDefaultTheme().WithA4Page()
    
    // Agregar TOC
    w.AddTOC(TOCOptions{
        Title:       "Contents",
        Depth:       3,
        PageNumbers: true,
        Hyperlinks:  true,
        TabLeader:   "dot",
    })
    
    // Agregar headings con bookmarks
    h1 := w.AddParagraph()
    h1.AddText("Introduction").Size("28").Bold()
    h1.AddBookmark("_Toc001")
    
    h2 := w.AddParagraph()
    h2.AddText("Background").Size("24").Bold()
    h2.AddBookmark("_Toc002")
    
    // Guardar y verificar que Word puede abrir el archivo
    w.SaveTo("test_toc.docx")
}
```

---

### Fase 3: Estilos de Word (Semana 2-3) 🔴

#### 3.1 Style Definitions
**Archivos a modificar:**
- `styles.go` - Agregar definiciones de Heading1-4
- `apistyle.go` (nuevo) - API para estilos personalizados

**Estructuras necesarias:**
```go
type StyleDefinition struct {
    StyleID   string  // "Heading1", "Heading2", etc.
    Name      string  // "Heading 1"
    Type      string  // "paragraph"
    BasedOn   string  // "Normal"
    NextStyle string  // "Normal"
    
    // Formato
    Font      string
    Size      string
    Color     string
    Bold      bool
    Italic    bool
    
    // Espaciado
    SpaceBefore int
    SpaceAfter  int
    
    // Outline level (para TOC)
    OutlineLvl  int  // 0=H1, 1=H2, 2=H3, 3=H4
}

func (d *Docx) AddStyleDefinition(style StyleDefinition) error {
    // Agregar estilo a styles.xml
}

func (d *Docx) WithHeadingStyles() *Docx {
    // Agregar Heading1-4 con formato profesional
    d.AddStyleDefinition(StyleDefinition{
        StyleID:     "Heading1",
        Name:        "Heading 1",
        Type:        "paragraph",
        Font:        "Calibri",
        Size:        "32",  // 16pt
        Color:       "2E75B6",
        Bold:        true,
        SpaceBefore: 480,  // 24pt
        SpaceAfter:  120,  // 6pt
        OutlineLvl:  0,
    })
    // ... Heading2, Heading3, Heading4
    return d
}
```

**Uso:**
```go
w := docx.New().WithDefaultTheme().WithA4Page().WithHeadingStyles()

h1 := w.AddParagraph()
h1.AddText("Introduction")
h1.Style("Heading1")  // ✅ Ahora funciona porque está definido
```

---

### Fase 4: Numeración y Referencias (Semana 3-4) 🟡

#### 4.1 Page Numbers en Headers/Footers
```go
func (d *Docx) AddPageNumbers(format string, alignment string) {
    footer := d.GetOrCreateFooter()
    p := footer.AddParagraph()
    p.Justification(alignment)
    
    // Campo PAGE
    p.AddField("PAGE", "1")
    
    if strings.Contains(format, "{{total}}") {
        p.AddText(" of ")
        p.AddField("NUMPAGES", "10")
    }
}
```

#### 4.2 Auto-Numbering para Figuras/Tablas
```go
func (d *Docx) AddFigureCaption(text string) {
    p := d.AddParagraph()
    p.AddField("SEQ Figure \\* ARABIC", "1")
    p.AddText(": " + text)
}
```

---

## 🔗 Integración con DocLang/SlideLang

### Actualizar go.mod en cli

```bash
cd ~/cli/doclang-cli

# Reemplazar fumiama/go-docx con nuestro fork
go mod edit -replace github.com/fumiama/go-docx=github.com/SlideLang/go-docx@slidelang-enhanced

# O directamente editar go.mod:
```

```go
// go.mod
module github.com/slidelang/doclang-cli

require (
    github.com/fumiama/go-docx v0.0.0-20250506085032-0c30fd09304b
)

replace github.com/fumiama/go-docx => github.com/SlideLang/go-docx slidelang-enhanced
```

```bash
go mod tidy
```

### Usar Nuevas Funcionalidades

```go
// doclang-cli/internal/generator/docx.go

func (g *DOCXGenerator) Generate(doc *ast.AST, outputFile string, opts GeneratorOptions) error {
    // Crear documento con estilos de heading
    w := docx.New().WithDefaultTheme().WithA4Page().WithHeadingStyles()
    
    // Agregar TOC si está habilitado
    if doc.FrontMatter != nil && doc.FrontMatter.TOC {
        w.AddTOC(docx.TOCOptions{
            Title:       "Tabla de Contenido",
            Depth:       3,
            PageNumbers: true,
            Hyperlinks:  true,
            TabLeader:   "dot",
        })
    }
    
    // Agregar page numbers
    w.AddPageNumbers("Page {{current}} of {{total}}", "center")
    
    // Renderizar contenido con bookmarks
    for _, slide := range doc.Slides {
        if slide.Title != "" {
            h1 := w.AddParagraph()
            h1.AddText(slide.Title)
            h1.Style("Heading1")  // ✅ Ahora funciona
            h1.AddBookmark(generateTOCBookmark(slide.Title))  // ✅ Para TOC
        }
    }
    
    return w.SaveTo(outputFile)
}
```

---

## 📊 Métricas de Éxito

### Checklist de Funcionalidades

- [ ] **Bookmarks**: Crear y referenciar bookmarks
- [ ] **Field Codes**: Soportar TOC, PAGEREF, PAGE, NUMPAGES, SEQ, REF
- [ ] **TOC Dinámico**: Generar TOC con números de página clickeables
- [ ] **Heading Styles**: H1-H4 nativos con OutlineLvl correcto
- [ ] **Page Numbers**: Headers/footers con numeración dinámica
- [ ] **References**: Referencias cruzadas funcionales
- [ ] **Auto-Numbering**: Figuras y tablas numeradas automáticamente

### Tests de Calidad

```bash
# Todos los tests pasan
go test ./...

# Documento de prueba abre en Word sin errores
# - TOC se puede actualizar (clic derecho > Actualizar campo)
# - Navigation Pane muestra estructura
# - Hyperlinks funcionan
# - Números de página correctos
# - Referencias actualizan al cambiar contenido
```

---

## 📚 Recursos

### Documentación de OOXML
- [Office Open XML Spec](http://www.ecma-international.org/publications/standards/Ecma-376.htm)
- [Word Fields Reference](https://support.microsoft.com/en-us/office/field-codes-in-word)

### Análisis de DOCX Existentes
```bash
# Crear DOCX en Word con TOC, bookmarks, estilos
# Descomprimir y analizar XML
unzip -q document.docx -d document_xml/
cd document_xml/word
cat document.xml | xmllint --format - > document_formatted.xml
cat styles.xml | xmllint --format - > styles_formatted.xml

# Ver estructura de TOC
grep -A 20 "w:fldChar" document_formatted.xml
grep -A 10 "w:bookmarkStart" document_formatted.xml
```

---

## 🤝 Contribuir

### Workflow
1. Feature branch desde `slidelang-enhanced`
2. Implementar con tests
3. PR a `slidelang-enhanced`
4. Review y merge
5. Tag releases: `v0.1.0-slidelang`, `v0.2.0-slidelang`, etc.

### Sincronizar con Upstream
```bash
# Traer cambios del original (si hay)
git fetch upstream
git merge upstream/master

# Resolver conflictos si hay
# Push a nuestro fork
git push origin slidelang-enhanced
```

---

## 🎯 Prioridades

### Corto Plazo (2 semanas)
1. ✅ Bookmarks básicos
2. ✅ Field codes (fldChar)
3. ✅ TOC con page numbers
4. ✅ Heading styles con OutlineLvl

### Medio Plazo (1 mes)
5. ✅ Page numbers en headers/footers
6. ✅ Referencias cruzadas
7. ✅ Auto-numbering figuras/tablas
8. ✅ Mejor API de estilos

### Largo Plazo (2-3 meses)
9. ✅ Sections con headers/footers diferentes
10. ✅ Custom styles avanzados
11. ✅ Bibliography support
12. ✅ Track changes support

---

## 📞 Contacto

**Proyecto**: https://github.com/SlideLang/go-docx  
**Issues**: https://github.com/SlideLang/go-docx/issues  
**Discussions**: https://github.com/SlideLang/go-docx/discussions

**Proyecto principal**: https://github.com/SlideLang/cli
