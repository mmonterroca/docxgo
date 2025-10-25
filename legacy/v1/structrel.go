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

import "fmt"

//nolint:revive,stylecheck
const (
	XMLNS_REL     = `http://schemas.openxmlformats.org/package/2006/relationships`
	REL_HYPERLINK = `http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink`
	REL_IMAGE     = `http://schemas.openxmlformats.org/officeDocument/2006/relationships/image`

	REL_TARGETMODE = "External"
)

// Relationships ...
type Relationships struct {
	Xmlns        string `xml:"xmlns,attr"`
	Relationship []Relationship
}

// Relationship ...
type Relationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr,omitempty"`
}

// AddExternalLink adds an external hyperlink relationship and returns its ID
func (r *Relationships) AddExternalLink(url string, relType string) string {
	// Generate new relationship ID
	rId := r.nextID()

	// Determine the full relationship type
	var fullType string
	switch relType {
	case "hyperlink":
		fullType = REL_HYPERLINK
	case "image":
		fullType = REL_IMAGE
	default:
		fullType = REL_HYPERLINK
	}

	// Create new relationship
	rel := Relationship{
		ID:         rId,
		Type:       fullType,
		Target:     url,
		TargetMode: REL_TARGETMODE,
	}

	r.Relationship = append(r.Relationship, rel)
	return rId
}

// nextID generates the next available relationship ID
func (r *Relationships) nextID() string {
	maxID := 0
	for _, rel := range r.Relationship {
		// Extract number from rId format (e.g., "rId4" -> 4)
		var num int
		if n, _ := fmt.Sscanf(rel.ID, "rId%d", &num); n == 1 {
			if num > maxID {
				maxID = num
			}
		}
	}
	return fmt.Sprintf("rId%d", maxID+1)
}
