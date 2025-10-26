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



package manager

import (
	"sync"
	"testing"

	"github.com/mmonterroca/docxgo/domain"
)

func TestNewStyleManager(t *testing.T) {
	sm := NewStyleManager()

	if sm == nil {
		t.Fatal("NewStyleManager() returned nil")
	}

	// Verify built-in styles are loaded
	if !sm.HasStyle(domain.StyleIDNormal) {
		t.Error("Normal style should be loaded")
	}

	if !sm.HasStyle(domain.StyleIDHeading1) {
		t.Error("Heading1 style should be loaded")
	}
}

func TestStyleManager_GetStyle(t *testing.T) {
	sm := NewStyleManager()

	tests := []struct {
		name      string
		styleID   string
		wantError bool
	}{
		{
			name:      "Normal",
			styleID:   domain.StyleIDNormal,
			wantError: false,
		},
		{
			name:      "Heading1",
			styleID:   domain.StyleIDHeading1,
			wantError: false,
		},
		{
			name:      "NonExistent",
			styleID:   "NonExistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, err := sm.GetStyle(tt.styleID)

			if tt.wantError {
				if err == nil {
					t.Error("GetStyle() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetStyle() error = %v", err)
				}
				if style == nil {
					t.Error("GetStyle() returned nil style")
				}
				if style.ID() != tt.styleID {
					t.Errorf("Style ID = %v, want %v", style.ID(), tt.styleID)
				}
			}
		})
	}
}

func TestStyleManager_AddStyle(t *testing.T) {
	sm := NewStyleManager()

	customStyle := newParagraphStyle("CustomStyle", "My Custom Style", false)

	// Add custom style
	if err := sm.AddStyle(customStyle); err != nil {
		t.Errorf("AddStyle() error = %v", err)
	}

	// Verify it was added
	if !sm.HasStyle("CustomStyle") {
		t.Error("Custom style should exist after adding")
	}

	// Try to add duplicate
	if err := sm.AddStyle(customStyle); err == nil {
		t.Error("AddStyle() should error on duplicate")
	}

	// Try to override built-in
	builtInOverride := newParagraphStyle(domain.StyleIDNormal, "Override", false)
	if err := sm.AddStyle(builtInOverride); err == nil {
		t.Error("AddStyle() should error when overriding built-in style")
	}
}

func TestStyleManager_RemoveStyle(t *testing.T) {
	sm := NewStyleManager()

	customStyle := newParagraphStyle("CustomStyle", "My Custom Style", false)
	sm.AddStyle(customStyle)

	// Remove custom style
	if err := sm.RemoveStyle("CustomStyle"); err != nil {
		t.Errorf("RemoveStyle() error = %v", err)
	}

	// Verify removed
	if sm.HasStyle("CustomStyle") {
		t.Error("Custom style should not exist after removal")
	}

	// Try to remove built-in
	if err := sm.RemoveStyle(domain.StyleIDNormal); err == nil {
		t.Error("RemoveStyle() should error when removing built-in style")
	}

	// Try to remove non-existent
	if err := sm.RemoveStyle("NonExistent"); err == nil {
		t.Error("RemoveStyle() should error for non-existent style")
	}
}

func TestStyleManager_ListStyles(t *testing.T) {
	sm := NewStyleManager()

	styles := sm.ListStyles()

	if len(styles) == 0 {
		t.Error("ListStyles() should return built-in styles")
	}

	// Verify some known styles are present
	found := make(map[string]bool)
	for _, style := range styles {
		found[style.ID()] = true
	}

	if !found[domain.StyleIDNormal] {
		t.Error("ListStyles() should include Normal style")
	}

	if !found[domain.StyleIDHeading1] {
		t.Error("ListStyles() should include Heading1 style")
	}
}

func TestStyleManager_ListStylesByType(t *testing.T) {
	sm := NewStyleManager()

	paragraphStyles := sm.ListStylesByType(domain.StyleTypeParagraph)
	characterStyles := sm.ListStylesByType(domain.StyleTypeCharacter)

	if len(paragraphStyles) == 0 {
		t.Error("ListStylesByType(Paragraph) should return styles")
	}

	if len(characterStyles) == 0 {
		t.Error("ListStylesByType(Character) should return styles")
	}

	// Verify types are correct
	for _, style := range paragraphStyles {
		if style.Type() != domain.StyleTypeParagraph {
			t.Errorf("Expected paragraph style, got %v", style.Type())
		}
	}

	for _, style := range characterStyles {
		if style.Type() != domain.StyleTypeCharacter {
			t.Errorf("Expected character style, got %v", style.Type())
		}
	}
}

func TestStyleManager_DefaultStyle(t *testing.T) {
	sm := NewStyleManager()

	// Test paragraph default
	paragraphDefault, err := sm.DefaultStyle(domain.StyleTypeParagraph)
	if err != nil {
		t.Errorf("DefaultStyle(Paragraph) error = %v", err)
	}
	if paragraphDefault.ID() != domain.StyleIDNormal {
		t.Errorf("Default paragraph style = %v, want %v", paragraphDefault.ID(), domain.StyleIDNormal)
	}

	// Test character default
	characterDefault, err := sm.DefaultStyle(domain.StyleTypeCharacter)
	if err != nil {
		t.Errorf("DefaultStyle(Character) error = %v", err)
	}
	if characterDefault.ID() != domain.StyleIDDefaultParagraphFont {
		t.Errorf("Default character style = %v, want %v", characterDefault.ID(), domain.StyleIDDefaultParagraphFont)
	}
}

func TestStyleManager_SetDefaultStyle(t *testing.T) {
	sm := NewStyleManager()

	// Set new default paragraph style
	if err := sm.SetDefaultStyle(domain.StyleTypeParagraph, domain.StyleIDHeading1); err != nil {
		t.Errorf("SetDefaultStyle() error = %v", err)
	}

	// Verify it was set
	defaultStyle, _ := sm.DefaultStyle(domain.StyleTypeParagraph)
	if defaultStyle.ID() != domain.StyleIDHeading1 {
		t.Errorf("Default style = %v, want %v", defaultStyle.ID(), domain.StyleIDHeading1)
	}

	// Try to set non-existent style
	if err := sm.SetDefaultStyle(domain.StyleTypeParagraph, "NonExistent"); err == nil {
		t.Error("SetDefaultStyle() should error for non-existent style")
	}

	// Try to set wrong type
	if err := sm.SetDefaultStyle(domain.StyleTypeParagraph, domain.StyleIDEmphasis); err == nil {
		t.Error("SetDefaultStyle() should error for type mismatch")
	}
}

func TestStyleManager_IsBuiltIn(t *testing.T) {
	sm := NewStyleManager()

	if !sm.IsBuiltIn(domain.StyleIDNormal) {
		t.Error("Normal should be built-in")
	}

	if !sm.IsBuiltIn(domain.StyleIDHeading1) {
		t.Error("Heading1 should be built-in")
	}

	if sm.IsBuiltIn("CustomStyle") {
		t.Error("CustomStyle should not be built-in")
	}
}

func TestStyleManager_Concurrency(t *testing.T) {
	sm := NewStyleManager()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = sm.GetStyle(domain.StyleIDNormal)
			_ = sm.HasStyle(domain.StyleIDHeading1)
			_ = sm.ListStyles()
		}()
	}

	// Concurrent custom style operations
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			customStyle := newParagraphStyle("Custom"+string(rune(id)), "Custom", false)
			_ = sm.AddStyle(customStyle)
		}()
	}

	wg.Wait()
}

func TestParagraphStyle_Properties(t *testing.T) {
	style := newParagraphStyle("TestStyle", "Test Style", false)

	// Test SetAlignment
	if err := style.SetAlignment(domain.AlignmentCenter); err != nil {
		t.Errorf("SetAlignment() error = %v", err)
	}
	if style.Alignment() != domain.AlignmentCenter {
		t.Errorf("Alignment() = %v, want %v", style.Alignment(), domain.AlignmentCenter)
	}

	// Test SetSpacingBefore
	if err := style.SetSpacingBefore(240); err != nil {
		t.Errorf("SetSpacingBefore() error = %v", err)
	}
	if style.SpacingBefore() != 240 {
		t.Errorf("SpacingBefore() = %v, want %v", style.SpacingBefore(), 240)
	}

	// Test SetOutlineLevel
	if err := style.SetOutlineLevel(1); err != nil {
		t.Errorf("SetOutlineLevel() error = %v", err)
	}
	if style.OutlineLevel() != 1 {
		t.Errorf("OutlineLevel() = %v, want %v", style.OutlineLevel(), 1)
	}

	// Test invalid outline level
	if err := style.SetOutlineLevel(10); err == nil {
		t.Error("SetOutlineLevel(10) should error")
	}
}

func TestCharacterStyle_Properties(t *testing.T) {
	style := newCharacterStyle("TestChar", "Test Char", false)

	// Test SetBold
	if err := style.SetBold(true); err != nil {
		t.Errorf("SetBold() error = %v", err)
	}
	if !style.Bold() {
		t.Error("Bold() should be true")
	}

	// Test SetItalic
	if err := style.SetItalic(true); err != nil {
		t.Errorf("SetItalic() error = %v", err)
	}
	if !style.Italic() {
		t.Error("Italic() should be true")
	}

	// Test SetColor
	blue := domain.Color{R: 0, G: 0, B: 255}
	if err := style.SetColor(blue); err != nil {
		t.Errorf("SetColor() error = %v", err)
	}
	if style.Color() != blue {
		t.Errorf("Color() = %v, want %v", style.Color(), blue)
	}

	// Test SetSize
	if err := style.SetSize(32); err != nil {
		t.Errorf("SetSize() error = %v", err)
	}
	if style.Size() != 32 {
		t.Errorf("Size() = %v, want %v", style.Size(), 32)
	}
}

func TestBuiltInStyles_Coverage(t *testing.T) {
	sm := NewStyleManager()

	// Test all heading styles exist
	for i := 1; i <= 9; i++ {
		styleID := ""
		switch i {
		case 1:
			styleID = domain.StyleIDHeading1
		case 2:
			styleID = domain.StyleIDHeading2
		case 3:
			styleID = domain.StyleIDHeading3
		case 4:
			styleID = domain.StyleIDHeading4
		case 5:
			styleID = domain.StyleIDHeading5
		case 6:
			styleID = domain.StyleIDHeading6
		case 7:
			styleID = domain.StyleIDHeading7
		case 8:
			styleID = domain.StyleIDHeading8
		case 9:
			styleID = domain.StyleIDHeading9
		}

		if !sm.HasStyle(styleID) {
			t.Errorf("Heading%d style should exist", i)
		}
	}

	// Test TOC styles exist
	for i := 1; i <= 9; i++ {
		styleID := ""
		switch i {
		case 1:
			styleID = domain.StyleIDTOC1
		case 2:
			styleID = domain.StyleIDTOC2
		case 3:
			styleID = domain.StyleIDTOC3
		case 4:
			styleID = domain.StyleIDTOC4
		case 5:
			styleID = domain.StyleIDTOC5
		case 6:
			styleID = domain.StyleIDTOC6
		case 7:
			styleID = domain.StyleIDTOC7
		case 8:
			styleID = domain.StyleIDTOC8
		case 9:
			styleID = domain.StyleIDTOC9
		}

		if !sm.HasStyle(styleID) {
			t.Errorf("TOC%d style should exist", i)
		}
	}

	// Test other important styles
	importantStyles := []string{
		domain.StyleIDTitle,
		domain.StyleIDSubtitle,
		domain.StyleIDQuote,
		domain.StyleIDHeader,
		domain.StyleIDFooter,
		domain.StyleIDEmphasis,
		domain.StyleIDStrong,
		domain.StyleIDHyperlink,
	}

	for _, styleID := range importantStyles {
		if !sm.HasStyle(styleID) {
			t.Errorf("Style %s should exist", styleID)
		}
	}
}
