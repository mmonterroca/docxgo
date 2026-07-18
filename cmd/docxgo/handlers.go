// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
	"github.com/mmonterroca/docxgo/v2/pkg/template"
	"github.com/mmonterroca/docxgo/v2/themes"
)

// server holds document sessions for RPC mode.
type server struct {
	mu   sync.Mutex
	docs map[string]domain.Document
	seq  int
}

func newServer() *server {
	return &server{
		docs: make(map[string]domain.Document),
	}
}

// nextID generates a monotonically increasing session ID.
func (s *server) nextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("doc-%d", s.seq)
}

// storeDoc stores a document under a new session ID and returns that ID.
func (s *server) storeDoc(doc domain.Document) string {
	id := s.nextID()
	s.mu.Lock()
	s.docs[id] = doc
	s.mu.Unlock()
	return id
}

// getDoc retrieves a document by session ID.
func (s *server) getDoc(id string) (domain.Document, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.docs[id]
	return d, ok
}

// removeDoc removes a document from the session map.
func (s *server) removeDoc(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.docs[id]
	if ok {
		delete(s.docs, id)
	}
	return ok
}

// dispatch routes a request to the appropriate handler.
func (s *server) dispatch(req *Request) Response {
	switch req.Method {
	// system.*
	case "system.ping":
		return s.handleSystemPing(req)
	case "system.version":
		return s.handleSystemVersion(req)
	case "system.capabilities":
		return s.handleSystemCapabilities(req)
	case "system.batch":
		return s.handleSystemBatch(req)
	// document.*
	case "document.create":
		return s.handleCreate(req)
	case "document.open":
		return s.handleOpen(req)
	case "document.save":
		return s.handleSave(req)
	case "document.validate":
		return s.handleValidate(req)
	case "document.inspect":
		return s.handleInspect(req)
	case "document.setMetadata":
		return s.handleSetMetadata(req)
	case "document.setLanguage":
		return s.handleSetLanguage(req)
	case "document.setBackgroundColor":
		return s.handleSetBackgroundColor(req)
	case "document.addContent":
		return s.handleAddContent(req)
	case "document.addPageBreak":
		return s.handleAddPageBreak(req)
	case "document.applyPatch":
		return s.handleApplyPatch(req)
	case "document.close":
		return s.handleClose(req)
	// paragraph.*
	case "paragraph.add":
		return s.handleParagraphAdd(req)
	case "paragraph.list":
		return s.handleParagraphList(req)
	// table.*
	case "table.add":
		return s.handleTableAdd(req)
	case "table.list":
		return s.handleTableList(req)
	// section.*
	case "section.add":
		return s.handleSectionAdd(req)
	// template.*
	case "template.inspect":
		return s.handleTemplateInspect(req)
	case "template.render":
		return s.handleTemplateRender(req)
	default:
		return errorResponse(req.ID, "METHOD_NOT_FOUND",
			fmt.Sprintf("unknown method: %s", req.Method), req.Method)
	}
}

// ─── Param types ─────────────────────────────────────────────────────────────

// createParams are the parameters for document.create.
type createParams struct {
	Options  *docOptions       `json:"options,omitempty"`
	Content  []json.RawMessage `json:"content,omitempty"`
	Output   string            `json:"output"` // "buffer" | "file"
	FilePath string            `json:"filePath,omitempty"`
}

// docOptions configures the document.
type docOptions struct {
	Title    string      `json:"title,omitempty"`
	Author   string      `json:"author,omitempty"`
	Subject  string      `json:"subject,omitempty"`
	PageSize interface{} `json:"pageSize,omitempty"` // string or object
	Margins  interface{} `json:"margins,omitempty"`  // string or object
	Theme    string      `json:"theme,omitempty"`
}

// openParams are the parameters for document.open.
type openParams struct {
	FilePath string `json:"filePath,omitempty"`
	Base64   string `json:"base64,omitempty"`
}

// saveParams are the parameters for document.save.
type saveParams struct {
	DocumentID string `json:"documentId"`
	Output     string `json:"output"` // "buffer" | "file"
	FilePath   string `json:"filePath,omitempty"`
}

// validateParams are the parameters for document.validate.
type validateParams struct {
	DocumentID string `json:"documentId"`
}

// inspectParams are the parameters for document.inspect.
type inspectParams struct {
	DocumentID string `json:"documentId"`
}

// setMetadataParams are the parameters for document.setMetadata.
type setMetadataParams struct {
	DocumentID string `json:"documentId"`
	metadataFields
}

// metadataFields are the document.setMetadata fields, shared with the
// setMetadata operation in document.applyPatch.
type metadataFields struct {
	Title       string   `json:"title,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Creator     string   `json:"creator,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Created     string   `json:"created,omitempty"`
	Modified    string   `json:"modified,omitempty"`
}

// metadataFromFields builds a domain.Metadata from the wire fields.
func metadataFromFields(f metadataFields) *domain.Metadata {
	return &domain.Metadata{
		Title:       f.Title,
		Subject:     f.Subject,
		Creator:     f.Creator,
		Description: f.Description,
		Keywords:    f.Keywords,
		Created:     f.Created,
		Modified:    f.Modified,
	}
}

// setLanguageParams are the parameters for document.setLanguage. Val,
// EastAsia, and Bidi are BCP 47 language tags (e.g. "es-MX", "en-US"); at
// least one must be set.
type setLanguageParams struct {
	DocumentID string `json:"documentId"`
	languageFields
}

// languageFields are the document.setLanguage fields, shared with the
// setLanguage operation in document.applyPatch.
type languageFields struct {
	Val      string `json:"val,omitempty"`
	EastAsia string `json:"eastAsia,omitempty"`
	Bidi     string `json:"bidi,omitempty"`
}

// languageFromFields builds a domain.Language from the wire fields.
func languageFromFields(f languageFields) *domain.Language {
	return &domain.Language{Val: f.Val, EastAsia: f.EastAsia, Bidi: f.Bidi}
}

// setBackgroundColorParams are the parameters for document.setBackgroundColor.
type setBackgroundColorParams struct {
	DocumentID string `json:"documentId"`
	Color      string `json:"color"` // hex string e.g. "#FF0000"
}

// applyBackgroundColor parses colorHex and sets it as doc's background color,
// shared between document.setBackgroundColor and the setBackgroundColor
// operation in document.applyPatch.
func applyBackgroundColor(doc domain.Document, colorHex string) error {
	color, err := parseHexColor(colorHex)
	if err != nil {
		return fmt.Errorf("invalid color: %w", err)
	}
	return doc.SetBackgroundColor(color)
}

// closeParams are the parameters for document.close.
type closeParams struct {
	DocumentID string `json:"documentId"`
}

// addContentParams are the parameters for document.addContent.
type addContentParams struct {
	DocumentID string            `json:"documentId"`
	Content    []json.RawMessage `json:"content"`
}

// addPageBreakParams are the parameters for document.addPageBreak.
type addPageBreakParams struct {
	DocumentID string `json:"documentId"`
}

// paragraphAddParams are the parameters for paragraph.add.
type paragraphAddParams struct {
	DocumentID string `json:"documentId"`
	paragraphItem
}

// paragraphListParams are the parameters for paragraph.list.
type paragraphListParams struct {
	DocumentID string `json:"documentId"`
}

// tableAddParams are the parameters for table.add.
type tableAddParams struct {
	DocumentID string `json:"documentId"`
	tableItem
}

// tableListParams are the parameters for table.list.
type tableListParams struct {
	DocumentID string `json:"documentId"`
}

// sectionAddParams are the parameters for section.add.
type sectionAddParams struct {
	DocumentID string `json:"documentId"`
	sectionItem
}

// applyPatchParams are the parameters for document.applyPatch.
type applyPatchParams struct {
	DocumentID string            `json:"documentId"`
	Operations []json.RawMessage `json:"operations"`
}

// patchOperation represents a single operation within document.applyPatch.
type patchOperation struct {
	Op string `json:"op"`
}

// templateInspectParams are the parameters for template.inspect.
type templateInspectParams struct {
	DocumentID     string `json:"documentId"`
	OpenDelimiter  string `json:"openDelimiter,omitempty"`
	CloseDelimiter string `json:"closeDelimiter,omitempty"`
}

// templateRenderParams are the parameters for template.render.
type templateRenderParams struct {
	DocumentID     string            `json:"documentId"`
	Data           map[string]string `json:"data"`
	StrictMode     bool              `json:"strictMode,omitempty"`
	OpenDelimiter  string            `json:"openDelimiter,omitempty"`
	CloseDelimiter string            `json:"closeDelimiter,omitempty"`
}

// batchRequest is a single request inside a system.batch call.
type batchRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// batchParams are the parameters for system.batch.
type batchParams struct {
	Requests []batchRequest `json:"requests"`
}

// ─── Content item types ──────────────────────────────────────────────────────

// contentItem is a generic content item with a type discriminator.
type contentItem struct {
	Type string `json:"type"`
}

// paragraphItem represents a paragraph content item.
type paragraphItem struct {
	Type          string          `json:"type"`
	Text          string          `json:"text,omitempty"`
	Runs          []runItem       `json:"runs,omitempty"`
	Style         string          `json:"style,omitempty"`
	Alignment     string          `json:"alignment,omitempty"`
	SpacingBefore int             `json:"spacingBefore,omitempty"`
	SpacingAfter  int             `json:"spacingAfter,omitempty"`
	LineSpacing   *lineSpacingDef `json:"lineSpacing,omitempty"`
	Indent        *indentDef      `json:"indent,omitempty"`
	Numbering     *numberingDef   `json:"numbering,omitempty"`
	Borders       *paraBodersDef  `json:"borders,omitempty"`
}

// runItem represents a text run within a paragraph.
type runItem struct {
	Text      string        `json:"text,omitempty"`
	Bold      bool          `json:"bold,omitempty"`
	Italic    bool          `json:"italic,omitempty"`
	Underline string        `json:"underline,omitempty"`
	Strike    bool          `json:"strike,omitempty"`
	Color     string        `json:"color,omitempty"`    // hex string
	FontSize  int           `json:"fontSize,omitempty"` // in points
	Font      string        `json:"font,omitempty"`
	Highlight string        `json:"highlight,omitempty"`
	Hyperlink *hyperlinkDef `json:"hyperlink,omitempty"`
	Field     *fieldDef     `json:"field,omitempty"`
	Image     *imageDef     `json:"image,omitempty"`
	Break     string        `json:"break,omitempty"` // "page", "column", "line"
}

// hyperlinkDef defines a hyperlink.
type hyperlinkDef struct {
	URL         string `json:"url"`
	DisplayText string `json:"displayText,omitempty"`
}

// fieldDef defines a document field.
type fieldDef struct {
	Type    string            `json:"type"`              // "pageNumber", "pageCount", "toc", "hyperlink", "styleRef"
	URL     string            `json:"url,omitempty"`     // for hyperlink fields
	Display string            `json:"display,omitempty"` // for hyperlink fields
	Style   string            `json:"style,omitempty"`   // for styleRef fields
	Options map[string]string `json:"options,omitempty"` // for TOC fields
}

// imageDef defines an image in a run.
type imageDef struct {
	Path     string `json:"path,omitempty"`
	Base64   string `json:"base64,omitempty"`
	Format   string `json:"format,omitempty"` // "png", "jpeg", etc.
	WidthPx  int    `json:"widthPx,omitempty"`
	HeightPx int    `json:"heightPx,omitempty"`
}

// tableItem represents a table content item.
type tableItem struct {
	Type      string         `json:"type"`
	Rows      []tableRowDef  `json:"rows,omitempty"`
	Alignment string         `json:"alignment,omitempty"`
	Style     string         `json:"style,omitempty"`
	Width     *tableWidthDef `json:"width,omitempty"`
}

// tableRowDef represents a table row.
type tableRowDef struct {
	Height int            `json:"height,omitempty"` // in twips
	Cells  []tableCellDef `json:"cells,omitempty"`
}

// tableCellDef represents a table cell.
type tableCellDef struct {
	Paragraphs        []json.RawMessage `json:"paragraphs,omitempty"`
	Width             int               `json:"width,omitempty"`
	VerticalAlignment string            `json:"verticalAlignment,omitempty"`
	Borders           *tableBordersDef  `json:"borders,omitempty"`
	Shading           string            `json:"shading,omitempty"` // hex color
	GridSpan          int               `json:"gridSpan,omitempty"`
}

// tableWidthDef represents table width.
type tableWidthDef struct {
	Type  string `json:"type,omitempty"` // "auto", "dxa", "pct"
	Value int    `json:"value,omitempty"`
}

// tableBordersDef represents table cell borders.
type tableBordersDef struct {
	Top    *borderStyleDef `json:"top,omitempty"`
	Left   *borderStyleDef `json:"left,omitempty"`
	Bottom *borderStyleDef `json:"bottom,omitempty"`
	Right  *borderStyleDef `json:"right,omitempty"`
}

// borderStyleDef represents a border style.
type borderStyleDef struct {
	Style string `json:"style,omitempty"` // "none", "single", "dotted", etc.
	Width int    `json:"width,omitempty"` // in eighths of a point
	Color string `json:"color,omitempty"` // hex string
}

// sectionItem represents a section content item.
type sectionItem struct {
	Type        string                       `json:"type"`
	PageSize    interface{}                  `json:"pageSize,omitempty"`    // string or object
	Margins     interface{}                  `json:"margins,omitempty"`     // string or object
	Orientation string                       `json:"orientation,omitempty"` // "portrait" | "landscape"
	Columns     int                          `json:"columns,omitempty"`
	BreakType   string                       `json:"breakType,omitempty"` // "nextPage", "continuous", "evenPage", "oddPage"
	Headers     map[string][]json.RawMessage `json:"headers,omitempty"`
	Footers     map[string][]json.RawMessage `json:"footers,omitempty"`
}

// lineSpacingDef defines line spacing.
type lineSpacingDef struct {
	Rule  string `json:"rule,omitempty"`  // "auto", "exact", "atLeast"
	Value int    `json:"value,omitempty"` // in twips
}

// indentDef defines paragraph indentation.
type indentDef struct {
	Left      int `json:"left,omitempty"`
	Right     int `json:"right,omitempty"`
	FirstLine int `json:"firstLine,omitempty"`
	Hanging   int `json:"hanging,omitempty"`
}

// numberingDef defines a numbering reference.
type numberingDef struct {
	ID    int `json:"id"`
	Level int `json:"level"`
}

// paraBodersDef defines paragraph borders.
type paraBodersDef struct {
	Top    *borderStyleDef `json:"top,omitempty"`
	Bottom *borderStyleDef `json:"bottom,omitempty"`
	Left   *borderStyleDef `json:"left,omitempty"`
	Right  *borderStyleDef `json:"right,omitempty"`
}

// ─── Handlers ────────────────────────────────────────────────────────────────

func (s *server) handleCreate(req *Request) Response {
	const op = "document.create"

	var params createParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc := docx.NewDocument()

	// Apply options
	if params.Options != nil {
		if rpcErr := applyDocOptions(doc, params.Options); rpcErr != nil {
			rpcErr.Operation = op
			return Response{ID: req.ID, Error: rpcErr}
		}
	}

	// Apply content
	if err := applyContent(doc, params.Content); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	docID := s.storeDoc(doc)
	result, rpcErr := serializeOutput(doc, params.Output, params.FilePath, docID)
	if rpcErr != nil {
		rpcErr.Operation = op
		return Response{ID: req.ID, Error: rpcErr}
	}
	return Response{ID: req.ID, Result: result}
}

func (s *server) handleOpen(req *Request) Response {
	const op = "document.open"

	var params openParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	var doc domain.Document
	var err error

	switch {
	case params.FilePath != "":
		doc, err = docx.OpenDocument(params.FilePath)
	case params.Base64 != "":
		data, decErr := base64.StdEncoding.DecodeString(params.Base64)
		if decErr != nil {
			return errorResponse(req.ID, errors.ErrCodeValidation, "invalid base64: "+decErr.Error(), op)
		}
		doc, err = docx.OpenDocumentFromBytes(data)
	default:
		return errorResponse(req.ID, errors.ErrCodeValidation, "filePath or base64 is required", op)
	}

	if err != nil {
		return errorResponse(req.ID, errors.ErrCodeIO, err.Error(), op)
	}

	docID := s.storeDoc(doc)
	return Response{ID: req.ID, Result: map[string]interface{}{
		"documentId": docID,
	}}
}

func (s *server) handleSave(req *Request) Response {
	const op = "document.save"

	var params saveParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	result, rpcErr := serializeOutput(doc, params.Output, params.FilePath, params.DocumentID)
	if rpcErr != nil {
		rpcErr.Operation = op
		return Response{ID: req.ID, Error: rpcErr}
	}
	return Response{ID: req.ID, Result: result}
}

func (s *server) handleValidate(req *Request) Response {
	const op = "document.validate"

	var params validateParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := doc.Validate(); err != nil {
		return Response{ID: req.ID, Result: map[string]interface{}{
			"valid":   false,
			"message": err.Error(),
		}}
	}
	return Response{ID: req.ID, Result: map[string]interface{}{
		"valid": true,
	}}
}

func (s *server) handleInspect(req *Request) Response {
	const op = "document.inspect"

	var params inspectParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	paragraphs := doc.Paragraphs()
	tables := doc.Tables()

	texts := make([]string, 0, len(paragraphs))
	for _, p := range paragraphs {
		if t := p.Text(); t != "" {
			texts = append(texts, t)
		}
	}

	result := map[string]interface{}{
		"paragraphCount": len(paragraphs),
		"tableCount":     len(tables),
		"text":           texts,
	}

	if meta := doc.Metadata(); meta != nil {
		result["metadata"] = map[string]interface{}{
			"title":       meta.Title,
			"subject":     meta.Subject,
			"creator":     meta.Creator,
			"description": meta.Description,
			"keywords":    meta.Keywords,
			"created":     meta.Created,
			"modified":    meta.Modified,
		}
	}

	if bgColor, hasBg := doc.BackgroundColor(); hasBg {
		result["backgroundColor"] = fmt.Sprintf("#%02X%02X%02X", bgColor.R, bgColor.G, bgColor.B)
	}

	if lang := doc.Language(); lang != nil {
		result["language"] = map[string]interface{}{
			"val":      lang.Val,
			"eastAsia": lang.EastAsia,
			"bidi":     lang.Bidi,
		}
	}

	return Response{ID: req.ID, Result: result}
}

func (s *server) handleSetMetadata(req *Request) Response {
	const op = "document.setMetadata"

	var params setMetadataParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := doc.SetMetadata(metadataFromFields(params.metadataFields)); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

// handleSetLanguage sets the document's default proofing language. Note:
// SetLanguage rejects documents opened via document.open (round-trip guard —
// see domain.Document.SetLanguage), so this only works on documents created
// via document.create. document.inspect reports the language of an opened
// document regardless, since the reader hydrates it separately.
func (s *server) handleSetLanguage(req *Request) Response {
	const op = "document.setLanguage"

	var params setLanguageParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := doc.SetLanguage(languageFromFields(params.languageFields)); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

func (s *server) handleSetBackgroundColor(req *Request) Response {
	const op = "document.setBackgroundColor"

	var params setBackgroundColorParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := applyBackgroundColor(doc, params.Color); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

// ─── Mutation handlers ────────────────────────────────────────────────────────

func (s *server) handleAddContent(req *Request) Response {
	const op = "document.addContent"

	var params addContentParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if len(params.Content) == 0 {
		return errorResponse(req.ID, errors.ErrCodeValidation, "content array is required and must not be empty", op)
	}

	if err := applyContent(doc, params.Content); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

func (s *server) handleAddPageBreak(req *Request) Response {
	const op = "document.addPageBreak"

	var params addPageBreakParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := doc.AddPageBreak(); err != nil {
		return errorResponse(req.ID, errors.ErrCodeInternal, err.Error(), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

func (s *server) handleParagraphAdd(req *Request) Response {
	const op = "paragraph.add"

	var params paragraphAddParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	para, err := doc.AddParagraph()
	if err != nil {
		return errorResponse(req.ID, errors.ErrCodeInternal, "failed to add paragraph: "+err.Error(), op)
	}

	if err := applyParagraph(para, params.paragraphItem); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	index := len(doc.Paragraphs()) - 1
	return Response{ID: req.ID, Result: map[string]interface{}{
		"ok":    true,
		"index": index,
	}}
}

func (s *server) handleParagraphList(req *Request) Response {
	const op = "paragraph.list"

	var params paragraphListParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	paragraphs := doc.Paragraphs()
	items := make([]map[string]interface{}, 0, len(paragraphs))
	for i, p := range paragraphs {
		item := map[string]interface{}{
			"index": i,
			"text":  p.Text(),
		}
		if s := p.Style(); s != nil && s.Name() != "" {
			item["style"] = s.Name()
		}
		items = append(items, item)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{
		"paragraphs": items,
		"count":      len(items),
	}}
}

func (s *server) handleTableAdd(req *Request) Response {
	const op = "table.add"

	var params tableAddParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := applyTable(doc, params.tableItem); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	index := len(doc.Tables()) - 1
	return Response{ID: req.ID, Result: map[string]interface{}{
		"ok":    true,
		"index": index,
	}}
}

func (s *server) handleTableList(req *Request) Response {
	const op = "table.list"

	var params tableListParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	tables := doc.Tables()
	items := make([]map[string]interface{}, 0, len(tables))
	for i, t := range tables {
		items = append(items, map[string]interface{}{
			"index":   i,
			"rows":    t.RowCount(),
			"columns": t.ColumnCount(),
		})
	}

	return Response{ID: req.ID, Result: map[string]interface{}{
		"tables": items,
		"count":  len(items),
	}}
}

func (s *server) handleSectionAdd(req *Request) Response {
	const op = "section.add"

	var params sectionAddParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if err := applySection(doc, params.sectionItem); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, err.Error(), op)
	}

	index := len(doc.Sections()) - 1
	return Response{ID: req.ID, Result: map[string]interface{}{
		"ok":    true,
		"index": index,
	}}
}

func (s *server) handleClose(req *Request) Response {
	const op = "document.close"

	var params closeParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	if params.DocumentID == "" {
		return errorResponse(req.ID, errors.ErrCodeValidation, "documentId is required", op)
	}

	if !s.removeDoc(params.DocumentID) {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{"ok": true}}
}

// ─── System handlers ──────────────────────────────────────────────────────────

func (s *server) handleSystemPing(req *Request) Response {
	return Response{ID: req.ID, Result: map[string]interface{}{
		"status": "ok",
	}}
}

func (s *server) handleSystemVersion(req *Request) Response {
	return Response{ID: req.ID, Result: map[string]interface{}{
		"name":            "docxgo",
		"version":         docx.Version,
		"protocolVersion": ProtocolVersion,
		"goVersion":       runtime.Version(),
		"platform":        runtime.GOOS,
		"arch":            runtime.GOARCH,
	}}
}

func (s *server) handleSystemCapabilities(req *Request) Response {
	return Response{ID: req.ID, Result: capabilitiesMap()}
}

func (s *server) handleSystemBatch(req *Request) Response {
	const op = "system.batch"

	var params batchParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	if len(params.Requests) == 0 {
		return errorResponse(req.ID, errors.ErrCodeValidation, "requests array is required and must not be empty", op)
	}

	results := make([]interface{}, 0, len(params.Requests))
	for i, br := range params.Requests {
		if br.Method == "system.batch" {
			results = append(results, map[string]interface{}{
				"error": map[string]interface{}{
					"code":    errors.ErrCodeValidation,
					"message": "system.batch cannot be nested",
				},
			})
			continue
		}

		subReq := &Request{
			ID:     i + 1,
			Method: br.Method,
			Params: br.Params,
		}
		resp := s.dispatch(subReq)

		if resp.Error != nil {
			results = append(results, map[string]interface{}{
				"error": resp.Error,
			})
		} else {
			results = append(results, map[string]interface{}{
				"result": resp.Result,
			})
		}
	}

	return Response{ID: req.ID, Result: map[string]interface{}{
		"responses": results,
	}}
}

// ─── Template handlers ───────────────────────────────────────────────────────

func (s *server) handleTemplateInspect(req *Request) Response {
	const op = "template.inspect"

	var params templateInspectParams
	raw := req.Params
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	opts := template.DefaultMergeOptions()
	if params.OpenDelimiter != "" {
		opts.OpenDelimiter = params.OpenDelimiter
	}
	if params.CloseDelimiter != "" {
		opts.CloseDelimiter = params.CloseDelimiter
	}
	pattern := template.BuildPattern(opts)
	placeholders := template.FindPlaceholdersCustom(doc, pattern)

	// Build unique names list (preserving first-seen order)
	seen := make(map[string]struct{})
	var names []string
	for _, p := range placeholders {
		if _, ok := seen[p.Name]; !ok {
			seen[p.Name] = struct{}{}
			names = append(names, p.Name)
		}
	}

	// Build detailed locations
	locations := make([]map[string]interface{}, 0, len(placeholders))
	for _, p := range placeholders {
		loc := map[string]interface{}{
			"name":      p.Name,
			"fullMatch": p.FullMatch,
			"location":  locationTypeName(p.Location.Type),
			"paragraph": p.Location.ParagraphIndex,
			"run":       p.Location.RunIndex,
		}
		// TableIndex/RowIndex/CellIndex are only meaningful for table-cell
		// placeholders; for a top-level paragraph, header, or footer they're
		// left at their zero value, so gate on the location type rather than
		// on TableIndex >= 0 (which is always true).
		if p.Location.Type == template.LocationTableCell {
			loc["table"] = p.Location.TableIndex
			loc["row"] = p.Location.RowIndex
			loc["cell"] = p.Location.CellIndex
		}
		locations = append(locations, loc)
	}

	return Response{ID: req.ID, Result: map[string]interface{}{
		"placeholders": names,
		"count":        len(names),
		"occurrences":  len(placeholders),
		"details":      locations,
	}}
}

func (s *server) handleTemplateRender(req *Request) Response {
	const op = "template.render"

	var params templateRenderParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if len(params.Data) == 0 {
		return errorResponse(req.ID, errors.ErrCodeValidation, "data map is required and must not be empty", op)
	}

	opts := template.DefaultMergeOptions()
	opts.StrictMode = params.StrictMode
	if params.OpenDelimiter != "" {
		opts.OpenDelimiter = params.OpenDelimiter
	}
	if params.CloseDelimiter != "" {
		opts.CloseDelimiter = params.CloseDelimiter
	}

	// Validate first. Note: in strict mode, MergeTemplate below already fails
	// hard (TEMPLATE_ERROR) on missing keys before this response is built, so
	// any issue that reaches a successful response here is by construction
	// non-fatal — report all of them as "warning", not their raw
	// ValidationSeverity, to avoid an ok=true response listing "error"
	// severities.
	validationErrors := template.ValidateTemplate(doc, template.MergeData(params.Data), opts)
	var warnings []map[string]interface{}
	for _, ve := range validationErrors {
		warnings = append(warnings, map[string]interface{}{
			"severity": "warning",
			"key":      ve.Key,
			"message":  ve.Message,
		})
	}

	// Perform the merge
	if err := template.MergeTemplate(doc, template.MergeData(params.Data), opts); err != nil {
		return errorResponseWithData(req.ID, "TEMPLATE_ERROR", err.Error(), op, map[string]interface{}{
			"category":  "merge",
			"retryable": false,
		})
	}

	result := map[string]interface{}{
		"ok": true,
	}
	if len(warnings) > 0 {
		result["warnings"] = warnings
	}

	return Response{ID: req.ID, Result: result}
}

// ─── ApplyPatch handler ──────────────────────────────────────────────────────

func (s *server) handleApplyPatch(req *Request) Response {
	const op = "document.applyPatch"

	var params applyPatchParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, errors.ErrCodeValidation, "invalid params: "+err.Error(), op)
	}

	doc, ok := s.getDoc(params.DocumentID)
	if !ok {
		return errorResponse(req.ID, errors.ErrCodeNotFound,
			fmt.Sprintf("document %q not found", params.DocumentID), op)
	}

	if len(params.Operations) == 0 {
		return errorResponse(req.ID, errors.ErrCodeValidation, "operations array is required and must not be empty", op)
	}

	applied := 0
	for i, raw := range params.Operations {
		var pOp patchOperation
		if err := json.Unmarshal(raw, &pOp); err != nil {
			return errorResponseWithData(req.ID, errors.ErrCodeValidation,
				fmt.Sprintf("invalid operation at index %d: %v", i, err), op,
				map[string]interface{}{"index": i, "applied": applied})
		}

		switch pOp.Op {
		case "appendParagraph":
			var pItem paragraphItem
			if err := json.Unmarshal(raw, &pItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid appendParagraph at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			para, err := doc.AddParagraph()
			if err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeInternal,
					fmt.Sprintf("failed to add paragraph at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := applyParagraph(para, pItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to apply paragraph at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "appendTable":
			var tItem tableItem
			if err := json.Unmarshal(raw, &tItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid appendTable at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := applyTable(doc, tItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to apply table at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "appendSection":
			var sItem sectionItem
			if err := json.Unmarshal(raw, &sItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid appendSection at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := applySection(doc, sItem); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to apply section at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "appendPageBreak":
			if err := doc.AddPageBreak(); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeInternal,
					fmt.Sprintf("failed to add page break at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "setMetadata":
			var mp struct {
				Op string `json:"op"`
				metadataFields
			}
			if err := json.Unmarshal(raw, &mp); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid setMetadata at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := doc.SetMetadata(metadataFromFields(mp.metadataFields)); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to set metadata at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "setBackgroundColor":
			var bcp struct {
				Op    string `json:"op"`
				Color string `json:"color"`
			}
			if err := json.Unmarshal(raw, &bcp); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid setBackgroundColor at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := applyBackgroundColor(doc, bcp.Color); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to set background color at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		case "setLanguage":
			var lp struct {
				Op string `json:"op"`
				languageFields
			}
			if err := json.Unmarshal(raw, &lp); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("invalid setLanguage at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}
			if err := doc.SetLanguage(languageFromFields(lp.languageFields)); err != nil {
				return errorResponseWithData(req.ID, errors.ErrCodeValidation,
					fmt.Sprintf("failed to set language at index %d: %v", i, err), op,
					map[string]interface{}{"index": i, "applied": applied})
			}

		default:
			return errorResponseWithData(req.ID, errors.ErrCodeValidation,
				fmt.Sprintf("unknown operation %q at index %d", pOp.Op, i), op,
				map[string]interface{}{"index": i, "op": pOp.Op, "applied": applied})
		}

		applied++
	}

	return Response{ID: req.ID, Result: map[string]interface{}{
		"ok":      true,
		"applied": applied,
	}}
}

// ─── Options application ──────────────────────────────────────────────────────

// applyDocOptions applies document options to a newly created document.
func applyDocOptions(doc domain.Document, opts *docOptions) *RPCError {
	// Metadata
	meta := &domain.Metadata{}
	if opts.Title != "" {
		meta.Title = opts.Title
	}
	if opts.Author != "" {
		meta.Creator = opts.Author
	}
	if opts.Subject != "" {
		meta.Subject = opts.Subject
	}
	if err := doc.SetMetadata(meta); err != nil {
		return &RPCError{Code: errors.ErrCodeValidation, Message: err.Error()}
	}

	// Page size and margins on the default section
	if opts.PageSize != nil || opts.Margins != nil {
		sec, err := doc.DefaultSection()
		if err != nil {
			return &RPCError{Code: errors.ErrCodeInternal, Message: "failed to get default section: " + err.Error()}
		}
		if opts.PageSize != nil {
			if err := sec.SetPageSize(parsePageSize(opts.PageSize)); err != nil {
				return &RPCError{Code: errors.ErrCodeValidation, Message: "failed to set page size: " + err.Error()}
			}
		}
		if opts.Margins != nil {
			if err := sec.SetMargins(parseMargins(opts.Margins)); err != nil {
				return &RPCError{Code: errors.ErrCodeValidation, Message: "failed to set margins: " + err.Error()}
			}
		}
	}

	// Theme
	if opts.Theme != "" {
		theme := lookupTheme(opts.Theme)
		if theme == nil {
			return &RPCError{
				Code:    errors.ErrCodeValidation,
				Message: fmt.Sprintf("unknown theme: %s", opts.Theme),
			}
		}
		if err := theme.ApplyTo(doc); err != nil {
			return &RPCError{Code: errors.ErrCodeValidation, Message: "theme error: " + err.Error()}
		}
	}

	return nil
}

// lookupTheme returns a themes.Theme by name.
func lookupTheme(name string) themes.Theme {
	switch strings.ToLower(name) {
	case "corporate":
		return themes.Corporate
	case "startup":
		return themes.Startup
	case "modern":
		return themes.Modern
	case "fintech":
		return themes.Fintech
	case "academic":
		return themes.Academic
	case "tech-presentation", "techpresentation":
		return themes.TechPresentation
	case "tech-darkmode", "techdarkmode":
		return themes.TechDarkMode
	}
	return nil
}

// ─── Content application ──────────────────────────────────────────────────────

// applyContent processes a content array and adds items to the document.
func applyContent(doc domain.Document, content []json.RawMessage) error {
	for _, raw := range content {
		var item contentItem
		if err := json.Unmarshal(raw, &item); err != nil {
			return fmt.Errorf("invalid content item: %w", err)
		}

		switch item.Type {
		case "paragraph":
			var pItem paragraphItem
			if err := json.Unmarshal(raw, &pItem); err != nil {
				return fmt.Errorf("invalid paragraph: %w", err)
			}
			para, err := doc.AddParagraph()
			if err != nil {
				return fmt.Errorf("failed to add paragraph: %w", err)
			}
			if err := applyParagraph(para, pItem); err != nil {
				return err
			}
		case "table":
			var tItem tableItem
			if err := json.Unmarshal(raw, &tItem); err != nil {
				return fmt.Errorf("invalid table: %w", err)
			}
			if err := applyTable(doc, tItem); err != nil {
				return err
			}
		case "section":
			var sItem sectionItem
			if err := json.Unmarshal(raw, &sItem); err != nil {
				return fmt.Errorf("invalid section: %w", err)
			}
			if err := applySection(doc, sItem); err != nil {
				return err
			}
		case "pageBreak":
			if err := doc.AddPageBreak(); err != nil {
				return fmt.Errorf("failed to add page break: %w", err)
			}
		default:
			return fmt.Errorf("unknown content type: %s", item.Type)
		}
	}
	return nil
}

// applyParagraph applies paragraph formatting and runs to a domain.Paragraph.
func applyParagraph(para domain.Paragraph, item paragraphItem) error {
	if item.Style != "" {
		_ = para.SetStyle(item.Style)
	}
	if item.Alignment != "" {
		if align, ok := parseAlignment(item.Alignment); ok {
			_ = para.SetAlignment(align)
		}
	}
	if item.SpacingBefore > 0 {
		_ = para.SetSpacingBefore(item.SpacingBefore)
	}
	if item.SpacingAfter > 0 {
		_ = para.SetSpacingAfter(item.SpacingAfter)
	}
	if item.LineSpacing != nil {
		_ = para.SetLineSpacing(parseLineSpacing(item.LineSpacing))
	}
	if item.Indent != nil {
		_ = para.SetIndent(domain.Indentation{
			Left:      item.Indent.Left,
			Right:     item.Indent.Right,
			FirstLine: item.Indent.FirstLine,
			Hanging:   item.Indent.Hanging,
		})
	}
	if item.Numbering != nil {
		_ = para.SetNumbering(domain.NumberingReference{
			ID:    item.Numbering.ID,
			Level: item.Numbering.Level,
		})
	}
	if item.Borders != nil {
		_ = para.SetBorders(parseParagraphBorders(item.Borders))
	}

	// If top-level "text" shortcut is provided and no explicit runs,
	// create a single run with that text.
	if item.Text != "" && len(item.Runs) == 0 {
		run, err := para.AddRun()
		if err != nil {
			return err
		}
		if err := run.SetText(item.Text); err != nil {
			return err
		}
	}

	for _, r := range item.Runs {
		if err := applyRun(para, r); err != nil {
			return err
		}
	}
	return nil
}

// applyRun applies a run item to a paragraph.
func applyRun(para domain.Paragraph, r runItem) error {
	if r.Hyperlink != nil {
		_, err := para.AddHyperlink(r.Hyperlink.URL, r.Hyperlink.DisplayText)
		return err
	}

	if r.Field != nil {
		return applyField(para, r.Field)
	}

	if r.Image != nil {
		return applyImage(para, r.Image)
	}

	if r.Break != "" {
		run, err := para.AddRun()
		if err != nil {
			return err
		}
		bt := parseBreakType(r.Break)
		return run.AddBreak(bt)
	}

	run, err := para.AddRun()
	if err != nil {
		return err
	}

	if r.Text != "" {
		if err := run.SetText(r.Text); err != nil {
			return err
		}
	}
	if r.Bold {
		_ = run.SetBold(true)
	}
	if r.Italic {
		_ = run.SetItalic(true)
	}
	if r.Strike {
		_ = run.SetStrike(true)
	}
	if r.Underline != "" {
		if u, ok := parseUnderline(r.Underline); ok {
			_ = run.SetUnderline(u)
		}
	}
	if r.Color != "" {
		if c, err := parseHexColor(r.Color); err == nil {
			_ = run.SetColor(c)
		}
	}
	if r.FontSize > 0 {
		_ = run.SetSize(r.FontSize * 2) // points → half-points
	}
	if r.Font != "" {
		_ = run.SetFont(domain.Font{Name: r.Font})
	}
	if r.Highlight != "" {
		if h, ok := parseHighlight(r.Highlight); ok {
			_ = run.SetHighlight(h)
		}
	}
	return nil
}

// applyField adds a field to a paragraph.
func applyField(para domain.Paragraph, f *fieldDef) error {
	run, err := para.AddRun()
	if err != nil {
		return err
	}
	switch strings.ToLower(f.Type) {
	case "pagenumber":
		return run.AddField(docx.NewPageNumberField())
	case "pagecount", "numpages":
		return run.AddField(docx.NewPageCountField())
	case "toc":
		return run.AddField(docx.NewTOCField(f.Options))
	case "hyperlink":
		return run.AddField(docx.NewHyperlinkField(f.URL, f.Display))
	case "styleref":
		return run.AddField(docx.NewStyleRefField(f.Style))
	default:
		return fmt.Errorf("unknown field type: %s", f.Type)
	}
}

// applyImage adds an image to a paragraph.
func applyImage(para domain.Paragraph, img *imageDef) error {
	// If base64 data is provided, use AddImageFromBytes directly (no temp file needed).
	if img.Base64 != "" {
		data, err := base64.StdEncoding.DecodeString(img.Base64)
		if err != nil {
			return fmt.Errorf("invalid image base64: %w", err)
		}
		format := domain.ImageFormat(img.Format)
		if format == "" {
			format = domain.ImageFormatPNG
		}
		if img.WidthPx > 0 || img.HeightPx > 0 {
			size := domain.NewImageSize(img.WidthPx, img.HeightPx)
			_, err = para.AddImageFromBytesWithSize(data, format, size)
			return err
		}
		_, err = para.AddImageFromBytes(data, format)
		return err
	}

	path := img.Path
	if path == "" {
		return fmt.Errorf("image requires path or base64")
	}

	if img.WidthPx > 0 || img.HeightPx > 0 {
		size := domain.NewImageSize(img.WidthPx, img.HeightPx)
		_, err := para.AddImageWithSize(path, size)
		return err
	}
	_, err := para.AddImage(path)
	return err
}

// applyTable adds a table to the document.
func applyTable(doc domain.Document, item tableItem) error {
	if len(item.Rows) == 0 {
		return nil
	}

	cols := 0
	for _, row := range item.Rows {
		if len(row.Cells) > cols {
			cols = len(row.Cells)
		}
	}
	if cols == 0 {
		return nil
	}

	table, err := doc.AddTable(len(item.Rows), cols)
	if err != nil {
		return fmt.Errorf("failed to add table: %w", err)
	}

	if item.Style != "" {
		_ = table.SetStyle(domain.TableStyle{Name: item.Style})
	}
	if item.Alignment != "" {
		if align, ok := parseAlignment(item.Alignment); ok {
			_ = table.SetAlignment(align)
		}
	}
	if item.Width != nil {
		_ = table.SetWidth(parseTableWidth(item.Width))
	}

	for rowIdx, rowDef := range item.Rows {
		row, err := table.Row(rowIdx)
		if err != nil {
			return fmt.Errorf("failed to get row %d: %w", rowIdx, err)
		}
		if rowDef.Height > 0 {
			_ = row.SetHeight(rowDef.Height)
		}
		for colIdx, cellDef := range rowDef.Cells {
			cell, err := row.Cell(colIdx)
			if err != nil {
				return fmt.Errorf("failed to get cell [%d,%d]: %w", rowIdx, colIdx, err)
			}
			if cellDef.Width > 0 {
				_ = cell.SetWidth(cellDef.Width)
			}
			if cellDef.VerticalAlignment != "" {
				if va, ok := parseVerticalAlignment(cellDef.VerticalAlignment); ok {
					_ = cell.SetVerticalAlignment(va)
				}
			}
			if cellDef.Borders != nil {
				_ = cell.SetBorders(parseTableBorders(cellDef.Borders))
			}
			if cellDef.Shading != "" {
				if c, err := parseHexColor(cellDef.Shading); err == nil {
					_ = cell.SetShading(c)
				}
			}
			if cellDef.GridSpan > 1 {
				_ = cell.SetGridSpan(cellDef.GridSpan)
			}
			for _, paraRaw := range cellDef.Paragraphs {
				var pItem paragraphItem
				if err := json.Unmarshal(paraRaw, &pItem); err != nil {
					return fmt.Errorf("invalid cell paragraph: %w", err)
				}
				para, err := cell.AddParagraph()
				if err != nil {
					return fmt.Errorf("failed to add cell paragraph: %w", err)
				}
				if err := applyParagraph(para, pItem); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// applySection adds a new section to the document.
func applySection(doc domain.Document, item sectionItem) error {
	bt := parseSectionBreakType(item.BreakType)
	sec, err := doc.AddSectionWithBreak(bt)
	if err != nil {
		return fmt.Errorf("failed to add section: %w", err)
	}

	if item.PageSize != nil {
		if err := sec.SetPageSize(parsePageSize(item.PageSize)); err != nil {
			return fmt.Errorf("failed to set page size: %w", err)
		}
	}
	if item.Margins != nil {
		if err := sec.SetMargins(parseMargins(item.Margins)); err != nil {
			return fmt.Errorf("failed to set margins: %w", err)
		}
	}
	switch item.Orientation {
	case "landscape":
		if err := sec.SetOrientation(domain.OrientationLandscape); err != nil {
			return fmt.Errorf("failed to set orientation: %w", err)
		}
	case "portrait":
		if err := sec.SetOrientation(domain.OrientationPortrait); err != nil {
			return fmt.Errorf("failed to set orientation: %w", err)
		}
	}
	if item.Columns > 0 {
		if err := sec.SetColumns(item.Columns); err != nil {
			return fmt.Errorf("failed to set columns: %w", err)
		}
	}

	for hType, content := range item.Headers {
		header, err := sec.Header(parseHeaderType(hType))
		if err != nil {
			return fmt.Errorf("failed to get header %q: %w", hType, err)
		}
		for _, paraRaw := range content {
			var pItem paragraphItem
			if err := json.Unmarshal(paraRaw, &pItem); err != nil {
				return fmt.Errorf("invalid header paragraph: %w", err)
			}
			para, err := header.AddParagraph()
			if err != nil {
				return fmt.Errorf("failed to add header paragraph: %w", err)
			}
			if err := applyParagraph(para, pItem); err != nil {
				return fmt.Errorf("failed to apply header paragraph: %w", err)
			}
		}
	}

	for fType, content := range item.Footers {
		footer, err := sec.Footer(parseFooterType(fType))
		if err != nil {
			return fmt.Errorf("failed to get footer %q: %w", fType, err)
		}
		for _, paraRaw := range content {
			var pItem paragraphItem
			if err := json.Unmarshal(paraRaw, &pItem); err != nil {
				return fmt.Errorf("invalid footer paragraph: %w", err)
			}
			para, err := footer.AddParagraph()
			if err != nil {
				return fmt.Errorf("failed to add footer paragraph: %w", err)
			}
			if err := applyParagraph(para, pItem); err != nil {
				return fmt.Errorf("failed to apply footer paragraph: %w", err)
			}
		}
	}

	return nil
}

// ─── Serialization ───────────────────────────────────────────────────────────

// serializeOutput serializes a document to the requested output format.
func serializeOutput(doc domain.Document, output, filePath, docID string) (interface{}, *RPCError) {
	switch output {
	case "file":
		if filePath == "" {
			return nil, &RPCError{Code: errors.ErrCodeValidation, Message: "filePath is required for output=file"}
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, &RPCError{Code: errors.ErrCodeIO, Message: "failed to create directory: " + err.Error()}
		}
		if err := doc.SaveAs(filePath); err != nil {
			return nil, &RPCError{Code: errors.ErrCodeIO, Message: err.Error()}
		}
		return map[string]interface{}{
			"filePath":   filePath,
			"documentId": docID,
		}, nil
	case "buffer", "":
		var buf bytes.Buffer
		if _, err := doc.WriteTo(&buf); err != nil {
			return nil, &RPCError{Code: errors.ErrCodeIO, Message: err.Error()}
		}
		return map[string]interface{}{
			"data":       base64.StdEncoding.EncodeToString(buf.Bytes()),
			"documentId": docID,
		}, nil
	default:
		return nil, &RPCError{
			Code:    errors.ErrCodeValidation,
			Message: fmt.Sprintf("unsupported output format: %q (use \"buffer\" or \"file\")", output),
		}
	}
}

// ─── Parse helpers ────────────────────────────────────────────────────────────

func parseHexColor(s string) (domain.Color, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return domain.Color{}, fmt.Errorf("invalid hex color: %q", s)
	}
	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return domain.Color{}, fmt.Errorf("invalid hex color: %w", err)
	}
	return domain.Color{
		R: uint8((val >> 16) & 0xFF),
		G: uint8((val >> 8) & 0xFF),
		B: uint8(val & 0xFF),
	}, nil
}

func parseAlignment(s string) (domain.Alignment, bool) {
	switch strings.ToLower(s) {
	case "left":
		return domain.AlignmentLeft, true
	case "center":
		return domain.AlignmentCenter, true
	case "right":
		return domain.AlignmentRight, true
	case "justify":
		return domain.AlignmentJustify, true
	case "distribute":
		return domain.AlignmentDistribute, true
	}
	return domain.AlignmentLeft, false
}

func parseUnderline(s string) (domain.UnderlineStyle, bool) {
	switch strings.ToLower(s) {
	case "none":
		return domain.UnderlineNone, true
	case "single":
		return domain.UnderlineSingle, true
	case "double":
		return domain.UnderlineDouble, true
	case "thick":
		return domain.UnderlineThick, true
	case "dotted":
		return domain.UnderlineDotted, true
	case "dashed":
		return domain.UnderlineDashed, true
	case "wave":
		return domain.UnderlineWave, true
	}
	return domain.UnderlineNone, false
}

func parseHighlight(s string) (domain.HighlightColor, bool) {
	switch strings.ToLower(s) {
	case "none":
		return domain.HighlightNone, true
	case "yellow":
		return domain.HighlightYellow, true
	case "green":
		return domain.HighlightGreen, true
	case "cyan":
		return domain.HighlightCyan, true
	case "magenta":
		return domain.HighlightMagenta, true
	case "blue":
		return domain.HighlightBlue, true
	case "red":
		return domain.HighlightRed, true
	case "darkblue":
		return domain.HighlightDarkBlue, true
	case "darkcyan":
		return domain.HighlightDarkCyan, true
	case "darkgreen":
		return domain.HighlightDarkGreen, true
	case "darkmagenta":
		return domain.HighlightDarkMagenta, true
	case "darkred":
		return domain.HighlightDarkRed, true
	case "darkyellow":
		return domain.HighlightDarkYellow, true
	case "darkgray":
		return domain.HighlightDarkGray, true
	case "lightgray":
		return domain.HighlightLightGray, true
	}
	return domain.HighlightNone, false
}

func parseBreakType(s string) domain.BreakType {
	switch strings.ToLower(s) {
	case "column":
		return domain.BreakTypeColumn
	case "line":
		return domain.BreakTypeLine
	default:
		return domain.BreakTypePage
	}
}

func parseLineSpacing(ls *lineSpacingDef) domain.LineSpacing {
	if ls == nil {
		return domain.LineSpacing{}
	}
	rule := domain.LineSpacingAuto
	switch strings.ToLower(ls.Rule) {
	case "exact":
		rule = domain.LineSpacingExact
	case "atleast":
		rule = domain.LineSpacingAtLeast
	}
	return domain.LineSpacing{Rule: rule, Value: ls.Value}
}

func parseParagraphBorders(b *paraBodersDef) domain.ParagraphBorders {
	if b == nil {
		return domain.ParagraphBorders{}
	}
	return domain.ParagraphBorders{
		Top:    parseBorderStyle(b.Top),
		Bottom: parseBorderStyle(b.Bottom),
		Left:   parseBorderStyle(b.Left),
		Right:  parseBorderStyle(b.Right),
	}
}

func parseTableBorders(b *tableBordersDef) domain.TableBorders {
	if b == nil {
		return domain.TableBorders{}
	}
	return domain.TableBorders{
		Top:    parseBorderStyle(b.Top),
		Left:   parseBorderStyle(b.Left),
		Bottom: parseBorderStyle(b.Bottom),
		Right:  parseBorderStyle(b.Right),
	}
}

func parseBorderStyle(b *borderStyleDef) domain.BorderStyle {
	if b == nil {
		return domain.BorderStyle{}
	}
	bs := domain.BorderStyle{Width: b.Width}
	if c, err := parseHexColor(b.Color); err == nil {
		bs.Color = c
	}
	switch strings.ToLower(b.Style) {
	case "single":
		bs.Style = domain.BorderSingle
	case "dotted":
		bs.Style = domain.BorderDotted
	case "dashed":
		bs.Style = domain.BorderDashed
	case "double":
		bs.Style = domain.BorderDouble
	case "triple":
		bs.Style = domain.BorderTriple
	case "thick":
		bs.Style = domain.BorderThick
	default:
		bs.Style = domain.BorderNone
	}
	return bs
}

func parseSectionBreakType(s string) domain.SectionBreakType {
	switch strings.ToLower(s) {
	case "continuous":
		return domain.SectionBreakTypeContinuous
	case "evenpage":
		return domain.SectionBreakTypeEvenPage
	case "oddpage":
		return domain.SectionBreakTypeOddPage
	default:
		return domain.SectionBreakTypeNextPage
	}
}

func parseHeaderType(s string) domain.HeaderType {
	switch strings.ToLower(s) {
	case "first":
		return domain.HeaderFirst
	case "even":
		return domain.HeaderEven
	default:
		return domain.HeaderDefault
	}
}

func parseFooterType(s string) domain.FooterType {
	switch strings.ToLower(s) {
	case "first":
		return domain.FooterFirst
	case "even":
		return domain.FooterEven
	default:
		return domain.FooterDefault
	}
}

func parseVerticalAlignment(s string) (domain.VerticalAlignment, bool) {
	switch strings.ToLower(s) {
	case "top":
		return domain.VerticalAlignTop, true
	case "center":
		return domain.VerticalAlignCenter, true
	case "bottom":
		return domain.VerticalAlignBottom, true
	}
	return domain.VerticalAlignTop, false
}

func parseTableWidth(w *tableWidthDef) domain.TableWidth {
	if w == nil {
		return domain.TableWidth{Type: domain.WidthAuto}
	}
	tw := domain.TableWidth{Value: w.Value}
	switch strings.ToLower(w.Type) {
	case "dxa":
		tw.Type = domain.WidthDXA
	case "pct":
		tw.Type = domain.WidthPct
	default:
		tw.Type = domain.WidthAuto
	}
	return tw
}

// parsePageSize parses a page size from a string preset or {width,height} object.
func parsePageSize(v interface{}) domain.PageSize {
	if v == nil {
		return domain.PageSizeLetter
	}
	switch t := v.(type) {
	case string:
		switch strings.ToUpper(t) {
		case "A4":
			return domain.PageSizeA4
		case "LETTER":
			return domain.PageSizeLetter
		case "LEGAL":
			return domain.PageSizeLegal
		case "A3":
			return domain.PageSizeA3
		case "TABLOID":
			return domain.PageSizeTableid
		}
	case map[string]interface{}:
		ps := domain.PageSize{}
		if w, ok := toInt(t["width"]); ok {
			ps.Width = w
		}
		if h, ok := toInt(t["height"]); ok {
			ps.Height = h
		}
		return ps
	}
	return domain.PageSizeLetter
}

// parseMargins parses margins from a string preset or {top,bottom,left,right} object.
func parseMargins(v interface{}) domain.Margins {
	if v == nil {
		return domain.DefaultMargins
	}
	switch t := v.(type) {
	case string:
		switch strings.ToLower(t) {
		case "normal":
			return domain.DefaultMargins
		case "narrow":
			return domain.Margins{Top: 720, Bottom: 720, Left: 720, Right: 720, Header: 360, Footer: 360}
		case "wide":
			return domain.Margins{Top: 1440, Bottom: 1440, Left: 2880, Right: 2880, Header: 720, Footer: 720}
		}
	case map[string]interface{}:
		m := domain.DefaultMargins
		if top, ok := toInt(t["top"]); ok {
			m.Top = top
		}
		if bottom, ok := toInt(t["bottom"]); ok {
			m.Bottom = bottom
		}
		if left, ok := toInt(t["left"]); ok {
			m.Left = left
		}
		if right, ok := toInt(t["right"]); ok {
			m.Right = right
		}
		if header, ok := toInt(t["header"]); ok {
			m.Header = header
		}
		if footer, ok := toInt(t["footer"]); ok {
			m.Footer = footer
		}
		return m
	}
	return domain.DefaultMargins
}

// toInt converts various JSON numeric types to int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := strconv.Atoi(string(n))
		return i, err == nil
	}
	return 0, false
}

// locationTypeName returns a human-readable name for a template.LocationType.
func locationTypeName(lt template.LocationType) string {
	switch lt {
	case template.LocationParagraph:
		return "paragraph"
	case template.LocationTableCell:
		return "tableCell"
	case template.LocationHeader:
		return "header"
	case template.LocationFooter:
		return "footer"
	default:
		return "unknown"
	}
}
