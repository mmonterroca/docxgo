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

func TestMediaString(t *testing.T) {
	m := Media{Name: "image1.jpg"}
	expected := MEDIA_FOLDER + "image1.jpg"
	got := m.String()

	if got != expected {
		t.Errorf("Media.String() = %q, want %q", got, expected)
	}
}

func TestMediaMethod(t *testing.T) {
	doc := New().WithDefaultTheme()
	media := Media{Name: "test.jpg", Data: []byte("test data")}
	doc.addMedia(media)

	// Test existing media
	m := doc.Media("test.jpg")
	if m == nil {
		t.Fatal("Media('test.jpg') returned nil")
	}
	if m.Name != "test.jpg" {
		t.Errorf("Expected media name 'test.jpg', got '%s'", m.Name)
	}

	// Test non-existent media
	m = doc.Media("nonexistent.jpg")
	if m != nil {
		t.Error("Media('nonexistent.jpg') should return nil")
	}
}
