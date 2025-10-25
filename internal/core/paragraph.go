/*
   Copyright (c) 2025 SlideLang

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

package core

import (
	"strings"

	"github.com/SlideLang/go-docx/domain"
	"github.com/SlideLang/go-docx/internal/manager"
	"github.com/SlideLang/go-docx/pkg/constants"
	"github.com/SlideLang/go-docx/pkg/errors"
)

// IDGenerator interface for testing purposes
type IDGenerator interface {
	NextParagraphID() string
	NextRunID() string
}

// paragraph implements the domain.Paragraph interface.
type paragraph struct {
	id            string
	runs          []domain.Run
	fields        []domain.Field
	styleName     string
	alignment     domain.Alignment
	indent        domain.Indentation
	spacingBefore int
	spacingAfter  int
	lineSpacing   domain.LineSpacing
	idGen         IDGenerator
	relManager    *manager.RelationshipManager
}

// NewParagraph creates a new Paragraph.
func NewParagraph(id string, idGen IDGenerator, relManager *manager.RelationshipManager) domain.Paragraph {
	return &paragraph{
		id:            id,
		runs:          make([]domain.Run, 0, constants.DefaultRunCapacity),
		fields:        make([]domain.Field, 0, 4),
		alignment:     domain.AlignmentLeft,
		indent:        domain.Indentation{},
		spacingBefore: constants.DefaultParagraphSpacing,
		spacingAfter:  constants.DefaultParagraphSpacing,
		lineSpacing:   domain.LineSpacing{Rule: domain.LineSpacingAuto, Value: constants.DefaultLineSpacing},
		idGen:         idGen,
		relManager:    relManager,
	}
}

// AddRun adds a new text run to the paragraph.
func (p *paragraph) AddRun() (domain.Run, error) {
	id := p.idGen.NextRunID()
	run := NewRun(id)
	p.runs = append(p.runs, run)
	return run, nil
}

// AddField adds a field to the paragraph.
func (p *paragraph) AddField(fieldType domain.FieldType) (domain.Field, error) {
	// TODO: Implement field creation
	return nil, errors.Unsupported("Paragraph.AddField", "fields not yet implemented")
}

// AddHyperlink adds a hyperlink to the paragraph.
func (p *paragraph) AddHyperlink(url, displayText string) (domain.Run, error) {
	if url == "" {
		return nil, errors.InvalidArgument("Paragraph.AddHyperlink", "url", url, "URL cannot be empty")
	}

	// Add relationship for hyperlink
	_, err := p.relManager.AddHyperlink(url)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	// Create run with hyperlink text
	run, err := p.AddRun()
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	text := displayText
	if text == "" {
		text = url
	}

	err = run.SetText(text)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	// Set hyperlink styling (blue, underlined)
	_ = run.SetColor(domain.ColorBlue)
	_ = run.SetUnderline(domain.UnderlineSingle)

	return run, nil
}

// Runs returns all runs in this paragraph.
func (p *paragraph) Runs() []domain.Run {
	// Return a copy to prevent external modification
	runs := make([]domain.Run, len(p.runs))
	copy(runs, p.runs)
	return runs
}

// Fields returns all fields in this paragraph.
func (p *paragraph) Fields() []domain.Field {
	fields := make([]domain.Field, len(p.fields))
	copy(fields, p.fields)
	return fields
}

// Text returns the plain text content of the paragraph.
func (p *paragraph) Text() string {
	var sb strings.Builder
	for _, run := range p.runs {
		sb.WriteString(run.Text())
	}
	return sb.String()
}

// Style returns the style applied to this paragraph.
func (p *paragraph) Style() domain.Style {
	// TODO: Implement style retrieval
	return nil
}

// SetStyle applies a named style to the paragraph.
func (p *paragraph) SetStyle(styleName string) error {
	if styleName == "" {
		return errors.InvalidArgument("Paragraph.SetStyle", "styleName", styleName, "style name cannot be empty")
	}
	p.styleName = styleName
	return nil
}

// Alignment returns the paragraph's horizontal alignment.
func (p *paragraph) Alignment() domain.Alignment {
	return p.alignment
}

// SetAlignment sets the paragraph's horizontal alignment.
func (p *paragraph) SetAlignment(align domain.Alignment) error {
	if align < domain.AlignmentLeft || align > domain.AlignmentDistribute {
		return errors.InvalidArgument("Paragraph.SetAlignment", "align", align, "invalid alignment value")
	}
	p.alignment = align
	return nil
}

// Indent returns the paragraph's indentation settings.
func (p *paragraph) Indent() domain.Indentation {
	return p.indent
}

// SetIndent sets the paragraph's indentation.
func (p *paragraph) SetIndent(indent domain.Indentation) error {
	// Validate indentation values
	if indent.Left < constants.MinIndent || indent.Left > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Left", indent.Left,
			"left indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	if indent.Right < constants.MinIndent || indent.Right > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Right", indent.Right,
			"right indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	if indent.FirstLine < 0 || indent.FirstLine > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.FirstLine", indent.FirstLine,
			"first line indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	if indent.Hanging < 0 || indent.Hanging > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Hanging", indent.Hanging,
			"hanging indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	if indent.FirstLine > 0 && indent.Hanging > 0 {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent", indent,
			"cannot have both first line indent and hanging indent")
	}

	p.indent = indent
	return nil
}

// SpacingBefore returns spacing before the paragraph (in twips).
func (p *paragraph) SpacingBefore() int {
	return p.spacingBefore
}

// SetSpacingBefore sets spacing before the paragraph.
func (p *paragraph) SetSpacingBefore(twips int) error {
	if twips < constants.MinSpacing || twips > constants.MaxSpacing {
		return errors.InvalidArgument("Paragraph.SetSpacingBefore", "twips", twips,
			"spacing must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.spacingBefore = twips
	return nil
}

// SpacingAfter returns spacing after the paragraph (in twips).
func (p *paragraph) SpacingAfter() int {
	return p.spacingAfter
}

// SetSpacingAfter sets spacing after the paragraph.
func (p *paragraph) SetSpacingAfter(twips int) error {
	if twips < constants.MinSpacing || twips > constants.MaxSpacing {
		return errors.InvalidArgument("Paragraph.SetSpacingAfter", "twips", twips,
			"spacing must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.spacingAfter = twips
	return nil
}

// LineSpacing returns the line spacing setting.
func (p *paragraph) LineSpacing() domain.LineSpacing {
	return p.lineSpacing
}

// SetLineSpacing sets the line spacing.
func (p *paragraph) SetLineSpacing(spacing domain.LineSpacing) error {
	if spacing.Rule < domain.LineSpacingAuto || spacing.Rule > domain.LineSpacingAtLeast {
		return errors.InvalidArgument("Paragraph.SetLineSpacing", "spacing.Rule", spacing.Rule,
			"invalid line spacing rule")
	}
	if spacing.Value < constants.MinLineSpacing || spacing.Value > constants.MaxLineSpacing {
		return errors.InvalidArgument("Paragraph.SetLineSpacing", "spacing.Value", spacing.Value,
			"line spacing value must be between 0 and 31680 twips")
	}
	p.lineSpacing = spacing
	return nil
}
