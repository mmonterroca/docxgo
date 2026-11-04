// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package reader

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// Element represents a generic XML element with nested children.
type Element struct {
	Name               xml.Name
	Attr               []xml.Attr
	Text               string
	Children           []*Element
	StartOffset        int
	ContentStartOffset int
	ContentEndOffset   int
	EndOffset          int
}

// findOrCreateChild attaches a child to the element, creating the slice lazily.
func (e *Element) addChild(child *Element) {
	if child == nil {
		return
	}
	e.Children = append(e.Children, child)
}

// parseXMLTree parses the provided XML bytes into an Element tree.
func parseXMLTree(data []byte) (*Element, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		startOffset := int(dec.InputOffset())
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return parseElement(dec, start, startOffset, false)
		}
	}
}

// xmlNamespace is the namespace bound to the reserved xml prefix.
const xmlNamespace = "http://www.w3.org/XML/1998/namespace"

// parseElement parses start and its content. preserve reports whether
// xml:space="preserve" is inherited from the ancestors of start.
func parseElement(dec *xml.Decoder, start xml.StartElement, startOffset int, preserve bool) (*Element, error) {
	elem := &Element{
		Name:               start.Name,
		Attr:               append([]xml.Attr(nil), start.Attr...),
		StartOffset:        startOffset,
		ContentStartOffset: int(dec.InputOffset()),
	}
	if space, ok := getAttrQName(elem, xmlNamespace, "space"); ok {
		preserve = space == "preserve"
	}

	// Every character data node, including whitespace-only ones. Between
	// child elements whitespace is only indentation, but in an element
	// without children, such as <w:t xml:space="preserve"> </w:t>, it can be
	// the content itself.
	var allText strings.Builder

	for {
		tokenStart := int(dec.InputOffset())
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseElement(dec, t, tokenStart, preserve)
			if err != nil {
				return nil, err
			}
			elem.addChild(child)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local && t.Name.Space == start.Name.Space {
				// Whitespace-only text is content only when preserve is set.
				// XML whitespace is space, tab, CR and LF, so a no-break
				// space always counts as content.
				text := allText.String()
				if len(elem.Children) == 0 && (preserve || strings.Trim(text, " \t\r\n") != "") {
					elem.Text = text
				}
				elem.ContentEndOffset = tokenStart
				elem.EndOffset = int(dec.InputOffset())
				return elem, nil
			}
		case xml.CharData:
			allText.Write(t)
			// Preserve meaningful text; ignore nodes that are entirely whitespace
			// (a leaf element may keep them, see allText).
			if len(bytes.TrimSpace([]byte(t))) == 0 {
				continue
			}
			elem.Text += string(t)
		}
	}
}
