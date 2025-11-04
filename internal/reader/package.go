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

// Package reader provides low-level helpers for loading DOCX archives into
// raw OOXML parts that can later be mapped to domain models.
package reader

import (
	"path/filepath"
	"strings"

	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
)

// Package represents the low-level parts that make up a DOCX archive.
// It focuses on raw OOXML payloads so higher layers can hydrate domain models
// without worrying about ZIP details.
type Package struct {
	// ContentTypes mirrors [Content_Types].xml.
	ContentTypes *xmlstructs.ContentTypes

	// RawParts keeps every part in the archive keyed by its canonical name.
	RawParts map[string][]byte

	// normalizedPaths allows lookup by normalized (lowercase, trimmed) names.
	normalizedPaths map[string]string

	// Core Word parts
	MainDocument          []byte
	DocumentRelationships []byte
	RootRelationships     []byte
	Styles                []byte
	Numbering             []byte
	FontTable             []byte
	Settings              []byte
	WebSettings           []byte
	ThemeParts            map[string][]byte
	CoreProperties        []byte
	AppProperties         []byte
	CustomProperties      []byte

	// Header/Footer content indexed by file name (e.g. "word/header1.xml").
	Headers map[string][]byte
	Footers map[string][]byte

	// Media assets keyed by archive path (e.g. "word/media/image1.png").
	Media map[string]*MediaPart

	// AdditionalParts captures any payload we do not process yet.
	AdditionalParts map[string][]byte

	// PackageSize is the total size of the original DOCX archive in bytes.
	PackageSize int64

	contentTypeOverrides map[string]string
}

// MediaPart represents a binary asset bundled inside the DOCX archive.
type MediaPart struct {
	Path        string
	Name        string
	ContentType string
	Data        []byte
}

// contentTypeFor returns the content type for a given part path.
func (p *Package) contentTypeFor(path string) string {
	if p == nil || p.ContentTypes == nil {
		return ""
	}

	normalized := normalizePartName(path)
	if ct, ok := p.contentTypeOverrides[normalized]; ok {
		return ct
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(normalized)), ".")
	if ext == "" {
		return ""
	}

	for _, def := range p.ContentTypes.Defaults {
		if def == nil {
			continue
		}
		if strings.EqualFold(def.Extension, ext) {
			return def.ContentType
		}
	}

	return ""
}

// lookupPart resolves a path to the actual stored file name.
func (p *Package) lookupPart(path string) (string, bool) {
	if p == nil {
		return "", false
	}
	if _, ok := p.RawParts[path]; ok {
		return path, true
	}
	normalized := normalizePartName(path)
	name, ok := p.normalizedPaths[normalized]
	return name, ok
}

// normalizePartName produces a canonical key for part lookup.
func normalizePartName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSpace(name)
	return strings.ToLower(name)
}
