package manager

import (
	"sync"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// tableStyle implements domain.Style for table styles.
type tableStyle struct {
	mu        sync.RWMutex
	id        string
	name      string
	basedOn   string
	next      string
	font      domain.Font
	isDefault bool
	isBuiltIn bool
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
