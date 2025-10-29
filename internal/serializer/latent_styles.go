// Package serializer converts domain objects into XML structures for OOXML serialization.
// It provides serializers for documents, paragraphs, runs, tables, and other document elements.
package serializer

/*
   Copyright (c) 2025 Misael Monterroca

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

import "github.com/mmonterroca/docxgo/internal/xml"

// defaultLatentStyles mirrors Microsoft Word's shipped latent style table so that
// generated documents include the same built-in quick styles metadata.
var defaultLatentStyles = &xml.LatentStyles{
	DefLockedState:    "0",
	DefUIPriority:     "99",
	DefSemiHidden:     "0",
	DefUnhideWhenUsed: "0",
	DefQFormat:        "0",
	Count:             "376",
	Exceptions: []*xml.LatentStyleException{
		{
			Name:       "Normal",
			UIPriority: "0",
			QFormat:    "1",
		},
		{
			Name:       "heading 1",
			UIPriority: "9",
			QFormat:    "1",
		},
		{
			Name:           "heading 2",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 3",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 4",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 5",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 6",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 7",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 8",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "heading 9",
			UIPriority:     "9",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 6",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 7",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 8",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index 9",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 1",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 2",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 3",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 4",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 5",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 6",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 7",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 8",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toc 9",
			UIPriority:     "39",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Normal Indent",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "footnote text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "annotation text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "header",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "footer",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "index heading",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "caption",
			UIPriority:     "35",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "table of figures",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "envelope address",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "envelope return",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "footnote reference",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "annotation reference",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "line number",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "page number",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "endnote reference",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "endnote text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "table of authorities",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "macro",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "toa heading",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Bullet",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Number",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Bullet 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Bullet 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Bullet 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Bullet 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Number 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Number 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Number 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Number 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Title",
			UIPriority: "10",
			QFormat:    "1",
		},
		{
			Name:           "Closing",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Signature",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Default Paragraph Font",
			UIPriority:     "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text Indent",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Continue",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Continue 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Continue 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Continue 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "List Continue 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Message Header",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Subtitle",
			UIPriority: "11",
			QFormat:    "1",
		},
		{
			Name:           "Salutation",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Date",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text First Indent",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text First Indent 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Note Heading",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text Indent 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Body Text Indent 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Block Text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Hyperlink",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "FollowedHyperlink",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Strong",
			UIPriority: "22",
			QFormat:    "1",
		},
		{
			Name:       "Emphasis",
			UIPriority: "20",
			QFormat:    "1",
		},
		{
			Name:           "Document Map",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Plain Text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "E-mail Signature",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Top of Form",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Bottom of Form",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Normal (Web)",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Acronym",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Address",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Cite",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Code",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Definition",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Keyboard",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Preformatted",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Sample",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Typewriter",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "HTML Variable",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Normal Table",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "annotation subject",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "No List",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Outline List 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Outline List 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Outline List 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Simple 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Simple 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Simple 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Classic 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Classic 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Classic 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Classic 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Colorful 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Colorful 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Colorful 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Columns 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Columns 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Columns 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Columns 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Columns 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 6",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 7",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Grid 8",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 4",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 5",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 6",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 7",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table List 8",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table 3D effects 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table 3D effects 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table 3D effects 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Contemporary",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Elegant",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Professional",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Subtle 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Subtle 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Web 1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Web 2",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Table Web 3",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Balloon Text",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Table Grid",
			UIPriority: "39",
		},
		{
			Name:           "Table Theme",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Placeholder Text",
			SemiHidden: "1",
		},
		{
			Name:       "No Spacing",
			UIPriority: "1",
			QFormat:    "1",
		},
		{
			Name:       "Light Shading",
			UIPriority: "60",
		},
		{
			Name:       "Light List",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3",
			UIPriority: "69",
		},
		{
			Name:       "Dark List",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 1",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 1",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 1",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 1",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 1",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 1",
			UIPriority: "65",
		},
		{
			Name:       "Revision",
			SemiHidden: "1",
		},
		{
			Name:       "List Paragraph",
			UIPriority: "34",
			QFormat:    "1",
		},
		{
			Name:       "Quote",
			UIPriority: "29",
			QFormat:    "1",
		},
		{
			Name:       "Intense Quote",
			UIPriority: "30",
			QFormat:    "1",
		},
		{
			Name:       "Medium List 2 Accent 1",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 1",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 1",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 1",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 1",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 1",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 1",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 1",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 2",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 2",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 2",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 2",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 2",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 2",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2 Accent 2",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 2",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 2",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 2",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 2",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 2",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 2",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 2",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 3",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 3",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 3",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 3",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 3",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 3",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2 Accent 3",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 3",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 3",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 3",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 3",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 3",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 3",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 3",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 4",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 4",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 4",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 4",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 4",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 4",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2 Accent 4",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 4",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 4",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 4",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 4",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 4",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 4",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 4",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 5",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 5",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 5",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 5",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 5",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 5",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2 Accent 5",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 5",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 5",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 5",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 5",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 5",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 5",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 5",
			UIPriority: "73",
		},
		{
			Name:       "Light Shading Accent 6",
			UIPriority: "60",
		},
		{
			Name:       "Light List Accent 6",
			UIPriority: "61",
		},
		{
			Name:       "Light Grid Accent 6",
			UIPriority: "62",
		},
		{
			Name:       "Medium Shading 1 Accent 6",
			UIPriority: "63",
		},
		{
			Name:       "Medium Shading 2 Accent 6",
			UIPriority: "64",
		},
		{
			Name:       "Medium List 1 Accent 6",
			UIPriority: "65",
		},
		{
			Name:       "Medium List 2 Accent 6",
			UIPriority: "66",
		},
		{
			Name:       "Medium Grid 1 Accent 6",
			UIPriority: "67",
		},
		{
			Name:       "Medium Grid 2 Accent 6",
			UIPriority: "68",
		},
		{
			Name:       "Medium Grid 3 Accent 6",
			UIPriority: "69",
		},
		{
			Name:       "Dark List Accent 6",
			UIPriority: "70",
		},
		{
			Name:       "Colorful Shading Accent 6",
			UIPriority: "71",
		},
		{
			Name:       "Colorful List Accent 6",
			UIPriority: "72",
		},
		{
			Name:       "Colorful Grid Accent 6",
			UIPriority: "73",
		},
		{
			Name:       "Subtle Emphasis",
			UIPriority: "19",
			QFormat:    "1",
		},
		{
			Name:       "Intense Emphasis",
			UIPriority: "21",
			QFormat:    "1",
		},
		{
			Name:       "Subtle Reference",
			UIPriority: "31",
			QFormat:    "1",
		},
		{
			Name:       "Intense Reference",
			UIPriority: "32",
			QFormat:    "1",
		},
		{
			Name:       "Book Title",
			UIPriority: "33",
			QFormat:    "1",
		},
		{
			Name:           "Bibliography",
			UIPriority:     "37",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "TOC Heading",
			UIPriority:     "39",
			QFormat:        "1",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:       "Plain Table 1",
			UIPriority: "41",
		},
		{
			Name:       "Plain Table 2",
			UIPriority: "42",
		},
		{
			Name:       "Plain Table 3",
			UIPriority: "43",
		},
		{
			Name:       "Plain Table 4",
			UIPriority: "44",
		},
		{
			Name:       "Plain Table 5",
			UIPriority: "45",
		},
		{
			Name:       "Grid Table Light",
			UIPriority: "40",
		},
		{
			Name:       "Grid Table 1 Light",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 1",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 1",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 1",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 1",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 1",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 1",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 1",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 2",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 2",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 2",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 2",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 2",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 2",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 2",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 3",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 3",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 3",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 3",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 3",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 3",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 3",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 4",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 4",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 4",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 4",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 4",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 4",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 4",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 5",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 5",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 5",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 5",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 5",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 5",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 5",
			UIPriority: "52",
		},
		{
			Name:       "Grid Table 1 Light Accent 6",
			UIPriority: "46",
		},
		{
			Name:       "Grid Table 2 Accent 6",
			UIPriority: "47",
		},
		{
			Name:       "Grid Table 3 Accent 6",
			UIPriority: "48",
		},
		{
			Name:       "Grid Table 4 Accent 6",
			UIPriority: "49",
		},
		{
			Name:       "Grid Table 5 Dark Accent 6",
			UIPriority: "50",
		},
		{
			Name:       "Grid Table 6 Colorful Accent 6",
			UIPriority: "51",
		},
		{
			Name:       "Grid Table 7 Colorful Accent 6",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 1",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 1",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 1",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 1",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 1",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 1",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 1",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 2",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 2",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 2",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 2",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 2",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 2",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 2",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 3",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 3",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 3",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 3",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 3",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 3",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 3",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 4",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 4",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 4",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 4",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 4",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 4",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 4",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 5",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 5",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 5",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 5",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 5",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 5",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 5",
			UIPriority: "52",
		},
		{
			Name:       "List Table 1 Light Accent 6",
			UIPriority: "46",
		},
		{
			Name:       "List Table 2 Accent 6",
			UIPriority: "47",
		},
		{
			Name:       "List Table 3 Accent 6",
			UIPriority: "48",
		},
		{
			Name:       "List Table 4 Accent 6",
			UIPriority: "49",
		},
		{
			Name:       "List Table 5 Dark Accent 6",
			UIPriority: "50",
		},
		{
			Name:       "List Table 6 Colorful Accent 6",
			UIPriority: "51",
		},
		{
			Name:       "List Table 7 Colorful Accent 6",
			UIPriority: "52",
		},
		{
			Name:           "Mention",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Smart Hyperlink",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Hashtag",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Unresolved Mention",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
		{
			Name:           "Smart Link",
			SemiHidden:     "1",
			UnhideWhenUsed: "1",
		},
	},
}
