/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
)

// pack receives a zip file writer (word documents are a zip with multiple xml inside)
// and writes the relevant files. Some of them come from the empty_constants file,
// others from the actual in-memory structure
func (f *Docx) pack(zipWriter *zip.Writer) (err error) {
	files := make(map[string]io.Reader, DefaultFileMapCapacity)

	if f.template != "" {
		for _, name := range f.tmpfslst {
			// Skip [Content_Types].xml - we'll generate it dynamically
			if name == "[Content_Types].xml" {
				continue
			}
			files[name], err = f.tmplfs.Open("xml/" + f.template + "/" + name)
			if err != nil {
				return
			}
		}
	} else {
		for _, name := range f.tmpfslst {
			// Skip [Content_Types].xml - we'll generate it dynamically
			if name == "[Content_Types].xml" {
				continue
			}
			files[name], err = f.tmplfs.Open(name)
			if err != nil {
				return
			}
		}
	}

	// Add sectPr at the end of the document body if it exists and not already present
	if f.sectPr != nil {
		// Check if sectPr is already at the end
		hasSectPr := false
		if len(f.Document.Body.Items) > 0 {
			_, hasSectPr = f.Document.Body.Items[len(f.Document.Body.Items)-1].(*SectPr)
		}

		// Only append if not already there
		if !hasSectPr {
			f.Document.Body.Items = append(f.Document.Body.Items, f.sectPr)
		}
	}

	files["word/_rels/document.xml.rels"] = marshaller{data: &f.docRelation}
	files["word/document.xml"] = marshaller{data: &f.Document}

	// Add headers
	if header, ok := f.headers[HeaderFooterDefault]; ok {
		files["word/header1.xml"] = marshaller{data: header}
	}
	if header, ok := f.headers[HeaderFooterFirst]; ok {
		files["word/header2.xml"] = marshaller{data: header}
	}
	if header, ok := f.headers[HeaderFooterEven]; ok {
		files["word/header3.xml"] = marshaller{data: header}
	}

	// Add footers
	if footer, ok := f.footers[HeaderFooterDefault]; ok {
		files["word/footer1.xml"] = marshaller{data: footer}
	}
	if footer, ok := f.footers[HeaderFooterFirst]; ok {
		files["word/footer2.xml"] = marshaller{data: footer}
	}
	if footer, ok := f.footers[HeaderFooterEven]; ok {
		files["word/footer3.xml"] = marshaller{data: footer}
	}

	for _, m := range f.media {
		files[m.String()] = bytes.NewReader(m.Data)
	}

	// Generate [Content_Types].xml dynamically to include headers/footers
	contentTypes := f.generateContentTypes()
	files["[Content_Types].xml"] = bytes.NewReader(contentTypes)

	for path, r := range files {
		w, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}

	return
}

type marshaller struct {
	data interface{}
	io.Reader
	io.WriterTo
}

// Read is fake and is to trigger io.WriterTo
func (m marshaller) Read(_ []byte) (int, error) {
	return 0, os.ErrInvalid
}

// WriteTo n is always 0 for we don't care that value
func (m marshaller) WriteTo(w io.Writer) (n int64, err error) {
	_, err = io.WriteString(w, xml.Header)
	if err != nil {
		return
	}
	err = xml.NewEncoder(w).Encode(m.data)
	return
}

// generateContentTypes creates the [Content_Types].xml file dynamically
// including entries for headers and footers if they exist
func (f *Docx) generateContentTypes() []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString("\n")
	buf.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	buf.WriteString("\n")

	// Default extensions
	buf.WriteString(`    <Default Extension="jpeg" ContentType="image/jpeg"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Default Extension="png" ContentType="image/png"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Default Extension="webp" ContentType="image/webp"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Default Extension="xml" ContentType="application/xml"/>`)
	buf.WriteString("\n")

	// Core document parts
	buf.WriteString(`    <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>`)
	buf.WriteString("\n")

	// Headers
	if _, ok := f.headers[HeaderFooterDefault]; ok {
		buf.WriteString(`    <Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`)
		buf.WriteString("\n")
	}
	if _, ok := f.headers[HeaderFooterFirst]; ok {
		buf.WriteString(`    <Override PartName="/word/header2.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`)
		buf.WriteString("\n")
	}
	if _, ok := f.headers[HeaderFooterEven]; ok {
		buf.WriteString(`    <Override PartName="/word/header3.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`)
		buf.WriteString("\n")
	}

	// Footers
	if _, ok := f.footers[HeaderFooterDefault]; ok {
		buf.WriteString(`    <Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
		buf.WriteString("\n")
	}
	if _, ok := f.footers[HeaderFooterFirst]; ok {
		buf.WriteString(`    <Override PartName="/word/footer2.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
		buf.WriteString("\n")
	}
	if _, ok := f.footers[HeaderFooterEven]; ok {
		buf.WriteString(`    <Override PartName="/word/footer3.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
		buf.WriteString("\n")
	}

	// Document properties
	buf.WriteString(`    <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	buf.WriteString("\n")
	buf.WriteString(`    <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	buf.WriteString("\n")

	buf.WriteString(`</Types>`)
	return buf.Bytes()
}
