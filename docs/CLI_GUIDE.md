# docxgo CLI Guide

The `docxgo` CLI binary exposes the full docxgo library API as a JSON-RPC service over stdin/stdout, enabling document creation and manipulation from any language.

## Table of Contents

- [Installation](#installation)
- [Execution Modes](#execution-modes)
- [JSON-RPC Protocol](#json-rpc-protocol)
- [Methods Reference](#methods-reference)
  - [system.ping](#systemping)
  - [system.version](#systemversion)
  - [system.capabilities](#systemcapabilities)
  - [system.batch](#systembatch)
  - [document.create](#documentcreate)
  - [document.open](#documentopen)
  - [document.save](#documentsave)
  - [document.validate](#documentvalidate)
  - [document.inspect](#documentinspect)
  - [document.setMetadata](#documentsetmetadata)
  - [document.setBackgroundColor](#documentsetbackgroundcolor)
  - [document.addContent](#documentaddcontent)
  - [document.addPageBreak](#documentaddpagebreak)
  - [document.applyPatch](#documentapplypatch)
  - [paragraph.add](#paragraphadd)
  - [paragraph.list](#paragraphlist)
  - [table.add](#tableadd)
  - [table.list](#tablelist)
  - [section.add](#sectionadd)
  - [template.inspect](#templateinspect)
  - [template.render](#templaterender)
  - [document.close](#documentclose)
- [Content Types](#content-types)
  - [Paragraph](#paragraph)
  - [Table](#table)
  - [Section](#section)
  - [Page Break](#page-break)
- [Shell Examples](#shell-examples)
- [Node.js Integration](#nodejs-integration)
- [Error Code Reference](#error-code-reference)

---

## Installation

Build the binary from source:

```bash
go install github.com/mmonterroca/docxgo/v2/cmd/docxgo@latest
```

Or build locally:

```bash
git clone https://github.com/mmonterroca/docxgo.git
cd docxgo
go build -o docxgo ./cmd/docxgo/
```

Verify the installation:

```bash
docxgo version
# 2.0.0-beta
```

---

## Execution Modes

### One-shot mode (`docxgo exec`)

Reads a single JSON-RPC request, executes it, writes the JSON response to stdout, and exits.

```bash
# From stdin:
echo '{"id":1,"method":"document.create","params":{...}}' | docxgo exec

# From flag:
docxgo exec --request '{"id":1,"method":"document.create","params":{...}}'
```

Ideal for: occasional calls via `child_process.execFile`, shell scripts, CI pipelines.

**Exit codes:**
- `0` — success
- `1` — error (the response will contain a JSON error object)

### RPC mode (`docxgo rpc`)

Reads newline-delimited JSON requests from stdin and writes newline-delimited JSON responses to stdout. The process stays alive between requests.

```bash
docxgo rpc
```

Diagnostic logs are written to **stderr** (not stdout), so they don't interfere with JSON responses.

Shutdown triggers:
- EOF on stdin (pipe closed)
- `SIGTERM` or `SIGINT` signal

Ideal for: high-frequency usage, Lambda warm starts, batch processing via `child_process.spawn`.

---

## JSON-RPC Protocol

### Request format

```json
{
  "id": 1,
  "method": "document.create",
  "params": { ... }
}
```

`id` can be any JSON value (number, string, null). It is echoed back in the response.

### Success response

```json
{
  "id": 1,
  "result": { ... }
}
```

### Error response

```json
{
  "id": 1,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "description of the error",
    "operation": "document.create"
  }
}
```

Some methods return enriched errors with a `data` field containing structured context:

```json
{
  "id": 1,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "unknown operation \"deleteAll\" at index 2",
    "operation": "document.applyPatch",
    "data": {
      "index": 2,
      "op": "deleteAll"
    }
  }
}
```

| `data` field | Type | Description |
|--------------|------|-------------|
| `index` | Number | Index of the failing operation (in batch/patch) |
| `category` | String | Error category (e.g. `"merge"`) |
| `retryable` | Boolean | Whether the error is retryable |
| `op` | String | The operation that failed |

---

## Methods Reference

### system.ping

Health check — verifies the RPC process is alive and responsive.

**Params:** None (or `{}`).

**Success result:**

```json
{ "status": "ok" }
```

---

### system.version

Returns version, protocol version, and platform information.

**Params:** None (or `{}`).

**Success result:**

```json
{
  "name": "docxgo",
  "version": "2.0.0-beta",
  "protocolVersion": "1.0",
  "goVersion": "go1.23.0",
  "platform": "darwin",
  "arch": "arm64"
}
```

---

### system.capabilities

Returns a map of supported features for the current binary.

**Params:** None (or `{}`).

**Success result:**

```json
{
  "rpc": true,
  "template": true,
  "mailMerge": true,
  "inspect": true,
  "validate": true,
  "batch": true,
  "applyPatch": true,
  "streaming": false,
  "partialUpdate": false
}
```

Use this for feature detection before calling advanced methods.

---

### system.batch

Executes multiple RPC requests in a single roundtrip. Each sub-request is processed sequentially. Nested `system.batch` calls are rejected.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `requests` | Array | Yes | Array of `{ method, params? }` objects |

**Success result:**

```json
{
  "responses": [
    { "result": { "status": "ok" } },
    { "result": { "name": "docxgo", "version": "2.0.0-beta", ... } },
    { "error": { "code": "NOT_FOUND", "message": "..." } }
  ]
}
```

Each entry in `responses` contains either `result` or `error`, matching the order of the input `requests` array.

**Example:**

```json
{
  "id": 1,
  "method": "system.batch",
  "params": {
    "requests": [
      { "method": "system.ping" },
      { "method": "document.create", "params": {
        "content": [{ "type": "paragraph", "runs": [{ "text": "Hello" }] }],
        "output": "buffer"
      }},
      { "method": "document.inspect", "params": { "documentId": "doc-1" } }
    ]
  }
}
```

---

### document.create

Creates a new Word document and returns it as base64 or saves it to a file.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `options` | Object | No | Document configuration (see below) |
| `content` | Array | No | Ordered list of content items |
| `output` | `"buffer"` \| `"file"` | No | Output format (default: `"buffer"`) |
| `filePath` | String | When `output="file"` | Path to write the .docx file |

**options fields:**

| Field | Type | Description |
|-------|------|-------------|
| `title` | String | Document title |
| `author` | String | Document author |
| `subject` | String | Document subject |
| `pageSize` | `"A4"` \| `"Letter"` \| `"Legal"` \| `"A3"` \| `"Tabloid"` \| `{width, height}` | Page dimensions (twips) |
| `margins` | `"normal"` \| `"narrow"` \| `"wide"` \| `{top, bottom, left, right}` | Page margins (twips) |
| `theme` | `"Corporate"` \| `"Startup"` \| `"Modern"` \| `"Fintech"` \| `"Academic"` | Apply a preset theme |

**Success result:**

```json
{
  "data": "<base64-encoded .docx>",
  "documentId": "doc-1"
}
```
(when `output="buffer"`)

```json
{
  "filePath": "/path/to/output.docx",
  "documentId": "doc-1"
}
```
(when `output="file"`)

The `documentId` can be used in subsequent RPC calls to modify or re-save the document.

**Example:**

```json
{
  "id": 1,
  "method": "document.create",
  "params": {
    "options": {
      "title": "My Report",
      "author": "Jane Smith",
      "pageSize": "A4",
      "margins": "normal",
      "theme": "Corporate"
    },
    "content": [
      {
        "type": "paragraph",
        "style": "Heading1",
        "runs": [{ "text": "Introduction" }]
      },
      {
        "type": "paragraph",
        "runs": [
          { "text": "This is " },
          { "text": "bold", "bold": true },
          { "text": " text." }
        ]
      }
    ],
    "output": "buffer"
  }
}
```

---

### document.open

Opens an existing .docx document and stores it in the session.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `filePath` | String | One of | Path to a .docx file on disk |
| `base64` | String | One of | Base64-encoded .docx bytes |

**Success result:**

```json
{ "documentId": "doc-2" }
```

---

### document.save

Saves or re-exports a document that was previously created or opened.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `output` | `"buffer"` \| `"file"` | No | Output format (default: `"buffer"`) |
| `filePath` | String | When `output="file"` | Destination file path |

**Success result:** Same structure as `document.create`.

---

### document.validate

Validates the document structure and returns the result.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:**

```json
{ "valid": true }
```

or

```json
{ "valid": false, "message": "document is empty" }
```

---

### document.inspect

Extracts metadata, text, and structural information from a document.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:**

```json
{
  "paragraphCount": 5,
  "tableCount": 1,
  "text": ["First paragraph text", "Second paragraph text"],
  "metadata": {
    "title": "My Document",
    "subject": "",
    "creator": "Jane Smith",
    "description": "",
    "keywords": null,
    "created": "",
    "modified": ""
  },
  "backgroundColor": "#E0F0FF"
}
```

(`metadata` is omitted when not set; `backgroundColor` is omitted when not set.)

---

### document.setMetadata

Updates the document metadata.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `title` | String | No | Document title |
| `subject` | String | No | Document subject |
| `creator` | String | No | Document author/creator |
| `description` | String | No | Document description |
| `keywords` | Array\<String\> | No | Keywords list |
| `created` | String | No | Creation date (ISO 8601) |
| `modified` | String | No | Modification date (ISO 8601) |

**Success result:** `{ "ok": true }`

---

### document.setBackgroundColor

Sets the page background color for the entire document.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `color` | String | Yes | Hex color string (e.g. `"#E8F0FE"` or `"E8F0FE"`) |

**Success result:** `{ "ok": true }`

---

### document.addContent

Appends content to an existing document session. Accepts the same content array format as `document.create`. This is the primary method for mutating documents that were opened via `document.open`.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `content` | Array | Yes | Ordered list of content items (same format as `document.create`) |

**Success result:** `{ "ok": true }`

**Example:**

```json
{
  "id": 5,
  "method": "document.addContent",
  "params": {
    "documentId": "doc-1",
    "content": [
      {
        "type": "paragraph",
        "runs": [{ "text": "Appended paragraph", "bold": true }]
      },
      { "type": "pageBreak" },
      {
        "type": "table",
        "rows": [
          { "cells": [{ "paragraphs": [{ "runs": [{ "text": "A1" }] }] }] }
        ]
      }
    ]
  }
}
```

---

### document.addPageBreak

Adds a page break to an existing document.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:** `{ "ok": true }`

---

### document.applyPatch

Applies a sequence of patch operations to an existing document sequentially. If any operation fails, subsequent operations are **not** applied and the error includes the failing index. Operations applied before the failure remain in effect (no rollback).

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `operations` | Array | Yes | Array of patch operation objects |

**Supported operations:**

| `op` value | Description | Additional fields |
|------------|-------------|-------------------|
| `appendParagraph` | Append a paragraph | Same fields as `paragraph.add` (style, alignment, runs, etc.) |
| `appendTable` | Append a table | Same fields as `table.add` (rows, style, alignment, width) |
| `appendSection` | Append a section break | Same fields as `section.add` (breakType, pageSize, orientation, etc.) |
| `appendPageBreak` | Append a page break | None |
| `setMetadata` | Set document metadata | Same fields as `document.setMetadata` (title, creator, etc.) |
| `setBackgroundColor` | Set background color | `color` (hex string) |

**Success result:**

```json
{ "ok": true, "applied": 3 }
```

**Error result (with enriched `data`):**

```json
{
  "code": "VALIDATION_ERROR",
  "message": "unknown operation \"deleteAll\" at index 2",
  "operation": "document.applyPatch",
  "data": { "index": 2, "op": "deleteAll" }
}
```

**Example:**

```json
{
  "id": 10,
  "method": "document.applyPatch",
  "params": {
    "documentId": "doc-1",
    "operations": [
      {
        "op": "appendParagraph",
        "style": "Heading1",
        "runs": [{ "text": "New Section" }]
      },
      { "op": "appendPageBreak" },
      {
        "op": "appendTable",
        "rows": [
          { "cells": [{ "paragraphs": [{ "runs": [{ "text": "A1" }] }] }] }
        ]
      },
      { "op": "setMetadata", "title": "Updated Title" },
      { "op": "setBackgroundColor", "color": "#F0F8FF" }
    ]
  }
}
```

---

### paragraph.add

Adds a single paragraph to an existing document. Supports the same paragraph properties as the content array (style, alignment, spacing, runs, etc.).

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `style` | String | No | Paragraph style name |
| `alignment` | String | No | `left`, `center`, `right`, `justify`, `distribute` |
| `spacingBefore` | Number | No | Spacing before (twips) |
| `spacingAfter` | Number | No | Spacing after (twips) |
| `lineSpacing` | Object | No | `{ "rule": "auto", "value": 360 }` |
| `indent` | Object | No | `{ "left", "right", "firstLine", "hanging" }` |
| `numbering` | Object | No | `{ "id": 1, "level": 0 }` |
| `borders` | Object | No | Paragraph borders |
| `runs` | Array | No | Text runs (same format as content paragraphs) |

**Success result:**

```json
{ "ok": true, "index": 3 }
```

`index` is the zero-based position of the new paragraph.

**Example:**

```json
{
  "id": 6,
  "method": "paragraph.add",
  "params": {
    "documentId": "doc-1",
    "style": "Heading1",
    "alignment": "center",
    "runs": [
      { "text": "New Section Title", "bold": true, "fontSize": 18 }
    ]
  }
}
```

---

### paragraph.list

Lists all paragraphs in a document with their text and style.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:**

```json
{
  "count": 3,
  "paragraphs": [
    { "index": 0, "text": "Introduction", "style": "Heading1" },
    { "index": 1, "text": "Some body text." },
    { "index": 2, "text": "" }
  ]
}
```

---

### table.add

Adds a table to an existing document. Uses the same table format as the content array.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `rows` | Array | Yes | Table rows (same format as content tables) |
| `style` | String | No | Table style name |
| `alignment` | String | No | Table alignment |
| `width` | Object | No | `{ "type": "dxa", "value": 9000 }` |

**Success result:**

```json
{ "ok": true, "index": 0 }
```

`index` is the zero-based position of the new table.

**Example:**

```json
{
  "id": 7,
  "method": "table.add",
  "params": {
    "documentId": "doc-1",
    "style": "TableGrid",
    "rows": [
      {
        "cells": [
          { "paragraphs": [{ "runs": [{ "text": "Name", "bold": true }] }] },
          { "paragraphs": [{ "runs": [{ "text": "Value", "bold": true }] }] }
        ]
      },
      {
        "cells": [
          { "paragraphs": [{ "runs": [{ "text": "Score" }] }] },
          { "paragraphs": [{ "runs": [{ "text": "95" }] }] }
        ]
      }
    ]
  }
}
```

---

### table.list

Lists all tables in a document with their dimensions.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:**

```json
{
  "count": 2,
  "tables": [
    { "index": 0, "rows": 3, "columns": 2 },
    { "index": 1, "rows": 5, "columns": 4 }
  ]
}
```

---

### section.add

Adds a new section to an existing document. Supports page size, margins, orientation, columns, and headers/footers.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `breakType` | String | No | `nextPage` (default), `continuous`, `evenPage`, `oddPage` |
| `pageSize` | String/Object | No | Page size preset or `{width, height}` |
| `margins` | String/Object | No | Margins preset or `{top, bottom, left, right}` |
| `orientation` | String | No | `portrait` or `landscape` |
| `columns` | Number | No | Number of text columns |
| `headers` | Object | No | Headers by type (`default`, `first`, `even`) |
| `footers` | Object | No | Footers by type (`default`, `first`, `even`) |

**Success result:**

```json
{ "ok": true, "index": 1 }
```

`index` is the zero-based position of the new section.

**Example:**

```json
{
  "id": 8,
  "method": "section.add",
  "params": {
    "documentId": "doc-1",
    "breakType": "nextPage",
    "pageSize": "A4",
    "orientation": "landscape",
    "columns": 2
  }
}
```

---

### template.inspect

Scans a document for template placeholders (default: `{{key}}`) and returns detailed information about each occurrence.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `openDelimiter` | String | No | Custom open delimiter (default: `"{{"`) |
| `closeDelimiter` | String | No | Custom close delimiter (default: `"}}"`) |

**Success result:**

```json
{
  "placeholders": ["Name", "Company", "Role"],
  "count": 3,
  "occurrences": 4,
  "details": [
    {
      "name": "Name",
      "fullMatch": "{{Name}}",
      "location": "paragraph",
      "paragraph": 0,
      "run": 0
    },
    {
      "name": "Company",
      "fullMatch": "{{Company}}",
      "location": "tableCell",
      "paragraph": 0,
      "run": 0,
      "table": 0,
      "row": 1,
      "cell": 0
    }
  ]
}
```

| Result field | Type | Description |
|--------------|------|-------------|
| `placeholders` | Array\<String\> | Unique placeholder names (first-seen order) |
| `count` | Number | Number of unique placeholders |
| `occurrences` | Number | Total number of placeholder instances |
| `details` | Array | Per-occurrence location details |

**Location types:** `paragraph`, `tableCell`, `header`, `footer`

**Example:**

```json
{
  "id": 11,
  "method": "template.inspect",
  "params": {
    "documentId": "doc-1"
  }
}
```

---

### template.render

Replaces template placeholders in the document with the provided data values. Optionally validates all placeholders are covered (strict mode).

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |
| `data` | Object | Yes | Key-value map of replacements (`{ "Name": "Alice" }`) |
| `strictMode` | Boolean | No | If `true`, fail on missing keys (default: `false`) |
| `openDelimiter` | String | No | Custom open delimiter (default: `"{{"`) |
| `closeDelimiter` | String | No | Custom close delimiter (default: `"}}"`) |

**Success result:**

```json
{ "ok": true }
```

With validation warnings:

```json
{
  "ok": true,
  "warnings": [
    { "severity": "warning", "key": "OptionalField", "message": "key OptionalField not found in data" }
  ]
}
```

**Error (strict mode, missing key):**

```json
{
  "code": "TEMPLATE_ERROR",
  "message": "template: missing keys: Code",
  "operation": "template.render",
  "data": { "category": "merge", "retryable": false }
}
```

**Example:**

```json
{
  "id": 12,
  "method": "template.render",
  "params": {
    "documentId": "doc-1",
    "data": {
      "Name": "Alice Johnson",
      "Company": "Acme Corp",
      "Date": "2025-01-15"
    },
    "strictMode": true
  }
}
```

---

### document.close

Removes a document from the session, freeing associated memory. Should be called when a document is no longer needed in RPC mode.

**Params:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `documentId` | String | Yes | Session document ID |

**Success result:** `{ "ok": true }`

---

## Content Types

Content items are passed as an array in `document.create` and `document.addContent` params. Each item has a `type` field.

### Paragraph

```json
{
  "type": "paragraph",
  "style": "Heading1",
  "alignment": "center",
  "spacingBefore": 240,
  "spacingAfter": 120,
  "lineSpacing": { "rule": "auto", "value": 360 },
  "indent": { "left": 720, "right": 0, "firstLine": 0, "hanging": 0 },
  "numbering": { "id": 1, "level": 0 },
  "borders": {
    "bottom": { "style": "single", "width": 6, "color": "#000000" }
  },
  "runs": [ ... ]
}
```

**Style values:** `Normal`, `Heading1`–`Heading9`, `Title`, `Subtitle`, `Quote`, `IntenseQuote`, `ListParagraph`, `Caption`, `BodyText`, `NoSpacing`, etc.

**Alignment values:** `left`, `center`, `right`, `justify`, `distribute`

**Line spacing rules:** `auto`, `exact`, `atLeast`

**Run fields:**

| Field | Type | Description |
|-------|------|-------------|
| `text` | String | Text content |
| `bold` | Boolean | Bold formatting |
| `italic` | Boolean | Italic formatting |
| `strike` | Boolean | Strikethrough |
| `underline` | `"single"` \| `"double"` \| `"thick"` \| `"dotted"` \| `"dashed"` \| `"wave"` \| `"none"` | Underline style |
| `color` | String | Text color (hex) |
| `fontSize` | Number | Font size in **points** |
| `font` | String | Font name (e.g. `"Arial"`) |
| `highlight` | String | Highlight color (see below) |
| `hyperlink` | `{ "url": "...", "displayText": "..." }` | Inline hyperlink |
| `field` | Object | Document field (see below) |
| `image` | Object | Image (see below) |
| `break` | `"page"` \| `"column"` \| `"line"` | Insert a break |

**Highlight colors:** `yellow`, `green`, `cyan`, `magenta`, `blue`, `red`, `darkBlue`, `darkCyan`, `darkGreen`, `darkMagenta`, `darkRed`, `darkYellow`, `darkGray`, `lightGray`, `none`

**Field object:**

```json
{
  "type": "pageNumber"
}
```

Field types: `pageNumber`, `pageCount`, `toc`, `hyperlink`, `styleRef`

TOC field with options:
```json
{
  "type": "toc",
  "options": { "levels": "1-3", "hyperlinks": "true" }
}
```

Hyperlink field:
```json
{
  "type": "hyperlink",
  "url": "https://example.com",
  "display": "Click here"
}
```

Style reference field (for running headers):
```json
{
  "type": "styleRef",
  "style": "Heading 1"
}
```

**Image object:**

```json
{
  "path": "/path/to/image.png",
  "widthPx": 400,
  "heightPx": 300
}
```

Or with base64 data:
```json
{
  "base64": "<base64-encoded image data>",
  "format": "png",
  "widthPx": 400,
  "heightPx": 300
}
```

---

### Table

```json
{
  "type": "table",
  "style": "TableGrid",
  "alignment": "left",
  "width": { "type": "dxa", "value": 9000 },
  "rows": [
    {
      "height": 400,
      "cells": [
        {
          "width": 3000,
          "verticalAlignment": "top",
          "shading": "#F0F0F0",
          "gridSpan": 2,
          "borders": {
            "top": { "style": "single", "width": 6, "color": "#000000" }
          },
          "paragraphs": [
            {
              "type": "paragraph",
              "runs": [{ "text": "Cell content", "bold": true }]
            }
          ]
        }
      ]
    }
  ]
}
```

**Table styles:** `TableNormal`, `TableGrid`, `PlainTable1`, `MediumShading1`, `LightShading`, `ColorfulList`

**Width types:** `auto`, `dxa` (twips), `pct` (percent × 50)

**Vertical alignment:** `top`, `center`, `bottom`

**Border styles:** `single`, `dotted`, `dashed`, `double`, `triple`, `thick`, `none`

---

### Section

```json
{
  "type": "section",
  "breakType": "nextPage",
  "pageSize": "A4",
  "margins": { "top": 1440, "bottom": 1440, "left": 1440, "right": 1440 },
  "orientation": "portrait",
  "columns": 2,
  "headers": {
    "default": [
      {
        "type": "paragraph",
        "alignment": "right",
        "runs": [{ "text": "My Document" }]
      }
    ]
  },
  "footers": {
    "default": [
      {
        "type": "paragraph",
        "alignment": "center",
        "runs": [{ "field": { "type": "pageNumber" } }]
      }
    ]
  }
}
```

**Break types:** `nextPage` (default), `continuous`, `evenPage`, `oddPage`

**Header/footer keys:** `default`, `first`, `even`

---

### Page Break

```json
{ "type": "pageBreak" }
```

---

## Shell Examples

### Create a simple document

```bash
echo '{
  "id": 1,
  "method": "document.create",
  "params": {
    "options": { "title": "Hello World" },
    "content": [
      {
        "type": "paragraph",
        "runs": [{ "text": "Hello, World!", "bold": true }]
      }
    ],
    "output": "file",
    "filePath": "/tmp/hello.docx"
  }
}' | docxgo exec
```

### Inspect an existing document

```bash
echo '{
  "id": 1,
  "method": "document.open",
  "params": { "filePath": "/path/to/existing.docx" }
}' | docxgo exec
# → {"id":1,"result":{"documentId":"doc-1"}}
```

### Get document info (two-step in exec mode, or one session in RPC)

Use RPC mode for multi-step workflows:

```bash
printf '{"id":1,"method":"document.open","params":{"filePath":"/path/to/doc.docx"}}\n{"id":2,"method":"document.inspect","params":{"documentId":"doc-1"}}\n' \
  | docxgo rpc 2>/dev/null
```

---

## Node.js Integration

### One-shot mode (exec)

```javascript
const { execFile } = require('child_process');
const path = require('path');

function createDocument(params) {
  return new Promise((resolve, reject) => {
    const request = JSON.stringify({ id: 1, method: 'document.create', params });
    execFile('docxgo', ['exec', '--request', request], (err, stdout, stderr) => {
      if (err) return reject(err);
      const response = JSON.parse(stdout);
      if (response.error) return reject(new Error(response.error.message));
      resolve(response.result);
    });
  });
}

// Usage:
createDocument({
  options: { title: 'My Doc', pageSize: 'A4' },
  content: [
    { type: 'paragraph', runs: [{ text: 'Hello!', bold: true }] }
  ],
  output: 'buffer'
}).then(result => {
  const buf = Buffer.from(result.data, 'base64');
  require('fs').writeFileSync('output.docx', buf);
  console.log('Created:', buf.length, 'bytes');
});
```

### RPC mode (spawn)

```javascript
const { spawn } = require('child_process');
const readline = require('readline');

class DocxgoRPC {
  constructor(binaryPath = 'docxgo') {
    this._proc = spawn(binaryPath, ['rpc']);
    this._rl = readline.createInterface({ input: this._proc.stdout });
    this._pending = new Map();
    this._rl.on('line', line => {
      const resp = JSON.parse(line);
      const resolve = this._pending.get(resp.id);
      if (resolve) {
        this._pending.delete(resp.id);
        resolve(resp);
      }
    });
    this._seq = 0;
  }

  call(method, params) {
    return new Promise((resolve, reject) => {
      const id = ++this._seq;
      this._pending.set(id, resp => {
        if (resp.error) reject(new Error(resp.error.message));
        else resolve(resp.result);
      });
      this._proc.stdin.write(JSON.stringify({ id, method, params }) + '\n');
    });
  }

  close() {
    this._proc.stdin.end();
  }
}

// Usage:
async function main() {
  const rpc = new DocxgoRPC();

  const result = await rpc.call('document.create', {
    options: { title: 'Batch Document', pageSize: 'Letter' },
    content: [
      { type: 'paragraph', style: 'Heading1', runs: [{ text: 'Report' }] },
      { type: 'paragraph', runs: [{ text: 'Generated via RPC.' }] }
    ],
    output: 'buffer'
  });

  const buf = Buffer.from(result.data, 'base64');
  require('fs').writeFileSync('report.docx', buf);
  console.log('Saved report.docx:', buf.length, 'bytes');

  rpc.close();
}

main().catch(console.error);
```

---

## Error Code Reference

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid input parameters or document structure |
| `NOT_FOUND` | Referenced document ID does not exist in the session |
| `IO_ERROR` | File read/write failure |
| `INTERNAL_ERROR` | Unexpected internal error |
| `METHOD_NOT_FOUND` | Unknown RPC method |
| `PARSE_ERROR` | Malformed JSON in the request |
| `TEMPLATE_ERROR` | Template merge/validation failure |
