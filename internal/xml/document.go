// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package xml provides XML structures and marshaling for Office Open XML (OOXML) format.
// This package defines the XML types used to serialize document elements into the
// DOCX file format, including documents, paragraphs, runs, tables, and styles.
package xml

import "encoding/xml"

// Document represents w:document element (main document).
type Document struct {
	XMLName    xml.Name    `xml:"w:document"`
	XMLnsW     string      `xml:"xmlns:w,attr"`
	XMLnsR     string      `xml:"xmlns:r,attr"`
	XMLnsWP    string      `xml:"xmlns:wp,attr,omitempty"`
	Background *Background `xml:"w:background,omitempty"`
	Body       *Body       `xml:"w:body"`
}

// Body represents w:body element.
type Body struct {
	XMLName xml.Name           `xml:"w:body"`
	Content []interface{}      `xml:",any"`
	SectPr  *SectionProperties `xml:"w:sectPr,omitempty"`
}

// Background represents w:background element.
type Background struct {
	XMLName xml.Name `xml:"w:background"`
	Color   string   `xml:"w:color,attr,omitempty"`
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
