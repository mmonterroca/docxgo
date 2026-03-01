package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ─── Protocol tests ───────────────────────────────────────────────────────────

func TestErrorResponse(t *testing.T) {
	r := errorResponse(42, "VALIDATION_ERROR", "bad input", "document.create")
	if r.ID != 42 {
		t.Errorf("expected ID=42, got %v", r.ID)
	}
	if r.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if r.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", r.Error.Code)
	}
	if r.Error.Message != "bad input" {
		t.Errorf("unexpected message: %s", r.Error.Message)
	}
	if r.Error.Operation != "document.create" {
		t.Errorf("unexpected operation: %s", r.Error.Operation)
	}
	if r.Result != nil {
		t.Error("result should be nil for error response")
	}
}

func TestResponseJSON(t *testing.T) {
	resp := Response{
		ID:     1,
		Result: map[string]interface{}{"ok": true},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !bytes.Contains(data, []byte(`"result"`)) {
		t.Error("expected result field in JSON")
	}
	if bytes.Contains(data, []byte(`"error"`)) {
		t.Error("error field should be absent when nil")
	}
}

// ─── Parse helper tests ───────────────────────────────────────────────────────

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input string
		r, g, b uint8
		wantErr bool
	}{
		{"#FF0000", 255, 0, 0, false},
		{"FF0000", 255, 0, 0, false},
		{"#000000", 0, 0, 0, false},
		{"#FFFFFF", 255, 255, 255, false},
		{"FFF", 255, 255, 255, false}, // shorthand
		{"invalid", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}
	for _, tt := range tests {
		c, err := parseHexColor(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseHexColor(%q): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseHexColor(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if c.R != tt.r || c.G != tt.g || c.B != tt.b {
			t.Errorf("parseHexColor(%q) = {%d,%d,%d}, want {%d,%d,%d}",
				tt.input, c.R, c.G, c.B, tt.r, tt.g, tt.b)
		}
	}
}

func TestParseAlignment(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"left", true},
		{"center", true},
		{"right", true},
		{"justify", true},
		{"distribute", true},
		{"CENTER", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		_, ok := parseAlignment(tt.input)
		if ok != tt.ok {
			t.Errorf("parseAlignment(%q) ok=%v, want %v", tt.input, ok, tt.ok)
		}
	}
}

func TestParseUnderline(t *testing.T) {
	for _, s := range []string{"none", "single", "double", "thick", "dotted", "dashed", "wave"} {
		_, ok := parseUnderline(s)
		if !ok {
			t.Errorf("parseUnderline(%q) should be valid", s)
		}
	}
	_, ok := parseUnderline("invalid")
	if ok {
		t.Error("parseUnderline(invalid) should return false")
	}
}

func TestParsePageSize(t *testing.T) {
	tests := []struct{ input string; wantW, wantH int }{
		{"A4", 11906, 16838},
		{"Letter", 12240, 15840},
		{"Legal", 12240, 20160},
		{"A3", 16838, 23811},
		{"Tabloid", 15840, 24480},
		{"unknown", 12240, 15840}, // defaults to Letter
	}
	for _, tt := range tests {
		ps := parsePageSize(tt.input)
		if ps.Width != tt.wantW || ps.Height != tt.wantH {
			t.Errorf("parsePageSize(%q) = {%d,%d}, want {%d,%d}",
				tt.input, ps.Width, ps.Height, tt.wantW, tt.wantH)
		}
	}
}

func TestParseMargins(t *testing.T) {
	// string presets
	normal := parseMargins("normal")
	if normal.Top != 1440 {
		t.Errorf("normal margin top: want 1440, got %d", normal.Top)
	}
	narrow := parseMargins("narrow")
	if narrow.Top != 720 {
		t.Errorf("narrow margin top: want 720, got %d", narrow.Top)
	}

	// object form
	custom := parseMargins(map[string]interface{}{
		"top": float64(500), "bottom": float64(600),
		"left": float64(700), "right": float64(800),
	})
	if custom.Top != 500 || custom.Bottom != 600 || custom.Left != 700 || custom.Right != 800 {
		t.Errorf("custom margins: got %+v", custom)
	}
}

// ─── Server / dispatch tests ─────────────────────────────────────────────────

func makeRequest(id interface{}, method string, params interface{}) *Request {
	raw, _ := json.Marshal(params)
	return &Request{ID: id, Method: method, Params: raw}
}

func TestDispatch_UnknownMethod(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "document.unknown", map[string]interface{}{}))
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != "METHOD_NOT_FOUND" {
		t.Errorf("expected METHOD_NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleCreate_BasicBuffer(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Hello World", "bold": true},
				},
			},
		},
		"output": "buffer",
	}
	resp := s.dispatch(makeRequest(1, "document.create", params))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if _, ok := result["data"]; !ok {
		t.Error("expected 'data' field in result")
	}
	if _, ok := result["documentId"]; !ok {
		t.Error("expected 'documentId' field in result")
	}

	// Verify the base64 data decodes to a valid zip (docx is a zip archive)
	dataStr, _ := result["data"].(string)
	docxBytes, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}
	// DOCX files start with PK (zip magic bytes)
	if len(docxBytes) < 2 || docxBytes[0] != 'P' || docxBytes[1] != 'K' {
		t.Error("result data does not look like a DOCX (zip) file")
	}
}

func TestHandleCreate_WithOptions(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"options": map[string]interface{}{
			"title":    "Test Doc",
			"author":   "Test Author",
			"pageSize": "A4",
			"margins":  "narrow",
		},
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Content"},
				},
			},
		},
		"output": "buffer",
	}
	resp := s.dispatch(makeRequest(2, "document.create", params))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleCreate_WithTable(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type":  "table",
				"style": "TableGrid",
				"rows": []interface{}{
					map[string]interface{}{
						"cells": []interface{}{
							map[string]interface{}{
								"paragraphs": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"runs": []interface{}{
											map[string]interface{}{"text": "Cell 1"},
										},
									},
								},
							},
							map[string]interface{}{
								"paragraphs": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"runs": []interface{}{
											map[string]interface{}{"text": "Cell 2"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"output": "buffer",
	}
	resp := s.dispatch(makeRequest(3, "document.create", params))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleCreate_SaveToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := dir + "/test.docx"

	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "File output test"},
				},
			},
		},
		"output":   "file",
		"filePath": outPath,
	}
	resp := s.dispatch(makeRequest(4, "document.create", params))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestHandleCreate_PageBreak(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Page 1"},
				},
			},
			map[string]interface{}{"type": "pageBreak"},
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Page 2"},
				},
			},
		},
		"output": "buffer",
	}
	resp := s.dispatch(makeRequest(5, "document.create", params))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleCreate_InvalidParams(t *testing.T) {
	s := newServer()
	req := &Request{ID: 99, Method: "document.create", Params: json.RawMessage(`not-json`)}
	resp := s.dispatch(req)
	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
}

func TestHandleCreate_UnknownContentType(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{"type": "unknownType"},
		},
		"output": "buffer",
	}
	resp := s.dispatch(makeRequest(6, "document.create", params))
	if resp.Error == nil {
		t.Fatal("expected error for unknown content type")
	}
}

func TestHandleOpen_InvalidParams(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(7, "document.open", map[string]interface{}{}))
	if resp.Error == nil {
		t.Fatal("expected error when no filePath or base64 provided")
	}
}

func TestHandleOpen_NonexistentFile(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(8, "document.open", map[string]interface{}{
		"filePath": "/nonexistent/path/file.docx",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestHandleSave_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(9, "document.save", map[string]interface{}{
		"documentId": "no-such-doc",
		"output":     "buffer",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for unknown documentId")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleValidate(t *testing.T) {
	s := newServer()
	// Create a document first
	createResp := s.dispatch(makeRequest(10, "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Hello"},
				},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	result := createResp.Result.(map[string]interface{})
	docID := result["documentId"].(string)

	// Validate it
	validateResp := s.dispatch(makeRequest(11, "document.validate", map[string]interface{}{
		"documentId": docID,
	}))
	if validateResp.Error != nil {
		t.Fatalf("validate failed: %+v", validateResp.Error)
	}
	vResult := validateResp.Result.(map[string]interface{})
	if valid, ok := vResult["valid"].(bool); !ok || !valid {
		t.Errorf("expected valid=true, got %v", vResult)
	}
}

func TestHandleInspect(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(12, "document.create", map[string]interface{}{
		"options": map[string]interface{}{"title": "My Title"},
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Inspect me"},
				},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	result := createResp.Result.(map[string]interface{})
	docID := result["documentId"].(string)

	inspectResp := s.dispatch(makeRequest(13, "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspectResp.Error != nil {
		t.Fatalf("inspect failed: %+v", inspectResp.Error)
	}
	iResult := inspectResp.Result.(map[string]interface{})
	if _, ok := iResult["paragraphCount"]; !ok {
		t.Error("expected paragraphCount in inspect result")
	}
	if _, ok := iResult["text"]; !ok {
		t.Error("expected text in inspect result")
	}
}

func TestHandleSetMetadata(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(14, "document.create", map[string]interface{}{
		"content": []interface{}{},
		"output":  "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	result := createResp.Result.(map[string]interface{})
	docID := result["documentId"].(string)

	setResp := s.dispatch(makeRequest(15, "document.setMetadata", map[string]interface{}{
		"documentId": docID,
		"title":      "Updated Title",
		"creator":    "Test Author",
	}))
	if setResp.Error != nil {
		t.Fatalf("setMetadata failed: %+v", setResp.Error)
	}
}

func TestHandleSetBackgroundColor(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(16, "document.create", map[string]interface{}{
		"content": []interface{}{},
		"output":  "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	result := createResp.Result.(map[string]interface{})
	docID := result["documentId"].(string)

	bgResp := s.dispatch(makeRequest(17, "document.setBackgroundColor", map[string]interface{}{
		"documentId": docID,
		"color":      "#E0F0FF",
	}))
	if bgResp.Error != nil {
		t.Fatalf("setBackgroundColor failed: %+v", bgResp.Error)
	}
}

func TestHandleSetBackgroundColor_InvalidColor(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(18, "document.create", map[string]interface{}{
		"content": []interface{}{},
		"output":  "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	result := createResp.Result.(map[string]interface{})
	docID := result["documentId"].(string)

	bgResp := s.dispatch(makeRequest(19, "document.setBackgroundColor", map[string]interface{}{
		"documentId": docID,
		"color":      "not-a-color",
	}))
	if bgResp.Error == nil {
		t.Fatal("expected error for invalid color")
	}
}

// ─── Integration test: create → open → inspect → save ────────────────────────

func TestIntegration_CreateOpenInspectSave(t *testing.T) {
	dir := t.TempDir()
	docPath := dir + "/integration.docx"

	s := newServer()

	// 1. Create document and save to file
	createResp := s.dispatch(makeRequest("i1", "document.create", map[string]interface{}{
		"options": map[string]interface{}{
			"title":   "Integration Test",
			"author":  "Tester",
			"pageSize": "A4",
		},
		"content": []interface{}{
			map[string]interface{}{
				"type":      "paragraph",
				"style":     "Heading1",
				"alignment": "center",
				"runs": []interface{}{
					map[string]interface{}{"text": "Integration Heading"},
				},
			},
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "Bold text", "bold": true},
					map[string]interface{}{"text": " and ", "bold": false},
					map[string]interface{}{"text": "italic text", "italic": true},
				},
			},
		},
		"output":   "file",
		"filePath": docPath,
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}

	// 2. Open the saved file
	openResp := s.dispatch(makeRequest("i2", "document.open", map[string]interface{}{
		"filePath": docPath,
	}))
	if openResp.Error != nil {
		t.Fatalf("open failed: %+v", openResp.Error)
	}
	openResult := openResp.Result.(map[string]interface{})
	docID := openResult["documentId"].(string)

	// 3. Inspect
	inspectResp := s.dispatch(makeRequest("i3", "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspectResp.Error != nil {
		t.Fatalf("inspect failed: %+v", inspectResp.Error)
	}
	iResult := inspectResp.Result.(map[string]interface{})
	count, ok1 := iResult["paragraphCount"].(int)
	count64, ok2 := iResult["paragraphCount"].(float64)
	if ok2 {
		count = int(count64)
		ok1 = true
	}
	if !ok1 || count < 2 {
		t.Errorf("expected at least 2 paragraphs, got %v", iResult["paragraphCount"])
	}

	// 4. Save as buffer
	saveResp := s.dispatch(makeRequest("i4", "document.save", map[string]interface{}{
		"documentId": docID,
		"output":     "buffer",
	}))
	if saveResp.Error != nil {
		t.Fatalf("save failed: %+v", saveResp.Error)
	}
	saveResult := saveResp.Result.(map[string]interface{})
	if _, ok := saveResult["data"]; !ok {
		t.Error("expected data in save result")
	}
}

// ─── runExec tests ────────────────────────────────────────────────────────────

func TestRunExec_WithRequestFlag(t *testing.T) {
	// Capture stdout by redirecting
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	req := `{"id":1,"method":"document.create","params":{"content":[{"type":"paragraph","runs":[{"text":"exec test"}]}],"output":"buffer"}}`
	code := runExec(req)

	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse response: %v\noutput: %s", err, output)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %+v", resp.Error)
	}
}

func TestRunExec_InvalidJSON(t *testing.T) {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runExec("not valid json")

	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v\noutput: %s", err, output)
	}
	if resp.Error == nil {
		t.Error("expected error in response")
	}
}

// ─── Open via base64 ──────────────────────────────────────────────────────────

func TestHandleOpen_Base64(t *testing.T) {
	s := newServer()

	// First create a document as a buffer
	createResp := s.dispatch(makeRequest("b1", "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "base64 open test"},
				},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	b64 := createResp.Result.(map[string]interface{})["data"].(string)

	// Open it via base64
	openResp := s.dispatch(makeRequest("b2", "document.open", map[string]interface{}{
		"base64": b64,
	}))
	if openResp.Error != nil {
		t.Fatalf("open via base64 failed: %+v", openResp.Error)
	}
	if _, ok := openResp.Result.(map[string]interface{})["documentId"]; !ok {
		t.Error("expected documentId in open result")
	}
}

// ─── document.close tests ────────────────────────────────────────────────────

func TestHandleClose(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest("c1", "document.create", map[string]interface{}{
		"content": []interface{}{},
		"output":  "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	// Close the document
	closeResp := s.dispatch(makeRequest("c2", "document.close", map[string]interface{}{
		"documentId": docID,
	}))
	if closeResp.Error != nil {
		t.Fatalf("close failed: %+v", closeResp.Error)
	}

	// Trying to inspect the closed document should fail
	inspectResp := s.dispatch(makeRequest("c3", "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspectResp.Error == nil {
		t.Fatal("expected error when inspecting closed document")
	}
	if inspectResp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", inspectResp.Error.Code)
	}
}

func TestHandleClose_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest("c4", "document.close", map[string]interface{}{
		"documentId": "nonexistent",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for unknown documentId")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

// ─── Strict output validation tests ──────────────────────────────────────────

func TestHandleCreate_InvalidOutput(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{
					map[string]interface{}{"text": "test"},
				},
			},
		},
		"output": "invalid_output",
	}
	resp := s.dispatch(makeRequest("o1", "document.create", params))
	if resp.Error == nil {
		t.Fatal("expected error for invalid output format")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

// ─── Unknown theme tests ─────────────────────────────────────────────────────

func TestHandleCreate_UnknownTheme(t *testing.T) {
	s := newServer()
	params := map[string]interface{}{
		"options": map[string]interface{}{
			"theme": "NonExistentTheme",
		},
		"content": []interface{}{},
		"output":  "buffer",
	}
	resp := s.dispatch(makeRequest("t1", "document.create", params))
	if resp.Error == nil {
		t.Fatal("expected error for unknown theme")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

// ─── Nil params tests ────────────────────────────────────────────────────────

func TestHandleCreate_NilParams(t *testing.T) {
	s := newServer()
	// Simulate a request with no params at all
	req := &Request{ID: "n1", Method: "document.create"}
	resp := s.dispatch(req)
	// Should succeed with empty document (no content, default buffer output)
	if resp.Error != nil {
		t.Fatalf("unexpected error for nil params: %+v", resp.Error)
	}
}

// ─── document.addContent tests ───────────────────────────────────────────────

func TestHandleAddContent(t *testing.T) {
	s := newServer()

	// Create and store a document
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "First"}},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	// Now add more content
	resp := s.dispatch(makeRequest(2, "document.addContent", map[string]interface{}{
		"documentId": docID,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Second"}},
			},
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Third"}},
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("addContent failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	// Verify via inspect
	inspResp := s.dispatch(makeRequest(3, "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspResp.Error != nil {
		t.Fatalf("inspect failed: %+v", inspResp.Error)
	}
	inspResult := inspResp.Result.(map[string]interface{})
	count := inspResult["paragraphCount"]
	// We should have at least 3 paragraphs
	if c, ok := count.(int); ok && c < 3 {
		t.Errorf("expected at least 3 paragraphs, got %d", c)
	}
}

func TestHandleAddContent_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "document.addContent", map[string]interface{}{
		"documentId": "nonexistent",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "x"}},
			},
		},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleAddContent_EmptyContent(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.addContent", map[string]interface{}{
		"documentId": docID,
		"content":    []interface{}{},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for empty content")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

// ─── document.addPageBreak tests ─────────────────────────────────────────────

func TestHandleAddPageBreak(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.addPageBreak", map[string]interface{}{
		"documentId": docID,
	}))
	if resp.Error != nil {
		t.Fatalf("addPageBreak failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
}

func TestHandleAddPageBreak_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "document.addPageBreak", map[string]interface{}{
		"documentId": "nonexistent",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

// ─── paragraph.add tests ────────────────────────────────────────────────────

func TestHandleParagraphAdd(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "paragraph.add", map[string]interface{}{
		"documentId": docID,
		"style":      "Heading1",
		"alignment":  "center",
		"runs": []interface{}{
			map[string]interface{}{"text": "My Heading", "bold": true},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("paragraph.add failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if _, ok := result["index"]; !ok {
		t.Error("expected index in result")
	}
}

func TestHandleParagraphAdd_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "paragraph.add", map[string]interface{}{
		"documentId": "nonexistent",
		"runs":       []interface{}{map[string]interface{}{"text": "x"}},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleParagraphAdd_WithFormatting(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "paragraph.add", map[string]interface{}{
		"documentId":    docID,
		"spacingBefore": 240,
		"spacingAfter":  120,
		"runs": []interface{}{
			map[string]interface{}{
				"text":      "Formatted",
				"bold":      true,
				"italic":    true,
				"underline": "single",
				"color":     "#FF0000",
				"fontSize":  14,
				"font":      "Arial",
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("paragraph.add with formatting failed: %+v", resp.Error)
	}
}

// ─── paragraph.list tests ───────────────────────────────────────────────────

func TestHandleParagraphList(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Alpha"}},
			},
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Beta"}},
			},
		},
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "paragraph.list", map[string]interface{}{
		"documentId": docID,
	}))
	if resp.Error != nil {
		t.Fatalf("paragraph.list failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	count, _ := result["count"].(int)
	if count < 2 {
		t.Errorf("expected at least 2 paragraphs, got %d", count)
	}
	paragraphs, ok := result["paragraphs"].([]map[string]interface{})
	if !ok {
		// try type assertion for []interface{}
		pList, ok2 := result["paragraphs"].([]interface{})
		if !ok2 {
			t.Fatalf("expected paragraphs array, got %T", result["paragraphs"])
		}
		if len(pList) < 2 {
			t.Errorf("expected at least 2 paragraphs, got %d", len(pList))
		}
		// Check first paragraph has text
		first, _ := pList[0].(map[string]interface{})
		if first["text"] != "Alpha" {
			t.Errorf("expected first paragraph text 'Alpha', got %v", first["text"])
		}
	} else {
		if paragraphs[0]["text"] != "Alpha" {
			t.Errorf("expected first paragraph text 'Alpha', got %v", paragraphs[0]["text"])
		}
	}
}

func TestHandleParagraphList_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "paragraph.list", map[string]interface{}{
		"documentId": "nonexistent",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

// ─── table.add tests ────────────────────────────────────────────────────────

func TestHandleTableAdd(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "table.add", map[string]interface{}{
		"documentId": docID,
		"rows": []interface{}{
			map[string]interface{}{
				"cells": []interface{}{
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "A1"}},
							},
						},
					},
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "B1"}},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"cells": []interface{}{
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "A2"}},
							},
						},
					},
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "B2"}},
							},
						},
					},
				},
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("table.add failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if _, ok := result["index"]; !ok {
		t.Error("expected index in result")
	}
}

func TestHandleTableAdd_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "table.add", map[string]interface{}{
		"documentId": "nonexistent",
		"rows": []interface{}{
			map[string]interface{}{
				"cells": []interface{}{
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "x"}},
							},
						},
					},
				},
			},
		},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

// ─── table.list tests ───────────────────────────────────────────────────────

func TestHandleTableList(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "table",
				"rows": []interface{}{
					map[string]interface{}{
						"cells": []interface{}{
							map[string]interface{}{
								"paragraphs": []interface{}{
									map[string]interface{}{
										"runs": []interface{}{map[string]interface{}{"text": "Cell"}},
									},
								},
							},
						},
					},
				},
			},
		},
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "table.list", map[string]interface{}{
		"documentId": docID,
	}))
	if resp.Error != nil {
		t.Fatalf("table.list failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	count, _ := result["count"].(int)
	if count < 1 {
		t.Errorf("expected at least 1 table, got %d", count)
	}
}

func TestHandleTableList_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "table.list", map[string]interface{}{
		"documentId": "nonexistent",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

// ─── section.add tests ──────────────────────────────────────────────────────

func TestHandleSectionAdd(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "section.add", map[string]interface{}{
		"documentId":  docID,
		"breakType":   "nextPage",
		"pageSize":    "A4",
		"orientation": "landscape",
		"columns":     2,
	}))
	if resp.Error != nil {
		t.Fatalf("section.add failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if _, ok := result["index"]; !ok {
		t.Error("expected index in result")
	}
}

func TestHandleSectionAdd_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "section.add", map[string]interface{}{
		"documentId": "nonexistent",
		"breakType":  "nextPage",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

// ─── Integration: open → add content → save ─────────────────────────────────

func TestIntegration_OpenAddContentSave(t *testing.T) {
	s := newServer()

	// Create initial document to file
	tmpDir := t.TempDir()
	filePath := tmpDir + "/test_add_content.docx"

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Original"}},
			},
		},
		"output":   "file",
		"filePath": filePath,
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}

	// Re-open the file
	openResp := s.dispatch(makeRequest(2, "document.open", map[string]interface{}{
		"filePath": filePath,
	}))
	if openResp.Error != nil {
		t.Fatalf("open failed: %+v", openResp.Error)
	}
	docID := openResp.Result.(map[string]interface{})["documentId"].(string)

	// Add content to the opened document
	addResp := s.dispatch(makeRequest(3, "document.addContent", map[string]interface{}{
		"documentId": docID,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"runs": []interface{}{map[string]interface{}{"text": "Appended via RPC"}},
			},
		},
	}))
	if addResp.Error != nil {
		t.Fatalf("addContent failed: %+v", addResp.Error)
	}

	// Add a paragraph directly
	paraResp := s.dispatch(makeRequest(4, "paragraph.add", map[string]interface{}{
		"documentId": docID,
		"runs":       []interface{}{map[string]interface{}{"text": "Direct paragraph"}},
	}))
	if paraResp.Error != nil {
		t.Fatalf("paragraph.add failed: %+v", paraResp.Error)
	}

	// Add a table
	tableResp := s.dispatch(makeRequest(5, "table.add", map[string]interface{}{
		"documentId": docID,
		"rows": []interface{}{
			map[string]interface{}{
				"cells": []interface{}{
					map[string]interface{}{
						"paragraphs": []interface{}{
							map[string]interface{}{
								"runs": []interface{}{map[string]interface{}{"text": "Cell"}},
							},
						},
					},
				},
			},
		},
	}))
	if tableResp.Error != nil {
		t.Fatalf("table.add failed: %+v", tableResp.Error)
	}

	// List paragraphs
	listResp := s.dispatch(makeRequest(6, "paragraph.list", map[string]interface{}{
		"documentId": docID,
	}))
	if listResp.Error != nil {
		t.Fatalf("paragraph.list failed: %+v", listResp.Error)
	}

	// List tables
	tListResp := s.dispatch(makeRequest(7, "table.list", map[string]interface{}{
		"documentId": docID,
	}))
	if tListResp.Error != nil {
		t.Fatalf("table.list failed: %+v", tListResp.Error)
	}

	// Save to a new file
	outPath := tmpDir + "/test_modified.docx"
	saveResp := s.dispatch(makeRequest(8, "document.save", map[string]interface{}{
		"documentId": docID,
		"output":     "file",
		"filePath":   outPath,
	}))
	if saveResp.Error != nil {
		t.Fatalf("save failed: %+v", saveResp.Error)
	}

	// Verify file exists and is non-empty
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() < 100 {
		t.Errorf("output file too small: %d bytes", info.Size())
	}

	// Close
	closeResp := s.dispatch(makeRequest(9, "document.close", map[string]interface{}{
		"documentId": docID,
	}))
	if closeResp.Error != nil {
		t.Fatalf("close failed: %+v", closeResp.Error)
	}
}

// ─── System method tests ──────────────────────────────────────────────────────

func TestHandleSystemPing(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "system.ping", nil))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", result["status"])
	}
}

func TestHandleSystemVersion(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "system.version", nil))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["name"] != "docxgo" {
		t.Errorf("expected name=docxgo, got %v", result["name"])
	}
	if result["protocolVersion"] != ProtocolVersion {
		t.Errorf("expected protocolVersion=%s, got %v", ProtocolVersion, result["protocolVersion"])
	}
	if result["goVersion"] == nil || result["goVersion"] == "" {
		t.Error("goVersion should be non-empty")
	}
	if result["platform"] == nil || result["platform"] == "" {
		t.Error("platform should be non-empty")
	}
	if result["arch"] == nil || result["arch"] == "" {
		t.Error("arch should be non-empty")
	}
}

func TestHandleSystemCapabilities(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "system.capabilities", nil))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]bool)
	if !result["rpc"] {
		t.Error("expected rpc capability to be true")
	}
	if !result["template"] {
		t.Error("expected template capability to be true")
	}
	if !result["mailMerge"] {
		t.Error("expected mailMerge capability to be true")
	}
	if !result["batch"] {
		t.Error("expected batch capability to be true")
	}
}

// ─── System batch tests ──────────────────────────────────────────────────────

func TestHandleSystemBatch_Basic(t *testing.T) {
	s := newServer()

	resp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{
			{"method": "system.ping"},
			{"method": "system.version"},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	responses := result["responses"].([]interface{})
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	// First should be system.ping result
	first := responses[0].(map[string]interface{})
	if first["result"] == nil {
		t.Error("first response should have result")
	}

	// Second should be system.version result
	second := responses[1].(map[string]interface{})
	if second["result"] == nil {
		t.Error("second response should have result")
	}
}

func TestHandleSystemBatch_WithError(t *testing.T) {
	s := newServer()

	resp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{
			{"method": "system.ping"},
			{"method": "document.inspect", "params": map[string]interface{}{"documentId": "nonexistent"}},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("batch itself should not error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	responses := result["responses"].([]interface{})
	second := responses[1].(map[string]interface{})
	if second["error"] == nil {
		t.Error("second response should have an error (doc not found)")
	}
}

func TestHandleSystemBatch_RejectsNested(t *testing.T) {
	s := newServer()

	resp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{
			{"method": "system.batch", "params": map[string]interface{}{
				"requests": []map[string]interface{}{{"method": "system.ping"}},
			}},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("batch itself should not error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	responses := result["responses"].([]interface{})
	first := responses[0].(map[string]interface{})
	if first["error"] == nil {
		t.Error("nested batch should be rejected with error")
	}
}

func TestHandleSystemBatch_EmptyRequests(t *testing.T) {
	s := newServer()

	resp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for empty requests")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

func TestHandleSystemBatch_Workflow(t *testing.T) {
	s := newServer()

	// Batch: create → add content → save to buffer in one batch
	resp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"method": "document.create",
				"params": map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "paragraph",
							"runs": []map[string]interface{}{
								{"text": "Batch created document"},
							},
						},
					},
					"output": "buffer",
				},
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("batch failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	responses := result["responses"].([]interface{})
	first := responses[0].(map[string]interface{})
	if first["error"] != nil {
		t.Fatalf("create in batch failed: %+v", first["error"])
	}
	createResult := first["result"].(map[string]interface{})
	if createResult["data"] == nil || createResult["data"] == "" {
		t.Error("expected base64 data from batch create")
	}
}

// ─── Template tests ──────────────────────────────────────────────────────────

func TestHandleTemplateInspect(t *testing.T) {
	s := newServer()

	// Create a document with placeholders
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"runs": []map[string]interface{}{
					{"text": "Hello {{name}}, your order {{orderId}} is ready."},
				},
			},
			{
				"type": "paragraph",
				"runs": []map[string]interface{}{
					{"text": "Total: {{total}}"},
				},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	createResult := createResp.Result.(map[string]interface{})
	docID := createResult["documentId"].(string)

	// Inspect template
	inspResp := s.dispatch(makeRequest(2, "template.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspResp.Error != nil {
		t.Fatalf("template.inspect failed: %+v", inspResp.Error)
	}
	result := inspResp.Result.(map[string]interface{})

	// The result types depend on how Go serializes — use JSON round-trip for reliable access
	resultJSON, _ := json.Marshal(result)
	var parsed struct {
		Placeholders []string                 `json:"placeholders"`
		Count        int                      `json:"count"`
		Occurrences  int                      `json:"occurrences"`
		Details      []map[string]interface{} `json:"details"`
	}
	json.Unmarshal(resultJSON, &parsed)

	if parsed.Count != 3 {
		t.Errorf("expected 3 unique placeholders, got %d", parsed.Count)
	}
	if parsed.Occurrences != 3 {
		t.Errorf("expected 3 occurrences, got %d", parsed.Occurrences)
	}

	names := make(map[string]bool)
	for _, p := range parsed.Placeholders {
		names[p] = true
	}
	for _, expected := range []string{"name", "orderId", "total"} {
		if !names[expected] {
			t.Errorf("expected placeholder %q not found", expected)
		}
	}
}

func TestHandleTemplateInspect_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "template.inspect", map[string]interface{}{
		"documentId": "nonexistent",
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleTemplateRender(t *testing.T) {
	s := newServer()

	// Create a document with a placeholder
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"runs": []map[string]interface{}{
					{"text": "Hello {{name}}!"},
				},
			},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	// Render template
	renderResp := s.dispatch(makeRequest(2, "template.render", map[string]interface{}{
		"documentId": docID,
		"data": map[string]interface{}{
			"name": "World",
		},
	}))
	if renderResp.Error != nil {
		t.Fatalf("template.render failed: %+v", renderResp.Error)
	}
	result := renderResp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}

	// Verify the document was modified by inspecting it
	inspResp := s.dispatch(makeRequest(3, "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspResp.Error != nil {
		t.Fatalf("inspect failed: %+v", inspResp.Error)
	}
	inspResult := inspResp.Result.(map[string]interface{})
	texts := inspResult["text"].([]string)
	found := false
	for _, txt := range texts {
		if strings.Contains(txt, "Hello World!") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected merged text 'Hello World!' in document, got texts: %v", texts)
	}
}

func TestHandleTemplateRender_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "template.render", map[string]interface{}{
		"documentId": "nonexistent",
		"data":       map[string]interface{}{"key": "val"},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleTemplateRender_EmptyData(t *testing.T) {
	s := newServer()

	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "paragraph", "runs": []map[string]interface{}{{"text": "No placeholders here"}}},
		},
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "template.render", map[string]interface{}{
		"documentId": docID,
		"data":       map[string]interface{}{},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for empty data")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

func TestHandleTemplateRender_StrictMode(t *testing.T) {
	s := newServer()

	// Create doc with placeholder that won't be satisfied
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "paragraph", "runs": []map[string]interface{}{{"text": "{{missing}}"}}},
		},
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "template.render", map[string]interface{}{
		"documentId": docID,
		"data":       map[string]interface{}{"other": "val"},
		"strictMode": true,
	}))
	if resp.Error == nil {
		t.Fatal("expected error in strict mode with missing key")
	}
	if resp.Error.Code != "TEMPLATE_ERROR" {
		t.Errorf("expected TEMPLATE_ERROR, got %s", resp.Error.Code)
	}
}

func TestHandleTemplateRender_WithWarnings(t *testing.T) {
	s := newServer()

	// Create doc with one placeholder, provide extra data key → should produce a warning
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "paragraph", "runs": []map[string]interface{}{{"text": "Hello {{name}}!"}}},
		},
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "template.render", map[string]interface{}{
		"documentId": docID,
		"data":       map[string]interface{}{"name": "Alice", "unused": "val"},
	}))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	// Should have a warning about the unused key
	if result["warnings"] != nil {
		warnings := result["warnings"].([]map[string]interface{})
		if len(warnings) == 0 {
			t.Error("expected at least one warning about unused key")
		}
	}
}

// ─── ApplyPatch tests ────────────────────────────────────────────────────────

func TestHandleApplyPatch_Basic(t *testing.T) {
	s := newServer()

	// Create empty doc
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"output": "buffer",
	}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	// Apply patch with multiple operations
	patchResp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{
			{
				"op":   "appendParagraph",
				"runs": []map[string]interface{}{{"text": "First paragraph", "bold": true}},
			},
			{
				"op":   "appendParagraph",
				"runs": []map[string]interface{}{{"text": "Second paragraph"}},
				"style": "Heading1",
			},
			{
				"op": "appendPageBreak",
			},
			{
				"op": "appendTable",
				"rows": []map[string]interface{}{
					{
						"cells": []map[string]interface{}{
							{"paragraphs": []map[string]interface{}{{"runs": []map[string]interface{}{{"text": "A1"}}}}},
							{"paragraphs": []map[string]interface{}{{"runs": []map[string]interface{}{{"text": "B1"}}}}},
						},
					},
				},
			},
		},
	}))
	if patchResp.Error != nil {
		t.Fatalf("applyPatch failed: %+v", patchResp.Error)
	}
	result := patchResp.Result.(map[string]interface{})
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	applied := result["applied"]
	if applied != 4 {
		t.Errorf("expected 4 applied operations, got %v", applied)
	}

	// Verify by inspecting
	inspResp := s.dispatch(makeRequest(3, "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	inspResult := inspResp.Result.(map[string]interface{})
	pCount := inspResult["paragraphCount"].(int)
	tCount := inspResult["tableCount"].(int)
	// At least 2 paragraphs + the generated ones from the page break and table
	if pCount < 2 {
		t.Errorf("expected at least 2 paragraphs, got %d", pCount)
	}
	if tCount < 1 {
		t.Errorf("expected at least 1 table, got %d", tCount)
	}
}

func TestHandleApplyPatch_NotFound(t *testing.T) {
	s := newServer()
	resp := s.dispatch(makeRequest(1, "document.applyPatch", map[string]interface{}{
		"documentId": "nonexistent",
		"operations": []map[string]interface{}{
			{"op": "appendParagraph", "runs": []map[string]interface{}{{"text": "test"}}},
		},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent document")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandleApplyPatch_EmptyOperations(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{"output": "buffer"}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for empty operations")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
}

func TestHandleApplyPatch_UnknownOp(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{"output": "buffer"}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{
			{"op": "unknownOp"},
		},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for unknown op")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", resp.Error.Code)
	}
	// Check that error has data with index
	if resp.Error.Data == nil {
		t.Error("expected error data with index")
	}
}

func TestHandleApplyPatch_SetMetadata(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{"output": "buffer"}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{
			{
				"op":    "setMetadata",
				"title": "Patched Title",
			},
			{
				"op":    "setBackgroundColor",
				"color": "#FF0000",
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("applyPatch failed: %+v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	if result["applied"] != 2 {
		t.Errorf("expected 2 applied, got %v", result["applied"])
	}
}

func TestHandleApplyPatch_AppendSection(t *testing.T) {
	s := newServer()
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{"output": "buffer"}))
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	resp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{
			{
				"op":        "appendSection",
				"breakType": "nextPage",
			},
		},
	}))
	if resp.Error != nil {
		t.Fatalf("applyPatch with section failed: %+v", resp.Error)
	}
}

// ─── RPCError Data field tests ────────────────────────────────────────────────

func TestErrorResponseWithData(t *testing.T) {
	r := errorResponseWithData(1, "VALIDATION_ERROR", "bad op", "document.applyPatch",
		map[string]interface{}{"index": 3, "retryable": false})
	if r.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if r.Error.Data == nil {
		t.Fatal("expected non-nil data")
	}
	if r.Error.Data["index"] != 3 {
		t.Errorf("expected index=3, got %v", r.Error.Data["index"])
	}
	if r.Error.Data["retryable"] != false {
		t.Errorf("expected retryable=false, got %v", r.Error.Data["retryable"])
	}

	// Verify it serializes correctly
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !bytes.Contains(data, []byte(`"data"`)) {
		t.Error("expected data field in JSON")
	}
}

// ─── Integration: template workflow ──────────────────────────────────────────

func TestIntegration_TemplateWorkflow(t *testing.T) {
	s := newServer()

	// 1. Create doc with template placeholders
	createResp := s.dispatch(makeRequest(1, "document.create", map[string]interface{}{
		"options": map[string]interface{}{"title": "Invoice Template"},
		"content": []map[string]interface{}{
			{"type": "paragraph", "runs": []map[string]interface{}{{"text": "Invoice for {{customer}}"}}},
			{"type": "paragraph", "runs": []map[string]interface{}{{"text": "Amount: {{amount}}"}}},
		},
		"output": "buffer",
	}))
	if createResp.Error != nil {
		t.Fatalf("create failed: %+v", createResp.Error)
	}
	docID := createResp.Result.(map[string]interface{})["documentId"].(string)

	// 2. Inspect template
	inspResp := s.dispatch(makeRequest(2, "template.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	if inspResp.Error != nil {
		t.Fatalf("inspect failed: %+v", inspResp.Error)
	}
	inspResult := inspResp.Result.(map[string]interface{})
	if inspResult["count"].(int) != 2 {
		t.Errorf("expected 2 placeholders, got %d", inspResult["count"])
	}

	// 3. Render template
	renderResp := s.dispatch(makeRequest(3, "template.render", map[string]interface{}{
		"documentId": docID,
		"data":       map[string]interface{}{"customer": "Acme Corp", "amount": "$1,234"},
	}))
	if renderResp.Error != nil {
		t.Fatalf("render failed: %+v", renderResp.Error)
	}

	// 4. Verify merged content
	docInspResp := s.dispatch(makeRequest(4, "document.inspect", map[string]interface{}{
		"documentId": docID,
	}))
	texts := docInspResp.Result.(map[string]interface{})["text"].([]string)
	foundCustomer := false
	foundAmount := false
	for _, ts := range texts {
		if strings.Contains(ts, "Acme Corp") {
			foundCustomer = true
		}
		if strings.Contains(ts, "$1,234") {
			foundAmount = true
		}
	}
	if !foundCustomer {
		t.Error("expected merged customer name in document")
	}
	if !foundAmount {
		t.Error("expected merged amount in document")
	}

	// 5. Save to file
	tmpDir := t.TempDir()
	saveResp := s.dispatch(makeRequest(5, "document.save", map[string]interface{}{
		"documentId": docID,
		"output":     "file",
		"filePath":   tmpDir + "/invoice.docx",
	}))
	if saveResp.Error != nil {
		t.Fatalf("save failed: %+v", saveResp.Error)
	}

	info, err := os.Stat(tmpDir + "/invoice.docx")
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() < 100 {
		t.Errorf("output file too small: %d bytes", info.Size())
	}

	// 6. Close
	s.dispatch(makeRequest(6, "document.close", map[string]interface{}{"documentId": docID}))
}

// ─── Integration: applyPatch + batch workflow ────────────────────────────────

func TestIntegration_ApplyPatchBatch(t *testing.T) {
	s := newServer()

	// Use batch to create, patch, and save
	batchResp := s.dispatch(makeRequest(1, "system.batch", map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"method": "document.create",
				"params": map[string]interface{}{"output": "buffer"},
			},
		},
	}))
	if batchResp.Error != nil {
		t.Fatalf("batch create failed: %+v", batchResp.Error)
	}
	responses := batchResp.Result.(map[string]interface{})["responses"].([]interface{})
	createResult := responses[0].(map[string]interface{})["result"].(map[string]interface{})
	docID := createResult["documentId"].(string)

	// Apply patch
	patchResp := s.dispatch(makeRequest(2, "document.applyPatch", map[string]interface{}{
		"documentId": docID,
		"operations": []map[string]interface{}{
			{"op": "appendParagraph", "runs": []map[string]interface{}{{"text": "Hello from patch"}}, "style": "Heading1"},
			{"op": "appendParagraph", "runs": []map[string]interface{}{{"text": "Body text"}}},
			{"op": "appendTable", "rows": []map[string]interface{}{
				{"cells": []map[string]interface{}{
					{"paragraphs": []map[string]interface{}{{"runs": []map[string]interface{}{{"text": "Col1"}}}}},
					{"paragraphs": []map[string]interface{}{{"runs": []map[string]interface{}{{"text": "Col2"}}}}},
				}},
			}},
			{"op": "setMetadata", "title": "Patched Doc"},
		},
	}))
	if patchResp.Error != nil {
		t.Fatalf("applyPatch failed: %+v", patchResp.Error)
	}
	if patchResp.Result.(map[string]interface{})["applied"] != 4 {
		t.Errorf("expected 4 applied, got %v", patchResp.Result.(map[string]interface{})["applied"])
	}

	// Save
	tmpDir := t.TempDir()
	saveResp := s.dispatch(makeRequest(3, "document.save", map[string]interface{}{
		"documentId": docID,
		"output":     "file",
		"filePath":   tmpDir + "/patched.docx",
	}))
	if saveResp.Error != nil {
		t.Fatalf("save failed: %+v", saveResp.Error)
	}

	// Close
	s.dispatch(makeRequest(4, "document.close", map[string]interface{}{"documentId": docID}))
}
