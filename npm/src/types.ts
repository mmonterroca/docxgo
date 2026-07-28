// ─── JSON-RPC Protocol Types ─────────────────────────────────────────────────

/** A JSON-RPC request sent to the docxgo binary. */
export interface RPCRequest {
  id: number;
  method: string;
  params?: Record<string, unknown>;
}

/** A JSON-RPC response from the docxgo binary. */
export interface RPCResponse<T = unknown> {
  id: number;
  result?: T;
  error?: RPCError;
}

/** An error in a JSON-RPC response. */
export interface RPCError {
  code: string;
  message: string;
  operation?: string;
  /** Additional structured data (e.g. { index, category, retryable }). */
  data?: Record<string, unknown>;
}

// ─── Document Options ────────────────────────────────────────────────────────

/** Page size presets or custom dimensions (in twips). */
export type PageSizePreset = 'A4' | 'Letter' | 'Legal' | 'A3' | 'Tabloid';
export interface PageSizeCustom {
  width: number;
  height: number;
}
export type PageSize = PageSizePreset | PageSizeCustom;

/** Margin presets or custom values (in twips). */
export type MarginPreset = 'normal' | 'narrow' | 'wide';
export interface MarginCustom {
  top?: number;
  bottom?: number;
  left?: number;
  right?: number;
  header?: number;
  footer?: number;
}
export type Margins = MarginPreset | MarginCustom;

/** Theme presets. */
export type ThemeName =
  | 'Corporate'
  | 'Startup'
  | 'Modern'
  | 'Fintech'
  | 'Academic'
  | 'TechPresentation'
  | 'TechDarkMode';

/** Document-level options. */
export interface DocumentOptions {
  title?: string;
  author?: string;
  subject?: string;
  pageSize?: PageSize;
  margins?: Margins;
  theme?: ThemeName;
}

// ─── Content Items ───────────────────────────────────────────────────────────

export type Alignment = 'left' | 'center' | 'right' | 'justify' | 'distribute';
export type UnderlineStyle = 'none' | 'single' | 'double' | 'thick' | 'dotted' | 'dashed' | 'wave';
export type HighlightColor =
  | 'none' | 'yellow' | 'green' | 'cyan' | 'magenta' | 'blue' | 'red'
  | 'darkBlue' | 'darkCyan' | 'darkGreen' | 'darkMagenta' | 'darkRed'
  | 'darkYellow' | 'darkGray' | 'lightGray';
export type BreakType = 'page' | 'column' | 'line';
export type SectionBreakType = 'nextPage' | 'continuous' | 'evenPage' | 'oddPage';
export type Orientation = 'portrait' | 'landscape';
export type VerticalAlign = 'top' | 'center' | 'bottom';
export type BorderStyleType = 'none' | 'single' | 'dotted' | 'dashed' | 'double' | 'triple' | 'thick';
export type HeaderFooterType = 'default' | 'first' | 'even';
export type WidthType = 'auto' | 'dxa' | 'pct';

/** A hyperlink definition. */
export interface HyperlinkDef {
  url: string;
  displayText?: string;
}

/** A field definition. */
export interface FieldDef {
  type: 'pageNumber' | 'pageCount' | 'toc' | 'hyperlink' | 'styleRef';
  url?: string;
  display?: string;
  style?: string;
  options?: Record<string, string>;
}

/** An image definition. */
export interface ImageDef {
  path?: string;
  base64?: string;
  format?: string;
  widthPx?: number;
  heightPx?: number;
}

/** A text run within a paragraph. */
export interface RunDef {
  text?: string;
  bold?: boolean;
  italic?: boolean;
  strike?: boolean;
  underline?: UnderlineStyle;
  color?: string;
  fontSize?: number;
  font?: string;
  highlight?: HighlightColor;
  hyperlink?: HyperlinkDef;
  field?: FieldDef;
  image?: ImageDef;
  break?: BreakType;
}

/** Line spacing definition. */
export interface LineSpacingDef {
  rule?: 'auto' | 'exact' | 'atLeast';
  value?: number;
}

/** Indentation definition. */
export interface IndentDef {
  left?: number;
  right?: number;
  firstLine?: number;
  hanging?: number;
}

/** Numbering/list reference. */
export interface NumberingDef {
  id: number;
  level: number;
}

/** Border style definition. */
export interface BorderDef {
  style?: BorderStyleType;
  width?: number;
  color?: string;
}

/** Paragraph borders. */
export interface ParagraphBordersDef {
  top?: BorderDef;
  bottom?: BorderDef;
  left?: BorderDef;
  right?: BorderDef;
}

/** A paragraph content item. */
export interface ParagraphItem {
  type: 'paragraph';
  /** Plain-text shortcut for a single unformatted run. Ignored when `runs` is given. */
  text?: string;
  style?: string;
  alignment?: Alignment;
  spacingBefore?: number;
  spacingAfter?: number;
  lineSpacing?: LineSpacingDef;
  indent?: IndentDef;
  numbering?: NumberingDef;
  borders?: ParagraphBordersDef;
  runs?: RunDef[];
}

/** Table cell borders. */
export interface TableBordersDef {
  top?: BorderDef;
  left?: BorderDef;
  bottom?: BorderDef;
  right?: BorderDef;
}

/** A table cell. */
export interface TableCellDef {
  paragraphs?: Array<Omit<ParagraphItem, 'type'>>;
  width?: number;
  verticalAlignment?: VerticalAlign;
  borders?: TableBordersDef;
  shading?: string;
  gridSpan?: number;
}

/** A table row. */
export interface TableRowDef {
  height?: number;
  cells?: TableCellDef[];
}

/** Table width definition. */
export interface TableWidthDef {
  type?: WidthType;
  value?: number;
}

/** A table content item. */
export interface TableItem {
  type: 'table';
  rows?: TableRowDef[];
  alignment?: Alignment;
  style?: string;
  width?: TableWidthDef;
}

/** A section content item. */
export interface SectionItem {
  type: 'section';
  breakType?: SectionBreakType;
  pageSize?: PageSize;
  margins?: Margins;
  orientation?: Orientation;
  columns?: number;
  headers?: Partial<Record<HeaderFooterType, Array<Omit<ParagraphItem, 'type'>>>>;
  footers?: Partial<Record<HeaderFooterType, Array<Omit<ParagraphItem, 'type'>>>>;
}

/** A page break content item. */
export interface PageBreakItem {
  type: 'pageBreak';
}

/** Any content item. */
export type ContentItem = ParagraphItem | TableItem | SectionItem | PageBreakItem;

// ─── Method Params ───────────────────────────────────────────────────────────

export interface CreateParams {
  options?: DocumentOptions;
  content?: ContentItem[];
  output?: 'buffer' | 'file';
  filePath?: string;
}

export interface OpenParams {
  filePath?: string;
  base64?: string;
}

export interface SaveParams {
  documentId: string;
  output?: 'buffer' | 'file';
  filePath?: string;
}

export interface ValidateParams {
  documentId: string;
}

export interface InspectParams {
  documentId: string;
}

export interface SetMetadataParams {
  documentId: string;
  title?: string;
  subject?: string;
  creator?: string;
  description?: string;
  keywords?: string[];
  created?: string;
  modified?: string;
}

export interface SetBackgroundColorParams {
  documentId: string;
  color: string;
}

/**
 * Params for document.setLanguage. val, eastAsia, and bidi are BCP 47
 * language tags (e.g. "es-MX", "en-US"); at least one must be set.
 *
 * Note: the docxgo binary rejects this on a document opened via
 * `document.open` / `DocumentBuilder.open()` (round-trip guard) — it only
 * works on documents created via `document.create` / `DocumentBuilder.create()`.
 */
export interface SetLanguageParams {
  documentId: string;
  val?: string;
  eastAsia?: string;
  bidi?: string;
}

export interface AddContentParams {
  documentId: string;
  content: ContentItem[];
}

export interface AddPageBreakParams {
  documentId: string;
}

export interface ParagraphAddParams {
  documentId: string;
  /** Plain-text shortcut for a single unformatted run. Ignored when `runs` is given. */
  text?: string;
  style?: string;
  alignment?: Alignment;
  spacingBefore?: number;
  spacingAfter?: number;
  lineSpacing?: LineSpacingDef;
  indent?: IndentDef;
  numbering?: NumberingDef;
  borders?: ParagraphBordersDef;
  runs?: RunDef[];
}

export interface ParagraphListParams {
  documentId: string;
}

/** Replaces a body paragraph's content by index (as reported by paragraph.list).
 * Replaces content (text and runs) only — any of the inherited fields you
 * omit (style, alignment, indent, ...) leaves the paragraph's existing value
 * untouched rather than resetting it. */
export interface ParagraphSetTextParams extends ParagraphAddParams {
  index: number;
}

export interface TableAddParams {
  documentId: string;
  rows?: TableRowDef[];
  alignment?: Alignment;
  style?: string;
  width?: TableWidthDef;
}

export interface TableListParams {
  documentId: string;
  /** When true, adds each table's cell texts to the listing. */
  includeText?: boolean;
}

export interface TableCellRef {
  documentId: string;
  tableIndex: number;
  rowIndex: number;
  columnIndex: number;
}

export type TableGetCellParams = TableCellRef;

/**
 * Provide exactly one of `text` or `paragraphs` — supplying both, or neither,
 * is rejected. Both replace the cell's content, so `text: ''` and
 * `paragraphs: []` are meaningful requests to empty it.
 */
export interface TableSetCellParams extends TableCellRef {
  text?: string;
  paragraphs?: Array<Omit<ParagraphItem, 'type'>>;
}

export interface ReplaceTextParams {
  documentId: string;
  find: string;
  replace: string;
}

export interface SectionAddParams {
  documentId: string;
  breakType?: SectionBreakType;
  pageSize?: PageSize;
  margins?: Margins;
  orientation?: Orientation;
  columns?: number;
  headers?: Partial<Record<HeaderFooterType, Array<Omit<ParagraphItem, 'type'>>>>;
  footers?: Partial<Record<HeaderFooterType, Array<Omit<ParagraphItem, 'type'>>>>;
}

export interface CloseParams {
  documentId: string;
}

// ─── Method Results ──────────────────────────────────────────────────────────

export interface BufferResult {
  data: string;
  documentId: string;
}

export interface FileResult {
  filePath: string;
  documentId: string;
}

export type CreateResult = BufferResult | FileResult;
export type SaveResult = BufferResult | FileResult;

export interface OpenResult {
  documentId: string;
}

export interface ValidateResult {
  valid: boolean;
  message?: string;
  errors?: Array<{ message: string }>;
}

export interface InspectMetadata {
  title: string;
  subject: string;
  creator: string;
  description: string;
  keywords: string[] | null;
  created: string;
  modified: string;
}

/** A document's default proofing language, as reported by document.inspect. */
export interface LanguageInfo {
  val: string;
  eastAsia: string;
  bidi: string;
}

export interface InspectResult {
  paragraphCount: number;
  tableCount: number;
  text: string[];
  metadata?: InspectMetadata;
  backgroundColor?: string;
  language?: LanguageInfo;
}

export interface OkResult {
  ok: boolean;
}

export interface IndexResult {
  ok: boolean;
  index: number;
}

export interface ParagraphInfo {
  index: number;
  text: string;
  style?: string;
}

export interface ParagraphListResult {
  count: number;
  paragraphs: ParagraphInfo[];
}

export interface TableInfo {
  index: number;
  rows: number;
  columns: number;
  /** Present only when table.list was called with `includeText: true`. */
  cells?: string[][];
}

export interface TableListResult {
  count: number;
  tables: TableInfo[];
}

/** Result of table.getCell. */
export interface TableCellResult {
  /** The cell's paragraphs joined with "\n". */
  text: string;
  paragraphs: string[];
  paragraphCount: number;
}

/** Result of table.setCell. paragraphCount always equals the number of
 * paragraph items written — fewer than the cell had removes the trailing
 * ones, more appends new ones. */
export interface TableSetCellResult {
  ok: boolean;
  paragraphCount: number;
}

/** Result of document.replaceText. */
export interface ReplaceTextResult {
  replaced: number;
  skipped: number;
}

// ─── System Types ────────────────────────────────────────────────────────────

/** Result of system.ping. */
export interface PingResult {
  status: 'ok';
}

/** Result of system.version. */
export interface SystemVersionResult {
  name: string;
  version: string;
  protocolVersion: string;
  goVersion: string;
  platform: string;
  arch: string;
}

/** Result of system.capabilities — flat map of feature → enabled. */
export type SystemCapabilitiesResult = Record<string, boolean>;

/** A single request within a system.batch call. */
export interface BatchRequest {
  method: string;
  params?: Record<string, unknown>;
}

/** Params for system.batch. */
export interface BatchParams {
  requests: BatchRequest[];
}

/** A single response in a batch result. */
export interface BatchResponseEntry {
  result?: unknown;
  error?: RPCError;
}

/** Result of system.batch. */
export interface BatchResult {
  responses: BatchResponseEntry[];
}

// ─── Template Types ──────────────────────────────────────────────────────────

/** Params for template.inspect. */
export interface TemplateInspectParams {
  documentId: string;
  openDelimiter?: string;
  closeDelimiter?: string;
}

/** A placeholder location detail. */
export interface PlaceholderDetail {
  name: string;
  fullMatch: string;
  location: string;
  paragraph: number;
  run: number;
  table?: number;
  row?: number;
  cell?: number;
}

/** Result of template.inspect. */
export interface TemplateInspectResult {
  placeholders: string[];
  count: number;
  occurrences: number;
  details: PlaceholderDetail[];
}

/** Params for template.render. */
export interface TemplateRenderParams {
  documentId: string;
  data: Record<string, string>;
  strictMode?: boolean;
  openDelimiter?: string;
  closeDelimiter?: string;
}

/** A validation warning from template.render. */
export interface TemplateWarning {
  severity: 'warning' | 'info';
  key: string;
  message: string;
}

/** Result of template.render. */
export interface TemplateRenderResult {
  ok: boolean;
  warnings?: TemplateWarning[];
}

// ─── ApplyPatch Types ────────────────────────────────────────────────────────

/** Base patch operation. */
export interface PatchOperationBase {
  op: string;
}

/** Append a paragraph. */
export interface AppendParagraphOp extends PatchOperationBase {
  op: 'appendParagraph';
  text?: string;
  style?: string;
  alignment?: Alignment;
  spacingBefore?: number;
  spacingAfter?: number;
  lineSpacing?: LineSpacingDef;
  indent?: IndentDef;
  numbering?: NumberingDef;
  borders?: ParagraphBordersDef;
  runs?: RunDef[];
}

/** Append a table. */
export interface AppendTableOp extends PatchOperationBase {
  op: 'appendTable';
  rows?: TableRowDef[];
  alignment?: Alignment;
  style?: string;
  width?: TableWidthDef;
}

/** Append a section. */
export interface AppendSectionOp extends PatchOperationBase {
  op: 'appendSection';
  breakType?: SectionBreakType;
  pageSize?: PageSize;
  margins?: Margins;
  orientation?: Orientation;
  columns?: number;
}

/** Append a page break. */
export interface AppendPageBreakOp extends PatchOperationBase {
  op: 'appendPageBreak';
}

/** Set document metadata. */
export interface SetMetadataOp extends PatchOperationBase {
  op: 'setMetadata';
  title?: string;
  subject?: string;
  creator?: string;
  description?: string;
  keywords?: string[];
  created?: string;
  modified?: string;
}

/** Set document background color. */
export interface SetBackgroundColorOp extends PatchOperationBase {
  op: 'setBackgroundColor';
  color: string;
}

/** Set the document's default proofing language. See SetLanguageParams. */
export interface SetLanguageOp extends PatchOperationBase {
  op: 'setLanguage';
  val?: string;
  eastAsia?: string;
  bidi?: string;
}

/** Union of all patch operations. */
export type PatchOperation =
  | AppendParagraphOp
  | AppendTableOp
  | AppendSectionOp
  | AppendPageBreakOp
  | SetMetadataOp
  | SetBackgroundColorOp
  | SetLanguageOp;

/** Params for document.applyPatch. */
export interface ApplyPatchParams {
  documentId: string;
  operations: PatchOperation[];
}

/** Result of document.applyPatch. */
export interface ApplyPatchResult {
  ok: boolean;
  applied: number;
}
