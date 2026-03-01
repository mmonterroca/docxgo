# @mmonterroca/docxgo

Node.js wrapper for [docxgo](https://github.com/mmonterroca/docxgo) — create and manipulate Word (.docx) documents from JavaScript and TypeScript.

## Features

- **Three client modes**: Sync one-shot (`DocxgoExec`), async persistent (`DocxgoRPC`), and fluent builder (`DocumentBuilder`)
- **Template engine**: Inspect and render `{{placeholder}}` templates with strict mode validation
- **Patch operations**: Apply atomic multi-operation patches (`appendParagraph`, `setMetadata`, etc.)
- **Batch requests**: Execute multiple RPC calls in a single roundtrip
- **Full TypeScript support**: Complete type definitions for all RPC methods, options, and content types
- **Cross-platform binaries**: Automatic binary resolution for macOS, Linux, and Windows (x64 & arm64)
- **CJS + ESM**: Dual module output — works with `require()` and `import`

## Installation

```bash
npm install @mmonterroca/docxgo
```

Platform-specific binaries are installed automatically via `optionalDependencies`.

### Manual binary

If a pre-built binary is not available for your platform, build from source:

```bash
git clone https://github.com/mmonterroca/docxgo.git
cd docxgo
go build -o docxgo ./cmd/docxgo
```

Point the wrapper to the binary:

```bash
export DOCXGO_BIN=/path/to/docxgo
```

Or pass the path directly:

```ts
const doc = new DocumentBuilder({ binaryPath: '/path/to/docxgo' });
```

## Quick Start

### DocumentBuilder (recommended)

The fluent builder API is the easiest way to create documents:

```ts
import { DocumentBuilder } from '@mmonterroca/docxgo';

const doc = new DocumentBuilder();

// Create a document from scratch
await doc
  .setTitle('Quarterly Report')
  .setAuthor('Jane Smith')
  .addHeading('Q3 Results', 1)
  .addParagraph('Revenue exceeded targets by 15%.', { bold: true })
  .addTable([
    [{ text: 'Metric', bold: true }, { text: 'Value', bold: true }],
    [{ text: 'Revenue' }, { text: '$1.2M' }],
    [{ text: 'Growth' }, { text: '15%' }],
  ])
  .addPageBreak()
  .addHeading('Details', 2)
  .addParagraph('See the full breakdown below.')
  .createToFile('/tmp/report.docx');

doc.dispose();
```

### Open and modify existing documents

```ts
import { DocumentBuilder } from '@mmonterroca/docxgo';

const doc = new DocumentBuilder();

await doc.open('/path/to/existing.docx');

// Add new content
doc.addHeading('Appendix', 1);
doc.addParagraph('Added after the original content.');

await doc.saveToFile('/path/to/modified.docx');

// Inspect the document
const info = await doc.inspect();
console.log(info.metadata);
console.log(`${info.paragraphCount} paragraphs, ${info.tableCount} tables`);

await doc.closeDocument();
doc.dispose();
```

### Template rendering (mail merge)

```ts
import { DocumentBuilder } from '@mmonterroca/docxgo';

const doc = new DocumentBuilder();

// Create a template document
await doc
  .addHeading('Invoice for {{CustomerName}}')
  .addParagraph('Date: {{Date}}')
  .addParagraph('Total: {{Amount}}')
  .createToFile('/tmp/template.docx');
doc.reset();

// Open the template and inspect placeholders
await doc.open('/tmp/template.docx');

const placeholders = await doc.inspectTemplate();
console.log(placeholders.placeholders); // ['CustomerName', 'Date', 'Amount']
console.log(`${placeholders.count} unique, ${placeholders.occurrences} total`);

// Render with data
const result = await doc.renderTemplate({
  CustomerName: 'Acme Corp',
  Date: '2025-01-15',
  Amount: '$1,234.56',
});
console.log(result.ok); // true

await doc.saveToFile('/tmp/invoice.docx');
await doc.closeDocument();
doc.dispose();
```

### Patch operations

```ts
import { DocumentBuilder } from '@mmonterroca/docxgo';

const doc = new DocumentBuilder();
await doc.open('/path/to/existing.docx');

// Apply multiple atomic operations
const result = await doc.applyPatch([
  { op: 'appendParagraph', style: 'Heading1', runs: [{ text: 'New Section' }] },
  { op: 'appendPageBreak' },
  { op: 'appendTable', rows: [
    { cells: [{ paragraphs: [{ runs: [{ text: 'A' }] }] }] }
  ]},
  { op: 'setMetadata', title: 'Updated Title' },
  { op: 'setBackgroundColor', color: '#F0F8FF' },
]);
console.log(`Applied ${result.applied} operations`);

await doc.saveToFile('/path/to/patched.docx');
await doc.closeDocument();
doc.dispose();
```

### Get the document as a Buffer

```ts
const doc = new DocumentBuilder();

const buffer = await doc
  .addHeading('Hello World')
  .addParagraph('Created in Node.js!')
  .toBuffer();

// Use the buffer (e.g., send in HTTP response)
res.setHeader('Content-Type', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document');
res.send(buffer);

doc.dispose();
```

## API Reference

### `DocumentBuilder`

Fluent API wrapping a persistent RPC connection.

#### Constructor

```ts
new DocumentBuilder(options?: DocumentBuilderOptions)
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `binaryPath` | `string` | auto-detect | Path to the `docxgo` binary |
| `timeout` | `number` | `30000` | RPC call timeout in ms |

#### Document Options

| Method | Description |
|--------|-------------|
| `setOptions(opts)` | Set document-level options (title, author, pageSize, etc.) |
| `setTitle(title)` | Set document title |
| `setAuthor(author)` | Set document author |
| `setPageSize(size)` | Set page size (`'letter'`, `'A4'`, or `{width, height}`) |
| `setMargins(margins)` | Set page margins (`'normal'`, `'narrow'`, or custom) |

#### Content Building

| Method | Description |
|--------|-------------|
| `addParagraph(text, formatting?)` | Add a paragraph with optional run formatting |
| `addHeading(text, level?)` | Add a heading (1–9, default: 1) |
| `addFormattedParagraph(runs, options?)` | Add a paragraph with multiple formatted runs |
| `addTable(rows, options?)` | Add a table from a 2D array |
| `addRawTable(rows, options?)` | Add a table with full row/cell control |
| `addPageBreak()` | Insert a page break |
| `addSection(options?)` | Add a section break |
| `addContent(item)` | Add a raw `ContentItem` |
| `addContentItems(items)` | Add multiple raw content items |

#### Create & Save

| Method | Returns | Description |
|--------|---------|-------------|
| `create()` | `BufferResult` | Create document as base64 |
| `createToFile(path)` | `FileResult` | Create and write to file |
| `toBuffer()` | `Buffer` | Create as a Node.js Buffer |

#### Open & Modify

| Method | Returns | Description |
|--------|---------|-------------|
| `open(filePath)` | `string` | Open a .docx from path |
| `openFromBuffer(buffer)` | `string` | Open from Buffer |
| `openFromBase64(base64)` | `string` | Open from base64 |
| `appendContent()` | `void` | Flush queued content to document |
| `saveToFile(path)` | `FileResult` | Save to file |
| `saveToBuffer()` | `BufferResult` | Save as base64 |
| `inspect()` | `InspectResult` | Get document metadata & structure |
| `validate()` | `ValidateResult` | Validate document structure |
| `listParagraphs()` | `ParagraphListResult` | List all paragraphs |
| `listTables()` | `TableListResult` | List all tables |
| `closeDocument()` | `void` | Close and free document resources |

#### System & Discovery

| Method | Returns | Description |
|--------|---------|-------------|
| `ping()` | `PingResult` | Health check — verify RPC process is alive |
| `version()` | `SystemVersionResult` | Get binary version, protocol version, platform info |
| `capabilities()` | `SystemCapabilitiesResult` | Get map of supported features |
| `batch(requests)` | `BatchResult` | Execute multiple RPC calls in one roundtrip |

#### Template Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `inspectTemplate(options?)` | `TemplateInspectResult` | Find `{{placeholders}}` with location details |
| `renderTemplate(data, options?)` | `TemplateRenderResult` | Replace placeholders with data values |

#### Patch Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `applyPatch(operations)` | `ApplyPatchResult` | Apply atomic multi-operation patches |

Available patch operations: `appendParagraph`, `appendTable`, `appendSection`, `appendPageBreak`, `setMetadata`, `setBackgroundColor`.

#### Lifecycle

| Method | Description |
|--------|-------------|
| `reset()` | Clear builder state (keeps RPC connection) |
| `dispose()` | Close RPC connection — must be called when done |

---

### `DocxgoRPC`

Low-level async client that keeps a persistent RPC connection via `child_process.spawn`.

```ts
import { DocxgoRPC } from '@mmonterroca/docxgo';

const rpc = new DocxgoRPC({ timeout: 10000 });

const result = await rpc.call('document.create', {
  options: { title: 'Test' },
  content: [
    { type: 'paragraph', runs: [{ text: 'Hello from RPC!' }] },
  ],
  output: 'buffer',
});

console.log(result.data); // base64
rpc.close();
```

#### Methods

| Method | Description |
|--------|-------------|
| `call<T>(method, params?)` | Send an RPC request and await the response |
| `close()` | Gracefully close the connection |
| `kill()` | Force-kill the child process |

---

### `DocxgoExec`

Synchronous one-shot client. Each call spawns a new process.

Best for simple scripts or CLIs where async is unnecessary.

```ts
import { DocxgoExec } from '@mmonterroca/docxgo';

const exec = new DocxgoExec();

const result = exec.call('document.create', {
  options: { title: 'Quick Doc' },
  content: [
    { type: 'paragraph', runs: [{ text: 'Hello!' }] },
  ],
  output: 'buffer',
});

console.log(result.data); // base64
```

---

### Error Handling

All clients throw `DocxgoError` on failures:

```ts
import { DocxgoError } from '@mmonterroca/docxgo';

try {
  await doc.open('/nonexistent.docx');
} catch (err) {
  if (err instanceof DocxgoError) {
    console.error(`[${err.code}] ${err.message}`);
    // e.g. [OPEN_FAILED] failed to open document
  }
}
```

Some methods return enriched errors with a `data` field:

```ts
try {
  await doc.applyPatch([
    { op: 'appendParagraph', runs: [{ text: 'OK' }] },
    { op: 'unknownOp' as any },
  ]);
} catch (err) {
  if (err instanceof DocxgoError) {
    console.error(err.code);       // 'VALIDATION_ERROR'
    console.error(err.data?.index); // 1 — the failing operation index
    console.error(err.data?.op);    // 'unknownOp'
  }
}
```

Template errors include category info:

```ts
try {
  await doc.renderTemplate({ Name: 'Alice' }, { strictMode: true });
} catch (err) {
  if (err instanceof DocxgoError) {
    console.error(err.code);            // 'TEMPLATE_ERROR'
    console.error(err.data?.category);  // 'merge'
    console.error(err.data?.retryable); // false
  }
}
```

## RPC Methods

The following JSON-RPC methods are available:

| Method | Description |
|--------|-------------|
| `system.ping` | Health check |
| `system.version` | Get version and platform info |
| `system.capabilities` | Get supported features map |
| `system.batch` | Execute multiple requests in one call |
| `document.create` | Create a new document with content |
| `document.open` | Open an existing .docx file |
| `document.save` | Save an opened document |
| `document.close` | Close and free a document |
| `document.validate` | Validate document structure |
| `document.inspect` | Get document metadata and statistics |
| `document.setMetadata` | Update document metadata fields |
| `document.setBackgroundColor` | Set document background color |
| `document.addContent` | Append content items to an opened document |
| `document.addPageBreak` | Append a page break |
| `document.applyPatch` | Apply atomic multi-operation patches |
| `paragraph.add` | Add a single paragraph |
| `paragraph.list` | List all paragraphs |
| `table.add` | Add a single table |
| `table.list` | List all tables |
| `section.add` | Add a section break |
| `template.inspect` | Find template placeholders |
| `template.render` | Render template with data |

See the full [CLI Guide](../docs/CLI_GUIDE.md) for detailed parameter schemas.

## Content Types

### Paragraph

```ts
{
  type: 'paragraph',
  style: 'Heading1',          // optional
  alignment: 'center',        // optional: left|center|right|justify|both
  runs: [
    {
      text: 'Hello',
      bold: true,
      italic: true,
      underline: 'single',    // single|double|dotted|dashed|wave|thick
      color: 'FF0000',
      fontSize: 14,
      fontFamily: 'Arial',
      highlight: 'yellow',
    }
  ]
}
```

### Table

```ts
{
  type: 'table',
  style: 'TableGrid',    // optional
  rows: [
    {
      cells: [
        {
          paragraphs: [
            { runs: [{ text: 'Cell A1', bold: true }] }
          ],
          width: { value: 3000, type: 'dxa' },  // optional
          shading: 'E0E0E0',                     // optional
        }
      ]
    }
  ]
}
```

### Section Break

```ts
{
  type: 'section',
  breakType: 'nextPage',        // nextPage|continuous|evenPage|oddPage
  pageSize: { width: 12240, height: 15840 },
  orientation: 'landscape',
  margins: { top: 1440, bottom: 1440, left: 1800, right: 1800 },
}
```

### Page Break

```ts
{ type: 'pageBreak' }
```

## Development

```bash
# Install dependencies
cd npm && npm install

# Run tests (TypeScript, no build needed)
npm run test:src

# Build (CJS + ESM)
npm run build

# Type-check
npm run lint
```

## License

MIT — see [LICENSE](../LICENSE).
