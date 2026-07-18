// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package reader

import (
	"bytes"
	"encoding/xml"
)

// Element represents a generic XML element with nested children.
type Element struct {
	Name     xml.Name
	Attr     []xml.Attr
	Text     string
	Children []*Element
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
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return parseElement(dec, start)
		}
	}
}

func parseElement(dec *xml.Decoder, start xml.StartElement) (*Element, error) {
	elem := &Element{
		Name: start.Name,
		Attr: append([]xml.Attr(nil), start.Attr...),
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseElement(dec, t)
			if err != nil {
				return nil, err
			}
			elem.addChild(child)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local && t.Name.Space == start.Name.Space {
				return elem, nil
			}
		case xml.CharData:
			// Preserve meaningful text; ignore nodes that are entirely whitespace.
			if len(bytes.TrimSpace([]byte(t))) == 0 {
				continue
			}
			elem.Text += string(t)
		}
	}
}
