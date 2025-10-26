/*
MIT License

Copyright (c) 2025 Misael Montero
Copyright (c) 2020-2023 fumiama (original go-docx)

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
	XMLName      xml.Name `xml:"cp:coreProperties"`
	XMLnsCP      string   `xml:"xmlns:cp,attr"`
	XMLnsDC      string   `xml:"xmlns:dc,attr"`
	XMLnsDCTerms string   `xml:"xmlns:dcterms,attr"`
	XMLnsXSI     string   `xml:"xmlns:xsi,attr"`
	Title        string   `xml:"dc:title,omitempty"`
	Subject      string   `xml:"dc:subject,omitempty"`
	Creator      string   `xml:"dc:creator,omitempty"`
	Keywords     string   `xml:"cp:keywords,omitempty"`
	Description  string   `xml:"dc:description,omitempty"`
	Created      *DCDate  `xml:"dcterms:created,omitempty"`
	Modified     *DCDate  `xml:"dcterms:modified,omitempty"`
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
