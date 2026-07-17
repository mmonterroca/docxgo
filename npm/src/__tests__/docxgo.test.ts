import { describe, it, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { join } from 'path';
import { mkdtempSync, writeFileSync, readFileSync, existsSync, rmSync } from 'fs';
import { tmpdir } from 'os';

import { DocxgoRPC, DocxgoExec, DocumentBuilder, DocxgoError, resolveBinary } from '../index';
import type {
  BufferResult, FileResult, InspectResult, ParagraphListResult, TableListResult,
  PingResult, SystemVersionResult, SystemCapabilitiesResult, BatchResult,
  TemplateInspectResult, TemplateRenderResult, ApplyPatchResult,
} from '../types';

const BINARY = join(__dirname, '..', '..', 'bin', 'docxgo');

// ─── resolveBinary ───────────────────────────────────────────────────────────

describe('resolveBinary', () => {
  it('returns custom path when provided', () => {
    assert.equal(resolveBinary('/custom/path'), '/custom/path');
  });

  it('uses DOCXGO_BIN env var', () => {
    const orig = process.env.DOCXGO_BIN;
    process.env.DOCXGO_BIN = '/env/path/docxgo';
    try {
      assert.equal(resolveBinary(), '/env/path/docxgo');
    } finally {
      if (orig === undefined) delete process.env.DOCXGO_BIN;
      else process.env.DOCXGO_BIN = orig;
    }
  });

  it('finds binary in bin/ directory', () => {
    const resolved = resolveBinary();
    // Should find the binary we built
    assert.ok(typeof resolved === 'string');
    assert.ok(resolved.length > 0);
  });
});

// ─── DocxgoExec ──────────────────────────────────────────────────────────────

describe('DocxgoExec', () => {
  let exec: DocxgoExec;

  before(() => {
    exec = new DocxgoExec({ binaryPath: BINARY });
  });

  it('creates a document and returns base64', () => {
    const result = exec.call<BufferResult>('document.create', {
      content: [
        { type: 'paragraph', runs: [{ text: 'Hello from Node!' }] },
      ],
      output: 'buffer',
    });
    assert.ok(result.data);
    assert.ok(result.data.length > 100);
    assert.ok(result.documentId);
  });

  it('creates a document to file', () => {
    const tmpDir = mkdtempSync(join(tmpdir(), 'docxgo-test-'));
    const filePath = join(tmpDir, 'test.docx');
    const result = exec.call<FileResult>('document.create', {
      content: [
        { type: 'paragraph', runs: [{ text: 'File output test' }] },
      ],
      output: 'file',
      filePath,
    });
    assert.equal(result.filePath, filePath);
    assert.ok(existsSync(filePath));
    const stat = readFileSync(filePath);
    assert.ok(stat.length > 100);
    rmSync(tmpDir, { recursive: true });
  });

  it('throws DocxgoError on invalid method', () => {
    assert.throws(
      () => exec.call('nonexistent.method', {}),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'METHOD_NOT_FOUND');
        return true;
      },
    );
  });

  it('creates document with options and table', () => {
    const result = exec.call<BufferResult>('document.create', {
      options: { title: 'Test Doc', pageSize: 'A4' },
      content: [
        { type: 'paragraph', style: 'Heading1', runs: [{ text: 'Title' }] },
        {
          type: 'table',
          rows: [
            {
              cells: [
                { paragraphs: [{ runs: [{ text: 'A1', bold: true }] }] },
                { paragraphs: [{ runs: [{ text: 'B1' }] }] },
              ],
            },
          ],
        },
      ],
      output: 'buffer',
    });
    assert.ok(result.data.length > 100);
  });
});

// ─── DocxgoRPC ───────────────────────────────────────────────────────────────

describe('DocxgoRPC', () => {
  let rpc: DocxgoRPC;

  before(() => {
    rpc = new DocxgoRPC({ binaryPath: BINARY });
  });

  after(() => {
    rpc.close();
  });

  it('creates a document via RPC', async () => {
    const result = await rpc.call<BufferResult>('document.create', {
      content: [
        { type: 'paragraph', runs: [{ text: 'RPC Hello!' }] },
      ],
      output: 'buffer',
    });
    assert.ok(result.data);
    assert.ok(result.documentId);
  });

  it('opens, mutates, inspects, and saves a document', async () => {
    // Create initial document
    const createResult = await rpc.call<BufferResult>('document.create', {
      content: [
        { type: 'paragraph', runs: [{ text: 'Original' }] },
      ],
      output: 'buffer',
    });
    const docId = createResult.documentId;

    // Add content
    await rpc.call('document.addContent', {
      documentId: docId,
      content: [
        { type: 'paragraph', runs: [{ text: 'Added via RPC' }] },
      ],
    });

    // Add paragraph directly
    await rpc.call('paragraph.add', {
      documentId: docId,
      style: 'Heading1',
      runs: [{ text: 'New Heading' }],
    });

    // Add table
    await rpc.call('table.add', {
      documentId: docId,
      rows: [
        {
          cells: [
            { paragraphs: [{ runs: [{ text: 'Cell 1' }] }] },
            { paragraphs: [{ runs: [{ text: 'Cell 2' }] }] },
          ],
        },
      ],
    });

    // List paragraphs
    const paraList = await rpc.call<ParagraphListResult>('paragraph.list', {
      documentId: docId,
    });
    assert.ok(paraList.count >= 3);

    // List tables
    const tableList = await rpc.call<TableListResult>('table.list', {
      documentId: docId,
    });
    assert.ok(tableList.count >= 1);

    // Inspect
    const inspectResult = await rpc.call<InspectResult>('document.inspect', {
      documentId: docId,
    });
    assert.ok(inspectResult.paragraphCount >= 3);
    assert.ok(inspectResult.tableCount >= 1);

    // Save
    const tmpDir = mkdtempSync(join(tmpdir(), 'docxgo-rpc-'));
    const filePath = join(tmpDir, 'mutated.docx');
    await rpc.call('document.save', {
      documentId: docId,
      output: 'file',
      filePath,
    });
    assert.ok(existsSync(filePath));

    // Close
    await rpc.call('document.close', { documentId: docId });

    rmSync(tmpDir, { recursive: true });
  });

  it('rejects on error', async () => {
    await assert.rejects(
      () => rpc.call('nonexistent.method', {}),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'METHOD_NOT_FOUND');
        return true;
      },
    );
  });

  it('sets and reports the document language', async () => {
    const createResult = await rpc.call<BufferResult>('document.create', {
      content: [],
      output: 'buffer',
    });
    const docId = createResult.documentId;

    await rpc.call('document.setLanguage', { documentId: docId, val: 'es-MX' });

    const inspectResult = await rpc.call<InspectResult>('document.inspect', {
      documentId: docId,
    });
    assert.equal(inspectResult.language?.val, 'es-MX');

    await rpc.call('document.close', { documentId: docId });
  });

  it('rejects setLanguage on a round-trip-opened document', async () => {
    const createResult = await rpc.call<BufferResult>('document.create', {
      content: [],
      output: 'buffer',
    });
    const openResult = await rpc.call<{ documentId: string }>('document.open', {
      base64: createResult.data,
    });

    await assert.rejects(
      () => rpc.call('document.setLanguage', { documentId: openResult.documentId, val: 'en-US' }),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'VALIDATION_ERROR');
        return true;
      },
    );
  });
});

// ─── DocumentBuilder ─────────────────────────────────────────────────────────

describe('DocumentBuilder', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('creates a simple document to buffer', async () => {
    const result = await doc
      .setTitle('Builder Test')
      .addHeading('Hello Builder')
      .addParagraph('A simple paragraph.')
      .create();

    assert.ok(result.data);
    assert.ok(result.data.length > 100);
    assert.ok(result.documentId);
    doc.reset();
  });

  it('creates a document to file', async () => {
    const tmpDir = mkdtempSync(join(tmpdir(), 'docxgo-builder-'));
    const filePath = join(tmpDir, 'builder.docx');

    const result = await doc
      .setOptions({ title: 'File Test', pageSize: 'A4', author: 'Test' })
      .addHeading('Section 1', 1)
      .addParagraph('Body text', { bold: true })
      .addPageBreak()
      .addHeading('Section 2', 1)
      .createToFile(filePath);

    assert.equal(result.filePath, filePath);
    assert.ok(existsSync(filePath));

    doc.reset();
    rmSync(tmpDir, { recursive: true });
  });

  it('creates a document with tables', async () => {
    const buf = await doc
      .addHeading('Report')
      .addTable([
        [{ text: 'Name', bold: true }, { text: 'Score', bold: true }],
        ['Alice', '95'],
        ['Bob', '87'],
      ])
      .toBuffer();

    assert.ok(buf.length > 100);
    // Check it's a valid zip (docx)
    assert.equal(buf[0], 0x50); // P
    assert.equal(buf[1], 0x4B); // K
    doc.reset();
  });

  it('opens, modifies, and saves a document', async () => {
    // First create a doc
    const tmpDir = mkdtempSync(join(tmpdir(), 'docxgo-open-'));
    const original = join(tmpDir, 'original.docx');

    await doc
      .addHeading('Original Document')
      .addParagraph('First paragraph.')
      .createToFile(original);

    doc.reset();

    // Open and modify
    await doc.open(original);

    doc.addHeading('Appended Section', 2);
    doc.addParagraph('Added after opening.');

    // Inspect
    const paraList = await doc.listParagraphs();
    assert.ok(paraList.count >= 2);

    // Save to new file
    const modified = join(tmpDir, 'modified.docx');
    await doc.saveToFile(modified);
    assert.ok(existsSync(modified));

    // Close
    await doc.closeDocument();
    doc.reset();

    rmSync(tmpDir, { recursive: true });
  });

  it('opens from buffer', async () => {
    // Create a doc as buffer first
    const buf = await doc
      .addParagraph('Buffer test')
      .toBuffer();
    doc.reset();

    // Open from buffer
    const docId = await doc.openFromBuffer(buf);
    assert.ok(docId);

    const inspection = await doc.inspect();
    assert.ok(inspection.paragraphCount >= 1);

    await doc.closeDocument();
    doc.reset();
  });

  it('validates a document', async () => {
    const result = await doc
      .addParagraph('Valid doc')
      .create();
    doc.reset();

    // Open the doc
    await doc.openFromBase64(result.data);
    const validation = await doc.validate();
    assert.equal(validation.valid, true);

    await doc.closeDocument();
    doc.reset();
  });

  it('throws when no document is open', async () => {
    await assert.rejects(
      () => doc.inspect(),
      { message: 'No document is open. Call open() first.' },
    );
  });
});

// ─── System Methods ──────────────────────────────────────────────────────────

describe('System Methods (via DocumentBuilder)', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('system.ping returns ok', async () => {
    const result = await doc.ping();
    assert.equal(result.status, 'ok');
  });

  it('system.version returns version info', async () => {
    const result = await doc.version();
    assert.equal(result.name, 'docxgo');
    assert.ok(result.version);
    assert.ok(result.protocolVersion);
    assert.ok(result.goVersion);
    assert.ok(result.platform);
    assert.ok(result.arch);
  });

  it('system.capabilities returns features', async () => {
    const result = await doc.capabilities();
    assert.equal(result.rpc, true);
    assert.equal(result.template, true);
    assert.equal(result.batch, true);
    assert.equal(result.applyPatch, true);
  });
});

// ─── Batch Methods ───────────────────────────────────────────────────────────

describe('system.batch (via DocumentBuilder)', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('executes multiple requests in a single batch', async () => {
    const result = await doc.batch([
      { method: 'system.ping' },
      { method: 'system.version' },
    ]);

    assert.equal(result.responses.length, 2);
    assert.deepEqual((result.responses[0].result as PingResult).status, 'ok');
    assert.ok((result.responses[1].result as SystemVersionResult).version);
  });

  it('returns per-request error for invalid method', async () => {
    const result = await doc.batch([
      { method: 'system.ping' },
      { method: 'nonexistent.method' },
    ]);

    assert.equal(result.responses.length, 2);
    assert.ok(result.responses[0].result);
    assert.ok(result.responses[1].error);
    assert.equal(result.responses[1].error!.code, 'METHOD_NOT_FOUND');
  });

  it('rejects nested batch', async () => {
    const result = await doc.batch([
      { method: 'system.batch', params: { requests: [{ method: 'system.ping' }] } },
    ]);

    assert.equal(result.responses.length, 1);
    assert.ok(result.responses[0].error);
    assert.equal(result.responses[0].error!.code, 'VALIDATION_ERROR');
  });
});

// ─── Template Methods ────────────────────────────────────────────────────────

describe('Template Methods (via DocumentBuilder)', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('inspects template placeholders', async () => {
    // Create a document with placeholders
    const result = await doc
      .addParagraph('Hello {{Name}}, welcome to {{Company}}!')
      .addParagraph('Your role is {{Role}} at {{Company}}.')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    const inspection = await doc.inspectTemplate();
    assert.ok(inspection.placeholders.length >= 2);
    assert.ok(inspection.count >= 2);
    assert.ok(inspection.occurrences >= 3); // Company appears twice
    assert.ok(inspection.details.length >= 3);

    await doc.closeDocument();
    doc.reset();
  });

  it('renders template placeholders', async () => {
    // Create a document with placeholders
    const result = await doc
      .addParagraph('Hello {{Name}}, welcome!')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    const renderResult = await doc.renderTemplate({
      Name: 'Alice',
    });
    assert.equal(renderResult.ok, true);

    // Verify the text was replaced
    const inspection = await doc.inspect();
    const allText = inspection.text.join(' ');
    assert.ok(allText.includes('Alice'));
    assert.ok(!allText.includes('{{Name}}'));

    await doc.closeDocument();
    doc.reset();
  });

  it('template.render with strict mode rejects on missing keys', async () => {
    const result = await doc
      .addParagraph('Dear {{Name}}, your code is {{Code}}.')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    // Strict mode should reject when keys are missing
    await assert.rejects(
      () => doc.renderTemplate(
        { Name: 'Bob' },
        { strictMode: true },
      ),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'TEMPLATE_ERROR');
        assert.ok(err.data);
        assert.equal(err.data!.category, 'merge');
        return true;
      },
    );

    await doc.closeDocument();
    doc.reset();
  });
});

// ─── ApplyPatch Methods ──────────────────────────────────────────────────────

describe('document.applyPatch (via DocumentBuilder)', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('applies multiple patch operations', async () => {
    const result = await doc
      .addParagraph('Initial content')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    const patchResult = await doc.applyPatch([
      { op: 'appendParagraph', runs: [{ text: 'Patched paragraph' }] },
      { op: 'appendPageBreak' },
      { op: 'appendParagraph', style: 'Heading1', runs: [{ text: 'New Section' }] },
    ]);

    assert.equal(patchResult.ok, true);
    assert.equal(patchResult.applied, 3);

    // Verify content was added
    const inspection = await doc.inspect();
    assert.ok(inspection.paragraphCount >= 3);

    await doc.closeDocument();
    doc.reset();
  });

  it('applies setLanguage directly after create() without reopening', async () => {
    // create() tracks the new document's ID internally, so applyPatch()
    // can target it right away — no open()/openFromBase64() round-trip,
    // which would trip setLanguage's round-trip guard.
    await doc.addParagraph('Doc for direct language patch').create();

    const patchResult = await doc.applyPatch([{ op: 'setLanguage', val: 'es-MX' }]);
    assert.equal(patchResult.ok, true);
    assert.equal(patchResult.applied, 1);

    const inspection = await doc.inspect();
    assert.equal(inspection.language?.val, 'es-MX');

    await doc.closeDocument();
    doc.reset();
  });

  it('applies setMetadata operation', async () => {
    const result = await doc
      .addParagraph('Doc with metadata')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    const patchResult = await doc.applyPatch([
      { op: 'setMetadata', title: 'Patched Title', creator: 'Patch Author' },
    ]);

    assert.equal(patchResult.ok, true);
    assert.equal(patchResult.applied, 1);

    const inspection = await doc.inspect();
    assert.equal(inspection.metadata?.title, 'Patched Title');
    assert.equal(inspection.metadata?.creator, 'Patch Author');

    await doc.closeDocument();
    doc.reset();
  });

  it('rejects setLanguage operation on a round-trip-opened document', async () => {
    const result = await doc
      .addParagraph('Doc for language patch')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    await assert.rejects(
      () => doc.applyPatch([{ op: 'setLanguage', val: 'fr-FR' }]),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'VALIDATION_ERROR');
        assert.ok(err.data);
        assert.equal(err.data!.applied, 0);
        return true;
      },
    );

    await doc.closeDocument();
    doc.reset();
  });

  it('reports error with index for unknown operation', async () => {
    const result = await doc
      .addParagraph('Test')
      .create();
    doc.reset();

    await doc.openFromBase64(result.data);

    await assert.rejects(
      () => doc.applyPatch([
        { op: 'appendParagraph', runs: [{ text: 'OK' }] },
        { op: 'unknownOp' as any },
      ]),
      (err: unknown) => {
        assert.ok(err instanceof DocxgoError);
        assert.equal(err.code, 'VALIDATION_ERROR');
        assert.ok(err.data);
        assert.equal(err.data!.index, 1);
        return true;
      },
    );

    await doc.closeDocument();
    doc.reset();
  });
});

// ─── Integration: Template + Patch + Batch ───────────────────────────────────

describe('Integration: Template + Patch + Batch workflow', () => {
  let doc: DocumentBuilder;

  before(() => {
    doc = new DocumentBuilder({ binaryPath: BINARY });
  });

  after(() => {
    doc.dispose();
  });

  it('creates template, inspects, renders, patches, and saves', async () => {
    // 1. Create a template document
    const createResult = await doc
      .addHeading('Invoice for {{CustomerName}}')
      .addParagraph('Date: {{Date}}')
      .addParagraph('Amount: {{Amount}}')
      .create();
    doc.reset();

    // 2. Open it
    await doc.openFromBase64(createResult.data);

    // 3. Inspect placeholders
    const inspection = await doc.inspectTemplate();
    assert.ok(inspection.placeholders.includes('CustomerName'));
    assert.ok(inspection.placeholders.includes('Date'));
    assert.ok(inspection.placeholders.includes('Amount'));

    // 4. Render template
    const renderResult = await doc.renderTemplate({
      CustomerName: 'Acme Corp',
      Date: '2025-01-15',
      Amount: '$1,234.56',
    });
    assert.equal(renderResult.ok, true);

    // 5. Apply patch to add more content
    const patchResult = await doc.applyPatch([
      { op: 'appendParagraph', runs: [{ text: 'Thank you for your business!' }] },
      { op: 'setMetadata', title: 'Invoice #1234' },
    ]);
    assert.equal(patchResult.ok, true);
    assert.equal(patchResult.applied, 2);

    // 6. Verify final content
    const finalInspection = await doc.inspect();
    const allText = finalInspection.text.join(' ');
    assert.ok(allText.includes('Acme Corp'));
    assert.ok(allText.includes('Thank you'));
    assert.equal(finalInspection.metadata?.title, 'Invoice #1234');

    // 7. Save
    const tmpDir = mkdtempSync(join(tmpdir(), 'docxgo-integration-'));
    const filePath = join(tmpDir, 'invoice.docx');
    await doc.saveToFile(filePath);
    assert.ok(existsSync(filePath));

    await doc.closeDocument();
    doc.reset();
    rmSync(tmpDir, { recursive: true });
  });
});
