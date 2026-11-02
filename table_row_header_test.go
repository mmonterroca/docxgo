// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"bytes"
	"strings"
	"testing"
)

// TestTableRow_RepeatAsHeaderAndCantSplit writes w:tblHeader/w:cantSplit in
// w:trPr, only on the rows that asked for them, and reads them back.
func TestTableRow_RepeatAsHeaderAndCantSplit(t *testing.T) {
	doc := NewDocument()
	table, err := doc.AddTable(2, 2)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	header, _ := table.Row(0)
	if err := header.SetRepeatAsHeader(true); err != nil {
		t.Fatal(err)
	}
	_ = header.SetCantSplit(true)

	xml := documentXML(t, doc)
	if n := strings.Count(xml, "<w:tblHeader></w:tblHeader>"); n != 1 {
		t.Errorf("want one w:tblHeader, got %d:\n%s", n, xml)
	}
	if n := strings.Count(xml, "<w:cantSplit></w:cantSplit>"); n != 1 {
		t.Errorf("want one w:cantSplit, got %d", n)
	}
	if strings.Index(xml, "<w:cantSplit>") > strings.Index(xml, "<w:tblHeader>") {
		t.Error("want cantSplit before tblHeader")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	rows := doc2.Tables()[0].Rows()
	if !rows[0].RepeatAsHeader() || !rows[0].CantSplit() {
		t.Error("header row flags lost on reopen")
	}
	if rows[1].RepeatAsHeader() || rows[1].CantSplit() {
		t.Error("data row gained flags on reopen")
	}
}
