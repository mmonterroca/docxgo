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
	"testing"
)

func TestReferTarget(t *testing.T) {
	doc := New().WithDefaultTheme()
	rid := doc.addLinkRelation("https://example.com")

	target, err := doc.ReferTarget(rid)
	if err != nil {
		t.Fatalf("ReferTarget(%q) failed: %v", rid, err)
	}
	if target != "https://example.com" {
		t.Errorf("ReferTarget(%q) = %q, want %q", rid, target, "https://example.com")
	}

	// Test non-existent RID
	target, err = doc.ReferTarget("nonexistent")
	if err != ErrRefIDNotFound {
		t.Errorf("ReferTarget('nonexistent') error = %v, want %v", err, ErrRefIDNotFound)
	}
	if target != "" {
		t.Errorf("ReferTarget('nonexistent') = %q, want empty string", target)
	}
}

func TestReferID(t *testing.T) {
	doc := New().WithDefaultTheme()
	rid1 := doc.addLinkRelation("https://example1.com")
	rid2 := doc.addLinkRelation("https://example2.com")

	// Test existing target
	id, err := doc.ReferID("https://example1.com")
	if err != nil {
		t.Fatalf("ReferID('https://example1.com') failed: %v", err)
	}
	if id != rid1 {
		t.Errorf("ReferID('https://example1.com') = %q, want %q", id, rid1)
	}

	id, err = doc.ReferID("https://example2.com")
	if err != nil {
		t.Fatalf("ReferID('https://example2.com') failed: %v", err)
	}
	if id != rid2 {
		t.Errorf("ReferID('https://example2.com') = %q, want %q", id, rid2)
	}

	// Test non-existent target
	id, err = doc.ReferID("https://nonexistent.com")
	if err != ErrRefIDNotFound {
		t.Errorf("ReferID('https://nonexistent.com') error = %v, want %v", err, ErrRefIDNotFound)
	}
	if id != "" {
		t.Errorf("ReferID('https://nonexistent.com') = %q, want empty string", id)
	}
}
