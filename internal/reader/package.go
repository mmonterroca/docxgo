// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package reader provides low-level helpers for loading DOCX archives into
// raw OOXML parts that can later be mapped to domain models.
package reader

import (
	"path"
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

	// Header/Footer content indexed by archive path (e.g. "word/header1.xml").
	Headers map[string][]byte
	Footers map[string][]byte

	// Header/Footer relationship parts, indexed by archive path (e.g.
	// "word/_rels/header1.xml.rels"). Kept separate from AdditionalParts so
	// they can be matched up with their owning header/footer by name instead
	// of being preserved as an opaque, unrelated blob.
	HeaderRels map[string][]byte
	FooterRels map[string][]byte

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

// Which kind of part a .rels file belongs to.
const (
	relsKindHeader = "header"
	relsKindFooter = "footer"
)

// headerFooterRelsKind reports whether a normalized part name is the
// relationships file of a header or a footer, and which.
//
// Decided by shape -- a "_rels" directory, a ".rels" suffix, and an owner
// named headerN.xml or footerN.xml -- rather than by a literal path prefix,
// because a relationship target may name a subdirectory: the part and the
// _rels directory beside it can sit anywhere under word/, so
// "word/headers/_rels/header1.xml.rels" is as legal as
// "word/_rels/header1.xml.rels" and both have to land in the same bucket.
//
// word/_rels/document.xml.rels and the package's own _rels/.rels are not
// header or footer rels and fall through to their existing handling.
func headerFooterRelsKind(normalized string) (string, bool) {
	if !strings.HasSuffix(normalized, ".rels") {
		return "", false
	}
	if path.Base(path.Dir(normalized)) != "_rels" {
		return "", false
	}
	owner := strings.TrimSuffix(path.Base(normalized), ".rels")
	switch {
	case strings.HasPrefix(owner, relsKindHeader):
		return relsKindHeader, true
	case strings.HasPrefix(owner, relsKindFooter):
		return relsKindFooter, true
	}
	return "", false
}
