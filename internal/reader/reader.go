// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package reader

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"

	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

const (
	opLoadFromPath   = "reader.LoadPackageFromPath"
	opLoadFromBytes  = "reader.LoadPackageFromBytes"
	opLoadFromStream = "reader.LoadPackageFromStream"
	opLoadFromZip    = "reader.loadFromZip"
)

// LoadPackageFromPath reads a DOCX archive from disk and returns its raw parts.
func LoadPackageFromPath(path string) (*Package, error) {
	if path == "" {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opLoadFromPath, "path cannot be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, opLoadFromPath)
	}
	defer func() {
		_ = file.Close() // Best effort close
	}()

	info, err := file.Stat()
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, opLoadFromPath)
	}

	pkg, err := LoadPackage(file, info.Size())
	if err != nil {
		return nil, err
	}
	pkg.PackageSize = info.Size()
	return pkg, nil
}

// LoadPackage loads a DOCX archive from an io.ReaderAt / size pair.
func LoadPackage(r io.ReaderAt, size int64) (*Package, error) {
	if r == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opLoadFromStream, "reader cannot be nil")
	}

	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, opLoadFromStream)
	}

	pkg, err := loadFromZip(zr)
	if err != nil {
		return nil, err
	}
	pkg.PackageSize = size
	return pkg, nil
}

// LoadPackageFromBytes reads a DOCX archive from an in-memory byte slice.
func LoadPackageFromBytes(data []byte) (*Package, error) {
	if len(data) == 0 {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opLoadFromBytes, "data cannot be empty")
	}
	reader := bytes.NewReader(data)
	pkg, err := LoadPackage(reader, int64(len(data)))
	if err != nil {
		return nil, err
	}
	pkg.PackageSize = int64(len(data))
	return pkg, nil
}

// LoadPackageFromStream reads a DOCX archive from an io.Reader by buffering its content.
func LoadPackageFromStream(r io.Reader) (*Package, error) {
	if r == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opLoadFromStream, "reader cannot be nil")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeIO, opLoadFromStream)
	}
	return LoadPackageFromBytes(data)
}

func loadFromZip(zr *zip.Reader) (*Package, error) {
	if zr == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opLoadFromZip, "zip reader cannot be nil")
	}

	parts := make(map[string][]byte, len(zr.File))
	normalized := make(map[string]string, len(zr.File))

	for _, file := range zr.File {
		if file == nil {
			continue
		}
		if file.FileInfo().IsDir() {
			continue
		}

		name := normalizePartName(file.Name)
		if name == "" {
			continue
		}

		data, err := readZipFile(file)
		if err != nil {
			return nil, errors.WrapWithContext(err, opLoadFromZip, map[string]interface{}{"part": file.Name})
		}

		// Preserve the original name as recorded in the archive for future writes.
		parts[file.Name] = data
		normalized[name] = file.Name
	}

	pkg := &Package{
		RawParts:             parts,
		normalizedPaths:      normalized,
		Headers:              make(map[string][]byte),
		Footers:              make(map[string][]byte),
		HeaderRels:           make(map[string][]byte),
		FooterRels:           make(map[string][]byte),
		Media:                make(map[string]*MediaPart),
		AdditionalParts:      make(map[string][]byte),
		ThemeParts:           make(map[string][]byte),
		contentTypeOverrides: make(map[string]string),
	}

	if err := pkg.populateContentTypes(); err != nil {
		return nil, err
	}

	pkg.extractKnownParts()

	return pkg, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rc.Close() // Best effort close
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *Package) populateContentTypes() error {
	name, ok := p.lookupPart(constants.PathContentTypes)
	if !ok {
		return errors.Errorf(errors.ErrCodeInvalidState, opLoadFromZip, "%s not found", constants.PathContentTypes)
	}

	data := p.RawParts[name]
	if len(data) == 0 {
		return errors.Errorf(errors.ErrCodeInvalidState, opLoadFromZip, "%s is empty", constants.PathContentTypes)
	}

	var ct xmlstructs.ContentTypes
	if err := xml.Unmarshal(data, &ct); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeXML, opLoadFromZip)
	}

	p.ContentTypes = &ct

	for _, override := range ct.Overrides {
		if override == nil {
			continue
		}
		path := normalizePartName(strings.TrimPrefix(override.PartName, "/"))
		if path == "" {
			continue
		}
		p.contentTypeOverrides[path] = override.ContentType
	}

	return nil
}

func (p *Package) extractKnownParts() {
	extract := func(path string) []byte {
		name, ok := p.lookupPart(path)
		if !ok {
			return nil
		}
		return p.RawParts[name]
	}

	p.RootRelationships = extract(constants.PathRels)
	p.DocumentRelationships = extract(constants.PathDocRels)
	p.MainDocument = extract(constants.PathDocument)
	p.Styles = extract(constants.PathStyles)
	p.Numbering = extract(constants.PathNumbering)
	p.FontTable = extract(constants.PathFontTable)
	p.Settings = extract(constants.PathSettings)
	p.WebSettings = extract(constants.PathWebSettings)

	// Theme could have multiple parts (theme1.xml, theme2.xml, etc.).
	for key, original := range p.normalizedPaths {
		if !strings.HasPrefix(key, normalizePartName(constants.PathTheme)) {
			continue
		}
		p.ThemeParts[original] = p.RawParts[original]
	}

	p.CoreProperties = extract(constants.PathCoreProps)
	p.AppProperties = extract(constants.PathAppProps)
	p.CustomProperties = extract(constants.PathCustomProps)

	for key, original := range p.normalizedPaths {
		// Header/footer .rels parts (e.g. "word/_rels/header1.xml.rels") must
		// be checked before the plain header/footer prefixes below: they live
		// under "word/_rels/", not "word/header"/"word/footer", so they never
		// collide with those, but must still be claimed here rather than
		// falling through to AdditionalParts as an opaque, unrelated blob.
		if strings.HasPrefix(key, normalizePartName(constants.PathHeaderRelsPrefix)) {
			p.HeaderRels[original] = p.RawParts[original]
			continue
		}
		if strings.HasPrefix(key, normalizePartName(constants.PathFooterRelsPrefix)) {
			p.FooterRels[original] = p.RawParts[original]
			continue
		}
		if strings.HasPrefix(key, normalizePartName(constants.PathHeaderPrefix)) {
			p.Headers[original] = p.RawParts[original]
			continue
		}
		if strings.HasPrefix(key, normalizePartName(constants.PathFooterPrefix)) {
			p.Footers[original] = p.RawParts[original]
			continue
		}
		if strings.HasPrefix(key, normalizePartName(constants.PathMediaPrefix)) {
			ct := p.contentTypeFor(original)
			name := filepath.Base(original)
			p.Media[original] = &MediaPart{
				Path:        original,
				Name:        name,
				ContentType: ct,
				Data:        p.RawParts[original],
			}
			continue
		}
	}

	// Record additional parts we have not categorized yet.
	for key, original := range p.normalizedPaths {
		if p.isKnownPart(key) {
			continue
		}
		p.AdditionalParts[original] = p.RawParts[original]
	}
}

func (p *Package) isKnownPart(normalized string) bool {
	switch {
	case normalized == normalizePartName(constants.PathContentTypes):
		return true
	case normalized == normalizePartName(constants.PathRels):
		return true
	case normalized == normalizePartName(constants.PathDocRels):
		return true
	case normalized == normalizePartName(constants.PathDocument):
		return true
	case normalized == normalizePartName(constants.PathStyles):
		return true
	case normalized == normalizePartName(constants.PathNumbering):
		return true
	case normalized == normalizePartName(constants.PathFontTable):
		return true
	case normalized == normalizePartName(constants.PathSettings):
		return true
	case normalized == normalizePartName(constants.PathWebSettings):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathTheme)):
		return true
	case normalized == normalizePartName(constants.PathCoreProps):
		return true
	case normalized == normalizePartName(constants.PathAppProps):
		return true
	case normalized == normalizePartName(constants.PathCustomProps):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathHeaderRelsPrefix)):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathFooterRelsPrefix)):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathHeaderPrefix)):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathFooterPrefix)):
		return true
	case strings.HasPrefix(normalized, normalizePartName(constants.PathMediaPrefix)):
		return true
	default:
		return false
	}
}
