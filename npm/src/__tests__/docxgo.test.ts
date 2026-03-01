import { describe, it, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { join } from 'path';
import { mkdtempSync, writeFileSync, readFileSync, existsSync, rmSync } from 'fs';
import { tmpdir } from 'os';

import { DocxgoRPC, DocxgoExec, DocumentBuilder, DocxgoError, resolveBinary } from '../index';
import type { BufferResult, FileResult, InspectResult, ParagraphListResult, TableListResult } from '../types';

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
