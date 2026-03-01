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
export type ThemeName = 'Corporate' | 'Startup' | 'Modern' | 'Fintech' | 'Academic';

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

export interface AddContentParams {
  documentId: string;
  content: ContentItem[];
}

export interface AddPageBreakParams {
  documentId: string;
}

export interface ParagraphAddParams {
  documentId: string;
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

export interface TableAddParams {
  documentId: string;
  rows?: TableRowDef[];
  alignment?: Alignment;
  style?: string;
  width?: TableWidthDef;
}

export interface TableListParams {
  documentId: string;
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

export interface InspectResult {
  paragraphCount: number;
  tableCount: number;
  text: string[];
  metadata?: InspectMetadata;
  backgroundColor?: string;
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
}

export interface TableListResult {
  count: number;
  tables: TableInfo[];
}
