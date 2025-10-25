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

// Package docx provides functionality for creating and manipulating
// Microsoft Word (.docx) documents.
//
// This is v2 of the go-docx library, a complete rewrite with improved
// architecture, better error handling, and comprehensive OOXML support.
//
// Example usage:
//
//	doc := docx.NewDocument()
//	para, _ := doc.AddParagraph()
//	run, _ := para.AddRun()
//	run.SetText("Hello, World!")
//	run.SetBold(true)
//	doc.SaveAs("hello.docx")
package docx

import (
	"github.com/SlideLang/go-docx/v2/domain"
	"github.com/SlideLang/go-docx/v2/internal/core"
)

// NewDocument creates a new empty Word document.
func NewDocument() domain.Document {
	return core.NewDocument()
}

// Version returns the library version.
const Version = "2.0.0-alpha"
