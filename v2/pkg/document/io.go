package document

/*
   Copyright (c) 2025 SlideLang

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

import (
	"io"
	"os"

	"github.com/SlideLang/go-docx/v2/internal/serializer"
	"github.com/SlideLang/go-docx/v2/internal/writer"
	"github.com/SlideLang/go-docx/v2/pkg/errors"
)

// WriteTo writes the document to the given writer in .docx format.
func (d *document) WriteTo(w io.Writer) error {
	// Serialize domain objects to XML structures
	ser := serializer.NewDocumentSerializer()

	xmlDoc, err := ser.SerializeDocument(d)
	if err != nil {
		return errors.Wrap(err, errors.ErrSerialization, "serialize document")
	}

	// Create ZIP writer
	zipWriter := writer.NewZipWriter(w)
	defer zipWriter.Close()

	// Write document structure
	rels := d.relManager.ToXML()
	coreProps := nil // TODO: Add metadata support
	appProps := nil  // TODO: Add metadata support

	if err := zipWriter.WriteDocument(xmlDoc, rels, coreProps, appProps); err != nil {
		return errors.Wrap(err, errors.ErrIO, "write document")
	}

	return nil
}

// SaveAs saves the document to a file.
func (d *document) SaveAs(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, errors.ErrIO, "create file")
	}
	defer f.Close()

	return d.WriteTo(f)
}
