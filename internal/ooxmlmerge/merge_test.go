// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca

package ooxmlmerge

import (
	"bytes"
	"testing"
)

const (
	wordNS          = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	strictWordNS    = "http://purl.oclc.org/ooxml/wordprocessingml/main"
	relsNS          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	strictRelsNS    = "http://purl.oclc.org/ooxml/officeDocument/relationships"
	drawingNS       = "http://schemas.openxmlformats.org/drawingml/2006/main"
	strictDrawingNS = "http://purl.oclc.org/ooxml/drawingml/main"
	pictureNS       = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	strictPictureNS = "http://purl.oclc.org/ooxml/drawingml/picture"
	wordDrawingNS   = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	strictWordDraw  = "http://purl.oclc.org/ooxml/drawingml/wordprocessingDrawing"
)

var testNamespaces = map[string]string{"w": wordNS, "r": relsNS, "x": "urn:test"}

func mustParse(t testing.TB, raw string) *Node {
	t.Helper()
	node, err := Parse([]byte(raw), testNamespaces)
	if err != nil {
		t.Fatalf("Parse(%q): %v", raw, err)
	}
	return node
}

func TestMergeAttributesUsesResolvedQName(t *testing.T) {
	source := mustParse(t, `<w:pgMar w:top="1440" x:top="opaque" w:gutter="72"/>`)
	snapshot := mustParse(t, `<w:pgMar w:top="1440"/>`)
	current := mustParse(t, `<w:pgMar w:top="1500"/>`)
	policy := &Policy{
		Mode:            MergeAttributes,
		OwnedAttributes: []QName{Name(wordNS, "top")},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	for _, want := range [][]byte{[]byte(`w:top="1500"`), []byte(`x:top="opaque"`), []byte(`w:gutter="72"`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
}

func TestMergeAttributesPreservesOpaqueAttributeLexicalBytes(t *testing.T) {
	source := mustParse(t, "<w:pgMar\n  w:top = '1440'   x:opaque = '&#x41;' w:gutter=\"72\" >opaque</w:pgMar >")
	snapshot := mustParse(t, `<w:pgMar w:top="1440"/>`)
	current := mustParse(t, `<w:pgMar w:top="1500"/>`)
	policy := &Policy{Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "top")}}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	want := "<w:pgMar\n  w:top = '1500'   x:opaque = '&#x41;' w:gutter=\"72\" >opaque</w:pgMar >"
	if string(got) != want {
		t.Fatalf("Merge = %q, want byte-preserved %q", got, want)
	}
}

func TestDefaultNamespaceDoesNotApplyToAttributes(t *testing.T) {
	node, err := Parse([]byte(`<root xmlns="urn:elements" value="plain"/>`), nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if node.Name.Namespace != "urn:elements" {
		t.Fatalf("element namespace = %q, want urn:elements", node.Name.Namespace)
	}
	if got, ok := node.Attr(Name("", "value")); !ok || got != "plain" {
		t.Fatalf("unprefixed attribute = %q, %v; want plain, true", got, ok)
	}
	if _, ok := node.Attr(Name("urn:elements", "value")); ok {
		t.Fatal("default namespace was incorrectly applied to an unprefixed attribute")
	}
}

func TestRenderInContextAvoidsConflictingGeneratedPrefix(t *testing.T) {
	generated := mustParse(t, `<w:pgSz w:w="12240" w:h="15840"><r:child r:id="rId1"/></w:pgSz>`)
	context := map[string]string{
		"x":   wordNS,
		"rel": relsNS,
		"w":   "urn:not-wordprocessingml",
		"r":   "urn:not-relationships",
	}

	got, err := RenderInContext(generated, context)
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	want := `<x:pgSz x:w="12240" x:h="15840"><rel:child rel:id="rId1"/></x:pgSz>`
	if string(got) != want {
		t.Fatalf("RenderInContext = %q, want %q", got, want)
	}
}

func TestRenderInContextDeclaresCollisionFreePrefix(t *testing.T) {
	generated := mustParse(t, `<w:pgSz w:w="12240"/>`)
	got, err := RenderInContext(generated, map[string]string{"w": "urn:not-wordprocessingml"})
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	if want := `<w1:pgSz w1:w="12240" xmlns:w1="` + wordNS + `"/>`; string(got) != want {
		t.Fatalf("RenderInContext = %q, want %q", got, want)
	}
}

func TestMergeTreatsStrictAndTransitionalQNamesAsEquivalent(t *testing.T) {
	source, err := Parse([]byte(`<s:pgMar xmlns:s="`+strictWordNS+`" s:top="1440" s:gutter="72"/>`), nil)
	if err != nil {
		t.Fatalf("Parse source: %v", err)
	}
	snapshot := mustParse(t, `<w:pgMar w:top="1440"/>`)
	current := mustParse(t, `<w:pgMar w:top="1500"/>`)
	policy := &Policy{Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "top")}}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	want := `<s:pgMar xmlns:s="` + strictWordNS + `" s:top="1500" s:gutter="72"/>`
	if string(got) != want {
		t.Fatalf("Merge = %q, want %q", got, want)
	}
}

func TestRenderInStrictContextReusesStrictBinding(t *testing.T) {
	generated := mustParse(t, `<w:pgSz w:w="12240" w:h="15840"/>`)
	got, err := RenderInContext(generated, map[string]string{"s": strictWordNS})
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	if want := `<s:pgSz s:w="12240" s:h="15840"/>`; string(got) != want {
		t.Fatalf("RenderInContext = %q, want %q", got, want)
	}
}

func TestRenderInStrictContextDoesNotReuseTransitionalBinding(t *testing.T) {
	generated, err := Parse([]byte(`<w:p r:id="rId1"/>`), map[string]string{
		"w": wordNS,
		"r": relsNS,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := RenderInContextWithMainNamespace(generated, map[string]string{
		"s":   strictWordNS,
		"rel": strictRelsNS,
		"w":   wordNS,
		"r":   relsNS,
	}, strictWordNS)
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	if want := `<s:p rel:id="rId1"/>`; string(got) != want {
		t.Fatalf("RenderInContext = %q, want %q", got, want)
	}
}

func TestRenderInTransitionalContextDoesNotUseUnusedStrictBinding(t *testing.T) {
	generated, err := Parse([]byte(`<w:p r:id="rId1"/>`), map[string]string{
		"w": wordNS,
		"r": relsNS,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := RenderInContextWithMainNamespace(generated, map[string]string{
		"w":   wordNS,
		"r":   relsNS,
		"s":   strictWordNS,
		"rel": strictRelsNS,
	}, wordNS)
	if err != nil {
		t.Fatalf("RenderInContextWithMainNamespace: %v", err)
	}
	if want := `<w:p r:id="rId1"/>`; string(got) != want {
		t.Fatalf("RenderInContextWithMainNamespace = %q, want %q", got, want)
	}
}

func TestMergeUsesSourceQNameToAnchorFamilyWithMixedDeclarations(t *testing.T) {
	source, err := Parse([]byte(`<w:root xmlns:w="`+wordNS+`" xmlns:s="`+strictWordNS+`"><w:a/></w:root>`), nil)
	if err != nil {
		t.Fatalf("Parse source: %v", err)
	}
	snapshot := mustParse(t, `<w:root><w:a/></w:root>`)
	current := mustParse(t, `<w:root><w:a/><w:b/></w:root>`)
	bKey := PolicyKey(Name(wordNS, "b"))
	policy := &Policy{
		Mode:       MergeChildren,
		Children:   map[string]*Policy{bKey: {Mode: Replace}},
		ChildRanks: map[string]int{bKey: 1},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !bytes.Contains(got, []byte(`<w:b`)) || bytes.Contains(got, []byte(`<s:b`)) {
		t.Fatalf("merged Transitional source used the unused Strict binding: %s", got)
	}
}

func TestRenderInStrictContextConvertsGraphicDataPictureURI(t *testing.T) {
	generated, err := Parse([]byte(`<a:graphicData uri="`+pictureNS+`"><pic:pic/></a:graphicData>`), map[string]string{
		"a":   drawingNS,
		"pic": pictureNS,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := RenderInContextWithMainNamespace(generated, map[string]string{"s": strictWordNS}, strictWordNS)
	if err != nil {
		t.Fatalf("RenderInContextWithMainNamespace: %v", err)
	}
	if !bytes.Contains(got, []byte(`uri="`+strictPictureNS+`"`)) {
		t.Fatalf("Strict graphicData retained Transitional picture URI: %s", got)
	}
	if bytes.Contains(got, []byte(`uri="`+pictureNS+`"`)) {
		t.Fatalf("Strict graphicData contains Transitional picture URI: %s", got)
	}
}

func TestRenderInStrictDefaultNamespaceDeclaresStrictQualifiedPrefix(t *testing.T) {
	generated := mustParse(t, `<w:pgSz w:w="12240"/>`)
	got, err := RenderInContext(generated, map[string]string{"": strictWordNS})
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	want := `<w:pgSz w:w="12240" xmlns:w="` + strictWordNS + `"/>`
	if string(got) != want {
		t.Fatalf("RenderInContext = %q, want %q", got, want)
	}
}

func TestRenderInStrictDocumentUsesStrictVariantForNewOOXMLFamilies(t *testing.T) {
	generated, err := Parse([]byte(`<w:root r:id="rId1"><wp:inline><a:graphic><pic:pic/></a:graphic></wp:inline></w:root>`), map[string]string{
		"w": wordNS, "r": relsNS, "wp": wordDrawingNS, "a": drawingNS, "pic": pictureNS,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := RenderInContext(generated, map[string]string{"s": strictWordNS})
	if err != nil {
		t.Fatalf("RenderInContext: %v", err)
	}
	for _, want := range []string{strictRelsNS, strictWordDraw, strictDrawingNS, strictPictureNS} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("rendered Strict fragment is missing namespace %q: %s", want, got)
		}
	}
	for _, unwanted := range []string{wordNS, relsNS, wordDrawingNS, drawingNS, pictureNS} {
		if bytes.Contains(got, []byte(unwanted)) {
			t.Errorf("rendered Strict fragment contains Transitional namespace %q: %s", unwanted, got)
		}
	}
}

func TestMergeAttributesCanReplaceContentAndDropConflicts(t *testing.T) {
	source := mustParse(t, `<w:cols w:num="2" w:equalWidth="0" w:sep="1"><w:col w:w="3000"/><w:col w:w="7000"/></w:cols>`)
	snapshot := mustParse(t, `<w:cols w:num="2" w:space="720"></w:cols>`)
	current := mustParse(t, `<w:cols w:num="3" w:space="720"></w:cols>`)
	policy := &Policy{
		Mode:              MergeAttributes,
		OwnedAttributes:   []QName{Name(wordNS, "num")},
		DropAttributes:    []QName{Name(wordNS, "equalWidth")},
		UseCurrentContent: true,
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	for _, want := range [][]byte{[]byte(`w:num="3"`), []byte(`w:sep="1"`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
	for _, unwanted := range [][]byte{[]byte(`equalWidth`), []byte(`<w:col `), []byte(`<w:col>`)} {
		if bytes.Contains(got, unwanted) {
			t.Errorf("merged fragment retained %s: %s", unwanted, got)
		}
	}
}

func TestMergeAttributesCurrentContentUsesSourceNamespaceContext(t *testing.T) {
	source, err := Parse([]byte(`<s:cols xmlns:s="`+strictWordNS+`"><s:col/></s:cols>`), nil)
	if err != nil {
		t.Fatalf("Parse source: %v", err)
	}
	snapshot := mustParse(t, `<w:cols/>`)
	current := mustParse(t, `<w:cols><w:col w:w="100"/></w:cols>`)
	policy := &Policy{Mode: MergeAttributes, UseCurrentContent: true}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	merged, err := Parse(got, nil)
	if err != nil {
		t.Fatalf("Parse merged XML: %v\n%s", err, got)
	}
	if len(merged.Children) != 1 || merged.Children[0].Name.Namespace != strictWordNS {
		t.Fatalf("merged content did not use the Strict source namespace: %s", got)
	}
	if bytes.Contains(got, []byte(`<w:`)) {
		t.Fatalf("merged Strict content retained the generated Transitional prefix: %s", got)
	}
}

func TestMergeChildrenDropsOnlyNamedSourceChildren(t *testing.T) {
	source := mustParse(t, `<w:cols w:num="2" w:equalWidth="0"><w:col w:w="3000"/><x:balance x:mode="keep"/><w:col w:w="7000"/></w:cols>`)
	snapshot := mustParse(t, `<w:cols w:num="2"></w:cols>`)
	current := mustParse(t, `<w:cols w:num="3"></w:cols>`)
	colKey := PolicyKey(Name(wordNS, "col"))
	policy := &Policy{
		Mode:               MergeChildren,
		OwnedAttributes:    []QName{Name(wordNS, "num")},
		DropAttributes:     []QName{Name(wordNS, "equalWidth")},
		DropSourceChildren: []QName{Name(wordNS, "col")},
		Children:           map[string]*Policy{colKey: {Mode: Replace}},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	for _, want := range [][]byte{[]byte(`w:num="3"`), []byte(`<x:balance x:mode="keep"/>`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
	for _, unwanted := range [][]byte{[]byte(`equalWidth`), []byte(`<w:col `), []byte(`<w:col>`)} {
		if bytes.Contains(got, unwanted) {
			t.Errorf("merged fragment retained %s: %s", unwanted, got)
		}
	}
}

func TestMergeChildrenPreservesUnknownAndMergesManagedChildren(t *testing.T) {
	source := mustParse(t, `<w:tblBorders><w:top w:val="single" w:sz="8" w:space="12"/><w:unknown/><w:bottom w:val="single" w:sz="8" w:space="13"/></w:tblBorders>`)
	snapshot := mustParse(t, `<w:tblBorders><w:top w:val="single" w:sz="8"></w:top><w:bottom w:val="single" w:sz="8"></w:bottom></w:tblBorders>`)
	current := mustParse(t, `<w:tblBorders><w:top w:val="single" w:sz="16"></w:top><w:bottom w:val="single" w:sz="8"></w:bottom></w:tblBorders>`)
	borderPolicy := &Policy{
		Mode:            MergeAttributes,
		OwnedAttributes: []QName{Name(wordNS, "val"), Name(wordNS, "sz"), Name(wordNS, "color")},
	}
	policy := &Policy{
		Mode: MergeChildren,
		Children: map[string]*Policy{
			PolicyKey(Name(wordNS, "top")):    borderPolicy,
			PolicyKey(Name(wordNS, "bottom")): borderPolicy,
		},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	for _, want := range [][]byte{[]byte(`w:sz="16"`), []byte(`w:space="12"`), []byte(`w:space="13"`), []byte(`<w:unknown/>`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
}

func TestMergeChildrenMergesContainerAttributes(t *testing.T) {
	source := mustParse(t, `<w:root w:owned="old" w:drop="remove" x:opaque="keep"><w:child/></w:root>`)
	snapshot := mustParse(t, `<w:root w:owned="old" w:drop="remove"><w:child/></w:root>`)
	current := mustParse(t, `<w:root w:owned="new"><w:child/></w:root>`)
	policy := &Policy{
		Mode:            MergeChildren,
		OwnedAttributes: []QName{Name(wordNS, "owned")},
		DropAttributes:  []QName{Name(wordNS, "drop")},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	for _, want := range [][]byte{[]byte(`w:owned="new"`), []byte(`x:opaque="keep"`), []byte(`<w:child/>`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
	if bytes.Contains(got, []byte(`w:drop=`)) {
		t.Errorf("merged fragment retained dropped container attribute: %s", got)
	}
}

func TestMergeChildrenAddsAndRemovesManagedChildren(t *testing.T) {
	source := mustParse(t, `<w:sectPr><w:type w:val="nextPage"/><w:opaque/></w:sectPr>`)
	snapshot := mustParse(t, `<w:sectPr><w:type w:val="nextPage"></w:type></w:sectPr>`)
	current := mustParse(t, `<w:sectPr><w:pgSz w:w="12240" w:h="15840"></w:pgSz></w:sectPr>`)
	attrPolicy := &Policy{Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "val"), Name(wordNS, "w"), Name(wordNS, "h")}}
	policy := &Policy{
		Mode: MergeChildren,
		Children: map[string]*Policy{
			PolicyKey(Name(wordNS, "type")): attrPolicy,
			PolicyKey(Name(wordNS, "pgSz")): attrPolicy,
		},
		ChildRanks: map[string]int{
			PolicyKey(Name(wordNS, "type")): 0,
			PolicyKey(Name(wordNS, "pgSz")): 1,
		},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if bytes.Contains(got, []byte(`<w:type`)) {
		t.Fatalf("merged fragment retained removed type: %s", got)
	}
	for _, want := range [][]byte{[]byte(`<w:pgSz`), []byte(`<w:opaque/>`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
}

func TestMergeChildrenDoesNotUseOpaqueChildAsOrderingBarrier(t *testing.T) {
	source := mustParse(t, `<w:sectPr><w:pgSz/><x:opaque/><w:pgMar/></w:sectPr>`)
	snapshot := mustParse(t, `<w:sectPr><w:pgSz/><w:pgMar/></w:sectPr>`)
	current := mustParse(t, `<w:sectPr><w:pgSz/><w:pgMar/><w:cols/></w:sectPr>`)
	pgSzKey := PolicyKey(Name(wordNS, "pgSz"))
	pgMarKey := PolicyKey(Name(wordNS, "pgMar"))
	colsKey := PolicyKey(Name(wordNS, "cols"))
	policy := &Policy{
		Mode: MergeChildren,
		Children: map[string]*Policy{
			pgSzKey:  {Mode: MergeAttributes},
			pgMarKey: {Mode: MergeAttributes},
			colsKey:  {Mode: Replace},
		},
		ChildRanks: map[string]int{pgSzKey: 5, pgMarKey: 6, colsKey: 11},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	marginAt := bytes.Index(got, []byte(`<w:pgMar`))
	columnsAt := bytes.Index(got, []byte(`<w:cols`))
	if marginAt < 0 || columnsAt < 0 || marginAt > columnsAt {
		t.Fatalf("inserted columns precede the lower-ranked page margins: %s", got)
	}
	if !bytes.Contains(got, []byte(`<x:opaque/>`)) {
		t.Fatalf("merged content lost the opaque child: %s", got)
	}
}

func TestPreservePolicyKeepsSourceWhenCurrentProjectionIsAbsent(t *testing.T) {
	source := mustParse(t, `<w:opaque x:value="keep"/>`)
	got, err := Merge(source, nil, nil, &Policy{Mode: Preserve})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !bytes.Equal(got, source.Bytes()) {
		t.Fatalf("Merge = %q, want preserved %q", got, source.Bytes())
	}
}

func TestMergeChildrenMatchesRepeatedChildrenByPosition(t *testing.T) {
	source := mustParse(t, `<w:tblGrid><w:gridCol w:w="100"/><w:gridCol w:w="200" x:hint="keep"/><w:gridCol w:w="300"/></w:tblGrid>`)
	snapshot := mustParse(t, `<w:tblGrid><w:gridCol w:w="100"></w:gridCol><w:gridCol w:w="200"></w:gridCol><w:gridCol w:w="300"></w:gridCol></w:tblGrid>`)
	current := mustParse(t, `<w:tblGrid><w:gridCol w:w="100"></w:gridCol><w:gridCol w:w="250"></w:gridCol><w:gridCol w:w="300"></w:gridCol></w:tblGrid>`)
	gridColKey := PolicyKey(Name(wordNS, "gridCol"))
	policy := &Policy{
		Mode:       MergeChildren,
		Children:   map[string]*Policy{gridColKey: {Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "w")}}},
		ChildRanks: map[string]int{gridColKey: 0},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	merged := mustParse(t, string(got))
	if len(merged.Children) != 3 {
		t.Fatalf("merged grid has %d columns, want 3: %s", len(merged.Children), got)
	}
	for index, want := range []string{"100", "250", "300"} {
		if value, _ := merged.Children[index].Attr(Name(wordNS, "w")); value != want {
			t.Errorf("column %d width = %q, want %q", index, value, want)
		}
	}
	if hint, _ := merged.Children[1].Attr(Name("urn:test", "hint")); hint != "keep" {
		t.Errorf("second column hint = %q, want keep", hint)
	}
}

func TestMergeChildrenCustomKeyCollapsesDuplicateManagedChildren(t *testing.T) {
	source := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId1"/><w:opaque/><w:headerReference w:type="default" r:id="rId2"/></w:sectPr>`)
	snapshot := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId1"></w:headerReference></w:sectPr>`)
	current := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId3"></w:headerReference></w:sectPr>`)
	headerKey := PolicyKey(Name(wordNS, "headerReference"))
	policy := &Policy{
		Mode: MergeChildren,
		Children: map[string]*Policy{
			headerKey: {
				Mode:            MergeAttributes,
				OwnedAttributes: []QName{Name(wordNS, "type"), Name(relsNS, "id")},
			},
		},
		ChildKey: func(node *Node) string {
			typeValue, _ := node.Attr(Name(wordNS, "type"))
			return PolicyKey(node.Name) + "\x00type=" + typeValue
		},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if count := bytes.Count(got, []byte(`<w:headerReference`)); count != 1 {
		t.Fatalf("merged fragment has %d default header references, want 1: %s", count, got)
	}
	for _, want := range [][]byte{[]byte(`r:id="rId3"`), []byte(`<w:opaque/>`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
}

func TestMergeChildrenCustomKeyPreservesDuplicatesWhenThatKeyIsUnchanged(t *testing.T) {
	source := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId1"/><w:headerReference w:type="default" r:id="rId2"/><w:pgMar w:top="1440" w:gutter="72"/></w:sectPr>`)
	snapshot := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId1"></w:headerReference><w:pgMar w:top="1440"></w:pgMar></w:sectPr>`)
	current := mustParse(t, `<w:sectPr><w:headerReference w:type="default" r:id="rId1"></w:headerReference><w:pgMar w:top="1500"></w:pgMar></w:sectPr>`)
	headerKey := PolicyKey(Name(wordNS, "headerReference"))
	marginKey := PolicyKey(Name(wordNS, "pgMar"))
	policy := &Policy{
		Mode: MergeChildren,
		Children: map[string]*Policy{
			headerKey: {Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "type"), Name(relsNS, "id")}},
			marginKey: {Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "top")}},
		},
		ChildKey: func(node *Node) string {
			typeValue, _ := node.Attr(Name(wordNS, "type"))
			return PolicyKey(node.Name) + "\x00type=" + typeValue
		},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if count := bytes.Count(got, []byte(`<w:headerReference`)); count != 2 {
		t.Fatalf("merged fragment has %d default header references, want the 2 unchanged source references: %s", count, got)
	}
	for _, want := range [][]byte{[]byte(`r:id="rId1"`), []byte(`r:id="rId2"`), []byte(`w:top="1500"`), []byte(`w:gutter="72"`)} {
		if !bytes.Contains(got, want) {
			t.Errorf("merged fragment lost %s: %s", want, got)
		}
	}
}

func TestApplyEditsUsesOriginalOffsets(t *testing.T) {
	got, err := ApplyEdits([]byte("abcdef"), []Edit{
		{Start: 1, End: 3, Replacement: []byte("X")},
		{Start: 4, End: 6, Replacement: []byte("YZ")},
	})
	if err != nil {
		t.Fatalf("ApplyEdits: %v", err)
	}
	if string(got) != "aXdYZ" {
		t.Fatalf("ApplyEdits = %q, want %q", got, "aXdYZ")
	}
}

func TestApplyEditsPreservesEqualOffsetInsertionOrder(t *testing.T) {
	got, err := ApplyEdits([]byte("ac"), []Edit{
		{Start: 1, End: 1, Replacement: []byte("1")},
		{Start: 1, End: 1, Replacement: []byte("2")},
	})
	if err != nil {
		t.Fatalf("ApplyEdits: %v", err)
	}
	if string(got) != "a12c" {
		t.Fatalf("ApplyEdits = %q, want %q", got, "a12c")
	}
}

func TestApplyEditsHandlesInsertionAtReplacementStart(t *testing.T) {
	got, err := ApplyEdits([]byte("abcde"), []Edit{
		{Start: 1, End: 1, Replacement: []byte("X")},
		{Start: 1, End: 3, Replacement: []byte("Y")},
	})
	if err != nil {
		t.Fatalf("ApplyEdits: %v", err)
	}
	if string(got) != "aXYde" {
		t.Fatalf("ApplyEdits = %q, want %q", got, "aXYde")
	}
}

func TestSplicePolicyUsesSourceRelativeOffsets(t *testing.T) {
	source := mustParse(t, `<w:root><w:a/>gap<w:b/></w:root>`)
	snapshot := mustParse(t, `<w:root><w:a/><w:b/></w:root>`)
	current := mustParse(t, `<w:root><w:a/><w:c/><w:b/></w:root>`)
	policy := &Policy{
		Mode: Splice,
		SpliceEdits: func(source, _, _ *Node) ([]Edit, error) {
			b := First(source, Name(wordNS, "b"))
			return []Edit{{
				Start:       b.Start - source.Start,
				End:         b.Start - source.Start,
				Replacement: []byte(`<w:c/>`),
			}}, nil
		},
	}

	got, err := Merge(source, snapshot, current, policy)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if want := `<w:root><w:a/>gap<w:c/><w:b/></w:root>`; string(got) != want {
		t.Fatalf("Merge = %q, want %q", got, want)
	}
}

func FuzzMergeAttributes(f *testing.F) {
	for _, value := range []string{"plain", "with space", `quote"value`, "áéí", "&<>"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		source := mustParse(t, `<w:pgMar w:top="1" x:top="opaque"/>`)
		snapshot := mustParse(t, `<w:pgMar w:top="1"/>`)
		current := mustParse(t, `<w:pgMar w:top="2"/>`)
		policy := &Policy{Mode: MergeAttributes, OwnedAttributes: []QName{Name(wordNS, "top")}}
		source.Attributes = append(source.Attributes, Attribute{Name: QName{Prefix: "x", Namespace: "urn:test", Local: "seed"}, Value: value})
		if _, err := Merge(source, snapshot, current, policy); err != nil {
			t.Fatalf("Merge: %v", err)
		}
	})
}

func FuzzMergeNamespaceAndWhitespace(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(1), uint8(2))
	f.Add(uint8(2), uint8(3))
	f.Fuzz(func(t *testing.T, prefixSeed, whitespaceSeed uint8) {
		prefixes := []string{"w", "word", "x"}
		whitespace := []string{"", " ", "\n", "\r\n  "}
		prefix := prefixes[int(prefixSeed)%len(prefixes)]
		gap := whitespace[int(whitespaceSeed)%len(whitespace)]
		raw := `<` + prefix + `:root xmlns:` + prefix + `="` + wordNS + `">` + gap + `<` + prefix + `:child ` + prefix + `:value="1"/>` + gap + `</` + prefix + `:root>`
		node, err := Parse([]byte(raw), nil)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if node.Name.Namespace != wordNS || len(node.Children) != 1 {
			t.Fatalf("resolved tree = %+v", node)
		}
		if !bytes.Equal(node.Bytes(), []byte(raw)) {
			t.Fatalf("parsed bytes changed: %q", node.Bytes())
		}
	})
}
