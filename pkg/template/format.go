/*
MIT License

Copyright (c) 2025 Misael Monterroca <misael@monterroca.com>

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

package template

import "github.com/mmonterroca/docxgo/v2/domain"

// formatsEqual returns true if two runs have identical visible formatting.
// It compares all 8 formatting attributes that affect text appearance.
// This is used by ConsolidateRuns to decide if adjacent runs can be merged.
func formatsEqual(a, b domain.Run) bool {
	return a.Font() == b.Font() &&
		a.Color() == b.Color() &&
		a.Size() == b.Size() &&
		a.Bold() == b.Bold() &&
		a.Italic() == b.Italic() &&
		a.Underline() == b.Underline() &&
		a.Strike() == b.Strike() &&
		a.Highlight() == b.Highlight()
}

// isTextOnly returns true if the run contains only text (no fields, breaks, or images).
// Runs with non-text content must not be merged during consolidation.
func isTextOnly(r domain.Run) bool {
	return len(r.Fields()) == 0 &&
		len(r.Breaks()) == 0 &&
		r.Image() == nil
}
