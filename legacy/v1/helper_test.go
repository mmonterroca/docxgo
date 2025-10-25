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

func TestBytesToString(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty", []byte{}, ""},
		{"simple", []byte("hello"), "hello"},
		{"with spaces", []byte("hello world"), "hello world"},
		{"unicode", []byte("你好世界"), "你好世界"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := BytesToString(tc.input)
			if got != tc.want {
				t.Errorf("BytesToString(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestGetInt64(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"simple integer", "123", 123, false},
		{"negative integer", "-456", -456, false},
		{"float as string", "123.45", 123, false},
		{"zero", "0", 0, false},
		{"large number", "9223372036854775807", 9223372036854775807, false},
		{"with spaces", "  789  ", 789, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetInt64(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetInt64(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("GetInt64(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"simple integer", "123", 123, false},
		{"negative integer", "-456", -456, false},
		{"float as string", "123.45", 123, false},
		{"zero", "0", 0, false},
		{"with spaces", "  789  ", 789, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetInt(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetInt(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("GetInt(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
