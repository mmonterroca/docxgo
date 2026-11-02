// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"strings"
	"testing"
)

// TestTOCField_CallerResultSurvivesDirty: a TOC field with a caller-set,
// multi-line result stays dirty (Word rebuilds it on open) and is written
// line by line with w:br, instead of the "Table of Contents" placeholder.
func TestTOCField_CallerResultSurvivesDirty(t *testing.T) {
	doc := NewDocument()
	p, _ := doc.AddParagraph()
	r, _ := p.AddRun()
	field := NewTOCField(map[string]string{"levels": "1-3"})
	field.(interface{ SetResult(string) }).SetResult("Intro\n    Detail")
	field.(interface{ MarkDirty() }).MarkDirty()
	if err := r.AddField(field); err != nil {
		t.Fatalf("AddField: %v", err)
	}

	xml := documentXML(t, doc)
	if strings.Contains(xml, "Table of Contents") {
		t.Errorf("placeholder replaced the caller's result:\n%s", xml)
	}
	if !strings.Contains(xml, `w:dirty="true"`) {
		t.Error("TOC field is no longer dirty")
	}
	intro := strings.Index(xml, "<w:t>Intro</w:t>")
	br := strings.Index(xml, "<w:br>")
	detail := strings.Index(xml, `<w:t xml:space="preserve">    Detail</w:t>`)
	if intro < 0 || br < intro || detail < br {
		t.Errorf("want Intro, w:br, indented Detail in order:\n%s", xml)
	}
}
