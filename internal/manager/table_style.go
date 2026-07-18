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

package manager

import (
	"sync"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// tableStyle implements domain.Style and domain.TableStyleDef for table styles.
type tableStyle struct {
	mu               sync.RWMutex
	id               string
	name             string
	basedOn          string
	next             string
	font             domain.Font
	isDefault        bool
	isBuiltIn        bool
	tableBorders     domain.TableLevelBorders
	hasTableBorders  bool
	cellMarginTop    int
	cellMarginLeft   int
	cellMarginBottom int
	cellMarginRight  int
}

// newTableStyle creates a new table style.
// Note: builtIn parameter is used in tests to create custom styles.
func newTableStyle(id, name string, builtIn bool) *tableStyle {
	return &tableStyle{
		id:        id,
		name:      name,
		next:      "",
		font:      domain.Font{Name: constants.DefaultFontName},
		isBuiltIn: builtIn,
	}
}

// ID returns the style identifier.
func (ts *tableStyle) ID() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.id
}

// Name returns the localized style name.
func (ts *tableStyle) Name() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.name
}

// Type identifies the style as a table style.
func (ts *tableStyle) Type() domain.StyleType {
	return domain.StyleTypeTable
}

// BasedOn returns the parent style identifier.
func (ts *tableStyle) BasedOn() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.basedOn
}

// SetBasedOn sets the parent style identifier.
func (ts *tableStyle) SetBasedOn(styleID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.basedOn = styleID
	return nil
}

// Next returns the recommended follow-on style identifier.
func (ts *tableStyle) Next() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.next
}

// SetNext sets the recommended follow-on style identifier.
func (ts *tableStyle) SetNext(styleID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.next = styleID
	return nil
}

// Font returns the default font metadata associated with the style.
func (ts *tableStyle) Font() domain.Font {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.font
}

// SetFont updates the default font metadata for the style.
func (ts *tableStyle) SetFont(font domain.Font) error {
	if font.Name == "" {
		return errors.NewValidationError(
			"TableStyle.SetFont",
			"font.Name",
			"",
			"font name cannot be empty",
		)
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.font = font
	return nil
}

// IsDefault reports whether this style is the default for tables.
func (ts *tableStyle) IsDefault() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.isDefault
}

// SetDefault marks this style as default for its type.
func (ts *tableStyle) SetDefault(isDefault bool) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.isDefault = isDefault
	return nil
}

// IsCustom reports whether the style is user-defined.
func (ts *tableStyle) IsCustom() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return !ts.isBuiltIn
}

// TableBorders returns the table-level borders for this style.
func (ts *tableStyle) TableBorders() domain.TableLevelBorders {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.tableBorders
}

// SetTableBorders sets the table-level borders for this style.
func (ts *tableStyle) SetTableBorders(borders domain.TableLevelBorders) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tableBorders = borders
	ts.hasTableBorders = true
	return nil
}

// HasTableBorders returns whether this style defines table borders.
func (ts *tableStyle) HasTableBorders() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.hasTableBorders
}

// CellMargins returns the default cell margins for this style.
func (ts *tableStyle) CellMargins() (top, left, bottom, right int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.cellMarginTop, ts.cellMarginLeft, ts.cellMarginBottom, ts.cellMarginRight
}

// SetCellMargins sets the default cell margins for this style.
func (ts *tableStyle) SetCellMargins(top, left, bottom, right int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.cellMarginTop = top
	ts.cellMarginLeft = left
	ts.cellMarginBottom = bottom
	ts.cellMarginRight = right
	return nil
}
