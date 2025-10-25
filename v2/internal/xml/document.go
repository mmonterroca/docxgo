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

package xml

import "encoding/xml"

// Document represents w:document element (main document).
type Document struct {
	XMLName xml.Name `xml:"w:document"`
	XMLnsW  string   `xml:"xmlns:w,attr"`
	XMLnsR  string   `xml:"xmlns:r,attr"`
	Body    *Body    `xml:"w:body"`
}

// Body represents w:body element.
type Body struct {
	XMLName    xml.Name     `xml:"w:body"`
	Paragraphs []*Paragraph `xml:"w:p,omitempty"`
	Tables     []*Table     `xml:"w:tbl,omitempty"`
}

// Relationships represents the .rels file.
type Relationships struct {
	XMLName       xml.Name        `xml:"Relationships"`
	Xmlns         string          `xml:"xmlns,attr"`
	Relationships []*Relationship `xml:"Relationship"`
}

// Relationship represents a single relationship.
type Relationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr,omitempty"`
}

// ContentTypes represents [Content_Types].xml
type ContentTypes struct {
	XMLName   xml.Name    `xml:"Types"`
	Xmlns     string      `xml:"xmlns,attr"`
	Defaults  []*Default  `xml:"Default"`
	Overrides []*Override `xml:"Override"`
}

// Default represents a default content type.
type Default struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// Override represents an override content type.
type Override struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// CoreProperties represents core document properties.
type CoreProperties struct {
	XMLName     xml.Name `xml:"cp:coreProperties"`
	XMLnsCP     string   `xml:"xmlns:cp,attr"`
	XMLnsDC     string   `xml:"xmlns:dc,attr"`
	XMLnsDCTerms string  `xml:"xmlns:dcterms,attr"`
	XMLnsXSI    string   `xml:"xmlns:xsi,attr"`
	Title       string   `xml:"dc:title,omitempty"`
	Subject     string   `xml:"dc:subject,omitempty"`
	Creator     string   `xml:"dc:creator,omitempty"`
	Keywords    string   `xml:"cp:keywords,omitempty"`
	Description string   `xml:"dc:description,omitempty"`
	Created     *DCDate  `xml:"dcterms:created,omitempty"`
	Modified    *DCDate  `xml:"dcterms:modified,omitempty"`
}

// DCDate represents a Dublin Core date.
type DCDate struct {
	Type  string `xml:"xsi:type,attr"`
	Value string `xml:",chardata"`
}

// AppProperties represents app.xml properties.
type AppProperties struct {
	XMLName     xml.Name `xml:"Properties"`
	Xmlns       string   `xml:"xmlns,attr"`
	Application string   `xml:"Application"`
	DocSecurity int      `xml:"DocSecurity"`
	Lines       int      `xml:"Lines"`
	Paragraphs  int      `xml:"Paragraphs"`
	Company     string   `xml:"Company,omitempty"`
}
