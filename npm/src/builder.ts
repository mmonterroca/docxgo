import { DocxgoRPC, DocxgoError } from './rpc';
import type {
  DocumentOptions,
  ContentItem,
  ParagraphItem,
  TableItem,
  SectionItem,
  RunDef,
  TableRowDef,
  Alignment,
  SectionBreakType,
  PageSize,
  Margins,
  Orientation,
  HeaderFooterType,
  CreateResult,
  BufferResult,
  FileResult,
  SaveResult,
  OpenResult,
  InspectResult,
  ValidateResult,
  OkResult,
  IndexResult,
  ParagraphListResult,
  TableListResult,
} from './types';

export { DocxgoError };

export interface DocumentBuilderOptions {
  /** Explicit path to the docxgo binary. */
  binaryPath?: string;
  /** RPC call timeout in ms (default: 30000). */
  timeout?: number;
}

/**
 * Fluent builder API for creating and manipulating Word documents.
 *
 * Uses a persistent RPC connection for efficient multi-step operations.
 *
 * @example
 * ```ts
 * const doc = new DocumentBuilder();
 *
 * // Create from scratch
 * const result = await doc
 *   .setOptions({ title: 'Report', pageSize: 'A4' })
 *   .addHeading('Introduction', 1)
 *   .addParagraph('This is the first paragraph.', { bold: true })
 *   .addTable([
 *     [{ text: 'Name' }, { text: 'Value' }],
 *     [{ text: 'Score' }, { text: '95' }],
 *   ])
 *   .create();
 *
 * // Save as buffer
 * const buf = Buffer.from(result.data, 'base64');
 *
 * // Or save to file
 * await doc.reset()
 *   .addHeading('Hello')
 *   .createToFile('/tmp/hello.docx');
 *
 * doc.dispose();
 * ```
 */
export class DocumentBuilder {
  private rpc: DocxgoRPC;
  private options: DocumentOptions = {};
  private content: ContentItem[] = [];
  private documentId: string | null = null;

  constructor(opts?: DocumentBuilderOptions) {
    this.rpc = new DocxgoRPC({
      binaryPath: opts?.binaryPath,
      timeout: opts?.timeout,
    });
  }

  // ─── Document options ────────────────────────────────────────────────

  /** Set document-level options (title, author, pageSize, etc.). */
  setOptions(opts: DocumentOptions): this {
    this.options = { ...this.options, ...opts };
    return this;
  }

  /** Set the document title. */
  setTitle(title: string): this {
    this.options.title = title;
    return this;
  }

  /** Set the document author. */
  setAuthor(author: string): this {
    this.options.author = author;
    return this;
  }

  /** Set the page size. */
  setPageSize(size: PageSize): this {
    this.options.pageSize = size;
    return this;
  }

  /** Set page margins. */
  setMargins(margins: Margins): this {
    this.options.margins = margins;
    return this;
  }

  // ─── Content building ────────────────────────────────────────────────

  /** Add a raw content item. */
  addContent(item: ContentItem): this {
    this.content.push(item);
    return this;
  }

  /** Add multiple content items at once. */
  addContentItems(items: ContentItem[]): this {
    this.content.push(...items);
    return this;
  }

  /**
   * Add a paragraph with text and optional run formatting.
   *
   * @param text The text content.
   * @param formatting Optional run-level formatting (bold, italic, color, etc.).
   */
  addParagraph(text: string, formatting?: Omit<RunDef, 'text'>): this {
    const run: RunDef = { text, ...formatting };
    this.content.push({
      type: 'paragraph',
      runs: [run],
    });
    return this;
  }

  /**
   * Add a heading paragraph (Heading1–Heading9).
   *
   * @param text The heading text.
   * @param level Heading level (1–9, default: 1).
   */
  addHeading(text: string, level: number = 1): this {
    const style = `Heading${Math.max(1, Math.min(9, level))}`;
    this.content.push({
      type: 'paragraph',
      style,
      runs: [{ text }],
    });
    return this;
  }

  /**
   * Add a paragraph with multiple formatted runs.
   *
   * @param runs Array of run definitions.
   * @param options Paragraph-level options (style, alignment, etc.).
   */
  addFormattedParagraph(runs: RunDef[], options?: Omit<ParagraphItem, 'type' | 'runs'>): this {
    this.content.push({
      type: 'paragraph',
      ...options,
      runs,
    });
    return this;
  }

  /** Add a page break. */
  addPageBreak(): this {
    this.content.push({ type: 'pageBreak' });
    return this;
  }

  /**
   * Add a simple table from a 2D array of cell contents.
   *
   * Each cell can be a string or a RunDef for formatting.
   *
   * @param rows 2D array of cell contents.
   * @param options Table-level options (style, alignment, etc.).
   */
  addTable(
    rows: Array<Array<string | RunDef>>,
    options?: Omit<TableItem, 'type' | 'rows'>,
  ): this {
    const tableRows: TableRowDef[] = rows.map((row) => ({
      cells: row.map((cell) => {
        const run: RunDef = typeof cell === 'string' ? { text: cell } : cell;
        return {
          paragraphs: [{ runs: [run] }],
        };
      }),
    }));
    this.content.push({
      type: 'table',
      rows: tableRows,
      ...options,
    });
    return this;
  }

  /**
   * Add a table from raw TableRowDef data.
   *
   * @param rows Raw row definitions with full cell control.
   * @param options Table-level options.
   */
  addRawTable(rows: TableRowDef[], options?: Omit<TableItem, 'type' | 'rows'>): this {
    this.content.push({
      type: 'table',
      rows,
      ...options,
    });
    return this;
  }

  /**
   * Add a section break.
   *
   * @param options Section options (pageSize, orientation, headers, etc.).
   */
  addSection(options?: Omit<SectionItem, 'type'>): this {
    this.content.push({
      type: 'section',
      ...options,
    });
    return this;
  }

  // ─── Create / Save ───────────────────────────────────────────────────

  /**
   * Create the document and return it as a base64 buffer.
   *
   * @returns Result containing `data` (base64) and `documentId`.
   */
  async create(): Promise<BufferResult> {
    return this.rpc.call<BufferResult>('document.create', {
      options: this.options,
      content: this.content,
      output: 'buffer',
    });
  }

  /**
   * Create the document and save it to a file.
   *
   * @param filePath The output .docx file path.
   * @returns Result containing `filePath` and `documentId`.
   */
  async createToFile(filePath: string): Promise<FileResult> {
    return this.rpc.call<FileResult>('document.create', {
      options: this.options,
      content: this.content,
      output: 'file',
      filePath,
    });
  }

  /**
   * Create the document and return it as a Node.js Buffer.
   *
   * @returns The document as a Buffer.
   */
  async toBuffer(): Promise<Buffer> {
    const result = await this.create();
    return Buffer.from(result.data, 'base64');
  }

  // ─── Open / Modify workflow ──────────────────────────────────────────

  /**
   * Open an existing document from a file path.
   *
   * @param filePath Path to the .docx file.
   * @returns The document ID for subsequent operations.
   */
  async open(filePath: string): Promise<string> {
    const result = await this.rpc.call<OpenResult>('document.open', { filePath });
    this.documentId = result.documentId;
    return this.documentId;
  }

  /**
   * Open an existing document from a base64-encoded buffer.
   *
   * @param base64 Base64-encoded .docx bytes.
   * @returns The document ID for subsequent operations.
   */
  async openFromBase64(base64: string): Promise<string> {
    const result = await this.rpc.call<OpenResult>('document.open', { base64 });
    this.documentId = result.documentId;
    return this.documentId;
  }

  /**
   * Open an existing document from a Buffer.
   *
   * @param buffer The .docx file as a Buffer.
   * @returns The document ID for subsequent operations.
   */
  async openFromBuffer(buffer: Buffer): Promise<string> {
    return this.openFromBase64(buffer.toString('base64'));
  }

  /**
   * Append content to the currently opened document.
   * Flushes all queued content items that were added via builder methods
   * after the `open()` call.
   */
  async appendContent(): Promise<void> {
    this.requireDocumentId();
    if (this.content.length === 0) return;
    await this.rpc.call('document.addContent', {
      documentId: this.documentId,
      content: this.content,
    });
    this.content = [];
  }

  /**
   * Save the currently opened document to a file.
   *
   * @param filePath The output .docx file path.
   */
  async saveToFile(filePath: string): Promise<FileResult> {
    this.requireDocumentId();
    // Flush any pending content first
    if (this.content.length > 0) {
      await this.appendContent();
    }
    return this.rpc.call<FileResult>('document.save', {
      documentId: this.documentId,
      output: 'file',
      filePath,
    });
  }

  /**
   * Save the currently opened document as a base64 buffer.
   */
  async saveToBuffer(): Promise<BufferResult> {
    this.requireDocumentId();
    if (this.content.length > 0) {
      await this.appendContent();
    }
    return this.rpc.call<BufferResult>('document.save', {
      documentId: this.documentId,
      output: 'buffer',
    });
  }

  /**
   * Inspect the currently opened document.
   */
  async inspect(): Promise<InspectResult> {
    this.requireDocumentId();
    return this.rpc.call<InspectResult>('document.inspect', {
      documentId: this.documentId,
    });
  }

  /**
   * Validate the currently opened document.
   */
  async validate(): Promise<ValidateResult> {
    this.requireDocumentId();
    return this.rpc.call<ValidateResult>('document.validate', {
      documentId: this.documentId,
    });
  }

  /**
   * List all paragraphs in the currently opened document.
   */
  async listParagraphs(): Promise<ParagraphListResult> {
    this.requireDocumentId();
    return this.rpc.call<ParagraphListResult>('paragraph.list', {
      documentId: this.documentId,
    });
  }

  /**
   * List all tables in the currently opened document.
   */
  async listTables(): Promise<TableListResult> {
    this.requireDocumentId();
    return this.rpc.call<TableListResult>('table.list', {
      documentId: this.documentId,
    });
  }

  /**
   * Close the currently opened document, freeing resources.
   */
  async closeDocument(): Promise<void> {
    if (this.documentId) {
      await this.rpc.call('document.close', { documentId: this.documentId });
      this.documentId = null;
    }
  }

  // ─── Lifecycle ───────────────────────────────────────────────────────

  /**
   * Reset the builder state for building a new document.
   * Does NOT close the RPC connection.
   */
  reset(): this {
    this.options = {};
    this.content = [];
    this.documentId = null;
    return this;
  }

  /**
   * Dispose the builder and close the RPC connection.
   * Should be called when you're done using the builder.
   */
  dispose(): void {
    this.rpc.close();
  }

  /**
   * Make a raw RPC call. For advanced usage.
   */
  rawCall<T = unknown>(method: string, params?: Record<string, unknown>): Promise<T> {
    return this.rpc.call<T>(method, params);
  }

  // ─── Internal ────────────────────────────────────────────────────────

  private requireDocumentId(): void {
    if (!this.documentId) {
      throw new Error('No document is open. Call open() first.');
    }
  }
}
