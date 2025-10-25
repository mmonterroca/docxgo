/*
   Copyright (c) 2025 SlideLang

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package constants provides OOXML constants and measurements for go-docx v2.
package constants

// Measurement conversions
const (
	// TwipsPerInch is the number of twips in one inch (1440).
	// A twip is 1/1440 of an inch, or 1/20 of a point.
	TwipsPerInch = 1440

	// TwipsPerPoint is the number of twips in one point (20).
	TwipsPerPoint = 20

	// TwipsPerCentimeter is approximately 567 twips per cm.
	TwipsPerCentimeter = 567

	// PointsPerInch is 72 points per inch.
	PointsPerInch = 72

	// EMUsPerInch is the number of English Metric Units in one inch (914400).
	EMUsPerInch = 914400

	// EMUsPerTwip is EMUsPerInch / TwipsPerInch (635).
	EMUsPerTwip = 635
)

// Default capacities for slices to reduce allocations
const (
	DefaultParagraphCapacity = 64
	DefaultRunCapacity       = 32
	DefaultTableCapacity     = 16
	DefaultRowCapacity       = 8
	DefaultCellCapacity      = 8
	DefaultStyleCapacity     = 64
	DefaultRelCapacity       = 32
	DefaultMediaCapacity     = 16
)

// OOXML Namespaces
const (
	// Main document namespace
	NamespaceMain = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

	// Relationships namespace
	NamespaceRelationships = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"

	// Package relationships namespace
	NamespacePackageRels = "http://schemas.openxmlformats.org/package/2006/relationships"

	// Drawing namespace
	NamespaceDrawing = "http://schemas.openxmlformats.org/drawingml/2006/main"

	// Picture namespace
	NamespacePicture = "http://schemas.openxmlformats.org/drawingml/2006/picture"

	// WordprocessingDrawing namespace
	NamespaceWordprocessingDrawing = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"

	// Content Types namespace
	NamespaceContentTypes = "http://schemas.openxmlformats.org/package/2006/content-types"

	// Core Properties namespace
	NamespaceCoreProperties = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"

	// Extended Properties namespace
	NamespaceExtendedProperties = "http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"

	// Dublin Core namespace
	NamespaceDC = "http://purl.org/dc/elements/1.1/"

	// Dublin Core Terms namespace
	NamespaceDCTerms = "http://purl.org/dc/terms/"
)

// OOXML Relationship Types
const (
	RelTypeDocument            = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	RelTypeStyles              = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	RelTypeNumbering           = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	RelTypeFontTable           = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"
	RelTypeSettings            = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	RelTypeWebSettings         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/webSettings"
	RelTypeTheme               = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
	RelTypeImage               = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	RelTypeHyperlink           = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	RelTypeHeader              = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
	RelTypeFooter              = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
	RelTypeFootnotes           = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"
	RelTypeEndnotes            = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/endnotes"
	RelTypeComments            = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"
	RelTypeCoreProperties      = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	RelTypeExtendedProperties  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties"
	RelTypeCustomProperties    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/custom-properties"
	RelTypeCustomXML           = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/customXml"
	RelTypeCustomXMLProperties = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/customXmlProps"
)

// Content Types
const (
	ContentTypeDocument           = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	ContentTypeStyles             = "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"
	ContentTypeNumbering          = "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"
	ContentTypeFontTable          = "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"
	ContentTypeSettings           = "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"
	ContentTypeWebSettings        = "application/vnd.openxmlformats-officedocument.wordprocessingml.webSettings+xml"
	ContentTypeTheme              = "application/vnd.openxmlformats-officedocument.theme+xml"
	ContentTypeHeader             = "application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"
	ContentTypeFooter             = "application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"
	ContentTypeFootnotes          = "application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml"
	ContentTypeEndnotes           = "application/vnd.openxmlformats-officedocument.wordprocessingml.endnotes+xml"
	ContentTypeComments           = "application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml"
	ContentTypeCoreProperties     = "application/vnd.openxmlformats-package.core-properties+xml"
	ContentTypeExtendedProperties = "application/vnd.openxmlformats-officedocument.extended-properties+xml"
	ContentTypeCustomProperties   = "application/vnd.openxmlformats-officedocument.custom-properties+xml"
	ContentTypeCustomXML          = "application/xml"
	ContentTypeRelationships      = "application/vnd.openxmlformats-package.relationships+xml"
	ContentTypePNG                = "image/png"
	ContentTypeJPEG               = "image/jpeg"
	ContentTypeGIF                = "image/gif"
	ContentTypeBMP                = "image/bmp"
	ContentTypeTIFF               = "image/tiff"
	ContentTypeWMF                = "image/x-wmf"
	ContentTypeEMF                = "image/x-emf"
)

// File paths within .docx archive
const (
	PathContentTypes = "[Content_Types].xml"
	PathRels         = "_rels/.rels"
	PathDocRels      = "word/_rels/document.xml.rels"
	PathDocument     = "word/document.xml"
	PathStyles       = "word/styles.xml"
	PathNumbering    = "word/numbering.xml"
	PathFontTable    = "word/fontTable.xml"
	PathSettings     = "word/settings.xml"
	PathWebSettings  = "word/webSettings.xml"
	PathTheme        = "word/theme/theme1.xml"
	PathCoreProps    = "docProps/core.xml"
	PathAppProps     = "docProps/app.xml"
	PathCustomProps  = "docProps/custom.xml"
	PathMediaPrefix  = "word/media/"
	PathHeaderPrefix = "word/header"
	PathFooterPrefix = "word/footer"
)

// Default font sizes (in half-points)
const (
	DefaultFontSize         = 22 // 11pt
	DefaultTitleSize        = 32 // 16pt
	DefaultHeading1Size     = 32 // 16pt
	DefaultHeading2Size     = 26 // 13pt
	DefaultHeading3Size     = 24 // 12pt
	DefaultSmallFontSize    = 18 // 9pt
	DefaultFootnoteFontSize = 20 // 10pt
)

// Default fonts
const (
	DefaultFontName      = "Calibri"
	DefaultMonospaceName = "Courier New"
	DefaultSerifName     = "Times New Roman"
	DefaultSansSerifName = "Arial"
)

// Default spacing values (in twips)
const (
	DefaultParagraphSpacing = 0
	DefaultLineSpacing      = 240 // Single spacing
	DefaultIndent           = 0
	DefaultFirstLineIndent  = 0
	DefaultHangingIndent    = 0
)

// Validation limits
const (
	// Font size limits (in half-points)
	MinFontSize = 2    // 1pt
	MaxFontSize = 3276 // 1638pt

	// Indentation limits (in twips)
	MinIndent = -31680 // -22 inches
	MaxIndent = 31680  // 22 inches

	// Spacing limits (in twips)
	MinSpacing = 0
	MaxSpacing = 31680 // 22 inches

	// Line spacing limits
	MinLineSpacing = 0
	MaxLineSpacing = 31680

	// Table dimensions
	MinTableRows = 1
	MaxTableRows = 1000
	MinTableCols = 1
	MaxTableCols = 63

	// Page columns
	MinColumns = 1
	MaxColumns = 10 // Maximum columns per page

	// Color component limits
	MinColorValue = 0
	MaxColorValue = 255
)

// Special IDs
const (
	IDPrefixParagraph = "para"
	IDPrefixRun       = "run"
	IDPrefixTable     = "tbl"
	IDPrefixRow       = "row"
	IDPrefixCell      = "cell"
	IDPrefixImage     = "img"
	IDPrefixShape     = "shp"
	IDPrefixRel       = "rId"
	IDPrefixBookmark  = "bm"
	IDPrefixComment   = "cmt"
	IDPrefixFootnote  = "fn"
	IDPrefixEndnote   = "en"
)

// OOXML string values for alignment
const (
	AlignmentValueLeft       = "left"
	AlignmentValueCenter     = "center"
	AlignmentValueRight      = "right"
	AlignmentValueJustify    = "both"
	AlignmentValueDistribute = "distribute"
)

// OOXML string values for vertical alignment
const (
	VerticalAlignmentValueTop    = "top"
	VerticalAlignmentValueCenter = "center"
	VerticalAlignmentValueBottom = "bottom"
)

// OOXML string values for underline styles
const (
	UnderlineValueNone   = "none"
	UnderlineValueSingle = "single"
	UnderlineValueDouble = "double"
	UnderlineValueThick  = "thick"
	UnderlineValueDotted = "dotted"
	UnderlineValueDashed = "dash"
	UnderlineValueWave   = "wave"
)

// OOXML string values for border styles
const (
	BorderValueNone   = "none"
	BorderValueSingle = "single"
	BorderValueDotted = "dotted"
	BorderValueDashed = "dashed"
	BorderValueDouble = "double"
	BorderValueTriple = "triple"
	BorderValueThick  = "thick"
)

// OOXML string values for line spacing rules
const (
	LineSpacingRuleAuto    = "auto"
	LineSpacingRuleExact   = "exact"
	LineSpacingRuleAtLeast = "atLeast"
)

// OOXML string values for width types
const (
	WidthTypeAuto = "auto"
	WidthTypeDXA  = "dxa"
	WidthTypePct  = "pct"
)

// OOXML string values for highlight colors
const (
	HighlightValueNone        = "none"
	HighlightValueYellow      = "yellow"
	HighlightValueGreen       = "green"
	HighlightValueCyan        = "cyan"
	HighlightValueMagenta     = "magenta"
	HighlightValueBlue        = "blue"
	HighlightValueRed         = "red"
	HighlightValueDarkBlue    = "darkBlue"
	HighlightValueDarkCyan    = "darkCyan"
	HighlightValueDarkGreen   = "darkGreen"
	HighlightValueDarkMagenta = "darkMagenta"
	HighlightValueDarkRed     = "darkRed"
	HighlightValueDarkYellow  = "darkYellow"
	HighlightValueDarkGray    = "darkGray"
	HighlightValueLightGray   = "lightGray"
)

// Field codes
const (
	FieldCodeTOC        = "TOC"
	FieldCodePageNumber = "PAGE"
	FieldCodeNumPages   = "NUMPAGES"
	FieldCodeDate       = "DATE"
	FieldCodeTime       = "TIME"
	FieldCodeStyleRef   = "STYLEREF"
	FieldCodeRef        = "REF"
	FieldCodeSeq        = "SEQ"
)
