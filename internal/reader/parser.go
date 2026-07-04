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
	stdpath "path"

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

	// HeaderRelationships and FooterRelationships hold each header/footer
	// part's own relationships, keyed the same way as HeaderTrees/FooterTrees
	// (e.g. "word/header1.xml"). Relationship IDs are scoped per-part in
	// OOXML, so these must not be merged into DocumentRelationships.
	HeaderRelationships map[string]*xmlstructs.Relationships
	FooterRelationships map[string]*xmlstructs.Relationships

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

	parsed.HeaderRelationships, err = partRelationships(pkg, pkg.Headers)
	if err != nil {
		return nil, err
	}
	parsed.FooterRelationships, err = partRelationships(pkg, pkg.Footers)
	if err != nil {
		return nil, err
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

// relsPathFor returns the archive path of a part's relationships file,
// e.g. "word/header1.xml" -> "word/_rels/header1.xml.rels".
func relsPathFor(partPath string) string {
	return stdpath.Join(stdpath.Dir(partPath), "_rels", stdpath.Base(partPath)+".rels")
}

// partRelationships parses the per-part .rels file for each entry in parts
// (header or footer bodies keyed by archive path) and returns them keyed the
// same way. Parts without a .rels file are simply omitted.
func partRelationships(pkg *Package, parts map[string][]byte) (map[string]*xmlstructs.Relationships, error) {
	out := make(map[string]*xmlstructs.Relationships, len(parts))
	for name := range parts {
		relsName, ok := pkg.lookupPart(relsPathFor(name))
		if !ok {
			continue
		}
		data := pkg.RawParts[relsName]
		if len(data) == 0 {
			continue
		}
		var rels xmlstructs.Relationships
		if err := decodeXML(data, &rels, relsName); err != nil {
			return nil, err
		}
		out[name] = &rels
	}
	return out, nil
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
