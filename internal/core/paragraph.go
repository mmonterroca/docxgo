/*
MIT License

Copyright (c) 2025 Misael Monterroca
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

package core

import (
	"path/filepath"
	"strings"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// IDGenerator interface for testing purposes
type IDGenerator interface {
	NextParagraphID() string
	NextRunID() string
	NextImageID() string
	GenerateID(prefix string) string
}

// paragraph implements the domain.Paragraph interface.
type paragraph struct {
	id            string
	runs          []domain.Run
	fields        []domain.Field
	images        []domain.Image
	styleName     string
	alignment     domain.Alignment
	indent        domain.Indentation
	spacingBefore int
	spacingAfter  int
	lineSpacing   domain.LineSpacing
	idGen         IDGenerator
	relManager    *manager.RelationshipManager
	bookmarkID    string // ID for bookmark (if this paragraph needs one for TOC)
	bookmarkName  string // Name for bookmark (e.g., "_Toc123456")
	mediaManager  *manager.MediaManager
}

// NewParagraph creates a new Paragraph.
func NewParagraph(id string, idGen IDGenerator, relManager *manager.RelationshipManager, mediaManager *manager.MediaManager) domain.Paragraph {
	return &paragraph{
		id:            id,
		runs:          make([]domain.Run, 0, constants.DefaultRunCapacity),
		fields:        make([]domain.Field, 0, 4),
		images:        make([]domain.Image, 0, 4),
		alignment:     domain.AlignmentLeft,
		indent:        domain.Indentation{},
		spacingBefore: constants.DefaultParagraphSpacing,
		spacingAfter:  constants.DefaultParagraphSpacing,
		lineSpacing:   domain.LineSpacing{Rule: domain.LineSpacingAuto, Value: constants.DefaultLineSpacing},
		idGen:         idGen,
		relManager:    relManager,
		mediaManager:  mediaManager,
	}
}

// AddRun adds a new text run to the paragraph.
func (p *paragraph) AddRun() (domain.Run, error) {
	id := p.idGen.NextRunID()
	run := NewRun(id, p.relManager)
	p.runs = append(p.runs, run)
	return run, nil
}

// AddField adds a field to the paragraph.
// Deprecated: Use AddRun() and run.AddField() instead for better control.
func (p *paragraph) AddField(_ domain.FieldType) (domain.Field, error) {
	return nil, errors.Unsupported("Paragraph.AddField", "use AddRun() and run.AddField() instead")
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

// AddImage adds an image to the paragraph from a file path.
func (p *paragraph) AddImage(path string) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImage(id, path)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImage")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageWithSize adds an image with custom dimensions.
func (p *paragraph) AddImageWithSize(path string, size domain.ImageSize) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageWithSize(id, path, size)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageWithSize")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageWithPosition adds an image with custom positioning.
func (p *paragraph) AddImageWithPosition(path string, size domain.ImageSize, pos domain.ImagePosition) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageWithPosition(id, path, size, pos)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageWithPosition")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// attachImage registers the image with the media and relationship managers and appends it as a drawing run.
func (p *paragraph) attachImage(img domain.Image, sourceName string) error {
	if p.mediaManager == nil {
		return errors.InvalidState("Paragraph.attachImage", "media manager not initialized")
	}

	if sourceName == "" {
		sourceName = img.ID()
	}

	_, mediaPath, err := p.mediaManager.Add(img.Data(), sourceName)
	if err != nil {
		return errors.Wrap(err, "Paragraph.attachImage")
	}

	relativePath := strings.TrimPrefix(mediaPath, "word/")
	if docxImg, ok := img.(*docxImage); ok {
		docxImg.setTarget(relativePath)
	}

	relID, err := p.relManager.AddImage(relativePath)
	if err != nil {
		return errors.Wrap(err, "Paragraph.attachImage")
	}

	if docxImg, ok := img.(*docxImage); ok {
		docxImg.SetRelationshipID(relID)
	}

	run := NewRun(p.idGen.NextRunID(), p.relManager)
	if setter, ok := run.(interface{ setImage(domain.Image) }); ok {
		setter.setImage(img)
	}

	p.runs = append(p.runs, run)
	p.images = append(p.images, img)
	return nil
}

// Images returns all images in this paragraph.
func (p *paragraph) Images() []domain.Image {
	images := make([]domain.Image, len(p.images))
	copy(images, p.images)
	return images
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
// Note: Currently only returns the style name, not a full Style object.
// For now, use SetStyle() to apply styles and track the name yourself if needed.
func (p *paragraph) Style() domain.Style {
	// Style retrieval from the style manager is not yet implemented.
	// Return nil for now - users should track the applied style name themselves.
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

// StyleName returns the style name applied to this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) StyleName() string {
	return p.styleName
}

// SetBookmark sets a bookmark for this paragraph (used for TOC).
// This is an internal method used when generating TOC.
func (p *paragraph) SetBookmark(id, name string) {
	p.bookmarkID = id
	p.bookmarkName = name
}

// BookmarkID returns the bookmark ID for this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) BookmarkID() string {
	return p.bookmarkID
}

// BookmarkName returns the bookmark name for this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) BookmarkName() string {
	return p.bookmarkName
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
