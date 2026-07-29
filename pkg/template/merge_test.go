// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"bytes"
	"strings"
	"testing"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
)

func TestMergeTemplate_SinglePlaceholder(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello {{name}}!")

	err := MergeTemplate(doc, MergeData{"name": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "Hello John!" {
		t.Errorf("expected 'Hello John!', got %q", text)
	}
}

func TestMergeTemplate_MultiplePlaceholders(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{first}} {{last}}")

	err := MergeTemplate(doc, MergeData{"first": "John", "last": "Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", text)
	}
}

func TestMergeTemplate_RepeatedPlaceholder(t *testing.T) {
	doc := core.NewDocument()

	for i := 0; i < 3; i++ {
		p, _ := doc.AddParagraph()
		r, _ := p.AddRun()
		r.SetText("Name: {{name}}")
	}

	err := MergeTemplate(doc, MergeData{"name": "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, p := range doc.Paragraphs() {
		text := p.Runs()[0].Text()
		if text != "Name: Alice" {
			t.Errorf("paragraph %d: expected 'Name: Alice', got %q", i, text)
		}
	}
}

func TestMergeTemplate_PreservesFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{name}}")
	r.SetBold(true)
	r.SetItalic(true)
	r.SetFont(domain.Font{Name: "Arial"})
	r.SetSize(24)

	err := MergeTemplate(doc, MergeData{"name": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	run := para.Runs()[0]
	if run.Text() != "John" {
		t.Errorf("expected 'John', got %q", run.Text())
	}
	if !run.Bold() {
		t.Error("expected bold to be preserved")
	}
	if !run.Italic() {
		t.Error("expected italic to be preserved")
	}
	if run.Font().Name != "Arial" {
		t.Errorf("expected font 'Arial', got %q", run.Font().Name)
	}
	if run.Size() != 24 {
		t.Errorf("expected size 24, got %d", run.Size())
	}
}

func TestMergeTemplate_InTableCell(t *testing.T) {
	doc := core.NewDocument()
	table, _ := doc.AddTable(1, 1)
	cell, _ := table.Rows()[0].Cell(0)
	para, _ := cell.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{value}}")

	err := MergeTemplate(doc, MergeData{"value": "CellData"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "CellData" {
		t.Errorf("expected 'CellData', got %q", text)
	}
}

func TestMergeTemplate_MissingKey_Lenient(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello {{name}}, you are {{role}}.")

	err := MergeTemplate(doc, MergeData{"name": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "Hello John, you are {{role}}." {
		t.Errorf("expected unreplaced placeholder, got %q", text)
	}
}

func TestMergeTemplate_MissingKey_Strict(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello {{name}}, you are {{role}}.")

	opts := MergeOptions{
		OpenDelimiter:  "{{",
		CloseDelimiter: "}}",
		StrictMode:     true,
	}
	err := MergeTemplate(doc, MergeData{"name": "John"}, opts)
	if err == nil {
		t.Fatal("expected error for missing key in strict mode")
	}
	if !strings.Contains(err.Error(), "role") {
		t.Errorf("expected error to mention 'role', got: %v", err)
	}
}

// TestMergeTemplate_SkipsPreservedHeaderAndFooter is an end-to-end regression
// test for the case where MergeTemplate silently replaced placeholder text
// in a header/footer whose bytes WriteTo then writes verbatim, discarding
// the edit on save. On a document whose headers/footers were preserved from
// a round-trip open, a header/footer placeholder must be left untouched (so
// re-running the merge later still finds it) and, in strict mode, reported
// as missing rather than silently dropped.
func TestMergeTemplate_SkipsPreservedHeaderAndFooter(t *testing.T) {
	doc := core.NewDocument()
	bodyPara, _ := doc.AddParagraph()
	bodyRun, _ := bodyPara.AddRun()
	bodyRun.SetText("Body: {{name}}")

	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	headerPara, _ := header.AddParagraph()
	headerRun, _ := headerPara.AddRun()
	headerRun.SetText("Header: {{name}}")

	templatePath := t.TempDir() + "/preserved_header.docx"
	if err := doc.SaveAs(templatePath); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	opened, err := docx.OpenDocument(templatePath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	opts := MergeOptions{OpenDelimiter: "{{", CloseDelimiter: "}}", StrictMode: true}
	err = MergeTemplate(opened, MergeData{"name": "Alice"}, opts)
	if err == nil {
		t.Fatal("expected strict-mode error for the header placeholder that cannot be persisted")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention 'name', got: %v", err)
	}

	var buf bytes.Buffer
	if _, err := opened.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	savedBytes := buf.Bytes()

	docXML := string(extractDocumentXML(t, savedBytes))
	if !strings.Contains(docXML, "Body: Alice") {
		t.Errorf("document.xml: expected merged body text, got %s", docXML)
	}

	headerXML := string(extractZipPart(t, savedBytes, "word/header1.xml"))
	if !strings.Contains(headerXML, "Header: {{name}}") {
		t.Errorf("header1.xml: expected untouched placeholder, got %s", headerXML)
	}
	if strings.Contains(headerXML, "Alice") {
		t.Errorf("header1.xml: unexpected merged value — header should be preserved verbatim, got %s", headerXML)
	}
}

func TestMergeTemplate_EmptyValue(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Name: {{name}}")

	err := MergeTemplate(doc, MergeData{"name": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "Name: " {
		t.Errorf("expected 'Name: ', got %q", text)
	}
}

func TestMergeTemplate_SpecialChars(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{content}}")

	err := MergeTemplate(doc, MergeData{"content": "A < B & C > D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "A < B & C > D" {
		t.Errorf("expected 'A < B & C > D', got %q", text)
	}
}

func TestMergeTemplate_CustomDelimiters(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello ${name}!")

	opts := MergeOptions{
		OpenDelimiter:  "${",
		CloseDelimiter: "}",
	}
	err := MergeTemplate(doc, MergeData{"name": "World"}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", text)
	}
}

func TestMergeTemplate_NoPlaceholders(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Just plain text.")

	err := MergeTemplate(doc, MergeData{"name": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := para.Runs()[0].Text()
	if text != "Just plain text." {
		t.Errorf("expected unchanged text, got %q", text)
	}
}

func TestMergeTemplate_SplitPlaceholder(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()

	// Simulate Word splitting "{{name}}" across runs
	r1, _ := para.AddRun()
	r1.SetText("Dear {{")
	r2, _ := para.AddRun()
	r2.SetText("name")
	r3, _ := para.AddRun()
	r3.SetText("}}, welcome!")

	err := MergeTemplate(doc, MergeData{"name": "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After consolidation + replacement, should be a single run
	runs := para.Runs()
	var fullText string
	for _, run := range runs {
		fullText += run.Text()
	}
	if fullText != "Dear Alice, welcome!" {
		t.Errorf("expected 'Dear Alice, welcome!', got %q", fullText)
	}
}

func TestMergeTemplate_EmptyDocument(t *testing.T) {
	doc := core.NewDocument()

	err := MergeTemplate(doc, MergeData{"name": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMergeTemplate_NilData(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("Hello {{name}}!")

	err := MergeTemplate(doc, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Placeholder should remain unreplaced
	text := para.Runs()[0].Text()
	if text != "Hello {{name}}!" {
		t.Errorf("expected unchanged text with nil data, got %q", text)
	}
}

func TestValidateTemplate_AllKeysPresent(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{first}} {{last}}")

	errs := ValidateTemplate(doc, MergeData{"first": "John", "last": "Doe"})
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateTemplate_MissingKeys(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{first}} {{last}} {{email}}")

	errs := ValidateTemplate(doc, MergeData{"first": "John"})

	errorKeys := make(map[string]bool)
	for _, e := range errs {
		if e.Severity == SeverityError {
			errorKeys[e.Key] = true
		}
	}

	if !errorKeys["last"] {
		t.Error("expected missing key error for 'last'")
	}
	if !errorKeys["email"] {
		t.Error("expected missing key error for 'email'")
	}
}

func TestValidateTemplate_UnusedKeys(t *testing.T) {
	doc := core.NewDocument()
	para, _ := doc.AddParagraph()
	r, _ := para.AddRun()
	r.SetText("{{name}}")

	errs := ValidateTemplate(doc, MergeData{"name": "John", "age": "30", "city": "NYC"})

	warningKeys := make(map[string]bool)
	for _, e := range errs {
		if e.Severity == SeverityWarning {
			warningKeys[e.Key] = true
		}
	}

	if !warningKeys["age"] {
		t.Error("expected unused key warning for 'age'")
	}
	if !warningKeys["city"] {
		t.Error("expected unused key warning for 'city'")
	}
}
