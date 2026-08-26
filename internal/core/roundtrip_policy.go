// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"github.com/mmonterroca/docxgo/v2/internal/ooxmlmerge"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

var roundTripNamespaces = map[string]string{
	"w":  constants.NamespaceMain,
	"r":  constants.NamespaceRelationships,
	"wp": constants.NamespaceWordprocessingDrawing,
}

func wordName(local string) ooxmlmerge.QName {
	return ooxmlmerge.Name(constants.NamespaceMain, local)
}

func relationshipName(local string) ooxmlmerge.QName {
	return ooxmlmerge.Name(constants.NamespaceRelationships, local)
}

func attributePolicy(names ...ooxmlmerge.QName) *ooxmlmerge.Policy {
	return &ooxmlmerge.Policy{Mode: ooxmlmerge.MergeAttributes, OwnedAttributes: names}
}

func sectionReferenceKey(node *ooxmlmerge.Node) string {
	if node == nil {
		return ""
	}
	typeValue, _ := node.Attr(wordName("type"))
	return ooxmlmerge.PolicyKey(node.Name) + "\x00type=" + typeValue
}

func qNameChildKey(node *ooxmlmerge.Node) string {
	if node == nil {
		return ""
	}
	return ooxmlmerge.PolicyKey(node.Name)
}

func sectionPropertiesPolicy() *ooxmlmerge.Policy {
	pgMar := attributePolicy(
		wordName("top"), wordName("right"), wordName("bottom"),
		wordName("left"), wordName("header"), wordName("footer"),
	)
	pgSz := attributePolicy(wordName("w"), wordName("h"), wordName("orient"))
	colKey := ooxmlmerge.PolicyKey(wordName("col"))
	cols := &ooxmlmerge.Policy{
		Mode:               ooxmlmerge.MergeChildren,
		OwnedAttributes:    []ooxmlmerge.QName{wordName("num")},
		DropSourceChildren: []ooxmlmerge.QName{wordName("col")},
		Children: map[string]*ooxmlmerge.Policy{
			colKey: {Mode: ooxmlmerge.Replace},
		},
		ChildRanks: map[string]int{colKey: 0},
	}
	cols.DropAttributes = []ooxmlmerge.QName{wordName("equalWidth")}

	children := map[string]*ooxmlmerge.Policy{
		ooxmlmerge.PolicyKey(wordName("headerReference")): attributePolicy(wordName("type"), relationshipName("id")),
		ooxmlmerge.PolicyKey(wordName("footerReference")): attributePolicy(wordName("type"), relationshipName("id")),
		ooxmlmerge.PolicyKey(wordName("type")):            attributePolicy(wordName("val")),
		ooxmlmerge.PolicyKey(wordName("pgSz")):            pgSz,
		ooxmlmerge.PolicyKey(wordName("pgMar")):           pgMar,
		ooxmlmerge.PolicyKey(wordName("cols")):            cols,
	}
	ranks := make(map[string]int, len(sectionPropertyRanks))
	for name, rank := range sectionPropertyRanks {
		ranks[ooxmlmerge.PolicyKey(wordName(name))] = rank
	}
	return &ooxmlmerge.Policy{
		Mode:       ooxmlmerge.MergeChildren,
		Children:   children,
		ChildRanks: ranks,
		ChildKey:   sectionReferenceKey,
	}
}

func tableBordersPolicy() *ooxmlmerge.Policy {
	border := attributePolicy(wordName("val"), wordName("sz"), wordName("color"))
	children := make(map[string]*ooxmlmerge.Policy)
	ranks := make(map[string]int)
	for index, name := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		key := ooxmlmerge.PolicyKey(wordName(name))
		children[key] = border
		ranks[key] = index
	}
	return &ooxmlmerge.Policy{
		Mode:       ooxmlmerge.MergeChildren,
		Children:   children,
		ChildRanks: ranks,
		ChildKey:   qNameChildKey,
	}
}

func tablePropertiesPolicy() *ooxmlmerge.Policy {
	children := map[string]*ooxmlmerge.Policy{
		ooxmlmerge.PolicyKey(wordName("tblStyle")):   attributePolicy(wordName("val")),
		ooxmlmerge.PolicyKey(wordName("tblW")):       attributePolicy(wordName("type"), wordName("w")),
		ooxmlmerge.PolicyKey(wordName("jc")):         attributePolicy(wordName("val")),
		ooxmlmerge.PolicyKey(wordName("tblBorders")): tableBordersPolicy(),
		ooxmlmerge.PolicyKey(wordName("tblLook")):    attributePolicy(wordName("val")),
	}
	ranks := make(map[string]int, len(tablePropertyRanks))
	for name, rank := range tablePropertyRanks {
		ranks[ooxmlmerge.PolicyKey(wordName(name))] = rank
	}
	return &ooxmlmerge.Policy{
		Mode:       ooxmlmerge.MergeChildren,
		Children:   children,
		ChildRanks: ranks,
		ChildKey:   qNameChildKey,
	}
}

func tableGridPolicy() *ooxmlmerge.Policy {
	gridColKey := ooxmlmerge.PolicyKey(wordName("gridCol"))
	return &ooxmlmerge.Policy{
		Mode:       ooxmlmerge.MergeChildren,
		Children:   map[string]*ooxmlmerge.Policy{gridColKey: attributePolicy(wordName("w"))},
		ChildRanks: map[string]int{gridColKey: 0},
	}
}

func mergeRoundTripFragment(source, snapshot, current []byte, policy *ooxmlmerge.Policy, namespaces map[string]string) ([]byte, error) {
	sourceNode, err := ooxmlmerge.Parse(source, namespaces)
	if err != nil {
		return nil, err
	}
	var snapshotNode *ooxmlmerge.Node
	if len(snapshot) > 0 {
		snapshotNode, err = ooxmlmerge.Parse(snapshot, roundTripNamespaces)
		if err != nil {
			return nil, err
		}
	}
	var currentNode *ooxmlmerge.Node
	if len(current) > 0 {
		currentNode, err = ooxmlmerge.Parse(current, roundTripNamespaces)
		if err != nil {
			return nil, err
		}
	}
	merged, err := ooxmlmerge.Merge(sourceNode, snapshotNode, currentNode, policy)
	if err != nil {
		return nil, err
	}
	return merged, nil
}
