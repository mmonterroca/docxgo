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
	"fmt"
	"strconv"
)

// BytesToString converts a byte slice to a string.
//
// DEPRECATED: This function previously used unsafe.Pointer for zero-copy conversion,
// which could lead to data races and undefined behavior when the underlying byte slice
// is modified. Use standard string(b) conversion instead for safety.
//
// For performance-critical code where profiling shows this conversion is a bottleneck,
// consider using strings.Builder or sync.Pool to reduce allocations rather than
// relying on unsafe operations.
func BytesToString(b []byte) string {
	return string(b)
}

// StringToBytes converts a string to a byte slice.
//
// DEPRECATED: This function previously used unsafe.Pointer for zero-copy conversion,
// which violates Go's memory safety guarantees and can cause crashes if the string's
// backing array is deallocated. Use standard []byte(s) conversion instead.
//
// For performance-critical code, consider reusing byte slices with sync.Pool or
// using io.Writer interfaces to avoid unnecessary copies.
func StringToBytes(s string) []byte {
	return []byte(s)
}

// GetInt64 from string
func GetInt64(s string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return v, nil
	}
	v2, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return int64(v2), nil
	}
	_, err = fmt.Sscanf(s, "%d", &v)
	return v, err
}

// GetInt from string
func GetInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err == nil {
		return v, nil
	}
	v2, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return int(v2), nil
	}
	_, err = fmt.Sscanf(s, "%d", &v)
	return v, err
}
