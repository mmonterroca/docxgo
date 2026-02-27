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
