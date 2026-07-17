// @mmonterroca/docxgo — Node.js wrapper for docxgo
//
// This package provides a TypeScript/JavaScript client for creating and
// manipulating Word documents via the docxgo CLI binary.

export { DocxgoRPC, DocxgoError } from './rpc';
export type { DocxgoRPCOptions } from './rpc';

export { DocxgoExec } from './exec';
export type { DocxgoExecOptions } from './exec';

export { DocumentBuilder } from './builder';
export type { DocumentBuilderOptions } from './builder';

export { resolveBinary } from './binary';

// Re-export all types
export type {
  // Protocol
  RPCRequest,
  RPCResponse,
  RPCError,
  // Document options
  DocumentOptions,
  PageSizePreset,
  PageSizeCustom,
  PageSize,
  MarginPreset,
  MarginCustom,
  Margins,
  ThemeName,
  // Content items
  Alignment,
  UnderlineStyle,
  HighlightColor,
  BreakType,
  SectionBreakType,
  Orientation,
  VerticalAlign,
  BorderStyleType,
  HeaderFooterType,
  WidthType,
  RunDef,
  HyperlinkDef,
  FieldDef,
  ImageDef,
  LineSpacingDef,
  IndentDef,
  NumberingDef,
  BorderDef,
  ParagraphBordersDef,
  ParagraphItem,
  TableBordersDef,
  TableCellDef,
  TableRowDef,
  TableWidthDef,
  TableItem,
  SectionItem,
  PageBreakItem,
  ContentItem,
  // Params
  CreateParams,
  OpenParams,
  SaveParams,
  ValidateParams,
  InspectParams,
  SetMetadataParams,
  SetBackgroundColorParams,
  SetLanguageParams,
  AddContentParams,
  AddPageBreakParams,
  ParagraphAddParams,
  ParagraphListParams,
  TableAddParams,
  TableListParams,
  SectionAddParams,
  CloseParams,
  // Results
  BufferResult,
  FileResult,
  CreateResult,
  SaveResult,
  OpenResult,
  ValidateResult,
  InspectResult,
  InspectMetadata,
  LanguageInfo,
  OkResult,
  IndexResult,
  ParagraphInfo,
  ParagraphListResult,
  TableInfo,
  TableListResult,
  // System
  PingResult,
  SystemVersionResult,
  SystemCapabilitiesResult,
  BatchRequest,
  BatchParams,
  BatchResponseEntry,
  BatchResult,
  // Template
  TemplateInspectParams,
  PlaceholderDetail,
  TemplateInspectResult,
  TemplateRenderParams,
  TemplateWarning,
  TemplateRenderResult,
  // ApplyPatch
  PatchOperationBase,
  AppendParagraphOp,
  AppendTableOp,
  AppendSectionOp,
  AppendPageBreakOp,
  SetMetadataOp,
  SetBackgroundColorOp,
  SetLanguageOp,
  PatchOperation,
  ApplyPatchParams,
  ApplyPatchResult,
} from './types';
