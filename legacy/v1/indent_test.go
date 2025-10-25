/*
   Copyright (c) 2025 SlideLang Enhanced Fork

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

import "testing"

func TestParagraphIndent(t *testing.T) {
	doc := New()
	
	// Test basic left indent
	p1 := doc.AddParagraph()
	p1.AddText("Left indented paragraph")
	p1.Indent(720, 0, 0) // 0.5 inch left indent
	
	if p1.Properties == nil {
		t.Fatal("Expected properties to be initialized")
	}
	if p1.Properties.Ind == nil {
		t.Fatal("Expected Ind to be initialized")
	}
	if p1.Properties.Ind.Left != 720 {
		t.Errorf("Expected Left indent of 720, got %d", p1.Properties.Ind.Left)
	}
	
	// Test first line indent
	p2 := doc.AddParagraph()
	p2.AddText("First line indented paragraph")
	p2.Indent(0, 360, 0) // 0.25 inch first line indent
	
	if p2.Properties.Ind.FirstLine != 360 {
		t.Errorf("Expected FirstLine indent of 360, got %d", p2.Properties.Ind.FirstLine)
	}
	
	// Test hanging indent
	p3 := doc.AddParagraph()
	p3.AddText("Hanging indent paragraph")
	p3.Indent(720, 0, 360) // 0.5 inch left, 0.25 inch hanging
	
	if p3.Properties.Ind.Left != 720 {
		t.Errorf("Expected Left indent of 720, got %d", p3.Properties.Ind.Left)
	}
	if p3.Properties.Ind.Hanging != 360 {
		t.Errorf("Expected Hanging indent of 360, got %d", p3.Properties.Ind.Hanging)
	}
	
	t.Log("✅ Paragraph indentation working correctly")
}
