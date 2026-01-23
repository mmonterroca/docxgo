# Troubleshooting DOCX Validation Errors

This document describes the OOXML validation issues encountered during go-docx v2 development and the solutions implemented. It serves as a guide for diagnosing and resolving similar issues in the future.

## Table of Contents

1. [Diagnostic Tools](#diagnostic-tools)
2. [Issue 1: Invalid Hyperlink RelationshipID](#issue-1-invalid-hyperlink-relationshipid)
3. [Issue 2: Empty wp:align Value](#issue-2-empty-wpalign-value)
4. [Known and Tolerable Errors](#known-and-tolerable-errors)
5. [General Diagnostic Workflow](#general-diagnostic-workflow)

---

## Diagnostic Tools

### DocxValidator (C#)

Location: `DocxValidator/`

```powershell
cd DocxValidator
dotnet run -- "path/to/document.docx"
```

This validator uses the OpenXML SDK to detect schema errors. It provides:
- Total error count
- Description of each error
- Affected document part (document.xml, header1.xml, etc.)
- Exact XPath of the problematic element

### Manual DOCX Extraction

A .docx file is a ZIP archive. To inspect the internal XML:

```powershell
# Extract contents
Expand-Archive -Path "document.docx" -DestinationPath "document_debug" -Force

# Search for specific patterns
Select-String -Path "document_debug\word\document.xml" -Pattern "wp:positionV" -Context 2,2
```

### Document Comparison

To identify what changed during a round-trip:

```powershell
# Extract original and generated
Expand-Archive -Path "original.docx" -DestinationPath "original_debug" -Force
Expand-Archive -Path "generated.docx" -DestinationPath "generated_debug" -Force

# Compare specific files
Compare-Object (Get-Content "original_debug\word\document.xml") `
               (Get-Content "generated_debug\word\document.xml")
```

---

## Issue 1: Invalid Hyperlink RelationshipID

### Symptom

```
The relationship 'rId37' was not found in 'word/document.xml'.
```

Word cannot open the document because it references relationship IDs that don't exist.

### Root Cause

When reading an existing document, external hyperlinks preserve their original `relationshipID` (e.g., `rId25`). However, during serialization, the `AddField()` method in `run.go` **always** called `AddHyperlink()`, which generated a **new** relationship ID (e.g., `rId37`), overwriting the preserved one.

This caused:
- The XML referenced `rId37`
- But `document.xml.rels` only had the original `rId25`

### Diagnosis

1. Search for hyperlink IDs in the generated document:
   ```powershell
   Select-String -Path "generated_debug\word\document.xml" -Pattern 'w:hyperlink r:id="rId'
   ```

2. Check which relationships exist:
   ```powershell
   Get-Content "generated_debug\word\_rels\document.xml.rels"
   ```

3. If the IDs don't match, the problem is in serialization.

### Solution

**File:** `internal/core/run.go` (lines 238-257)

Before generating a new relationship, check if a preserved one already exists:

```go
// Check if this hyperlink already has a relationship ID (preserved from read)
existingRelID, hasExistingRelID := accessor.GetProperty("relationshipID")
if hasExistingRelID && existingRelID != "" {
    // Already has a relationship ID, skip creating new one
    // The preserved ID will be used during serialization
} else {
    // No existing relationship ID, create a new one
    if relMgr := accessor.RelationshipManager(); relMgr != nil {
        relID, err := relMgr.AddHyperlink(link)
        if err != nil {
            return fmt.Errorf("failed to add hyperlink relationship: %w", err)
        }
        accessor.SetProperty("relationshipID", relID)
    }
}
```

### Related Files

- `internal/reader/reconstruct.go` - `hydrateHyperlink()`: Preserves `originalRelID` via `SetProperty`
- `internal/serializer/serializer.go` - `expandRunWithFields()`: Uses the preserved relationshipID
- `internal/core/run.go` - `AddField()`: Where the duplicate ID was being generated

---

## Issue 2: Empty wp:align Value

### Symptom

```
The element 'wp:align' has invalid value ''
Part: document.xml
Path: /w:document[1]/w:body[1]/w:p[4]/w:r[1]/w:drawing[1]/wp:anchor[1]/wp:positionV[1]/wp:align[1]
```

### Root Cause

In OOXML, `wp:positionH` and `wp:positionV` elements can contain:
- `<wp:posOffset>` - Numeric offset in EMUs (can be 0)
- `<wp:align>` - Named alignment ("left", "center", "top", etc.)

**But only one of them, never both, and never empty.**

The problem occurred when:
1. The original document had `<wp:posOffset>0</wp:posOffset>` (explicit offset of 0)
2. When reading, `pos.OffsetY = 0` (which is Go's default value)
3. When serializing, the condition `if pos.OffsetY != 0` was FALSE
4. So it fell through to the `else` which tried to use `pos.VAlign` (which was empty)
5. Result: `<wp:align></wp:align>` - **invalid**

### Diagnosis

1. Validate the original document:
   ```powershell
   cd DocxValidator
   dotnet run -- "original.docx"
   ```

2. If the original does NOT have the error but the generated one DOES, the problem is in the round-trip.

3. Compare the positionV XML structure:
   ```powershell
   Select-String -Path "original_debug\word\document.xml" -Pattern "wp:positionV" -Context 0,3
   Select-String -Path "generated_debug\word\document.xml" -Pattern "wp:positionV" -Context 0,3
   ```

### Solution

Two changes are needed:

#### 1. Add Flags to ImagePosition

**File:** `domain/image.go`

```go
type ImagePosition struct {
    Type       ImagePositionType
    HAlign     HorizontalAlign
    VAlign     VerticalAlign
    OffsetX    int
    OffsetY    int
    UseOffsetX bool  // True if OffsetX should be used (even if 0)
    UseOffsetY bool  // True if OffsetY should be used (even if 0)
    WrapText   TextWrapType
    ZOrder     int
    BehindText bool
}
```

#### 2. Set Flags When Reading

**File:** `internal/reader/reconstruct.go` - `buildFloatingPosition()`

```go
if positionH := findChild(elem, "positionH"); positionH != nil {
    if align := findChild(positionH, "align"); align != nil {
        if mapped, ok := mapHorizontalAlignValue(strings.TrimSpace(align.Text)); ok {
            pos.HAlign = mapped
        }
    }
    if offset, ok := parseChildInt(positionH, "posOffset"); ok {
        pos.OffsetX = offset
        pos.UseOffsetX = true  // ← Mark that we should use offset
    }
}

if positionV := findChild(elem, "positionV"); positionV != nil {
    if align := findChild(positionV, "align"); align != nil {
        if mapped, ok := mapVerticalAlignValue(strings.TrimSpace(align.Text)); ok {
            pos.VAlign = mapped
        }
    }
    if offset, ok := parseChildInt(positionV, "posOffset"); ok {
        pos.OffsetY = offset
        pos.UseOffsetY = true  // ← Mark that we should use offset
    }
}
```

#### 3. Use Flags When Serializing

**File:** `internal/xml/drawing_helper.go` - `NewFloatingDrawing()`

```go
// Set horizontal position
anchor.PositionH = &PositionH{
    RelativeFrom: convertHAlign(pos.HAlign),
}
// Use offset if explicitly set (even if 0), otherwise use alignment if provided
if pos.UseOffsetX || pos.OffsetX != 0 {
    offset := pos.OffsetX
    anchor.PositionH.PosOffset = &offset
} else if pos.HAlign != "" {
    align := string(pos.HAlign)
    anchor.PositionH.Align = &align
}

// Set vertical position
anchor.PositionV = &PositionV{
    RelativeFrom: convertVAlign(pos.VAlign),
}
// Use offset if explicitly set (even if 0), otherwise use alignment if provided
if pos.UseOffsetY || pos.OffsetY != 0 {
    offset := pos.OffsetY
    anchor.PositionV.PosOffset = &offset
} else if pos.VAlign != "" {
    align := string(pos.VAlign)
    anchor.PositionV.Align = &align
}
```

### Lesson Learned

> **In Go, the zero value (`0`, `""`, `false`) is indistinguishable from "not set".**
> 
> When a value of 0 is semantically different from "no value", an additional boolean flag is needed to track whether the value was explicitly set.

---

## Known and Tolerable Errors

### tblLook in Headers (typically 60 errors)

```
The 'http://schemas.openxmlformats.org/wordprocessingml/2006/main:firstRow' attribute is not declared.
Part: /word/headerX.xml
```

These errors come from original documents created by Word and are **tolerated by Word**. They are Word 2010+ attributes that the OpenXML validator doesn't recognize because it uses Word 2007 schemas.

**Action:** Ignore if present in the original document.

---

## General Diagnostic Workflow

```
┌──────────────────────────────────────┐
│ 1. Validate original document        │
│    DocxValidator original.docx       │
└──────────────────┬───────────────────┘
                   ↓
┌──────────────────────────────────────┐
│ 2. Validate generated document       │
│    DocxValidator generated.docx      │
└──────────────────┬───────────────────┘
                   ↓
┌──────────────────────────────────────┐
│ 3. Compare errors                    │
│    Are there new errors?             │
└──────────────────┬───────────────────┘
                   ↓
        ┌─────────┴─────────┐
        │ YES               │ NO
        ↓                   ↓
┌───────────────────┐  ┌───────────────────┐
│ 4. Extract both   │  │ Errors are from   │
│    documents      │  │ original, ignore  │
└────────┬──────────┘  └───────────────────┘
         ↓
┌──────────────────────────────────────┐
│ 5. Search for error XPath            │
│    in both XMLs                      │
└──────────────────┬───────────────────┘
         ↓
┌──────────────────────────────────────┐
│ 6. Identify the difference           │
│    What changed during round-trip?   │
└──────────────────┬───────────────────┘
         ↓
┌──────────────────────────────────────┐
│ 7. Locate in code                    │
│    grep_search for the element       │
└──────────────────┬───────────────────┘
         ↓
┌──────────────────────────────────────┐
│ 8. Trace read → serialize flow       │
│    Where is the value lost/changed?  │
└──────────────────────────────────────┘
```

---

## Pre-Release Validation Checklist

- [ ] Run DocxValidator on all examples
- [ ] Compare errors with original documents
- [ ] Verify no new errors are introduced
- [ ] Test opening in Word
- [ ] Run all unit tests

---

## References

- [ECMA-376 Office Open XML](https://www.ecma-international.org/publications-and-standards/standards/ecma-376/)
- [DrawingML Positioning](https://docs.microsoft.com/en-us/dotnet/api/documentformat.openxml.drawing.wordprocessing)
- [OpenXML SDK Validation](https://docs.microsoft.com/en-us/office/open-xml/how-to-validate-a-word-processing-document)

---

*Document created: January 2026*
*Last updated: January 2026*
*Author: mmonterroca*
