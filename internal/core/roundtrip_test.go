// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca

package core

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/internal/ooxmlmerge"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestEnsureRequiredDocumentNamespacesRecognizesSpacedDeclaration(t *testing.T) {
	source := []byte(`<w:document xmlns:w = "` + constants.NamespaceMain + `"><w:body/></w:document>`)
	root, err := ooxmlmerge.Parse(source, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if usesUnboundPrefix(root, "w") {
		t.Fatal("spaced namespace declaration did not bind the w prefix")
	}
	got := ensureRequiredDocumentNamespaces(source)
	if !bytes.Equal(got, source) {
		t.Fatalf("namespace normalization changed a fully bound document:\ngot:  %s\nwant: %s", got, source)
	}
}

func TestEnsureRequiredDocumentNamespacesBindsGeneratedPrefix(t *testing.T) {
	source := []byte(`<x:document xmlns:x="` + constants.NamespaceMain + `"><x:body><w:p/></x:body></x:document>`)
	got := string(ensureRequiredDocumentNamespaces(source))
	declaration := `xmlns:w="` + constants.NamespaceMain + `"`
	if strings.Count(got, declaration) != 1 {
		t.Fatalf("generated prefix declaration count = %d, want 1:\n%s", strings.Count(got, declaration), got)
	}
}

func TestEnsureRequiredDocumentNamespacesUsesStrictFamilyForStrictDocument(t *testing.T) {
	source := []byte(`<s:document xmlns:s="` + constants.NamespaceMainStrict + `"><s:body><w:p><w:r r:id="rId1"><wp:inline/></w:r></w:p></s:body></s:document>`)
	got := string(ensureRequiredDocumentNamespaces(source))
	for prefix, namespace := range map[string]string{
		"w":  constants.NamespaceMainStrict,
		"r":  constants.NamespaceRelationshipsStrict,
		"wp": constants.NamespaceWordprocessingDrawingStrict,
	} {
		declaration := `xmlns:` + prefix + `="` + namespace + `"`
		if strings.Count(got, declaration) != 1 {
			t.Errorf("Strict declaration %s count = %d, want 1:\n%s", declaration, strings.Count(got, declaration), got)
		}
	}
	if strings.Contains(got, constants.NamespaceMain+`"`) {
		t.Fatalf("Strict document received a Transitional namespace:\n%s", got)
	}
}

func TestRemoveUnbalancedRangeMarkersSupportsStrictOOXML(t *testing.T) {
	original := []byte(`<s:document xmlns:s="` + constants.NamespaceMainStrict + `"><s:body><s:bookmarkStart s:id="7"/><s:bookmarkEnd s:id="7"/></s:body></s:document>`)
	changed := []byte(`<s:document xmlns:s="` + constants.NamespaceMainStrict + `"><s:body><s:bookmarkStart s:id="7"/></s:body></s:document>`)

	got, err := removeUnbalancedRangeMarkers(original, changed)
	if err != nil {
		t.Fatalf("removeUnbalancedRangeMarkers: %v", err)
	}
	if bytes.Contains(got, []byte(`bookmarkStart`)) {
		t.Fatalf("unbalanced Strict range marker survived: %s", got)
	}
}

func TestTablePropertiesPolicyCollapsesChangedSingletonDuplicates(t *testing.T) {
	source := []byte(`<w:tblPr><w:tblW w:type="dxa" w:w="1000"/><w:tblW w:type="dxa" w:w="2000" w:opaque="drop-with-duplicate"/></w:tblPr>`)
	snapshot := []byte(`<w:tblPr><w:tblW w:type="dxa" w:w="1000"></w:tblW></w:tblPr>`)
	current := []byte(`<w:tblPr><w:tblW w:type="dxa" w:w="3000"></w:tblW></w:tblPr>`)

	got, err := mergeRoundTripFragment(source, snapshot, current, tablePropertiesPolicy(), roundTripNamespaces)
	if err != nil {
		t.Fatalf("mergeRoundTripFragment: %v", err)
	}
	if count := bytes.Count(got, []byte(`<w:tblW`)); count != 1 {
		t.Fatalf("merged table properties have %d tblW children, want 1: %s", count, got)
	}
	for _, want := range [][]byte{[]byte(`w:type="dxa"`), []byte(`w:w="3000"`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged table properties lost %s: %s", want, got)
		}
	}
}
