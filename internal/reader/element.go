/*
MIT License

Copyright (c) 2025 Misael Monterroca

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
