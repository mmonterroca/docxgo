/*
MIT License

Copyright (c) 2025 Misael Monterroca <misael@monterroca.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package template

import (
	"strings"
	"testing"

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
