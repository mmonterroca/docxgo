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
	"encoding/xml"

	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

const (
	opParsePackage = "reader.ParsePackage"
)

// ParsedPackage holds strongly typed OOXML structures extracted from a Package.
type ParsedPackage struct {
	Package *Package

	DocumentTree *Element
	StylesTree   *Element
	HeaderTrees  map[string]*Element
	FooterTrees  map[string]*Element

	RootRelationships     *xmlstructs.Relationships
	DocumentRelationships *xmlstructs.Relationships

	CorePropertiesTree *Element
	AppPropertiesTree  *Element
	CustomProperties   []byte

	ThemeParts  map[string][]byte
	Numbering   []byte
	FontTable   []byte
	Settings    []byte
	WebSettings []byte
}

// ParsePackage converts the raw byte-oriented Package into typed OOXML structures.
func ParsePackage(pkg *Package) (*ParsedPackage, error) {
	if pkg == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opParsePackage, "package cannot be nil")
	}

	parsed := &ParsedPackage{
		Package:          pkg,
		HeaderTrees:      make(map[string]*Element, len(pkg.Headers)),
		FooterTrees:      make(map[string]*Element, len(pkg.Footers)),
		CustomProperties: pkg.CustomProperties,
		ThemeParts:       make(map[string][]byte, len(pkg.ThemeParts)),
		Numbering:        pkg.Numbering,
		FontTable:        pkg.FontTable,
		Settings:         pkg.Settings,
		WebSettings:      pkg.WebSettings,
	}

	if len(pkg.MainDocument) == 0 {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opParsePackage, "%s missing", constants.PathDocument)
	}

	docTree, err := parseXMLTree(pkg.MainDocument)
	if err != nil {
		return nil, xmlPartError(constants.PathDocument, err)
	}
	parsed.DocumentTree = docTree

	if len(pkg.RootRelationships) > 0 {
		var rels xmlstructs.Relationships
		if err := decodeXML(pkg.RootRelationships, &rels, constants.PathRels); err != nil {
			return nil, err
		}
		parsed.RootRelationships = &rels
	}

	if len(pkg.DocumentRelationships) > 0 {
		var rels xmlstructs.Relationships
		if err := decodeXML(pkg.DocumentRelationships, &rels, constants.PathDocRels); err != nil {
			return nil, err
		}
		parsed.DocumentRelationships = &rels
	}

	if len(pkg.Styles) > 0 {
		stylesTree, err := parseXMLTree(pkg.Styles)
		if err != nil {
			return nil, xmlPartError(constants.PathStyles, err)
		}
		parsed.StylesTree = stylesTree
	}

	if len(pkg.CoreProperties) > 0 {
		coreTree, err := parseXMLTree(pkg.CoreProperties)
		if err != nil {
			return nil, xmlPartError(constants.PathCoreProps, err)
		}
		parsed.CorePropertiesTree = coreTree
	}

	if len(pkg.AppProperties) > 0 {
		appTree, err := parseXMLTree(pkg.AppProperties)
		if err != nil {
			return nil, xmlPartError(constants.PathAppProps, err)
		}
		parsed.AppPropertiesTree = appTree
	}

	for name, data := range pkg.Headers {
		if len(data) == 0 {
			continue
		}
		tree, err := parseXMLTree(data)
		if err != nil {
			return nil, xmlPartError(name, err)
		}
		parsed.HeaderTrees[name] = tree
	}

	for name, data := range pkg.Footers {
		if len(data) == 0 {
			continue
		}
		tree, err := parseXMLTree(data)
		if err != nil {
			return nil, xmlPartError(name, err)
		}
		parsed.FooterTrees[name] = tree
	}

	for name, data := range pkg.ThemeParts {
		if len(data) == 0 {
			continue
		}
		parsed.ThemeParts[name] = data
	}

	return parsed, nil
}

func decodeXML(data []byte, dest interface{}, part string) error {
	if len(data) == 0 {
		return errors.Errorf(errors.ErrCodeInvalidState, opParsePackage, "%s is empty", part)
	}

	if err := xml.Unmarshal(data, dest); err != nil {
		return xmlPartError(part, err)
	}

	return nil
}

func xmlPartError(part string, err error) error {
	if err == nil {
		return nil
	}
	return &errors.DocxError{
		Code:    errors.ErrCodeXML,
		Op:      opParsePackage,
		Err:     err,
		Context: map[string]interface{}{"part": part},
	}
}
